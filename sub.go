// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
)

// Sub holds an ephermal state of a log being tranformed to a new log.
type Sub struct {
	settings LogSettings
	log      *Log
}

type LogSettings struct {
	prefillMsg        string
	prefillData       []func(xopbase.Prefilling)
	minimumLogLevel   xopconst.Level               // XXX implement check
	stackFramesWanted [xopconst.AlertLevel + 1]int // indexed
}

// DefaultSettings are the settings that are used if no setting changes
// are made. Debug logs are excluded. Alert and Error level log lines
// get stack traces.
var DefaultSettings = func() LogSettings {
	var settings LogSettings
	settings.stackFramesWanted[xopconst.AlertLevel] = 20
	settings.stackFramesWanted[xopconst.ErrorLevel] = 10
	settings.minimumLogLevel = xopconst.TraceLevel
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

// Sub is the first step in creating a sub-Log from the current log.
// Sub allows log settings to be modified.  The returned value must
// be used.  It is used by a call to sub.Log(), sub.Fork(), or
// sub.Step().
func (l *Log) Sub() *Sub {
	return &Sub{
		settings: l.settings.Copy(),
		log:      l,
	}
}

// Fork creates a new Log that does not need to be terminated because
// it is assumed to be done with the current log is finished.  The new log
// has its own span.
func (s *Sub) Fork(msg string, mods ...SeedModifier) *Log {
	seed := s.log.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&s.log.span.forkCounter, 1))
	seed.spanSequenceCode += "." + base26(counter)
	return s.log.newChildLog(seed.spanSeed, msg, s.settings)
}

// Step creates a new log that does not need to be terminated -- it
// represents the continued execution of the current log but doing
// something that is different and should be in a fresh span. The expectation
// is that there is a parent log that is creating various sub-logs using
// Step over and over as it does different things.
func (s *Sub) Step(msg string, mods ...SeedModifier) *Log {
	seed := s.log.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&s.log.span.stepCounter, 1))
	seed.spanSequenceCode += "." + strconv.Itoa(counter)
	return s.log.newChildLog(seed.spanSeed, msg, s.settings)
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (s *Sub) StackFrames(level xopconst.Level, count int) *Sub {
	s.settings.StackFrames(level, count)
	return s
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (s *LogSettings) StackFrames(level xopconst.Level, frameCount int) {
	for _, l := range xopconst.LevelValues() {
		current := s.stackFramesWanted[l]
		if l <= level && current > frameCount {
			s.stackFramesWanted[l] = frameCount
		}
		if l >= level && current < frameCount {
			s.stackFramesWanted[l] = frameCount
		}
	}
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (s *Sub) MinLevel(level xopconst.Level) *Sub {
	s.settings.MinLevel(level)
	return s
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (s *LogSettings) MinLevel(level xopconst.Level) {
	s.minimumLogLevel = level
}

func (s *Sub) PrefillText(m string) *Sub {
	s.settings.PrefillText(m)
	return s
}

func (s *LogSettings) PrefillText(m string) {
	s.prefillMsg = m
}

func (s *Sub) NoPrefill() *Sub {
	s.settings.NoPrefill()
	return s
}

func (s *LogSettings) NoPrefill() {
	s.prefillData = nil
	s.prefillMsg = ""
}

func (l *Log) sendPrefill() {
	if l.settings.prefillData == nil && l.settings.prefillMsg == "" {
		l.prefilled = l.span.base.NoPrefill()
	}
	prefilling := l.span.base.StartPrefill()
	for _, f := range l.settings.prefillData {
		f(prefilling)
	}
	l.prefilled = prefilling.PrefillComplete(l.settings.prefillMsg)
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillAny(k string, v interface{}) *Sub {
	s.settings.PrefillAny(k, v)
	return s
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *LogSettings) PrefillAny(k string, v interface{}) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Any(k, v)
	})
}

// PrefillBool is used to set a data element that is included on every log
// line.
// PrefillBool is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillBool(k string, v bool) *Sub {
	s.settings.PrefillBool(k, v)
	return s
}

func (s *LogSettings) PrefillBool(k string, v bool) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Bool(k, v)
	})
}

// PrefillDuration is used to set a data element that is included on every log
// line.
// PrefillDuration is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillDuration(k string, v time.Duration) *Sub {
	s.settings.PrefillDuration(k, v)
	return s
}

func (s *LogSettings) PrefillDuration(k string, v time.Duration) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Duration(k, v)
	})
}

// PrefillError is used to set a data element that is included on every log
// line.
// PrefillError is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillError(k string, v error) *Sub {
	s.settings.PrefillError(k, v)
	return s
}

func (s *LogSettings) PrefillError(k string, v error) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Error(k, v)
	})
}

// PrefillFloat64 is used to set a data element that is included on every log
// line.
// PrefillFloat64 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillFloat64(k string, v float64) *Sub {
	s.settings.PrefillFloat64(k, v)
	return s
}

func (s *LogSettings) PrefillFloat64(k string, v float64) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Float64(k, v)
	})
}

// PrefillInt is used to set a data element that is included on every log
// line.
// PrefillInt is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillInt(k string, v int64) *Sub {
	s.settings.PrefillInt(k, v)
	return s
}

func (s *LogSettings) PrefillInt(k string, v int64) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Int(k, v)
	})
}

// PrefillLink is used to set a data element that is included on every log
// line.
// PrefillLink is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillLink(k string, v trace.Trace) *Sub {
	s.settings.PrefillLink(k, v)
	return s
}

func (s *LogSettings) PrefillLink(k string, v trace.Trace) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Link(k, v)
	})
}

// PrefillStr is used to set a data element that is included on every log
// line.
// PrefillStr is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillStr(k string, v string) *Sub {
	s.settings.PrefillStr(k, v)
	return s
}

func (s *LogSettings) PrefillStr(k string, v string) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Str(k, v)
	})
}

// PrefillTime is used to set a data element that is included on every log
// line.
// PrefillTime is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillTime(k string, v time.Time) *Sub {
	s.settings.PrefillTime(k, v)
	return s
}

func (s *LogSettings) PrefillTime(k string, v time.Time) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Time(k, v)
	})
}

// PrefillUint is used to set a data element that is included on every log
// line.
// PrefillUint is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillUint(k string, v uint64) *Sub {
	s.settings.PrefillUint(k, v)
	return s
}

func (s *LogSettings) PrefillUint(k string, v uint64) {
	s.prefillData = append(s.prefillData, func(line xopbase.Prefilling) {
		line.Uint(k, v)
	})
}
