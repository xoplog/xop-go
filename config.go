package xop

import (
	golog "log"
	"time"
)

type Config struct {
	UseB3      bool // Zipkin
	FlushDelay time.Duration

	// ErrorReporter provides a way to choose the behavior
	// for when underlying log functions throw an error.
	// Generally speaking, needing to check errors when
	// generating logs is a non-starter because the cost is
	// too high.  It would discourage logging.  That said,
	// there is a an error, we don't want to completely
	// ignore it.
	ErrorReporter func(error)
}

// LogLinkPrefix can be set during init(). It should not be modified
// afterwards. It can be used by base loggers (like xoptest) to prefix
// a traceheader to turn it into a link.
var LogLinkPrefix string

var DefaultConfig = Config{
	FlushDelay: time.Minute * 5,
	ErrorReporter: func(err error) {
		golog.Print("Error from zop", err)
	},
}

type ConfigModifier func(*Config)

func WithFlushDelay(d time.Duration) SeedModifier {
	return func(s *Seed) {
		s.config.FlushDelay = d
	}
}

func WithB3(b bool) SeedModifier {
	return func(s *Seed) {
		s.config.UseB3 = b
	}
}

func WithConfig(config Config) SeedModifier {
	return func(s *Seed) {
		s.config = config
	}
}

func WithConfigChanges(mods ...ConfigModifier) SeedModifier {
	return func(s *Seed) {
		for _, mod := range mods {
			mod(&s.config)
		}
	}
}

func (logger *Logger) Config() Config {
	return logger.span.seed.config
}
