// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"context"
	"time"

	"github.com/muir/xop-go/trace"
)

// Seed is used to create a Log.
type Seed struct {
	spanSeed
	settings LogSettings
}

func (seed Seed) Copy(mods ...SeedModifier) Seed {
	return Seed{
		spanSeed: seed.spanSeed.copy(true),
		settings: seed.settings.Copy(),
	}
}

// SeedReactiveCallback is used to modify seeds as they are just sprouting
// The selfIndex parameter can be used with WithReactiveReplaced or
// WithReactiveRemoved.
type SeedReactiveCallback func(ctx context.Context, seed Seed, selfIndex int, nameOrDescription string, isChildSpan bool) Seed

type seedReactiveCallbacks []SeedReactiveCallback

func (cbs seedReactiveCallbacks) Copy() seedReactiveCallbacks {
	n := make(seedReactiveCallbacks, len(cbs))
	copy(n, cbs)
	return n
}

type spanSeed struct {
	traceBundle      trace.Bundle
	spanSequenceCode string
	description      string
	loggers          loggers
	config           Config
	flushDelay       time.Duration
	reactive         seedReactiveCallbacks
	ctx              context.Context
}

func (s spanSeed) copy(withHistory bool) spanSeed {
	n := s
	n.loggers = s.loggers.Copy(withHistory)
	n.traceBundle = s.traceBundle.Copy()
	n.reactive = s.reactive.Copy()
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
		spanSeed: span.seed.copy(false),
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

// WithReactive provides a callback that is used to modify seeds as they
// are in the process of sprouting.  Just as a seed is being used to create
// a request of sub-span, all reactive functions will be called.  Such
// functions must return a valid seed.  The seed they start with will be
// valid, so they can simply return that seed.
func WithReactive(f SeedReactiveCallback) SeedModifier {
	return func(s *Seed) {
		s.reactive = append(s.reactive, f)
	}
}

func WithReactiveReplaced(index int, f SeedReactiveCallback) SeedModifier {
	return func(s *Seed) {
		s.reactive[index] = f
	}
}

func WithReactiveRemoved(index int) SeedModifier {
	return func(s *Seed) {
		s.reactive = append(s.reactive[:index], s.reactive[index+1:]...)
	}
}

// WithContext puts a context into the seed.  That context will be
// passed through to the base layer Request and Seed functions.
func WithContext(ctx context.Context) SeedModifier {
	return func(s *Seed) {
		s.ctx = ctx
	}
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
