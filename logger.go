package xop

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopnum"
	"github.com/muir/xop-go/xoputil"
)

type Log struct {
	request       *Log              // The ancestor request-level Log
	span          *span             // span associated with this Log
	capSpan       *Span             // Span associated with this Log
	parent        *Log              // immediate parent Log
	shared        *shared           // shared between logs with same request-level Log
	settings      LogSettings       // settings for this Log
	prefilled     xopbase.Prefilled // XXX double-check,  in request?
	nonSpanSubLog bool              // true if created by log.Sub().Log()
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
	description      string
	stepCounter      int32 //nolint:structcheck // false report
	forkCounter      int32 //nolint:structcheck // false report
	detached         bool
	dependentLock    sync.Mutex
	activeDependents map[int32]*Log
	doneCount        int32
	knownActive      int32
	logNumber        int32
}

// shared is common between the loggers that share a search index
type shared struct {
	FlushLock          sync.Mutex // protects Flush() (can be held for a longish period)
	FlusherLock        sync.RWMutex
	FlushTimer         *time.Timer
	FlushDelay         time.Duration
	FlushActive        int32                      // 1 == timer is running, 0 = timer is not running
	Flushers           map[string]xopbase.Request // key is xopbase.Logger.ID() // TODO: change key to int?
	Description        string
	LogCount           int32
	ActiveDetached     map[int32]*Log
	WaitingForDetached bool // true only when request is Done but is not yet flushed due to detached
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
			description: descriptionOrName,
			knownActive: 1,
		},
		shared: shared{
			FlushActive:    1,
			FlushDelay:     seed.config.FlushDelay,
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
	DebugPrint("starting timer", seed.config.FlushDelay)
	log.shared.FlushTimer = time.AfterFunc(seed.config.FlushDelay, log.timerFlush)
	if !log.span.buffered {
		DebugPrint("stopping timer")
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
			request:       sub.log.request,
			span:          sub.log.span,
			capSpan:       sub.log.capSpan,
			parent:        sub.log.parent,
			shared:        sub.log.shared,
			settings:      sub.settings,
			nonSpanSubLog: true,
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
	log.hasActivity(false)
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
			detached:    detached,
			description: description,
			knownActive: 1,
			logNumber:   atomic.AddInt32(&old.shared.LogCount, 1),
		},
	}
	alloc.Span.span = &alloc.span
	alloc.Span.log = &alloc.Log
	alloc.Log.capSpan = &alloc.Span
	alloc.Log.span = &alloc.span
	log := &alloc.Log

	log.span.base = old.span.base.Span(time.Now(), spanSeed.traceBundle, description, log.span.seed.spanSequenceCode)
	if len(spanSeed.loggers.Added) == 0 && len(spanSeed.loggers.Removed) == 0 {
		log.span.buffered = old.span.buffered
		log.span.referencesKept = old.span.referencesKept
	} else {
		DebugPrint("adjusting set of flusher", description, log.span.logNumber)
		spanSet := make(map[string]xopbase.Span)
		if baseSpans, ok := log.span.base.(baseSpans); ok {
			for _, baseSpan := range baseSpans {
				spanSet[baseSpan.ID()] = baseSpan
			}
		} else {
			spanSet[log.span.base.ID()] = log.span.base
		}
		for _, removed := range spanSeed.loggers.Removed {
			id := removed.ID()
			DebugPrint("remove flusher", id)
			delete(spanSet, id)
		}
		ts := time.Now()
		for _, added := range spanSeed.loggers.Added {
			id := added.ID()
			if _, ok := spanSet[id]; ok {
				DebugPrint("ignoring additional flusher, in span set", id)
				continue
			}
			if func() bool {
				log.shared.FlusherLock.RLock()
				defer log.shared.FlusherLock.RUnlock()
				_, ok := log.shared.Flushers[id]
				return ok
			}() {
				DebugPrint("ignoring additional flusher, already in flusher set", id)
				continue
			}
			req := added.Request(ts, log.request.span.seed.traceBundle, log.shared.Description)
			spanSet[id] = req
			req.SetErrorReporter(log.span.seed.config.ErrorReporter)
			DebugPrint("adding flusher to flusher set", id)
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
	log.sendPrefill()
	log.addMyselfAsDependent()
	return log
}

func (log *Log) addMyselfAsDependent() bool {
	if log == log.request {
		return false
	}
	if log.span.detached {
		log.Info().String("span", log.span.seed.traceBundle.Trace.SpanID().String()).Msg("XXX is detached")
		log.request.span.dependentLock.Lock()
		defer log.request.span.dependentLock.Unlock()
		log.shared.ActiveDetached[log.span.logNumber] = log
		return false
	}
	log.Info().String("span", log.span.seed.traceBundle.Trace.SpanID().String()).Msg("XXX is attached")
	log.parent.span.dependentLock.Lock()
	defer log.parent.span.dependentLock.Unlock()
	if log.parent.span.activeDependents == nil {
		log.parent.span.activeDependents = make(map[int32]*Log)
	}
	DebugPrint("add to active deps", log.span.description, ":", log.span.logNumber)
	log.parent.span.activeDependents[log.span.logNumber] = log
	return len(log.parent.span.activeDependents) == 1
}

func (log *Log) hasActivity(startFlusher bool) {
	was := atomic.SwapInt32(&log.span.knownActive, 1)
	if was == 0 {
		DebugPrint("now has activity!", log.span.description, log.span.logNumber)
		if log.addMyselfAsDependent() {
			log.parent.hasActivity(false)
		}
		if startFlusher {
			wasFlushing := atomic.SwapInt32(&log.shared.FlushActive, 1)
			if wasFlushing == 0 {
				DebugPrint("restarting timer", log.shared.FlushDelay)
				log.shared.FlushTimer.Reset(log.shared.FlushDelay)
			}
			if wasDone := atomic.LoadInt32(&log.span.doneCount); wasDone != 0 {
				log.Error().Static("XOP: log was already done, but was used again")
			}
		}
	}
}

// Done is used to indicate that a log is complete.  Buffered base loggers
// (implementing xopbase.Logger) wait for Done to be called before flushing their data.
//
// Requests (Log's created by seed.Request()) and detached logs (created by
// log.Sub().Detach().Fork() or log.Sub().Detah().Step()) are considered
// top-level logs.  A call to Done() on a top-level log marks all
// sub-spans as Done() if they were not already marked done.
//
// Calling Done() on a log that is already done generates a logged error
// message.
//
// Non-detached sub-spans (created by log.Sub().Fork() and log.Sub().Step())
// are done when either Done is called on the sub-span or when their associated
// top-level log is done.
//
// Sub-logs that are not spans (created by log.Sub().Log()) should not use
// Done().  Any call to Done on such a sub-log will log an error and otherwise
// be ignored.
//
// When all the logs associated with a Request are done, the Flush() is
// automatically triggered.  The Flush() call can be
// synchronous or asynchronous depending on the SynchronousFlush settings
// of the request-level Log.
//
// Any logging activity after a Done() causes an error to be logged and may
// trigger a call to Flush().
func (log *Log) Done() {
	if log.nonSpanSubLog {
		log.Error().Static("XOP: invalid call to Done() in non-span sub-log")
		return
	}
	DebugPrint("starting Done {", log.span.description, log.span.logNumber)
	log.done(true, time.Now())
	DebugPrint("done with Done }", log.span.description, log.span.logNumber)
}

func (log *Log) recursiveDone(done bool, now time.Time) (count int32) {
	DebugPrint("recursive done,", done, ",", log.span.description, log.span.logNumber)
	if done {
		atomic.StoreInt32(&log.span.knownActive, 0)
		count = atomic.AddInt32(&log.span.doneCount, 1)
		log.span.base.Done(time.Now(), true)
	} else {
		if atomic.SwapInt32(&log.span.knownActive, 0) == 1 {
			log.span.base.Done(now, false)
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
		DebugPrint("dep of", log.span.logNumber, ":", dep.span.description, dep.span.logNumber)
		dep.done(done, now)
	}
	return
}

func (log *Log) done(explicit bool, now time.Time) {
	postCount := log.recursiveDone(true, now)
	if postCount > 1 && explicit {
		DebugPrint("donecount=", postCount, "logging error")
		log.Error().Static("XOP: Done() called on log object when it was already Done()")
	}
	if log.span.detached {
		if func() bool {
			log.request.span.dependentLock.Lock()
			defer log.request.span.dependentLock.Unlock()
			delete(log.shared.ActiveDetached, log.span.logNumber)
			if log.shared.WaitingForDetached &&
				len(log.shared.ActiveDetached) == 0 &&
				len(log.request.span.activeDependents) == 0 {
				log.shared.WaitingForDetached = false
				return true
			}
			return false
		}() {
			DebugPrint("request was waiting, now we can flush")
			log.request.flush()
		}
		DebugPrint("we're detached, finished done")
		return
	}
	if log.parent == log {
		DebugPrint("in done, we're the request!")
		if func() bool {
			log.span.dependentLock.Lock()
			defer log.span.dependentLock.Unlock()
			if len(log.span.activeDependents) != 0 {
				return false
			}
			if len(log.shared.ActiveDetached) != 0 {
				DebugPrint("we have detached that are not yet done, waiting for them before flushing")
				log.shared.WaitingForDetached = true
				return false
			}
			return true
		}() {
			DebugPrint("...and we're flushing")
			log.request.flush()
			DebugPrint("...done flushing")
		}
		return
	}
	log.parent.span.dependentLock.Lock()
	defer log.parent.span.dependentLock.Unlock()
	DebugPrint("delete from active deps", log.span.description, ":", log.span.logNumber)
	delete(log.parent.span.activeDependents, log.span.logNumber)
}

// timerFlush is only called by log.shared.FlushTimer
func (log *Log) timerFlush() {
	DebugPrint("timer flush!")
	log.Flush()
}

func (log *Log) flush() {
	if log.settings.synchronousFlushWhenDone {
		log.Flush()
	} else {
		DebugPrint("doing async flush")
		go func() {
			SmallSleepForTesting()
			log.Flush()
		}()
	}
}

func (log *Log) Flush() {
	DebugPrint("begin flush {", Stack())
	now := time.Now()
	log.request.detachedDone(now)
	log.request.recursiveDone(false, now)
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
		DebugPrint("flushing", flusher.ID())
		go func(flusher xopbase.Request) {
			defer wg.Done()
			flusher.Flush()
		}(flusher)
	}
	wg.Wait()
	DebugPrint("done flush }")
}

func (log *Log) detachedDone(now time.Time) {
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
		_ = dep.recursiveDone(false, now)
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

const stackFramesToExclude = 4

func (log *Log) logLine(level xopnum.Level) *Line {
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
					log.settings.stackFramesWanted[xopnum.AlertLevel])
			} else {
				ll.pc = ll.pc[:cap(ll.pc)]
			}
			n := runtime.Callers(stackFramesToExclude, ll.pc)
			ll.pc = ll.pc[:n]
		}
	} else {
		ll = &Line{
			log: log,
		}
		if !skip && log.settings.stackFramesWanted[level] != 0 {
			ll.pc = make([]uintptr, log.settings.stackFramesWanted[level],
				log.settings.stackFramesWanted[xopnum.AlertLevel])
			n := runtime.Callers(stackFramesToExclude, ll.pc)
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

func (log *Log) Line(level xopnum.Level) *Line { return log.logLine(level) }
func (log *Log) Debug() *Line                  { return log.Line(xopnum.DebugLevel) }
func (log *Log) Trace() *Line                  { return log.Line(xopnum.TraceLevel) }
func (log *Log) Info() *Line                   { return log.Line(xopnum.InfoLevel) }
func (log *Log) Warn() *Line                   { return log.Line(xopnum.WarnLevel) }
func (log *Log) Error() *Line {
	log.notBoring()
	return log.Line(xopnum.ErrorLevel)
}
func (log *Log) Alert() *Line {
	log.notBoring()
	return log.Line(xopnum.AlertLevel)
}

func (line *Line) Msgs(v ...interface{}) { line.Msg(fmt.Sprint(v...)) }
