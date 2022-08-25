package xop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	seed := NewSeed()
	assert.Equal(t, DefaultConfig.UseB3, seed.config.UseB3)
	assert.Equal(t, DefaultConfig.FlushDelay, seed.config.FlushDelay)

	seed = NewSeed(WithConfig(DefaultConfig), WithB3(true))
	assert.Equal(t, true, seed.config.UseB3)
	assert.Equal(t, DefaultConfig.FlushDelay, seed.config.FlushDelay)

	seed = NewSeed(WithB3(true), WithFlushDelay(time.Second))
	assert.Equal(t, true, seed.config.UseB3)
	assert.Equal(t, time.Second, seed.config.FlushDelay)
}

func TestConfigModifier(t *testing.T) {
	configModifier := func(cfg *Config) {
		cfg.UseB3 = true
	}
	seed := NewSeed(WithConfigChanges(configModifier))
	assert.Equal(t, true, seed.config.UseB3)
}

func TestLogConfig(t *testing.T) {
	myLog := NewSeed().Request("myLog")
	assert.Equal(t, DefaultConfig.FlushDelay, myLog.Config().FlushDelay)
	assert.Equal(t, DefaultConfig.UseB3, myLog.Config().UseB3)
}
