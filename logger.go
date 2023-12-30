package xop

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/mohae/deepcopy"
)

type Logger struct {
	request       *Logger     // The ancestor request-level Logger
	span          *span       // span associated with this Logger
	capSpan       *Span       // Span associated with this Logger
	parent        *Logger     // immediate parent Logger
	shared        *shared     // shared between logs with same request-level Logger
	settings      LogSettings // settings for this Logger
	nonSpanSubLog bool        // true if created by log.Sub().Logger()
	prefilled     xopbase.Prefilled
}

type Span struct {
	*span
	logger *Logger
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
	activeDependents map[int32]*Logger
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
	Flushers           map[string]xopbase.Request // key is xopbase.Logger.ID()
	Description        string
	LogCount           int32
	ActiveDetached     map[int32]*Logger
	WaitingForDetached bool // true only when request is Done but is not yet flushed due to detached
}

type singleAllocRequest struct {
	Logger Logger
	shared shared
	Span   Span
	span   span
}

func (seed Seed) request(descriptionOrName string, now time.Time) *Logger {
	alloc := singleAllocRequest{
		Logger: Logger{
			settings: seed.settings.Copy(),
		},
		span: span{
			seed:        seed.spanSeed.copy(false),
			description: descriptionOrName,
			knownActive: 1,
		},
		shared: shared{
			FlushActive:    1,
			FlushDelay:     seed.config.FlushDelay,
			Description:    descriptionOrName,
			ActiveDetached: make(map[int32]*Logger),
		},
	}
	alloc.Span.span = &alloc.span
	alloc.Span.logger = &alloc.Logger
	alloc.Logger.capSpan = &alloc.Span
	alloc.Logger.span = &alloc.span
	alloc.Logger.request = &alloc.Logger
	alloc.Logger.parent = &alloc.Logger
	alloc.Logger.shared = &alloc.shared
	logger := &alloc.Logger

	combinedBaseRequest, flushers := logger.span.seed.loggers.List.startRequests(seed.ctx, now, logger.span.seed.traceBundle, descriptionOrName, logger.span.seed.sourceInfo)
	logger.shared.Flushers = flushers
	combinedBaseRequest.SetErrorReporter(seed.config.ErrorReporter)
	logger.span.referencesKept = logger.span.seed.loggers.List.ReferencesKept()
	logger.span.buffered = logger.span.seed.loggers.List.Buffered()
	logger.span.base = combinedBaseRequest.(xopbase.Span)
	logger.sendPrefill()
	debugPrint("starting timer", seed.config.FlushDelay)
	logger.shared.FlushTimer = time.AfterFunc(seed.config.FlushDelay, logger.timerFlush)
	runtime.SetFinalizer(&alloc, final)
	if !logger.span.buffered {
		debugPrint("stopping timer")
		logger.shared.FlushTimer.Stop()
		logger.shared.FlushActive = 0
	}
	return logger
}

// Logger creates a new Logger that does not need to be terminated because
// it is assumed to be done with the current logger is finished.  The new logger
// shares a span with its parent logger. It can have different settings from its
// parent logger.
func (sub *Sub) Logger() *Logger {
	type singleAlloc struct {
		Logger Logger
		Span   Span
	}
	alloc := singleAlloc{
		Logger: Logger{
			request:       sub.logger.request,
			span:          sub.logger.span,
			capSpan:       sub.logger.capSpan,
			parent:        sub.logger.parent,
			shared:        sub.logger.shared,
			settings:      sub.settings,
			nonSpanSubLog: true,
		},
		Span: Span{
			span: sub.logger.span,
		},
	}
	alloc.Span.logger = &alloc.Logger
	alloc.Logger.capSpan = &alloc.Span
	logger := &alloc.Logger
	logger.sendPrefill()
	logger.hasActivity(false)
	return logger
}

func (old *Logger) newChildLog(seed Seed, description string, detached bool) *Logger {
	now := time.Now()
	seed = seed.react(false, description, now)

	type singleAlloc struct {
		Logger Logger
		Span   Span
		span   span
	}
	alloc := singleAlloc{
		Logger: Logger{
			request:  old.request,
			parent:   old,
			shared:   old.shared,
			settings: seed.settings,
		},
		span: span{
			seed:        seed.spanSeed,
			detached:    detached,
			description: description,
			knownActive: 1,
			logNumber:   atomic.AddInt32(&old.shared.LogCount, 1),
		},
	}
	alloc.Span.span = &alloc.span
	alloc.Span.logger = &alloc.Logger
	alloc.Logger.capSpan = &alloc.Span
	alloc.Logger.span = &alloc.span
	logger := &alloc.Logger

	logger.span.base = old.span.base.Span(seed.ctx, now, seed.traceBundle, description, logger.span.seed.spanSequenceCode)
	if len(seed.loggers.Added) == 0 && len(seed.loggers.Removed) == 0 {
		logger.span.buffered = old.span.buffered
		logger.span.referencesKept = old.span.referencesKept
	} else {
		debugPrint("adjusting set of flusher", description, logger.span.logNumber)
		spanSet := make(map[string]xopbase.Span)
		if baseSpans, ok := logger.span.base.(baseSpans); ok {
			for _, baseSpan := range baseSpans {
				spanSet[baseSpan.ID()] = baseSpan
			}
		} else {
			spanSet[logger.span.base.ID()] = logger.span.base
		}
		for _, removed := range seed.loggers.Removed {
			id := removed.ID()
			debugPrint("remove flusher", id)
			delete(spanSet, id)
		}
		ts := now
		for _, added := range seed.loggers.Added {
			id := added.ID()
			if _, ok := spanSet[id]; ok {
				debugPrint("ignoring additional flusher, in span set", id)
				continue
			}
			if func() bool {
				logger.shared.FlusherLock.RLock()
				defer logger.shared.FlusherLock.RUnlock()
				_, ok := logger.shared.Flushers[id]
				return ok
			}() {
				debugPrint("ignoring additional flusher, already in flusher set", id)
				continue
			}
			req := added.Request(logger.request.span.seed.ctx, ts, logger.request.span.seed.traceBundle, logger.shared.Description, logger.request.span.seed.sourceInfo)
			spanSet[id] = req
			req.SetErrorReporter(logger.span.seed.config.ErrorReporter)
			debugPrint("adding flusher to flusher set", id)
			func() {
				logger.shared.FlusherLock.Lock()
				defer logger.shared.FlusherLock.Unlock()
				logger.shared.Flushers[id] = req
			}()
		}
		if len(spanSet) == 1 {
			for _, baseSpan := range spanSet {
				logger.span.base = baseSpan
			}
		} else {
			spans := make(baseSpans, 0, len(spanSet))
			for _, baseSpan := range spanSet {
				spans = append(spans, baseSpan)
			}
			logger.span.base = spans
		}
		logger.span.buffered = logger.span.seed.loggers.List.Buffered()
		logger.span.referencesKept = logger.span.seed.loggers.List.ReferencesKept()
	}
	logger.span.base.Boring(true)
	logger.sendPrefill()
	logger.addMyselfAsDependent()
	return logger
}

func (logger *Logger) addMyselfAsDependent() bool {
	if logger == logger.request {
		return false
	}
	if logger.span.detached {
		logger.request.span.dependentLock.Lock()
		defer logger.request.span.dependentLock.Unlock()
		logger.shared.ActiveDetached[logger.span.logNumber] = logger
		return false
	}
	logger.parent.span.dependentLock.Lock()
	defer logger.parent.span.dependentLock.Unlock()
	if logger.parent.span.activeDependents == nil {
		logger.parent.span.activeDependents = make(map[int32]*Logger)
	}
	debugPrint("add to active deps", logger.span.description, ":", logger.span.logNumber)
	logger.parent.span.activeDependents[logger.span.logNumber] = logger
	return len(logger.parent.span.activeDependents) == 1
}

func (logger *Logger) hasActivity(startFlusher bool) {
	was := atomic.SwapInt32(&logger.span.knownActive, 1)
	if was == 0 {
		debugPrint("now has activity!", logger.span.description, logger.span.logNumber)
		if logger.addMyselfAsDependent() {
			logger.parent.hasActivity(false)
		}
		if startFlusher {
			wasFlushing := atomic.SwapInt32(&logger.shared.FlushActive, 1)
			if wasFlushing == 0 {
				debugPrint("restarting timer", logger.shared.FlushDelay)
				logger.shared.FlushTimer.Reset(logger.shared.FlushDelay)
			}
			if wasDone := atomic.LoadInt32(&logger.span.doneCount); wasDone != 0 {
				logger.Error().Msg("XOP: logger was already done, but was used again")
			}
			logger.done(false, time.Now())
		}
	}
}

// Done is used to indicate that a logger is complete.  Buffered base loggers
// (implementing xopbase.Logger) wait for Done to be called before flushing their data.
//
// Requests (Logger's created by seed.Request()) and detached logs (created by
// logger.Sub().Detach().Fork() or logger.Sub().Detah().Step()) are considered
// top-level logs.  A call to Done() on a top-level logger marks all
// sub-spans as Done() if they were not already marked done.
//
// Calling Done() on a logger that is already done generates a logged error
// message.
//
// Non-detached sub-spans (created by logger.Sub().Fork() and logger.Sub().Step())
// are done when either Done is called on the sub-span or when their associated
// top-level logger is done.
//
// Sub-logs that are not spans (created by logger.Sub().Logger()) should not use
// Done().  Any call to Done on such a sub-logger will log an error and otherwise
// be ignored.
//
// When all the logs associated with a Request are done, the Flush() is
// automatically triggered.  The Flush() call can be
// synchronous or asynchronous depending on the SynchronousFlush settings
// of the request-level Logger.
//
// Any logging activity after a Done() causes an error to be logged and may
// trigger a call to Flush().
func (logger *Logger) Done() {
	if logger.nonSpanSubLog {
		logger.Error().Msg("XOP: invalid call to Done() in non-span sub-logger")
		return
	}
	debugPrint("starting Done {", logger.span.description, logger.span.logNumber)
	logger.done(true, time.Now())
	debugPrint("done with Done }", logger.span.description, logger.span.logNumber)
}

func (logger *Logger) recursiveDone(done bool, now time.Time) (count int32) {
	debugPrint("recursive done,", done, ",", logger.span.description, logger.span.logNumber)
	if done {
		atomic.StoreInt32(&logger.span.knownActive, 0)
		count = atomic.AddInt32(&logger.span.doneCount, 1)
		logger.span.base.Done(time.Now(), true)
	} else {
		if atomic.SwapInt32(&logger.span.knownActive, 0) == 1 {
			logger.span.base.Done(now, false)
		}
	}
	deps := func() []*Logger {
		logger.span.dependentLock.Lock()
		defer logger.span.dependentLock.Unlock()
		deps := make([]*Logger, 0, len(logger.span.activeDependents))
		for _, dep := range logger.span.activeDependents {
			deps = append(deps, dep)
		}
		return deps
	}()
	for _, dep := range deps {
		debugPrint("dep of", logger.span.logNumber, ":", dep.span.description, dep.span.logNumber)
		dep.done(done, now)
	}
	return
}

func (logger *Logger) done(explicit bool, now time.Time) {
	postCount := logger.recursiveDone(true, now)
	if postCount > 1 && explicit {
		debugPrint("donecount=", postCount, "logging error")
		logger.Error().Msg("XOP: Done() called on logger object when it was already Done()")
	}
	if logger.span.detached {
		if func() bool {
			logger.request.span.dependentLock.Lock()
			defer logger.request.span.dependentLock.Unlock()
			delete(logger.shared.ActiveDetached, logger.span.logNumber)
			if logger.shared.WaitingForDetached &&
				len(logger.shared.ActiveDetached) == 0 &&
				len(logger.request.span.activeDependents) == 0 {
				logger.shared.WaitingForDetached = false
				return true
			}
			return false
		}() {
			debugPrint("request was waiting, now we can flush")
			logger.request.flush()
		}
		debugPrint("we're detached, finished done")
		return
	}
	if logger.parent == logger {
		debugPrint("in done, we're the request!")
		if func() bool {
			logger.span.dependentLock.Lock()
			defer logger.span.dependentLock.Unlock()
			if len(logger.span.activeDependents) != 0 {
				return false
			}
			if len(logger.shared.ActiveDetached) != 0 {
				debugPrint("we have detached that are not yet done, waiting for them before flushing")
				logger.shared.WaitingForDetached = true
				return false
			}
			return true
		}() {
			debugPrint("...and we're flushing")
			logger.request.flush()
			debugPrint("...done flushing")
		}
		return
	}
	logger.parent.span.dependentLock.Lock()
	defer logger.parent.span.dependentLock.Unlock()
	debugPrint("delete from active deps", logger.span.description, ":", logger.span.logNumber)
	delete(logger.parent.span.activeDependents, logger.span.logNumber)
}

// timerFlush is only called by logger.shared.FlushTimer
func (logger *Logger) timerFlush() {
	debugPrint("timer flush!")
	logger.Flush()
}

func (logger *Logger) flush() {
	if logger.settings.synchronousFlushWhenDone {
		logger.Flush()
	} else {
		debugPrint("doing async flush")
		go func() {
			smallSleepForTesting()
			logger.Flush()
		}()
	}
}

func (logger *Logger) getFlushers() []xopbase.Request {
	logger.shared.FlusherLock.RLock()
	defer logger.shared.FlusherLock.RUnlock()
	requests := make([]xopbase.Request, 0, len(logger.shared.Flushers))
	for _, req := range logger.shared.Flushers {
		requests = append(requests, req)
	}
	return requests
}

func (logger *Logger) Flush() {
	debugPrint("begin flush {", stack())
	now := time.Now()
	logger.request.detachedDone(now)
	logger.request.recursiveDone(false, now)
	flushers := logger.getFlushers()
	logger.shared.FlushLock.Lock()
	defer logger.shared.FlushLock.Unlock()
	// Stop is is not thread-safe with respect to other calls to Stop
	logger.shared.FlushTimer.Stop()
	atomic.StoreInt32(&logger.shared.FlushActive, 0)
	var wg sync.WaitGroup
	wg.Add(len(flushers))
	for _, flusher := range flushers {
		debugPrint("flushing", flusher.ID())
		go func(flusher xopbase.Request) {
			defer wg.Done()
			flusher.Flush()
		}(flusher)
	}
	wg.Wait()
	debugPrint("done flush }")
}

func final(alloc *singleAllocRequest) {
	for _, flusher := range alloc.Logger.getFlushers() {
		flusher.Final()
	}
}

func (logger *Logger) detachedDone(now time.Time) {
	deps := func() []*Logger {
		logger.request.span.dependentLock.Lock()
		defer logger.request.span.dependentLock.Unlock()
		deps := make([]*Logger, 0, len(logger.shared.ActiveDetached))
		for _, dep := range logger.shared.ActiveDetached {
			deps = append(deps, dep)
		}
		return deps
	}()
	for _, dep := range deps {
		_ = dep.recursiveDone(false, now)
	}
}

// Marks this request as boring.  Any logger at the Alert or
// Error level will mark this request as not boring.
func (logger *Logger) Boring() {
	requestBoring := atomic.LoadInt32(&logger.request.span.boring)
	if requestBoring != 0 {
		return
	}
	logger.request.span.base.Boring(true)
	// There is chance that in the time we were sending that
	// boring=true, the the request became un-boring. If that
	// happened, we can't tell if we're currently marked as
	// boring, so let's make sure we're not boring by sending
	// a false
	requestStillBoring := atomic.LoadInt32(&logger.request.span.boring)
	if requestStillBoring != 0 {
		logger.request.span.base.Boring(false)
	}
	logger.hasActivity(true)
}

func (logger *Logger) notBoring() {
	spanBoring := atomic.AddInt32(&logger.span.boring, 1)
	if spanBoring == 1 {
		logger.span.base.Boring(false)
		requestBoring := atomic.AddInt32(&logger.request.span.boring, 1)
		if requestBoring == 1 {
			logger.request.span.base.Boring(false)
		}
		logger.hasActivity(true)
	}
}

// Line represents a single log event that is in progress.   All
// methods on Line either return Line or don't.  The methods that
// do not return line, like Msg() mark the end of life for that
// Line.  It should not be used in any way after that point.
//
// Nothing checks that Line isn't used after Msg().  Using line
// after Msg() probably won't panic, but will definitely open the
// door to confusing inconsistent logs and race conditions.
type Line struct {
	logger *Logger
	line   xopbase.Line
	pc     []uintptr
	stack  []runtime.Frame
	skip   bool
}

const stackFramesToExclude = 4

// logLine returns *Line, not Line.  Returning Line (and
// changing all the *Line methods to Line methods) is
// faster for some operations but overall it's slower.
func (logger *Logger) logLine(level xopnum.Level) *Line {
	skip := level < logger.settings.minimumLogLevel
	recycled := logger.span.linePool.Get()
	var ll *Line
	if recycled != nil {
		ll = recycled.(*Line)
		if ll.pc != nil {
			ll.pc = ll.pc[:0]
			ll.stack = ll.stack[:0]
		}
	} else {
		ll = &Line{
			logger: logger,
		}
	}
	if !skip && logger.settings.stackFramesWanted[level] != 0 {
		// collect program counters
		if ll.pc == nil || cap(ll.pc) < logger.settings.stackFramesWanted[level] {
			ll.pc = make([]uintptr,
				logger.settings.stackFramesWanted[level],
				logger.settings.stackFramesWanted[xopnum.AlertLevel])
		} else {
			ll.pc = ll.pc[:cap(ll.pc)]
		}
		n := runtime.Callers(stackFramesToExclude, ll.pc)
		ll.pc = ll.pc[:n]
		// collect stack frames
		if ll.stack == nil {
			ll.stack = make([]runtime.Frame, 0, len(ll.pc))
		}
		frames := runtime.CallersFrames(ll.pc)
		for {
			frame, more := frames.Next()
			if strings.Contains(frame.File, "/runtime/") {
				break
			}
			frame.File = logger.settings.stackFilenameRewrite(frame.File)
			if frame.File == "" {
				break
			}
			ll.stack = append(ll.stack, frame)
			if !more {
				break
			}
		}
	}
	ll.skip = skip
	if ll.skip {
		ll.line = xopbase.SkipLine
	} else {
		ll.line = logger.prefilled.Line(level, time.Now(), ll.stack)
	}
	return ll
}

// Template is an alternative to Msg() that sends a log line.  Template
// is a string that uses "{name}" substitutions from the data already
// sent with the line to format that data for human consumption.
//
// Depending on the base logger in use, Template can be cheaper or
// more expensive than Msgf. It's cheaper for base loggers that do not
// expand the message string but more expensive for ones that do, like
// xoptest.
//
// Unlike Msgf(), parameter to the message are preserved as data elements.
// Data elements do not have to be consumed by the template.
//
// The names used for "{name}" substitutions are restricted: they may
// not include any characters that would be escapsed in a JSON string.
// No double quote.  No linefeed.  No backslash.  Etc.
//
// Prefilled text (PrefillText()) will be prepended to the template.
func (line *Line) Template(template string) {
	line.line.Template(template)
	line.logger.span.linePool.Put(line)
	line.logger.hasActivity(true)
}

// Msg sends a log line.  After Msg(), no further use of the *Line
// is allowed.  Without calling Msg(), Template(), Msgf(), Msgs(),
// or Link(), Linkf(), Modelf() or Model(), the log line will not be sent or output.
func (line *Line) Msg(msg string) {
	line.line.Msg(msg)
	line.logger.span.linePool.Put(line)
	line.logger.hasActivity(true)
}

// Msgf sends a log line, using fmt.Sprintf()-style formatting.
func (line *Line) Msgf(msg string, v ...interface{}) {
	if !line.skip {
		line.Msg(fmt.Sprintf(msg, v...))
	}
}

// Model and Any serve similar roles: both can log an arbitrary
// data object.  Model terminates the log line where Any adds a key/value
// attribute to the log line.
//
// Prefer Model() over Any() when the point of the log line is the model.
// Prefer Any() over Model() when the model is just one of several key/value
// attributes attached to the log line.
func (line *Line) Model(obj interface{}, msg string) {
	if line.skip {
		line.Msg("")
		return
	}
	if line.logger.span.referencesKept {
		// TODO: make copy function configurable
		obj = deepcopy.Copy(obj)
	}
	line.ModelImmutable(obj, msg)
}

// ModelImmutable can be used to log something that is not going to be further modified
// after this call.
func (line *Line) ModelImmutable(obj interface{}, msg string) { // TODO: document
	if line.skip {
		line.Msg("")
		return
	}
	line.line.Model(msg, xopbase.ModelArg{
		Model: obj,
	})
	line.logger.span.linePool.Put(line)
	line.logger.hasActivity(true)
}

func (line *Line) Modelf(obj interface{}, msg string, v ...interface{}) { // TODO: document
	if line.skip {
		line.Msg("")
		return
	}
	line.Model(obj, fmt.Sprintf(msg, v...))
}

func (line *Line) ModelImmutablef(obj interface{}, msg string, v ...interface{}) { // TODO: document
	if line.skip {
		line.Msg("")
		return
	}
	line.ModelImmutable(obj, fmt.Sprintf(msg, v...))
}

func (line *Line) Linkf(link xoptrace.Trace, msg string, v ...interface{}) {
	if line.skip {
		line.Msg("")
		return
	}
	line.Link(link, fmt.Sprintf(msg, v...))
}
func (line *Line) Link(link xoptrace.Trace, msg string) {
	line.line.Link(msg, link)
	line.logger.span.linePool.Put(line)
	line.logger.hasActivity(true)
}

// Line starts a log line at the specified log level.  If the log level
// is below the minimum log level, the line will be discarded.
func (logger *Logger) Line(level xopnum.Level) *Line { return logger.logLine(level) }
func (logger *Logger) Debug() *Line                  { return logger.Line(xopnum.DebugLevel) }
func (logger *Logger) Trace() *Line                  { return logger.Line(xopnum.TraceLevel) }
func (logger *Logger) Log() *Line                    { return logger.Line(xopnum.LogLevel) }
func (logger *Logger) Info() *Line                   { return logger.Line(xopnum.InfoLevel) }
func (logger *Logger) Warn() *Line                   { return logger.Line(xopnum.WarnLevel) }
func (logger *Logger) Error() *Line {
	logger.notBoring()
	return logger.Line(xopnum.ErrorLevel)
}
func (logger *Logger) Alert() *Line {
	logger.notBoring()
	return logger.Line(xopnum.AlertLevel)
}

func (line *Line) Msgs(v ...interface{}) { line.Msg(fmt.Sprint(v...)) }
