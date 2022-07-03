package xoplog

type Config struct {
	UseB3 bool // Zipkin
}

type ConfigModifier func(*Config)

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
	return l.seed.config
}
