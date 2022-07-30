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
	prefillData       []func(xopbase.Line)
	minimumLogLevel   xopconst.Level
	stackFramesWanted [xopconst.AlertLevel + 1]int // indexed
}

// Sub is the first step in creating a sub-Log from the current log.
// Sub allows log settings to be modified.  The returned value must
// be used.  It is used by a call to sub.Log(), sub.Fork(), or
// sub.Step().
func (l *Log) Sub() *Sub {
	return &Settings{
		settings: l.settings,
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
	return l.newChildLog(seed.spanSeed, msg, s.settings)
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
func (s *LogSettings) StackFrames(level xopconst.Level, count int) {
	for _, l := range xopconst.AllLevels() {
		current := s.stackFramesWanted[l]
		if l <= level && current > level {
			s.stackFramesWanted = count
		}
		if l >= level && current < level {
			s.stackFramesWanted = count
		}
	}
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (s *Sub) MinLevel(level xopconst.Level) *Sub {
	s.settings.Level(level)
	return s
}

// MinLevel sets the minimum logging level below which logs will
// be discarded.  The default is that no logs are discarded.
func (s *LogSettings) MinLevel(level xopconst.Level) {
	s.minimumLoggingLevel = level
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
	s.settings.prefillData = nil
	s.settings.prefillMsg = ""
}

func (l *Log) sendPrefill() xopbase.Prefilled {
	if s.settings.prefillData == nil && s.settings.prefillMsg == "" {
		l.prefilled = log.span.base.NoPrefill()
	}
	line := log.span.base.StartPrefill()
	for _, f := range l.settings.prefillData {
		f(line)
	}
	l.prefilled = line.PrefillComplete()
}

// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillAny(k string, v interface{}) *Sub {
	s.settings.PrefillAny(k, v)
	return s
}

func (s *LogSettings) AnyPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Any(k, v)
	})
}

// PrefillBool is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillBool(k string, v bool) *Sub {
	s.settings.PrefillBool(k, v)
	return s
}

func (s *LogSettings) BoolPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Bool(k, v)
	})
}

// PrefillDuration is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillDuration(k string, v time.Duration) *Sub {
	s.settings.PrefillDuration(k, v)
	return s
}

func (s *LogSettings) DurationPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Duration(k, v)
	})
}

// PrefillError is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillError(k string, v error) *Sub {
	s.settings.PrefillError(k, v)
	return s
}

func (s *LogSettings) ErrorPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Error(k, v)
	})
}

// PrefillInt is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillInt(k string, v int64) *Sub {
	s.settings.PrefillInt(k, v)
	return s
}

func (s *LogSettings) IntPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Int(k, v)
	})
}

// PrefillLink is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillLink(k string, v trace.Trace) *Sub {
	s.settings.PrefillLink(k, v)
	return s
}

func (s *LogSettings) LinkPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Link(k, v)
	})
}

// PrefillStr is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillStr(k string, v string) *Sub {
	s.settings.PrefillStr(k, v)
	return s
}

func (s *LogSettings) StrPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Str(k, v)
	})
}

// PrefillTime is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillTime(k string, v time.Time) *Sub {
	s.settings.PrefillTime(k, v)
	return s
}

func (s *LogSettings) TimePrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Time(k, v)
	})
}

// PrefillUint is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) PrefillUint(k string, v uint64) *Sub {
	s.settings.PrefillUint(k, v)
	return s
}

func (s *LogSettings) UintPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Uint(k, v)
	})
}
