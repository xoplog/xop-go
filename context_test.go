package xop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntoContext(t *testing.T) {
	log := NewSeed().Request("myLog")
	ctx := context.TODO()
	ctxWithLog := log.IntoContext(ctx)
	assert.Equal(t, log, ctxWithLog.Value(contextKey))
}

func TestFromContext(t *testing.T) {
	log := NewSeed().Request("myLog")
	ctx := context.TODO()
	ctxLog, ok := FromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, ctxLog)
	ctxWithLog := log.IntoContext(ctx)
	ctxLog, ok = FromContext(ctxWithLog)
	assert.True(t, ok)
	assert.Equal(t, log, ctxLog)
}

func TestFromContextOrDefault(t *testing.T) {
	ctx := context.TODO()
	log := FromContextOrDefault(ctx)
	assert.Equal(t, Default, log)
	log = NewSeed().Request("myLog")
	ctxLog := log.IntoContext(ctx)
	assert.Equal(t, log, FromContextOrDefault(ctxLog))
}

func TestFromContextOrPanic(t *testing.T) {
	ctx := context.TODO()
	assert.Panics(t, func() { _ = FromContextOrPanic(ctx) })
	log := NewSeed().Request("myLog")
	ctxLog := log.IntoContext(ctx)
	assert.Equal(t, log, FromContextOrPanic(ctxLog))
}

func TestCustomFromContext(t *testing.T) {
	getDefaultLogFromContext := func(ctx context.Context) *Log {
		return Default
	}
	noOpAdjustSettings := func(sub *Sub) *Sub {
		return sub
	}
	customFromContextFun := CustomFromContext(getDefaultLogFromContext, noOpAdjustSettings)
	assert.Equal(t, Default, customFromContextFun(context.TODO()))
}
