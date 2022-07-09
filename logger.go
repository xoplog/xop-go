package xoplog

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"

	"github.com/mohae/deepcopy"
)

type Log struct {
	span    Span
	request *Span
	local   local   // one span in a request
	shared  *shared // shared between spans in a request
}

type Span struct {
	seed     Seed
	dataLock sync.Mutex // protects Data & SpanType (can only be held for short periods)
	data     []xop.Thing
	spanType xopconst.SpanType
	log      *Log // back to self
	base     xopbase.Span
	linePool sync.Pool
}

type local struct {
	ForkCounter int32
	StepCounter int32
	Created     time.Time
	InDirty     int32 // in shared.Dirty? 0 = false, 1 = true
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

	// Dirty holds spans that have modified data and need to be
	// written or re-written.  It does not not track logs that
	// need flushing. Protected by request.dataLock.
	Dirty []*Log
}

var DefaultFlushDelay = time.Minute * 5

func (s Seed) Request(description string) *Log {
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
			InDirty:   1,
			Created:   time.Now(),
			IsRequest: true,
		},
	}
	log.span.log = &log
	log.request = &log.span
	log.shared.Dirty = append(log.shared.Dirty, &log)
	log.shared.FlushTimer = time.AfterFunc(DefaultFlushDelay, log.timerFlush)
	log.shared.BaseRequest = log.span.seed.baseLoggers.AsOne.Request(log.span.seed.traceBundle)
	log.shared.ReferencesKept = log.span.seed.baseLoggers.AsOne.ReferencesKept()
	log.span.base = log.shared.BaseRequest.(xopbase.Span)
	return &log
}

func (old *Log) newChildLog(seed Seed) *Log {
	log := &Log{
		span: Span{
			seed: seed,
		},
		local: local{
			InDirty:   1,
			Created:   time.Now(),
			IsRequest: false,
		},
		shared:  old.shared,
		request: old.request,
	}
	log.span.log = log
	log.request.dataLock.Lock()
	defer log.request.dataLock.Unlock()
	log.shared.Dirty = append(log.shared.Dirty, log)
	return log
}

func (l *Log) setDirty() {
	wasInDirty := atomic.SwapInt32(&l.local.InDirty, 1)
	if wasInDirty == 0 {
		func() {
			l.request.dataLock.Lock()
			defer l.request.dataLock.Unlock()
			l.shared.Dirty = append(l.shared.Dirty, l)
		}()
		l.enableFlushTimer()
	}
}

func (l *Log) enableFlushTimer() {
	was := atomic.SwapInt32(&l.shared.FlushActive, 1)
	if was == 0 {
		l.shared.FlushTimer.Reset(l.shared.FlushDelay)
	}
}

// timerFlush is only called by log.shared.FlushTimer
func (l *Log) timerFlush() {
	l.Flush()
}

func (l *Log) Flush() {
	func() {
		l.shared.FlushLock.Lock()
		defer l.shared.FlushLock.Unlock()
		// Stop is is not thread-safe with respect to other calls to Stop
		l.shared.FlushTimer.Stop()
		atomic.StoreInt32(&l.shared.FlushActive, 0)
	}()
	var dirty []*Log
	func() {
		l.request.dataLock.Lock()
		defer l.request.dataLock.Unlock()
		dirty = l.shared.Dirty
		l.shared.Dirty = nil
		for _, log := range dirty {
			atomic.StoreInt32(&log.local.InDirty, 0)
		}
	}()

	l.shared.FlushLock.Lock()
	defer l.shared.FlushLock.Unlock()
	for _, log := range dirty {
		log.span.base.SpanInfo(log.span.spanType, log.span.data)
	}
	l.shared.BaseRequest.Flush()
}

func (l *Log) log(level xopconst.Level, msg string, values []xop.Thing) {
	line := l.LogLine(level)
	xopbase.LineThings(line.line, values)
	line.Msg(msg)
}

// TODO func (l *Log) Zap() like zap
// TODO func (l *Log) Sugar() like zap.Sugar
// TODO func (l *Log) Zero() like zerolog
// TODO func (l *Log) One() like onelog (or is Zero() good enough?)

// Done is used to single that a Log, Fork().Wait(), or Step().Wait() is
// done.  When all of the parts of a request are finished, the log is
// automatically flushed.
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
// something that is different and should be in a fresh span.
func (l *Log) Step(msg string, mods ...SeedModifier) *Log {
	seed := l.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&l.local.StepCounter, 1))
	seed.prefix += "." + strconv.Itoa(counter)
	return l.newChildLog(seed)
}

func (l *Log) Request() *Span {
	return l.request
}

func (l *Log) Span() *Span {
	return &l.span
}

func (s *Span) SetType(spanType xopconst.SpanType) {
	func() {
		s.dataLock.Lock()
		defer s.dataLock.Unlock()
		s.spanType = spanType
	}()
	s.log.setDirty()
}

func (s *Span) AddData(additionalData ...xop.Thing) {
	func() {
		s.dataLock.Lock()
		defer s.dataLock.Unlock()
		s.data = append(s.data, additionalData...)
	}()
	s.log.setDirty()
}

func (l *Log) LogThings(level xopconst.Level, msg string, values ...xop.Thing) {
	l.log(level, msg, values)
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

func (ll *LogLine) Msg(msg string) {
	ll.line.Msg(msg)
	ll.log.span.linePool.Put(ll)
	ll.log.enableFlushTimer()
}

func (l *Log) Debug() *LogLine { return l.LogLine(xopconst.DebugLevel) }
func (l *Log) Trace() *LogLine { return l.LogLine(xopconst.TraceLevel) }
func (l *Log) Info() *LogLine  { return l.LogLine(xopconst.InfoLevel) }
func (l *Log) Warn() *LogLine  { return l.LogLine(xopconst.WarnLevel) }
func (l *Log) Error() *LogLine { return l.LogLine(xopconst.ErrorLevel) }
func (l *Log) Alert() *LogLine { return l.LogLine(xopconst.AlertLevel) }

// TODO: generate these
// TODO: the rest of the set
func (ll *LogLine) Msgf(msg string, v ...interface{})   { ll.Msg(fmt.Sprintf(msg, v...)) }
func (ll *LogLine) Msgs(v ...interface{})               { ll.Msg(fmt.Sprint(v...)) }
func (ll *LogLine) Int(k string, v int) *LogLine        { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int8(k string, v int8) *LogLine      { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int16(k string, v int16) *LogLine    { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int32(k string, v int32) *LogLine    { ll.line.Int(k, int64(v)); return ll }
func (ll *LogLine) Int64(k string, v int64) *LogLine    { ll.line.Int(k, v); return ll }
func (ll *LogLine) Uint(k string, v uint) *LogLine      { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint8(k string, v uint8) *LogLine    { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint16(k string, v uint16) *LogLine  { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint32(k string, v uint32) *LogLine  { ll.line.Uint(k, uint64(v)); return ll }
func (ll *LogLine) Uint64(k string, v uint64) *LogLine  { ll.line.Uint(k, v); return ll }
func (ll *LogLine) Str(k string, v string) *LogLine     { ll.line.Str(k, v); return ll }
func (ll *LogLine) Bool(k string, v bool) *LogLine      { ll.line.Bool(k, v); return ll }
func (ll *LogLine) Time(k string, v time.Time) *LogLine { ll.line.Time(k, v); return ll }
func (ll *LogLine) Error(k string, v error) *LogLine    { ll.line.Error(k, v); return ll }

// AnyImmutable can be used to log something that is not going to be further modified
// after this call.
func (ll LogLine) AnyImmutable(k string, v interface{}) LogLine { ll.line.Any(k, v); return ll }

// Any can be used to log something that might be modified after this call.  If any base
// logger does not immediately serialize, then the object will be copied using
// github.com/mohae/deepcopy.Copy()
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

func (s *Span) CurrentPrefill() []xop.Thing {
	c := make([]xop.Thing, len(s.seed.prefill))
	copy(c, s.seed.prefill)
	return c
}

func copyMap(o map[string]interface{}) map[string]interface{} {
	n := make(map[string]interface{})
	for k, v := range o {
		n[k] = v
	}
	return n
}

func (s *Span) TraceState() trace.State     { return s.seed.traceBundle.State }
func (s *Span) TraceBaggage() trace.Baggage { return s.seed.traceBundle.Baggage }
func (s *Span) TraceParent() trace.Trace    { return s.seed.traceBundle.TraceParent.Copy() }
func (s *Span) Trace() trace.Trace          { return s.seed.traceBundle.Trace.Copy() }
func (s *Span) Bundle() trace.Bundle        { return s.seed.traceBundle.Copy() }
