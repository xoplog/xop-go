package xm

import (
	"time"
)

type Seed struct {
	config Config
	traceState
	prefix         string
	prefill        []Field
	prefillChanged bool
	description    string
	data           []Field
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

type SeedModifier func(*Seed)

func NewSeed(mods ...SeedModifier) Seed {
	seed := &Seed{}
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

func PrefilOnly(fields []Field) SeedModifier {
	return func(s *Seed) {
		s.prefill = fields
	}
}

func Data(f ...Field) SeedModifier {
	return func(s *Seed) {
		s.data = append(s.data, f...)
	}
}
