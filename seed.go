package xm

import (
	"time"

	"github.com/muir/xm/trace"
	"github.com/muir/xm/zap"
)

// Seed is used to create a Log
type Seed struct {
	config Config
	traceState
	prefix         string
	prefill        []zap.Field
	prefillChanged bool
	description    string
	data           map[string]interface{}
	baseLoggers    baseLoggers
	flushDelay     time.Duration
}

func (s Seed) Copy() Seed {
	n := s
	n.prefill = copyFields(s.prefill)
	n.baseLoggers = s.baseLoggers.CopyWithoutTrace()
	n.data = nil
	n.traceState = s.traceState.Copy()
	return n
}

func copyFields(from []zap.Field) []zap.Field {
	n := make([]zap.Field, len(from))
	copy(n, from)
	return n
}

type SeedModifier func(*Seed)

func NewSeed(mods ...SeedModifier) Seed {
	seed := &Seed{
		data: make(map[string]interface{}),
	}
	return seed.ApplyMods(mods)
}

func (l *Log) CopySeed(mods ...SeedModifier) Seed {
	seed := l.seed.Copy()
	return seed.ApplyMods(mods)
}

func (s Seed) ApplyMods(mods []SeedModifier) Seed {
	for _, mod := range mods {
		mod(&s)
	}
	return s
}

func PrefilOnly(fields []zap.Field) SeedModifier {
	return func(s *Seed) {
		s.prefill = fields
	}
}

func Data(overrides map[string]interface{}) SeedModifier {
	return func(s *Seed) {
		for k, v := range overrides {
			s.data[k] = v
		}
	}
}

func (s Seed) State() *trace.State { return &s.traceState.state }

func (s Seed) Trace() *trace.Trace {
	return &s.myTrace
}

func (s Seed) TraceParent() *trace.Trace {
	return &s.parentTrace
}

func (s Seed) Baggage() *trace.Baggage { return &s.traceState.baggage }

func (s Seed) SubSpan() Seed {
	s.parentTrace = s.myTrace.Copy()
	s.myTrace.RandomizeSpanId()
	return s
}

type traceState struct {
	parentTrace trace.Trace
	myTrace     trace.Trace
	state       trace.State
	baggage     trace.Baggage
}

func (t traceState) Copy() traceState {
	return traceState{
		parentTrace: t.parentTrace.Copy(),
		myTrace:     t.myTrace.Copy(),
		state:       t.state,
		baggage:     t.baggage,
	}
}
