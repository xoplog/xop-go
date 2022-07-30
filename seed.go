// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
)

// Seed is used to create a Log.
type Seed struct {
	spanSeed
	settings LogSettings
}

type spanSeed struct {
	traceBundle      trace.Bundle
	spanSequenceCode string
	description      string
	loggers          loggers
	config           Config
	flushDelay       time.Duration
}

func (s Seed) Copy() Seed {
	n := s
	n.loggers = s.loggers.Copy()
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
		spanSeed: spanSeed{
			config:      DefaultConfig,
			traceBundle: trace.NewBundle(),
		},
	}
	seed.rebuildAsOne()
	return seed.applyMods(mods)
}

func (s *Span) Seed(mods ...SeedModifier) Seed {
	seed := Seed{
		spanSeed: s.spanSeed.Copy(),
		settings: s.log.settings,
	}
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

func WithAdjustments(f func(*LogSettings)) SeedModifier {
	return func(s *Seed) {
		f(&seed.LogSettings)
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
