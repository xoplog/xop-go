package xop

import (
	"github.com/muir/xop-go/xopbase"
)

type loggers struct {
	Flushers baseRequests
	List     baseLoggers
	Removed  baseLoggers
	Added    baseLoggers
}

// Copy copies everything but Added & Removed.
// List, is a deep-ish copy.
func (s loggers) Copy() loggers {
	n := make(baseLoggers, len(s.List))
	copy(n, s.List)
	return loggers{
		Flushers: s.Flushers,
		List:     n,
	}
}

func WithoutBase(baseLoggerToRemove xopbase.Logger) SeedModifier {
	return func(s *Seed) {
		for i, baseLogger := range s.loggers.List {
			if baseLogger == baseLoggerToRemove {
				s.loggers.Removed = append(s.loggers.Removed, baseLogger)
				if i < len(s.loggers.List)-1 {
					s.loggers.List[i], s.loggers.List[len(s.loggers.List)-1] =
						s.loggers.List[len(s.loggers.List)-1], s.loggers.List[i]
				}
				s.loggers.List = s.loggers.List[:len(s.loggers.List)-1]
			}
		}
	}
}

func WithBase(baseLogger xopbase.Logger) SeedModifier {
	return func(s *Seed) {
		s.loggers.List = append(s.loggers.List, baseLogger)
		s.loggers.Added = append(s.loggers.Added, baseLogger)
	}
}
