package xoplog

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
)

type Log struct {
	span    Span
	request *Span
	local   local   // one span in a request
	shared  *shared // shared between spans in a request
}

type Span struct {
	seed      Seed
	dataLock  sync.Mutex // protects Data & SpanType
	data      []zop.Thing
	spanType  SpanType
	startTime time.Time
	endTime   int64 // unix nano
	log       *Log  // back to self
}

type local struct {
	ForkCounter    int32
	StepCounter    int32
	Created        time.Time
	InDirty        int32 // in shared.Dirty? 0 = false, 1 = true
	IsBufferParent bool
}

// shared is common between the loggers that share a search index
type shared struct {
	RefCount      int32
	UnflushedLogs int32
	FlushTimer    *time.Timer
	FlushDelay    time.Duration
	FlushActive   int32 // 1 == timer is running, 0 = timer is not running

	// Dirty holds spans that have modified data and need to be
	// written or re-written.  It does not not track logs that
	// need flushing. Protected by Span.dataLock.
	Dirty []*Log
}

var DefaultFlushDelay = time.Minute * 5

func (s Seed) Request(description string) *Log {
	s = s.Copy()
	s.myTrace.RebuildSetNonZero()
	log := Log{
		span: Span{
			seed:      s,
			data:      copyMap(s.data),
			startTime: time.Now(),
		},
		shared: &shared{
			RefCount:    1,
			FlushActive: 1,
			Index:       make(map[string][]string),
		},
		local: local{
			InDirty:        1,
			Created:        time.Now(),
			IsBufferParent: true,
		},
	}
	log.span.log = &log
	log.request = &log.span
	log.shared.Dirty = append(log.shared.Dirty, log)
	log.finishBaseLoggerChanges()
	log.shared.FlushTimer = time.AfterFunc(DefaultFlushDelay, log.timerFlush)
	return log
}

func (old *Log) newChildLog(seed Seed) *Log {
	log := &Log{
		span: &Span{
			seed:      seed,
			startTime: time.Now(),
		},
		local: local{
			InDirty:        1,
			Created:        time.Now(),
			IsBufferParent: false,
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

func (s *Span) touched() {
	wasInDirty := atomic.SwapInt32(&s.log.local.InDirty, 1)
	if wasInDirty == 0 {
		s.log.request.dataLock.Lock()
		defer s.log.request.dataLock.Unlock()
		s.log.shared.Dirty = append(l.shared.Dirty, l)
		if len(s.log.shared.Dirty) == 1 {
			s.log.enableFlushTimer()
		}
	}
}

func (l *Log) enableFlushTimer() {
	was := atomic.SwapInt32(&l.shared.FlushActive, 1)
	if was == 0 {
		l.shared.FlushTimer.Reset(l.shared.FlushDelay)
	}
}

func (l *Log) timerFlush() {
	atomic.StoreInt32(&l.shared.FlushActive, 0)
	l.Flush()
}

func (l *Log) Flush() {
	atomic.StoreInt32(&l.shared.UnflushedLogs, 0)
	func() {
		l.request.dataLock.Lock()
		defer l.request.dataLock.Unlock()
		for _, dirtyLog := range l.shared.Dirty {
			atomic.StoreInt32(&dirtyLog.local.InDirty, 0) // TODO: need atomic?
			var index map[string][]string
			var data map[string]interface{}
			if dirtyLog.local.IsBufferParent {
				index = dirtyLog.shared.Index
				data = dirtyLog.shared.Data
			} else {
				func() {
					dirtyLog.local.DataLock.Lock()
					defer dirtyLog.local.DataLock.Unlock()
					data = dirtyLog.local.Data
				}()
			}
			for _, baseLogger := range l.seed.baseLoggers.List {
				// XXX still need this?
				baseLogger.Buffered.Span(
					dirtyLog.seed.description,
					dirtyLog.seed.myTrace,
					dirtyLog.seed.parentTrace,
					index,
					data)
			}
		}
		l.shared.Dirty = l.shared.Dirty[:0]
	}()
	for _, baseLogger := range l.seed.baseLoggers.List {
		baseLogger.Buffered.Flush()
	}
}

func (l *Log) log(level xopconst.Level, msg string, values []xop.Thing) {
	unflushed := atomic.AddInt32(&l.shared.UnflushedLogs, 1)
	if unflushed == 1 {
		l.enableFlushTimer()
	}
	for _, baseLogger := range l.seed.baseLoggers.List {
		baseLogger.Prefilled.Log(level, msg, values)
	}
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
	now := time.Now().UnixNano()
	atomic.StoreInt64(&l.span.endTime, now)
	atomic.StoreInt64(&l.request.endTime, now)
	if remaining <= 0 {
		l.Flush()
	}
}

// Wait modifies (and returns) a Log to indicate that the overall
// request is not finished until an additional Done() is called.
func (l *Log) Wait() *Log {
	remaining := atomic.AddInt32(&l.shared.RefCount, 1)
	if remaining > 1 {
		return
	}
	// This indicates a bug in the code that is using the
	// logger.
	l.Warn("Too many calls to log.Done()")
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
	s.DataLock.Lock()
	defer s.DataLock.Unlock()
	s.SpanType = spanType
}

func (s *Span) AddData(additionalData ...xop.Thing) {
	func() {
		s.DataLock.Lock()
		defer s.DataLock.Unlock()
		s.Data = append(s.Data, additionalData...)
	}()
	l.touched()
}

func (l *Log) LogThings(level Level, msg string, values ...xop.Thing) { l.log(level, msg, values) }

func (l *Log) Debug() *LogLine { return l.logLine(DebugLevel) }
func (l *Log) Trace() *LogLine { return l.logLine(TraceLevel) }
func (l *Log) Info() *LogLine  { return l.logLine(InfoLevel) }
func (l *Log) Warn() *LogLine  { return l.logLine(WarnLevel) }
func (l *Log) Error() *LogLine { return l.logLine(ErrorLevel) }
func (l *Log) Alert() *LogLine { return l.logLine(AlertLevel) }

type LogLine struct {
	log          *Log
	pendingLines []PendingLine
}

func (l *Log) logLine(level xopconst.Level) {
	// TODO: Allocation
	ll := &LogLine{
		log: l,
	}
	for _, base := range l.seed.baseLoggers.List {
		ll.pendingLines = append(ll.pendingLines, base.Start(level))
	}
	return ll
}

// TODO: generate these
func (ll *LogLine) Int(name string, value int) *LogLine {
	for _, line := range ll.pendingLines {
		line.Int(name, value)
	}
	return ll
}
func (ll *LogLine) Str(name string, value string) *LogLine {
	for _, line := range ll.pendingLines {
		line.Str(name, value)
	}
	return ll
}

func (ll *LogLine) Msg(msg string) {
	for _, base := range ll.pendingLines {
		line.Msg(msg)
	}
}

func (ll *LogLine) Msgf(msg string, v ...interface{}) {
	ll.Msg(fmt.Sprintf(msg, v...))
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

func (s *Span) TraceState() trace.State     { return s.seed.state }
func (s *Span) TraceBaggage() trace.Baggage { return s.seed.baggage }
func (s *Span) TraceParent() trace.Trace    { return s.seed.parentTrace }
func (s *Span) Trace() trace.Trace          { return s.seed.myTrace }
