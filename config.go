package xoplog

import "time"

type Config struct {
	UseB3      bool // Zipkin
	FlushDelay time.Duration
	// TODO: Errorf func(msg string, v ...interface{})
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

func (l *Log) Config() Config {
	return l.span.seed.config
}
