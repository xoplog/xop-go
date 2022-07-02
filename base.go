package xop

import (
	"github.com/muir/xop/xopbase"
	"github.com/muir/xop/zap"
)

func (s baseLoggers) CopyWithoutTrace() baseLoggers {
	n := make([]baseLogger, len(s.List))
	for i, bl := range s.List {
		n[i] = baseLogger{
			Name: bl.Name,
			Base: bl.Base,
		}
	}
	return baseLoggers{
		List: n,
	}
}

func WithoutBaseLogger(name string) SeedModifier {
	return func(s *Seed) {
		for i, baseLogger := range s.baseLoggers.List {
			if baseLogger.Name == name {
				s.baseLoggers.Removed = append(s.baseLoggers.Removed, baseLogger)
				if i < len(s.baseLoggers.List)-1 {
					s.baseLoggers.List[i], s.baseLoggers.List[len(s.baseLoggers.List)-1] =
						s.baseLoggers.List[len(s.baseLoggers.List)-1], s.baseLoggers.List[i]
				}
				s.baseLoggers.List = s.baseLoggers.List[:len(s.baseLoggers.List)-1]
				break
			}
		}
	}
}

func WithBaseLogger(name string, writer xopbase.BaseLogger) SeedModifier {
	return func(s *Seed) {
		s.baseLoggers.List = append(s.baseLoggers.List, baseLogger{
			Name: name,
			Base: base,
		})
	}
}

func WithAdditionalPrefill(fields ...xopthing.Thing) SeedModifier {
	return func(s *Seed) {
		s.prefillChanged = true
		s.prefill = append(s.prefill, fields...)
	}
}

func WithOnlyPrefill(fields ...xopthing.Thing) SeedModifier {
	return func(s *Seed) {
		s.prefillChanged = true
		s.prefill = fields
	}
}

func (l *Log) finishBaseLoggerChanges() {
	for i, baseLogger := range l.seed.baseLoggers.List {
		if baseLogger.Buffered == nil {
			baseLogger.Buffered = baseLogger.Base.StartBuffer()
		} else if !l.seed.prefillChanged {
			continue
		}
		baseLogger.Prefilled = baseLogger.Buffered.Prefill(l.seed.myTrace, l.seed.prefill)
		l.seed.baseLoggers.List[i] = baseLogger
	}
	for _, baseLogger := range l.seed.baseLoggers.Removed {
		if baseLogger.Buffered != nil {
			baseLogger.Buffered.Flush()
		}
	}
	l.seed.baseLoggers.Removed = nil
	l.seed.prefillChanged = false
}

type baseLoggers struct {
	List    []baseLogger
	Removed []baseLogger
}

type baseLogger struct {
	Base     xopbase.BaseLogger
	MinLevel xopconst.Level
}
