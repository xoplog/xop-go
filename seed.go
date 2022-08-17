// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"time"

	"github.com/muir/xop-go/trace"
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

func (s spanSeed) Copy() spanSeed {
	n := s
	n.loggers = s.loggers.Copy()
	n.traceBundle = s.traceBundle.Copy()
	return n
}

type SeedModifier func(*Seed)

func NewSeed(mods ...SeedModifier) Seed {
	seed := &Seed{
		spanSeed: spanSeed{
			config:      DefaultConfig,
			traceBundle: trace.NewBundle(),
		},
		settings: DefaultSettings,
	}
	return seed.applyMods(mods)
}

// Seed provides a copy of the current span's seed, but the
// spanID is randomized.
func (span *Span) Seed(mods ...SeedModifier) Seed {
	seed := Seed{
		spanSeed: span.seed.Copy(),
		settings: span.log.settings.Copy(),
	}
	seed.spanSeed.traceBundle.Trace.RandomizeSpanID()
	return seed.applyMods(mods)
}

func (seed Seed) applyMods(mods []SeedModifier) Seed {
	for _, mod := range mods {
		mod(&seed)
	}
	return seed
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

func WithSettings(f func(*LogSettings)) SeedModifier {
	return func(s *Seed) {
		f(&s.settings)
	}
}

func (seed Seed) Trace() trace.Bundle {
	return seed.traceBundle
}

func (seed Seed) SubSpan() Seed {
	seed.traceBundle = seed.traceBundle.Copy()
	seed.traceBundle.Trace.RandomizeSpanID()
	return seed
}
