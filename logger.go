package xop

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"

	"github.com/mohae/deepcopy"
)

type Log struct {
	request   *Log // XXX changed from *Span
	span      *span
	capSpan   *Span   // added
	parent    *Log    // added
	shared    *shared // shared between spans in a request
	settings  LogSettings
	prefilled xopbase.Prefilled // XXX double-check,  in request?
}

type Span struct {
	*span
	log *Log
}

type span struct {
	seed             spanSeed
	referencesKept   bool
	base             xopbase.Span //nolint:structcheck // false report
	linePool         sync.Pool    //nolint:structcheck // false report
	boring           int32        // 0 = boring
	buffered         bool         //nolint:structcheck // false report
	stepCounter      int32        //nolint:structcheck // false report
	forkCounter      int32        //nolint:structcheck // false report
	spanNumber       int32
	detached         bool
	doneCount        int32
	knownActive      int32 // XXX added
	dependentLock    sync.Mutex
	activeDependents map[int32]*Log // XXX added
}

// shared is common between the loggers that share a search index
type shared struct {
	FlushLock      sync.Mutex // protects Flush() (can be held for a longish period)
	FlusherLock    sync.RWMutex
	FlushTimer     *time.Timer
	FlushDelay     time.Duration
	FlushActive    int32                      // 1 == timer is running, 0 = timer is not running
	Flushers       map[string]xopbase.Request // key is xopbase.Logger.ID() // XXX change key to int?
	Description    string
	Spans          map[int]*Log   // XXX added
	SpanCount      int32          // XXX added
	ActiveDetached map[int32]*Log // XXX added
}

func (seed Seed) Request(descriptionOrName string) *Log {
	seed.traceBundle.Trace.RebuildSetNonZero()

	type singleAlloc struct {
		Log    Log
		shared shared
		Span   Span
		span   span
	}
	alloc := singleAlloc{
		Log: Log{
			settings: seed.settings.Copy(),
			// XXX prefilled?
		},
		span: span{
			seed:        seed.spanSeed.Copy(),
			knownActive: 1,
		},
		shared: shared{
			FlushActive:    1,
			Description:    descriptionOrName,
			ActiveDetached: make(map[int32]*Log),
		},
	}
	alloc.Span.span = &alloc.span
	alloc.Span.log = &alloc.Log
	alloc.Log.capSpan = &alloc.Span
	alloc.Log.span = &alloc.span
	alloc.Log.request = &alloc.Log
	alloc.Log.parent = &alloc.Log
	alloc.Log.shared = &alloc.shared
	log := &alloc.Log

	combinedBaseRequest, flushers := log.span.seed.loggers.List.StartRequests(time.Now(), log.span.seed.traceBundle, descriptionOrName)
	log.shared.Flushers = flushers
	combinedBaseRequest.SetErrorReporter(seed.config.ErrorReporter)
	log.span.referencesKept = log.span.seed.loggers.List.ReferencesKept()
	log.span.buffered = log.span.seed.loggers.List.Buffered()
	log.span.base = combinedBaseRequest.(xopbase.Span)
	log.sendPrefill()
	log.shared.FlushTimer = time.AfterFunc(seed.config.FlushDelay, log.timerFlush)
	if !log.span.buffered {
		log.shared.FlushTimer.Stop()
		log.shared.FlushActive = 0
	}
	return log
}

// Log creates a new Log that does not need to be terminated because
// it is assumed to be done with the current log is finished.  The new log
// shares a span with its parent log. It can have different settings from its
// parent log.
func (sub *Sub) Log() *Log {
	type singleAlloc struct {
		Log  Log
		Span Span
	}
	alloc := singleAlloc{
		Log: Log{
			request:  sub.log.request,
			span:     sub.log.span,
			capSpan:  sub.log.capSpan,
			parent:   sub.log.parent,
			shared:   sub.log.shared,
			settings: sub.settings,
			// XXX prefilled?
		},
		Span: Span{
			span: sub.log.span,
		},
	}
	alloc.Span.log = &alloc.Log
	alloc.Log.capSpan = &alloc.Span
	log := &alloc.Log
	log.sendPrefill()
	return log
}

func (old *Log) newChildLog(spanSeed spanSeed, description string, settings LogSettings, detached bool) *Log {
	type singleAlloc struct {
		Log  Log
		Span Span
		span span
	}
	alloc := singleAlloc{
		Log: Log{
			request:  old.request,
			parent:   old,
			shared:   old.shared,
			settings: settings,
			// XXX prefilled?
		},
		span: span{
			seed:        spanSeed,
			knownActive: 1,
			spanNumber:  atomic.AddInt32(&old.shared.SpanCount, 1),
			detached:    detached,
		},
	}
	alloc.Span.span = &alloc.span
	alloc.Span.log = &alloc.Log
	alloc.Log.capSpan = &alloc.Span
	alloc.Log.span = &alloc.span
	log := &alloc.Log

	log.span.base = old.span.base.Span(time.Now(), spanSeed.traceBundle, description)
	if len(spanSeed.loggers.Added) == 0 && len(spanSeed.loggers.Removed) == 0 {
		log.span.buffered = old.span.buffered
		log.span.referencesKept = old.span.referencesKept
	} else {
		spanSet := make(map[string]xopbase.Span)
		if baseSpans, ok := log.span.base.(baseSpans); ok {
			for _, baseSpan := range baseSpans {
				spanSet[baseSpan.ID()] = baseSpan
			}
		}
		for _, removed := range spanSeed.loggers.Removed {
			delete(spanSet, removed.ID())
		}
		ts := time.Now()
		for _, added := range spanSeed.loggers.Added {
			id := added.ID()
			if _, ok := spanSet[id]; ok {
				continue
			}
			if func() bool {
				log.shared.FlusherLock.RLock()
				defer log.shared.FlusherLock.RUnlock()
				_, ok := log.shared.Flushers[id]
				return ok
			}() {
				continue
			}
			req := added.Request(ts, log.request.span.seed.traceBundle, log.shared.Description)
			req.SetErrorReporter(log.span.seed.config.ErrorReporter)
			func() {
				log.shared.FlusherLock.Lock()
				defer log.shared.FlusherLock.Unlock()
				log.shared.Flushers[id] = req
			}()
		}
		if len(spanSet) == 1 {
			for _, baseSpan := range spanSet {
				log.span.base = baseSpan
			}
		} else {
			spans := make(baseSpans, 0, len(spanSet))
			for _, baseSpan := range spanSet {
				spans = append(spans, baseSpan)
			}
			log.span.base = spans
		}
		log.span.buffered = log.span.seed.loggers.List.Buffered()
		log.span.referencesKept = log.span.seed.loggers.List.ReferencesKept()
	}
	log.span.base.Boring(true)
	log.Span().String(xopconst.SpanSequenceCode, log.span.seed.spanSequenceCode) // TODO: improve  (not efficient)
	log.sendPrefill()
	return log
}

func (log *Log) addMyselfAsDependent() bool {
	if log == log.request {
		return false
	}
	if log.span.detached {
		log.request.span.dependentLock.Lock()
		defer log.request.span.dependentLock.Unlock()
		log.shared.ActiveDetached[log.span.spanNumber] = log
		return false
	}
	log.parent.span.dependentLock.Lock()
	defer log.parent.span.dependentLock.Unlock()
	if log.parent.span.activeDependents == nil {
		log.parent.span.activeDependents = make(map[int32]*Log)
	}
	log.parent.span.activeDependents[log.span.spanNumber] = log
	return len(log.parent.span.activeDependents) == 1
}

func (log *Log) hasActivity(startFlusher bool) {
	was := atomic.SwapInt32(&log.span.knownActive, 1)
	if was == 0 {
		if log.addMyselfAsDependent() {
			log.parent.hasActivity(false)
		}
		if startFlusher {
			wasFlushing := atomic.SwapInt32(&log.shared.FlushActive, 1)
			if wasFlushing == 0 {
				log.shared.FlushTimer.Reset(log.shared.FlushDelay)
			}
		}
	}
}

// Done is used to indicate that a log is complete.  All non-Detach()ed
// logs that were created from this log are also considered complete.
//
// Any logging activity after a Done() causes an error to be logged and may
// trigger a call to Flush().
//
// Done can be synchronous or asynchronous depending on the SynchronousFlush
// settings.
func (log *Log) Done() {
	if log.settings.synchronousFlushWhenDone {
		log.done(true, true, time.Now())
	} else {
		go log.done(true, true, time.Now())
	}
}

func (log *Log) recursiveDone(done bool) (count int32) {
	if done {
		atomic.StoreInt32(&log.span.knownActive, 0)
		count = atomic.AddInt32(&log.span.doneCount, 1)
		log.span.base.Done(time.Now())
	} else {
		if atomic.SwapInt32(&log.span.knownActive, 0) == 1 {
			log.span.base.Done(time.Now())
		}
	}
	deps := func() []*Log {
		log.span.dependentLock.Lock()
		defer log.span.dependentLock.Unlock()
		deps := make([]*Log, 0, len(log.span.activeDependents))
		for _, dep := range log.span.activeDependents {
			deps = append(deps, dep)
		}
		return deps
	}()
	for _, dep := range deps {
		_ = dep.recursiveDone(done)
	}
	return
}

func (log *Log) done(explicit bool, doUp bool, now time.Time) {
	postCount := log.recursiveDone(true)
	if postCount > 1 && explicit {
		log.Error().Msg("Done() called on log object when it was already Done()")
	}
	if log.span.detached {
		if func() bool {
			log.request.span.dependentLock.Lock()
			defer log.request.span.dependentLock.Unlock()
			delete(log.shared.ActiveDetached, log.span.spanNumber)
			return len(log.shared.ActiveDetached) == 0 &&
				len(log.request.span.activeDependents) == 0
		}() {
			log.request.Flush()
		}
		return
	}
	if !doUp {
		return
	}
	if log.parent == log {
		// we're the request!
		if func() bool {
			log.span.dependentLock.Lock()
			defer log.span.dependentLock.Unlock()
			return len(log.shared.ActiveDetached) == 0 &&
				len(log.span.activeDependents) == 0
		}() {
			log.request.Flush()
		}
		return
	}
	if func() bool {
		log.parent.span.dependentLock.Lock()
		defer log.parent.span.dependentLock.Unlock()
		delete(log.parent.span.activeDependents, log.span.spanNumber)
		return len(log.parent.span.activeDependents) == 0
	}() {
		log.parent.done(explicit, doUp, now)
	}
}

// timerFlush is only called by log.shared.FlushTimer
func (log *Log) timerFlush() {
	log.Flush()
}

func (log *Log) Flush() {
	log.request.detachedDone()
	log.request.recursiveDone(false)
	flushers := func() []xopbase.Request {
		log.shared.FlusherLock.RLock()
		defer log.shared.FlusherLock.RUnlock()
		requests := make([]xopbase.Request, 0, len(log.shared.Flushers))
		for _, req := range log.shared.Flushers {
			requests = append(requests, req)
		}
		return requests
	}()
	log.shared.FlushLock.Lock()
	defer log.shared.FlushLock.Unlock()
	// Stop is is not thread-safe with respect to other calls to Stop
	log.shared.FlushTimer.Stop()
	atomic.StoreInt32(&log.shared.FlushActive, 0)
	var wg sync.WaitGroup
	wg.Add(len(flushers))
	for _, flusher := range flushers {
		go func(flusher xopbase.Request) {
			defer wg.Done()
			flusher.Flush()
		}(flusher)
	}
	wg.Wait()
}

func (log *Log) detachedDone() {
	deps := func() []*Log {
		log.request.span.dependentLock.Lock()
		defer log.request.span.dependentLock.Unlock()
		deps := make([]*Log, 0, len(log.shared.ActiveDetached))
		for _, dep := range log.shared.ActiveDetached {
			deps = append(deps, dep)
		}
		return deps
	}()
	for _, dep := range deps {
		_ = dep.recursiveDone(false)
	}
}

// Marks this request as boring.  Any log at the Alert or
// Error level will mark this request as not boring.
func (log *Log) Boring() {
	requestBoring := atomic.LoadInt32(&log.request.span.boring)
	if requestBoring != 0 {
		return
	}
	log.request.span.base.Boring(true)
	// There is chance that in the time we were sending that
	// boring=true, the the request became un-boring. If that
	// happened, we can't tell if we're currently marked as
	// boring, so let's make sure we're not boring by sending
	// a false
	requestStillBoring := atomic.LoadInt32(&log.request.span.boring)
	if requestStillBoring != 0 {
		log.request.span.base.Boring(false)
	}
	log.hasActivity(true)
}

func (log *Log) notBoring() {
	spanBoring := atomic.AddInt32(&log.span.boring, 1)
	if spanBoring == 1 {
		log.span.base.Boring(false)
		requestBoring := atomic.AddInt32(&log.request.span.boring, 1)
		if requestBoring == 1 {
			log.request.span.base.Boring(false)
		}
		log.hasActivity(true)
	}
}

type Line struct {
	log  *Log
	line xopbase.Line
	pc   []uintptr
	skip bool
}

func (log *Log) logLine(level xopconst.Level) *Line {
	skip := level < log.settings.minimumLogLevel
	recycled := log.span.linePool.Get()
	var ll *Line
	if recycled != nil {
		// TODO: try using Line instead of *Line
		ll = recycled.(*Line)
		if skip || log.settings.stackFramesWanted[level] == 0 {
			if ll.pc != nil {
				ll.pc = ll.pc[:0]
			}
		} else {
			if ll.pc == nil {
				ll.pc = make([]uintptr, log.settings.stackFramesWanted[level],
					log.settings.stackFramesWanted[xopconst.AlertLevel])
			} else {
				ll.pc = ll.pc[:cap(ll.pc)]
			}
			n := runtime.Callers(3, ll.pc)
			ll.pc = ll.pc[:n]
		}
	} else {
		ll = &Line{
			log: log,
		}
		if !skip && log.settings.stackFramesWanted[level] != 0 {
			ll.pc = make([]uintptr, log.settings.stackFramesWanted[level],
				log.settings.stackFramesWanted[xopconst.AlertLevel])
			n := runtime.Callers(3, ll.pc)
			ll.pc = ll.pc[:n]
		}
	}
	ll.skip = skip
	if ll.skip {
		ll.line = xoputil.SkipLine
	} else {
		ll.line = log.prefilled.Line(level, time.Now(), ll.pc)
	}
	return ll
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
func (line *Line) Template(template string) {
	line.line.Template(template)
	line.log.span.linePool.Put(line)
	line.log.hasActivity(true)
}

func (line *Line) Msg(msg string) {
	line.line.Msg(msg)
	line.log.span.linePool.Put(line)
	line.log.hasActivity(true)
}

// TODO: add Msgf to base loggers to pass through data elements
func (line *Line) Msgf(msg string, v ...interface{}) {
	if !line.skip {
		line.Msg(fmt.Sprintf(msg, v...))
	}
}

// Static is the same as Msg, but it hints that the supplied string is
// constant rather than something generated.  Since it's static, base
// loggers may keep them a dictionary and send references.
func (line *Line) Static(msg string) {
	line.line.Static(msg)
	line.log.span.linePool.Put(line)
	line.log.hasActivity(true)
}

func (log *Log) Line(level xopconst.Level) *Line { return log.logLine(level) }
func (log *Log) Debug() *Line                    { return log.Line(xopconst.DebugLevel) }
func (log *Log) Trace() *Line                    { return log.Line(xopconst.TraceLevel) }
func (log *Log) Info() *Line                     { return log.Line(xopconst.InfoLevel) }
func (log *Log) Warn() *Line                     { return log.Line(xopconst.WarnLevel) }
func (log *Log) Error() *Line {
	log.notBoring()
	return log.Line(xopconst.ErrorLevel)
}
func (log *Log) Alert() *Line {
	log.notBoring()
	return log.Line(xopconst.AlertLevel)
}

func (line *Line) Msgs(v ...interface{})                    { line.Msg(fmt.Sprint(v...)) }
func (line *Line) Int(k string, v int) *Line                { line.line.Int(k, int64(v)); return line }
func (line *Line) Int8(k string, v int8) *Line              { line.line.Int(k, int64(v)); return line }
func (line *Line) Int16(k string, v int16) *Line            { line.line.Int(k, int64(v)); return line }
func (line *Line) Int32(k string, v int32) *Line            { line.line.Int(k, int64(v)); return line }
func (line *Line) Int64(k string, v int64) *Line            { line.line.Int(k, v); return line }
func (line *Line) Uint(k string, v uint) *Line              { line.line.Uint(k, uint64(v)); return line }
func (line *Line) Uint8(k string, v uint8) *Line            { line.line.Uint(k, uint64(v)); return line }
func (line *Line) Uint16(k string, v uint16) *Line          { line.line.Uint(k, uint64(v)); return line }
func (line *Line) Uint32(k string, v uint32) *Line          { line.line.Uint(k, uint64(v)); return line }
func (line *Line) Uint64(k string, v uint64) *Line          { line.line.Uint(k, v); return line }
func (line *Line) String(k string, v string) *Line          { line.line.String(k, v); return line }
func (line *Line) Bool(k string, v bool) *Line              { line.line.Bool(k, v); return line }
func (line *Line) Time(k string, v time.Time) *Line         { line.line.Time(k, v); return line }
func (line *Line) Error(k string, v error) *Line            { line.line.Error(k, v); return line }
func (line *Line) Link(k string, v trace.Trace) *Line       { line.line.Link(k, v); return line }
func (line *Line) Duration(k string, v time.Duration) *Line { line.line.Duration(k, v); return line }
func (line *Line) Float64(k string, v float64) *Line        { line.line.Float64(k, v); return line }
func (line *Line) Float32(k string, v float32) *Line        { return line.Float64(k, float64(v)) }

func (line *Line) EmbeddedEnum(k xopconst.EmbeddedEnum) *Line {
	return line.Enum(k.EnumAttribute(), k)
}

func (line *Line) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) *Line {
	line.line.Enum(k, v)
	return line
}

// AnyImmutable can be used to log something that is not going to be further modified
// after this call.
func (line *Line) AnyImmutable(k string, v interface{}) *Line { line.line.Any(k, v); return line }

// Any can be used to log something that might be modified after this call.  If any base
// logger does not immediately serialize, then the object will be copied using
// https://github.com/mohae/deepcopy 's Copy().
func (line *Line) Any(k string, v interface{}) *Line {
	if line.skip {
		return line
	}
	if line.log.span.referencesKept {
		// TODO: make copy function configurable
		v = deepcopy.Copy(v)
	}
	line.line.Any(k, v)
	return line
}

// TODO: func (l *Log) Guage(name string, value float64, )
// TODO: func (l *Log) AdjustCounter(name string, value float64, )
// TODO: func (l *Log) Event
