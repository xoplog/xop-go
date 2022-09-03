// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopnum"
)

// Sub holds an ephermal state of a log being tranformed to a new log.
type Sub struct {
	detached bool
	settings LogSettings
	log      *Log
}

type Detaching struct {
	sub *Sub
}

type LogSettings struct {
	prefillMsg               string
	prefillData              []func(xopbase.Prefilling)
	minimumLogLevel          xopnum.Level
	stackFramesWanted        [xopnum.AlertLevel + 1]int // indexed
	tagLinesWithSpanSequence bool
	synchronousFlushWhenDone bool
}

// DefaultSettings are the settings that are used if no setting changes
// are made. Debug logs are excluded. Alert and Error level log lines
// get stack traces.
var DefaultSettings = func() LogSettings {
	var settings LogSettings
	settings.stackFramesWanted[xopnum.AlertLevel] = 20
	settings.stackFramesWanted[xopnum.ErrorLevel] = 10
	settings.minimumLogLevel = xopnum.TraceLevel
	settings.synchronousFlushWhenDone = true
	return settings
}()

func (settings LogSettings) Copy() LogSettings {
	if settings.prefillData != nil {
		n := make([]func(xopbase.Prefilling), len(settings.prefillData))
		copy(n, settings.prefillData)
		settings.prefillData = n
	}
	return settings
}

func (log *Log) Settings() LogSettings {
	return log.settings.Copy()
}

// Sub is a first step in creating a sub-Log from the current log.
// Sub allows log settings to be modified.  The returned value must
// be used.  It is used by a call to sub.Log(), sub.Fork(), or
// sub.Step().
//
// Logs created from Sub() are done when their parent is done.
func (log *Log) Sub() *Sub {
	return &Sub{
		settings: log.settings.Copy(),
		log:      log,
	}
}

// Detach followed by Fork() or Step() create a sub-span/log that is detached from
// it's parent.  A Done() on the parent does not imply Done() on the detached
// log.
func (sub Sub) Detach() *Detaching {
	sub.detached = true
	return &Detaching{
		sub: &sub,
	}
}

func (d *Detaching) Step(msg string, mods ...SeedModifier) *Log { return d.sub.Step(msg, mods...) }
func (d *Detaching) Fork(msg string, mods ...SeedModifier) *Log { return d.sub.Fork(msg, mods...) }

// Fork creates a new Log that does not need to be terminated because
// it is assumed to be done with the current log is finished.  The new log
// has its own span.
func (sub *Sub) Fork(msg string, mods ...SeedModifier) *Log {
	seed := sub.log.capSpan.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&sub.log.span.forkCounter, 1))
	seed.spanSequenceCode += "." + base26(counter-1)
	return sub.log.newChildLog(seed.spanSeed, msg, sub.settings, sub.detached)
}

// Step creates a new log that does not need to be terminated -- it
// represents the continued execution of the current log but doing
// something that is different and should be in a fresh span. The expectation
// is that there is a parent log that is creating various sub-logs using
// Step over and over as it does different things.
func (sub *Sub) Step(msg string, mods ...SeedModifier) *Log {
	seed := sub.log.capSpan.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&sub.log.span.stepCounter, 1))
	seed.spanSequenceCode += "." + strconv.Itoa(counter)
	return sub.log.newChildLog(seed.spanSeed, msg, sub.settings, sub.detached)
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (sub *Sub) StackFrames(level xopnum.Level, count int) *Sub {
	sub.settings.StackFrames(level, count)
	return sub
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (settings *LogSettings) StackFrames(level xopnum.Level, frameCount int) {
	for _, l := range xopnum.LevelValues() {
		current := settings.stackFramesWanted[l]
		if l <= level && current > frameCount {
			settings.stackFramesWanted[l] = frameCount
		}
		if l >= level && current < frameCount {
			settings.stackFramesWanted[l] = frameCount
		}
	}
}

// SynchronousFlush sets the behavior for any Flush()
// triggered by a call to Done().  When true, the
// call to Done() will not return until the Flush() is
// complete.
func (sub *Sub) SynchronousFlush(b bool) *Sub {
	sub.settings.SynchronousFlush(b)
	return sub
}

// SynchronousFlush sets the behavior for any Flush()
// triggered by a call to Done().  When true, the
// call to Done() will not return until the Flush() is
// complete.
func (settings *LogSettings) SynchronousFlush(b bool) {
	settings.synchronousFlushWhenDone = b
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (sub *Sub) MinLevel(level xopnum.Level) *Sub {
	sub.settings.MinLevel(level)
	return sub
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (settings *LogSettings) MinLevel(level xopnum.Level) {
	settings.minimumLogLevel = level
}

// TagLinesWithSpanSequence controls if the span sequence
// indicator (see Fork() and Step()) should be included in
// the prefill data on each line.
func (sub *Sub) TagLinesWithSpanSequence(b bool) *Sub {
	sub.settings.TagLinesWithSpanSequence(b)
	return sub
}

// TagLinesWithSpanSequence controls if the span sequence
// indicator (see Fork() and Step()) should be included in
// the prefill data on each line.
func (settings *LogSettings) TagLinesWithSpanSequence(b bool) {
	settings.tagLinesWithSpanSequence = b
}

func (sub *Sub) PrefillText(m string) *Sub {
	sub.settings.PrefillText(m)
	return sub
}

func (settings *LogSettings) PrefillText(m string) {
	settings.prefillMsg = m
}

func (sub *Sub) NoPrefill() *Sub {
	sub.settings.NoPrefill()
	return sub
}

func (settings *LogSettings) NoPrefill() {
	settings.prefillData = nil
	settings.prefillMsg = ""
}

func (log *Log) sendPrefill() {
	if log.settings.prefillData == nil && log.settings.prefillMsg == "" && !log.settings.tagLinesWithSpanSequence {
		log.prefilled = log.span.base.NoPrefill()
		return
	}
	prefilling := log.span.base.StartPrefill()
	for _, f := range log.settings.prefillData {
		f(prefilling)
	}
	if log.settings.tagLinesWithSpanSequence {
		prefilling.String(xopconst.SpanSequenceCode.Key(), log.span.seed.spanSequenceCode)
	}
	log.prefilled = prefilling.PrefillComplete(log.settings.prefillMsg)
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillAny(k string, v interface{}) *Sub {
	sub.settings.PrefillAny(k, v)
	return sub
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (settings *LogSettings) PrefillAny(k string, v interface{}) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Any(k, v)
	})
}

// PrefillBool is used to set a data element that is included on every log
// line.
// PrefillBool is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillBool(k string, v bool) *Sub {
	sub.settings.PrefillBool(k, v)
	return sub
}

func (settings *LogSettings) PrefillBool(k string, v bool) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Bool(k, v)
	})
}

// PrefillDuration is used to set a data element that is included on every log
// line.
// PrefillDuration is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillDuration(k string, v time.Duration) *Sub {
	sub.settings.PrefillDuration(k, v)
	return sub
}

func (settings *LogSettings) PrefillDuration(k string, v time.Duration) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Duration(k, v)
	})
}

// PrefillError is used to set a data element that is included on every log
// line.
// PrefillError is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillError(k string, v error) *Sub {
	sub.settings.PrefillError(k, v)
	return sub
}

func (settings *LogSettings) PrefillError(k string, v error) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Error(k, v)
	})
}

// PrefillFloat64 is used to set a data element that is included on every log
// line.
// PrefillFloat64 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillFloat64(k string, v float64) *Sub {
	sub.settings.PrefillFloat64(k, v)
	return sub
}

func (settings *LogSettings) PrefillFloat64(k string, v float64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Float64(k, v)
	})
}

// PrefillInt is used to set a data element that is included on every log
// line.
// PrefillInt is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillInt(k string, v int64) *Sub {
	sub.settings.PrefillInt(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt(k string, v int64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int(k, v)
	})
}

// PrefillLink is used to set a data element that is included on every log
// line.
// PrefillLink is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillLink(k string, v trace.Trace) *Sub {
	sub.settings.PrefillLink(k, v)
	return sub
}

func (settings *LogSettings) PrefillLink(k string, v trace.Trace) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Link(k, v)
	})
}

// PrefillString is used to set a data element that is included on every log
// line.
// PrefillString is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillString(k string, v string) *Sub {
	sub.settings.PrefillString(k, v)
	return sub
}

func (settings *LogSettings) PrefillString(k string, v string) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.String(k, v)
	})
}

// PrefillTime is used to set a data element that is included on every log
// line.
// PrefillTime is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillTime(k string, v time.Time) *Sub {
	sub.settings.PrefillTime(k, v)
	return sub
}

func (settings *LogSettings) PrefillTime(k string, v time.Time) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Time(k, v)
	})
}

// PrefillUint is used to set a data element that is included on every log
// line.
// PrefillUint is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (sub *Sub) PrefillUint(k string, v uint64) *Sub {
	sub.settings.PrefillUint(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint(k string, v uint64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint(k, v)
	})
}
