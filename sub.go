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

type Sub struct {
	settings LogSettings
	log      *Log
}

type LogSettings struct {
	prefillMsg      string
	prefillData     []func(xopbase.Line)
	minimumLogLevel Level
}

func (l *Log) Sub() *Sub {
	return &Settings{
		settings: l.settings,
		log:      l,
	}
}

func (s *Sub) Log() *Log {
}

// Fork creates a new Log that does not need to be terminated because
// it is assumed to be done with the current log is finished.
func (s *Sub) Fork(msg string, mods ...SeedModifier) *Log {
	seed := s.log.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&s.log.local.ForkCounter, 1))
	seed.spanSequenceCode += "." + base26(counter)
	return l.newChildLog(seed, s.settings, msg)
}

// Step creates a new log that does not need to be terminated -- it
// represents the continued execution of the current log but doing
// something that is different and should be in a fresh span. The expectation
// is that there is a parent log that is creating various sub-logs using
// Step over and over as it does different things.
func (s *Sub) Step(msg string, mods ...SeedModifier) *Log {
	seed := s.log.span.Seed(mods...).SubSpan()
	counter := int(atomic.AddInt32(&s.log.local.StepCounter, 1))
	seed.spanSequenceCode += "." + strconv.Itoa(counter)
	return s.log.newChildLog(seed, s.settings, msg)
}

func (s *Sub) Level(level xopconst.Level) *Sub {
	s.settings.Level(level)
	return s
}

func (s *LogSettings) Level(level xopconst.Level) {
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

func (s Seed) sendPrefill(log *Log) {
	if s.prefillData == nil && s.prefillMsg == "" {
		return
	}
	line := log.span.base.Line(xopconst.InfoLevel, time.Now(), nil)
	for _, f := range s.prefillData {
		f(line)
	}
	line.SetAsPrefill(s.prefillMsg)
}

// AnyPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) AnyPrefill(k string, v interface{}) *Sub {
	s.settings.AnyPrefill(k, v)
	return s
}

func (s *LogSettings) AnyPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Any(k, v)
	})
}

// BoolPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) BoolPrefill(k string, v bool) *Sub {
	s.settings.BoolPrefill(k, v)
	return s
}

func (s *LogSettings) BoolPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Bool(k, v)
	})
}

// DurationPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) DurationPrefill(k string, v time.Duration) *Sub {
	s.settings.DurationPrefill(k, v)
	return s
}

func (s *LogSettings) DurationPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Duration(k, v)
	})
}

// ErrorPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) ErrorPrefill(k string, v error) *Sub {
	s.settings.ErrorPrefill(k, v)
	return s
}

func (s *LogSettings) ErrorPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Error(k, v)
	})
}

// IntPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) IntPrefill(k string, v int64) *Sub {
	s.settings.IntPrefill(k, v)
	return s
}

func (s *LogSettings) IntPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Int(k, v)
	})
}

// LinkPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) LinkPrefill(k string, v trace.Trace) *Sub {
	s.settings.LinkPrefill(k, v)
	return s
}

func (s *LogSettings) LinkPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Link(k, v)
	})
}

// StrPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) StrPrefill(k string, v string) *Sub {
	s.settings.StrPrefill(k, v)
	return s
}

func (s *LogSettings) StrPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Str(k, v)
	})
}

// TimePrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) TimePrefill(k string, v time.Time) *Sub {
	s.settings.TimePrefill(k, v)
	return s
}

func (s *LogSettings) TimePrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Time(k, v)
	})
}

// UintPrefill is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Log() is called.
func (s *Sub) UintPrefill(k string, v uint64) *Sub {
	s.settings.UintPrefill(k, v)
	return s
}

func (s *LogSettings) UintPrefill() {
	s.prefillData = append(s.prefillData, func(line xopbase.Line) {
		line.Uint(k, v)
	})
}
