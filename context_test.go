package xop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntoContext(t *testing.T) {
	log := NewSeed().Request("myLog")
	ctx := context.Background()
	ctxWithLog := log.IntoContext(ctx)
	assert.Equal(t, log, ctxWithLog.Value(contextKey))
}

func TestFromContext(t *testing.T) {
	log := NewSeed().Request("myLog")
	ctx := context.Background()
	ctxLog, ok := FromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, ctxLog)
	ctxWithLog := log.IntoContext(ctx)
	ctxLog, ok = FromContext(ctxWithLog)
	assert.True(t, ok)
	assert.Equal(t, log, ctxLog)
}

func TestFromContextOrDefault(t *testing.T) {
	ctx := context.Background()
	log := FromContextOrDefault(ctx)
	assert.Equal(t, Default, log)
	log = NewSeed().Request("myLog")
	ctxLog := log.IntoContext(ctx)
	assert.Equal(t, log, FromContextOrDefault(ctxLog))
}

func TestFromContextOrPanic(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() { _ = FromContextOrPanic(ctx) })
	log := NewSeed().Request("myLog")
	ctxLog := log.IntoContext(ctx)
	assert.Equal(t, log, FromContextOrPanic(ctxLog))
}

func TestCustomFromContext(t *testing.T) {
	adjustSettings := func(sub *Sub) *Sub {
		return sub.PrefillText("banana")
	}
	customLogFun := CustomFromContext(FromContextOrDefault, adjustSettings)
	customLog := customLogFun(context.Background())
	defer customLog.Done()
	assert.NotEqual(t, Default, customLog, "logs are not equal")
	assert.NotEqual(t, Default.Settings(), customLog.Settings(), "default settings")
	assert.Equal(t, noFilenameFunc(Default.Sub().PrefillText("banana").Log().Settings()), noFilenameFunc(customLog.Settings()), "modified settings")
}

func noFilenameFunc(settings LogSettings) LogSettings {
	settings.StackFilenameRewrite(nil)
	return settings
}
