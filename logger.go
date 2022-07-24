package xoplog

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"

	"github.com/mohae/deepcopy"
)

type Log struct {
	span     Span
	request  *Span
	local    local   // one span in a request
	shared   *shared // shared between spans in a request
	buffered bool
}

type Span struct {
	seed     Seed
	dataLock sync.Mutex // protects Data & SpanType (can only be held for short periods)
	log      *Log       // back to self
	base     xopbase.Span
	linePool sync.Pool
	boring   int32 // 0 = boring
}

type local struct {
	ForkCounter int32
	StepCounter int32
	Created     time.Time
	IsRequest   bool
}

// shared is common between the loggers that share a search index
type shared struct {
	FlushLock      sync.Mutex // protects Flush() (can be held for a longish period)
	RefCount       int32
	UnflushedLogs  int32
	FlushTimer     *time.Timer
	FlushDelay     time.Duration
	FlushActive    int32 // 1 == timer is running, 0 = timer is not running
	BaseRequest    xopbase.Request
	ReferencesKept bool
}

var DefaultFlushDelay = time.Minute * 5

func (s Seed) Request(descriptionOrName string) *Log {
	s = s.Copy()
	s.traceBundle.Trace.RebuildSetNonZero()
	log := Log{
		span: Span{
			seed: s,
		},
		shared: &shared{
			RefCount:    1,
			FlushActive: 1,
		},
		local: local{
			Created:   time.Now(),
			IsRequest: true,
		},
	}
	log.span.log = &log
	log.request = &log.span
	log.shared.BaseRequest = log.span.seed.baseLoggers.AsOne.Request(log.span.seed.traceBundle, descriptionOrName)
	log.shared.ReferencesKept = log.span.seed.baseLoggers.AsOne.ReferencesKept()
	log.buffered = log.span.seed.baseLoggers.AsOne.Buffered()
	log.span.base = log.shared.BaseRequest.(xopbase.Span)
	s.sendPrefill(&log) // before turning on the timer so as to not create a race
	if log.buffered {
		log.shared.FlushTimer = time.AfterFunc(DefaultFlushDelay, log.timerFlush)
	}
	return &log
}

func (old *Log) newChildLog(seed Seed) *Log {
	log := &Log{
		span: Span{
			seed: seed,
		},
		local: local{
			Created:   time.Now(),
			IsRequest: false,
		},
		shared:  old.shared,
		request: old.request,
	}
	log.span.log = log
	log.span.base.Boring(true)
	seed.sendPrefill(log)
	return log
}

func (l *Log) enableFlushTimer() {
	if l.buffered {
		was := atomic.SwapInt32(&l.shared.FlushActive, 1)
		if was == 0 {
			l.shared.FlushTimer.Reset(l.shared.FlushDelay)
		}
	}
}

// timerFlush is only called by log.shared.FlushTimer
func (l *Log) timerFlush() {
	l.Flush()
}

func (l *Log) Flush() {
	l.shared.FlushLock.Lock()
	defer l.shared.FlushLock.Unlock()
	// Stop is is not thread-safe with respect to other calls to Stop
	l.shared.FlushTimer.Stop()
	atomic.StoreInt32(&l.shared.FlushActive, 0)
	l.shared.BaseRequest.Flush()
}

// Marks this request as boring.  Any log at the Alert or
// Error level will mark this request as not boring.
func (l *Log) Boring() {
	requestBoring := atomic.LoadInt32(&l.request.boring)
	if requestBoring != 0 {
		return
	}
	l.request.base.Boring(true)
	// There is chance that in the time we were sending that
	// boring=true, the the request became un-boring. If that
	// happened, we can't tell if we're currently marked as
	// boring, so let's make sure we're not boring by sending
	// a false
	requestStillBoring := atomic.LoadInt32(&l.request.boring)
	if requestStillBoring != 0 {
		l.request.base.Boring(false)
	}
	l.enableFlushTimer()
}

func (l *Log) notBoring() {
	spanBoring := atomic.AddInt32(&l.span.boring, 1)
	if spanBoring == 1 {
		l.span.base.Boring(false)
		requestBoring := atomic.AddInt32(&l.request.boring, 1)
		if requestBoring == 1 {
			l.request.base.Boring(false)
		}
		l.enableFlushTimer()
	}
}

// TODO func (l *Log) Zap() like zap
// TODO func (l *Log) Sugar() like zap.Sugar
// TODO func (l *Log) Zero() like zerolog
// TODO func (l *Log) One() like onelog (or is Zero() good enough?)

// Done is used to indicate that a seed.Reqeust(), log.Fork().Wait(), or
// log.Step().Wait() is done.  When all of the parts of a request are
// finished, the log is automatically flushed.
func (l *Log) Done() {
	remaining := atomic.AddInt32(&l.shared.RefCount, -1)
	if remaining <= 0 {
		l.Flush()
	} else {
		l.enableFlushTimer()
	}
}

// Wait modifies (and returns) a Log to indicate that the overall
// request is not finished until an additional Done() is called.
func (l *Log) Wait() *Log {
	remaining := atomic.AddInt32(&l.shared.RefCount, 1)
	if remaining > 1 {
		return l
	}
	// This indicates a bug in the code that is using the logger.
	l.Warn().Msg("Too many calls to log.Done()") // TODO: allow user to provide error maker
	l.shared.FlushTimer.Reset(DefaultFlushDelay)
	return l
}

// Fork creates a new Log that does not need to be terminated because
// it is assumed to be done with the current log is finished.
func (l *Log) Fork(msg string, mods ...SeedModifier) *Log {
	seed := l.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&l.local.ForkCounter, 1))
	seed.prefix += "." + base26(counter)
	return l.newChildLog(seed)
}

// Step creates a new log that does not need to be terminated -- it
// represents the continued execution of the current log bug doing
// something that is different and should be in a fresh span. The expectation
// is that there is a parent log that is creating various sub-logs using
// Step over and over as it does different things.
func (l *Log) Step(msg string, mods ...SeedModifier) *Log {
	seed := l.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&l.local.StepCounter, 1))
	seed.prefix += "." + strconv.Itoa(counter)
	return l.newChildLog(seed)
}

type LogLine struct {
	log  *Log
	line xopbase.Line
}

func (l *Log) LogLine(level xopconst.Level) *LogLine {
	recycled := l.span.linePool.Get()
	if recycled != nil {
		// TODO: try using LogLine instead of *LogLine
		ll := recycled.(*LogLine)
		ll.line.Recycle(level, time.Now())
		return ll
	}
	return &LogLine{
		log:  l,
		line: l.span.base.Line(level, time.Now()),
	}
}

// Template is an alternative to Msg() sends a log line.  Template
// is a string that uses "{name}" substitutions from the data already
// sent with the line to format that data for human consumption.
// Template is expected to be more expensive than Msg so it should
// be used somewhat sparingly.  Data elements do not have to be
// consumed by the template.
//
// The names used for "{name}" substitutions are restricted: they may
// not include any characters that would be escapsed in a JSON string.
// No double quote.  No linefeed.  No backslash.  Etc.
func (ll *LogLine) Template(template string) {
	ll.line.Template(template)
	ll.log.span.linePool.Put(ll)
	ll.log.enableFlushTimer()
}

func (ll *LogLine) Msg(msg string) {
	ll.line.Msg(msg)
	ll.log.span.linePool.Put(ll)
	ll.log.enableFlushTimer()
}

func (l *Log) Debug() *LogLine { return l.LogLine(xopconst.DebugLevel) }
func (l *Log) Trace() *LogLine { return l.LogLine(xopconst.TraceLevel) }
func (l *Log) Info() *LogLine  { return l.LogLine(xopconst.InfoLevel) }
func (l *Log) Warn() *LogLine  { return l.LogLine(xopconst.WarnLevel) }
func (l *Log) Error() *LogLine {
	l.notBoring()
	return l.LogLine(xopconst.ErrorLevel)
}
func (l *Log) Alert() *LogLine {
	l.notBoring()
	return l.LogLine(xopconst.AlertLevel)
}

// TODO: generate these
// TODO: the rest of the set
func (ll *LogLine) Msgf(msg string, v ...interface{})           { ll.Msg(fmt.Sprintf(msg, v...)) }
func (ll *LogLine) Msgs(v ...interface{})                       { ll.Msg(fmt.Sprint(v...)) }
func (ll *LogLine) Int(k string, v int) *LogLine                { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int8(k string, v int8) *LogLine              { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int16(k string, v int16) *LogLine            { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int32(k string, v int32) *LogLine            { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int64(k string, v int64) *LogLine            { ll.line.Int(k, v); return ll }
func (ll *LogLine) Uint(k string, v uint) *LogLine              { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint8(k string, v uint8) *LogLine            { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint16(k string, v uint16) *LogLine          { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint32(k string, v uint32) *LogLine          { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint64(k string, v uint64) *LogLine          { ll.line.Uint(k, v); return ll }
func (ll *LogLine) Str(k string, v string) *LogLine             { ll.line.Str(k, v); return ll }
func (ll *LogLine) Bool(k string, v bool) *LogLine              { ll.line.Bool(k, v); return ll }
func (ll *LogLine) Time(k string, v time.Time) *LogLine         { ll.line.Time(k, v); return ll }
func (ll *LogLine) Error(k string, v error) *LogLine            { ll.line.Error(k, v); return ll }
func (ll *LogLine) Link(k string, v trace.Trace) *LogLine       { ll.line.Link(k, v); return ll }
func (ll *LogLine) Duration(k string, v time.Duration) *LogLine { ll.line.Duration(k, v); return ll }

func (ll *LogLine) EmbeddedEnum(k xopconst.EmbeddedEnum) *LogLine {
	return ll.Enum(k.EnumAttribute(), k)
}

func (ll *LogLine) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) *LogLine {
	ll.line.Enum(k, v)
	return ll
}

// AnyImmutable can be used to log something that is not going to be further modified
// after this call.
func (ll *LogLine) AnyImmutable(k string, v interface{}) *LogLine { ll.line.Any(k, v); return ll }

// Any can be used to log something that might be modified after this call.  If any base
// logger does not immediately serialize, then the object will be copied using
// https://github.com/mohae/deepcopy 's Copy().
func (ll *LogLine) Any(k string, v interface{}) *LogLine {
	if ll.log.shared.ReferencesKept {
		// TODO: make copy function configurable
		v = deepcopy.Copy(v)
	}
	ll.line.Any(k, v)
	return ll
}

// TODO: func (l *Log) Guage(name string, value float64, )
// TODO: func (l *Log) AdjustCounter(name string, value float64, )
// TODO: func (l *Log) Event

func copyMap(o map[string]interface{}) map[string]interface{} {
	n := make(map[string]interface{})
	for k, v := range o {
		n[k] = v
	}
	return n
}
