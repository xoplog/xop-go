package xoplog

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
)

// Seed is used to create a Log.
type Seed struct {
	config         Config
	traceBundle    trace.Bundle
	prefix         string
	prefill        []xop.Thing
	prefillChanged bool
	description    string
	data           []xop.Thing
	baseLoggers    baseLoggers
	flushDelay     time.Duration
}

func (s Seed) Copy() Seed {
	n := s
	n.prefill = copyFields(s.prefill)
	n.baseLoggers = s.baseLoggers.copyWithoutTrace()
	n.data = nil
	n.traceBundle = s.traceBundle.Copy()
	return n
}

func copyFields(from []xop.Thing) []xop.Thing {
	n := make([]xop.Thing, len(from))
	copy(n, from)
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

func WithMorePrefill(fields ...xop.Thing) SeedModifier {
	return func(s *Seed) {
		s.prefillChanged = true
		s.prefill = append(s.prefill, fields...)
	}
}

func WithPrefill(fields ...xop.Thing) SeedModifier {
	return func(s *Seed) {
		s.prefill = fields
		s.prefillChanged = true
	}
}

func WithData(fields ...xop.Thing) SeedModifier {
	return func(s *Seed) {
		s.data = fields
	}
}

func (s Seed) Trace() trace.Bundle {
	return s.traceBundle
}

func (s Seed) SubSpan() Seed {
	s.traceBundle = s.traceBundle.Copy()
	s.traceBundle.Trace.RandomizeSpanId()
	return s
}
