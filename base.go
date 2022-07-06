package xoplog

import (
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type baseLoggers struct {
	List    []baseLogger
	Removed []baseLogger
}

type baseLogger struct {
	Name     string
	Base     xopbase.Logger
	MinLevel xopconst.Level
}

func (s baseLoggers) requests(bundle trace.Bundle) (xopbase.Requests, bool) {
	baseRequests := make(xopbase.Requests, len(s.List))
	var referencesKept bool
	for i, baseLogger := range s.List {
		baseRequests[i] = baseLogger.Base.Request(bundle)
		if baseLogger.Base.ReferencesKept() {
			referencesKept = true
		}
	}
	return baseRequests, referencesKept
}

func (s baseLoggers) copyWithoutTrace() baseLoggers {
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

func WithBaseLogger(name string, logger xopbase.Logger) SeedModifier {
	return func(s *Seed) {
		s.baseLoggers.List = append(s.baseLoggers.List, baseLogger{
			Name: name,
			Base: logger,
		})
	}
}
