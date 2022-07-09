package xoplog

import (
	"github.com/muir/xoplog/internal/multibase"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type baseLoggers struct {
	AsOne   xopbase.Logger
	List    []baseLogger
	Removed []baseLogger
}

type baseLogger struct {
	Name     string
	Base     xopbase.Logger
	MinLevel xopconst.Level
}

func (s baseLoggers) Copy() baseLoggers {
	n := make([]baseLogger, len(s.List))
	for i, bl := range s.List {
		n[i] = baseLogger{
			Name: bl.Name,
			Base: bl.Base,
		}
	}
	return baseLoggers{
		AsOne: s.AsOne,
		List:  n,
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
				s.rebuildAsOne()
			}
		}
	}
}

func (s *Seed) rebuildAsOne() {
	loggers := make([]xopbase.Logger, len(s.baseLoggers.List))
	for i, baseLogger := range s.baseLoggers.List {
		loggers[i] = baseLogger.Base
	}
	s.baseLoggers.AsOne = multibase.CombineLoggers(loggers)
}

func WithBaseLogger(name string, logger xopbase.Logger) SeedModifier {
	return func(s *Seed) {
		s.baseLoggers.List = append(s.baseLoggers.List, baseLogger{
			Name: name,
			Base: logger,
		})
		s.rebuildAsOne()
	}
}
