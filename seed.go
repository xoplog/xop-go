// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"context"
	"time"

	"github.com/xoplog/xop-go/trace"
)

// Seed is used to create a Log.
type Seed struct {
	spanSeed
	settings LogSettings
}

// Copy makes a deep copy of a seed and also randomizes
// the SpanID.
func (seed Seed) Copy(mods ...SeedModifier) Seed {
	n := Seed{
		spanSeed: seed.spanSeed.copy(true),
		settings: seed.settings.Copy(),
	}
	if !seed.spanSet {
		n.traceBundle.Trace.SpanID().SetRandom()
	}
	n = n.applyMods(mods)
	return n
}

// Seed provides a copy of the current span's seed, but the
// spanID is randomized.
func (span *Span) Seed(mods ...SeedModifier) Seed {
	n := Seed{
		spanSeed: span.seed.copy(false),
		settings: span.log.settings.Copy(),
	}
	if !span.seed.spanSet {
		n.traceBundle.Trace.SpanID().SetRandom()
	}
	n = n.applyMods(mods)
	return n
}

// SeedReactiveCallback is used to modify seeds as they are just sprouting.
type SeedReactiveCallback func(ctx context.Context, seed Seed, nameOrDescription string, isChildSpan bool) []SeedModifier

type seedReactiveCallbacks []SeedReactiveCallback

func (cbs seedReactiveCallbacks) Copy() seedReactiveCallbacks {
	n := make(seedReactiveCallbacks, len(cbs))
	copy(n, cbs)
	return n
}

type spanSeed struct {
	traceBundle          trace.Bundle
	spanSequenceCode     string
	description          string
	loggers              loggers
	config               Config
	flushDelay           time.Duration
	reactive             seedReactiveCallbacks
	ctx                  context.Context
	currentReactiveIndex int
	traceSet             bool
	spanSet              bool
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

func (seed Seed) applyMods(mods []SeedModifier) Seed {
	for _, mod := range mods {
		mod(&seed)
	}
	return seed
}

// Request creates a new top-level span (a request).  Use when
// starting something new, like receiving an http request or
// starting a cron job.
func (seed Seed) Request(descriptionOrName string) *Log {
	seed = seed.react(true, descriptionOrName)
	return seed.request(descriptionOrName)
}

// SubSpan creates a new top-level span (a request) but
// initializes the span/trace data as if it were a subspan.
// The traceID must already be set.  Use this with caution,
// it is meant for handing off from spans created elsewhere.
func (seed Seed) SubSpan(descriptionOrName string) *Log {
	seed = seed.react(false, descriptionOrName)
	return seed.request(descriptionOrName)
}

// WithReactive provides a callback that is used to modify seeds as they
// are in the process of sprouting.  Just as a seed is being used to create
// a request of sub-span, all reactive functions will be called.  Such
// functions must return a valid seed.  The seed they start with will be
// valid, so they can simply return that seed.
//
// When WithReactive is called from a SeedReactiveCallback, the new reactive
// function is only evaluated on descendent seeds.
func WithReactive(f SeedReactiveCallback) SeedModifier {
	return func(s *Seed) {
		s.reactive = append(s.reactive, f)
	}
}

// WithReactiveReplaced may only be used from within a call to a reactive
// function.  The current reactive function is the one that will be replaced.
// To remove a reactive function, call WithReactiveReplaced with nil.
func WithReactiveReplaced(f SeedReactiveCallback) SeedModifier {
	return func(s *Seed) {
		s.reactive[s.currentReactiveIndex] = f
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

func WithSpan(spanID [8]byte) SeedModifier {
	return func(s *Seed) {
		s.traceBundle.Trace.SpanID().Set(spanID)
		s.spanSet = true
	}
}

func WithTrace(trace trace.Trace) SeedModifier {
	return func(s *Seed) {
		s.traceBundle.Trace = trace
		s.traceSet = true
		s.spanSet = true
	}
}

func WithSettings(f func(*LogSettings)) SeedModifier {
	return func(s *Seed) {
		f(&s.settings)
	}
}

func CombineSeedModifiers(mods ...SeedModifier) SeedModifier {
	return func(s *Seed) {
		for _, f := range mods {
			f(s)
		}
	}
}

func (seed Seed) Bundle() trace.Bundle {
	return seed.traceBundle
}

func (seed Seed) react(isRequest bool, description string) Seed {
	if isRequest {
		if !seed.traceSet {
			seed.traceBundle.Trace.RebuildSetNonZero()
		}
	}
	seed.traceSet = false
	seed.spanSet = false
	if len(seed.reactive) == 0 {
		return seed
	}
	var nilSeen bool
	initialCount := len(seed.reactive)
	for i := 0; i < initialCount; i++ {
		f := seed.reactive[i]
		if f == nil {
			nilSeen = true
			i++
			continue
		}
		seed.currentReactiveIndex = i
		seed = seed.applyMods(f(seed.ctx, seed, description, !isRequest))
		if seed.reactive[i] == nil {
			nilSeen = true
		}
	}
	if nilSeen {
		reactive := make(seedReactiveCallbacks, 0, len(seed.reactive))
		for _, f := range seed.reactive {
			if f != nil {
				reactive = append(reactive, f)
			}
		}
		seed.reactive = reactive
	}
	return seed
}
