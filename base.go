package xoplog

import (
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type loggers struct {
	AsOne    baseLoggers
	Flushers baseRequests
	List     []baseLogger
	Removed  []baseLogger
	Added    []baseLogger
}

type baseLogger struct {
	Name     string // XXX don't need anymore
	Base     xopbase.Logger
	MinLevel xopconst.Level // XXX is this used anywhere?  Put where useful
}

func (s loggers) Copy() loggers {
	n := make([]baseLogger, len(s.List))
	for i, bl := range s.List {
		n[i] = baseLogger{
			Name: bl.Name,
			Base: bl.Base,
		}
	}
	return loggers{
		AsOne:    s.AsOne,
		Flushers: s.Flushers,
		List:     n,
		// Added & Removed are not included
	}
}

func WithoutBaseLogger(name string) SeedModifier {
	return func(s *Seed) {
		for i, baseLogger := range s.loggers.List {
			if baseLogger.Name == name {
				s.loggers.Removed = append(s.loggers.Removed, baseLogger)
				if i < len(s.loggers.List)-1 {
					s.loggers.List[i], s.loggers.List[len(s.loggers.List)-1] =
						s.loggers.List[len(s.loggers.List)-1], s.loggers.List[i]
				}
				s.loggers.List = s.loggers.List[:len(s.loggers.List)-1]
				s.rebuildAsOne()
			}
		}
	}
}

func (s *Seed) rebuildAsOne() {
	s.loggers.AsOne = make(baseLoggers, len(s.loggers.List))
	for i, baseLogger := range s.loggers.List {
		s.loggers.AsOne[i] = baseLogger.Base
	}
}

func WithBaseLogger(name string, logger xopbase.Logger) SeedModifier {
	return func(s *Seed) {
		baseLogger := baseLogger{
			Name: name,
			Base: logger,
		}
		s.loggers.List = append(s.loggers.List, baseLogger)
		s.loggers.Added = append(s.loggers.Added, baseLogger)
		s.rebuildAsOne()
	}
}
