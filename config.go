package xm

type Config struct {
	UseB3 bool // Zipkin
}

type ConfigModifier func(*Config)

func (s Seed) AlterConfig(mods ...ConfigModifier) {
	for _, mod := range mods {
		mod(&s.config)
	}
}

func (l *Logger) Config() Config {
	return l.seed.config
}
