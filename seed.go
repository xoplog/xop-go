// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoplog

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

// Seed is used to create a Log.
type Seed struct {
	config      Config
	traceBundle trace.Bundle
	prefix      string
	prefillMsg  string
	prefillData []func(xopbase.Line)
	description string
	baseLoggers baseLoggers
	flushDelay  time.Duration
}

func (s Seed) Copy() Seed {
	n := s
	n.baseLoggers = s.baseLoggers.Copy()
	n.traceBundle = s.traceBundle.Copy()
	n.prefillMsg = s.prefillMsg
	if s.prefillData != nil {
		n.prefillData = make([]func(xopbase.Line), len(s.prefillData))
		copy(n.prefillData, s.prefillData)
	}
	return n
}

type SeedModifier func(*Seed)

func NewSeed(mods ...SeedModifier) Seed {
	seed := &Seed{
		config: Config{
			FlushDelay: DefaultFlushDelay,
		},
		traceBundle: trace.NewBundle(),
	}
	seed.rebuildAsOne()
	return seed.applyMods(mods)
}

func (s *Span) Seed(mods ...SeedModifier) Seed {
	seed := s.seed.Copy()
	return seed.applyMods(mods)
}

func (s Seed) applyMods(mods []SeedModifier) Seed {
	for _, mod := range mods {
		mod(&s)
	}
	return s
}

func WithBundle(bundle trace.Bundle) SeedModifier {
	return func(s *Seed) {
		s.traceBundle = bundle
	}
}

func WithTrace(trace trace.Trace) SeedModifier {
	return func(s *Seed) {
		s.traceBundle.Trace = trace
	}
}

func WithNoPrefill() SeedModifier {
	return func(s *Seed) {
		s.prefillData = nil
		s.prefillMsg = ""
	}
}

func WithPrefillText(m string) SeedModifier {
	return func(s *Seed) {
		s.prefillMsg = m
	}
}

func (s Seed) Trace() trace.Bundle {
	return s.traceBundle
}

func (s Seed) SubSpan() Seed {
	s.traceBundle = s.traceBundle.Copy()
	s.traceBundle.Trace.RandomizeSpanID()
	return s
}

func (s Seed) sendPrefill(log *Log) {
	if s.prefillData == nil && s.prefillMsg == "" {
		return
	}
	line := log.span.base.Line(xopconst.InfoLevel, time.Now())
	for _, f := range s.prefillData {
		f(line)
	}
	line.SetAsPrefill(s.prefillMsg)
}

// WithAnyPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Any()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithAnyPrefill(k string, v interface{}) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Any(k, v)
		})
	}
}

// WithBoolPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Bool()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithBoolPrefill(k string, v bool) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Bool(k, v)
		})
	}
}

// WithDurationPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Duration()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithDurationPrefill(k string, v time.Duration) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Duration(k, v)
		})
	}
}

// WithErrorPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Error()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithErrorPrefill(k string, v error) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Error(k, v)
		})
	}
}

// WithIntPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Int()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithIntPrefill(k string, v int64) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Int(k, v)
		})
	}
}

// WithLinkPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Link()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithLinkPrefill(k string, v trace.Trace) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Link(k, v)
		})
	}
}

// WithStrPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Str()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithStrPrefill(k string, v string) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Str(k, v)
		})
	}
}

// WithTimePrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Time()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithTimePrefill(k string, v time.Time) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Time(k, v)
		})
	}
}

// WithUintPrefill adds a key/value pair that will be included as part
// of every log line in this span.  If there are no log lines in the
// span then this data will not be logged at all. Use Span.Uint()
// to add a span-level key/value pair that is logged once (with the
// span).
func WithUintPrefill(k string, v uint64) SeedModifier {
	return func(s *Seed) {
		s.prefillData = append(s.prefillData, func(line xopbase.Line) {
			line.Uint(k, v)
		})
	}
}
