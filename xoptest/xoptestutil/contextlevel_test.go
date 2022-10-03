package xoptestutil

import (
	"context"
	"os"
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopnum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var getLogger0 = xop.LevelAdjuster(xop.WithSkippedFrames(0))
var getContext0 = xop.ContextLevelAdjuster(xop.FromContextOrDefault, xop.WithSkippedFrames(0))

func TestAdjusterContext(t *testing.T) {
	var getContextError = xop.ContextLevelAdjuster(xop.FromContextOrDefault, xop.WithDefault(xopnum.ErrorLevel))
	var getContextInfo = xop.ContextLevelAdjuster(xop.FromContextOrDefault, xop.WithDefault(xopnum.InfoLevel))
	var getContextFoo = xop.ContextLevelAdjuster(xop.FromContextOrDefault, xop.WithEnvironment("XOPLEVEL_foo"))

	ctx := context.Background()
	want0 := xop.DefaultSettings.GetMinLevel()
	wantFoo := xop.DefaultSettings.GetMinLevel()
	wantInfo := xopnum.InfoLevel
	wantError := xopnum.ErrorLevel
	if lvl, ok := os.LookupEnv("XOPLEVEL_xoptestutil"); ok {
		want, err := xopnum.LevelString(lvl)
		require.NoError(t, err, "get level")
		t.Log("XOPLEVEL_xoptestutil set, now want", want)
		want0 = want
		wantInfo = want
		wantError = want
	}
	assert.Equal(t, want0, getContext0(ctx).Settings().GetMinLevel())
	assert.Equal(t, wantInfo, getContextInfo(ctx).Settings().GetMinLevel())
	assert.Equal(t, wantError, getContextError(ctx).Settings().GetMinLevel())

	if lvl, ok := os.LookupEnv("XOPLEVEL_foo"); ok {
		want, err := xopnum.LevelString(lvl)
		require.NoError(t, err, "get level")
		t.Log("XOPLEVEL_foo set, now want", want)
		wantFoo = want
	}

	assert.Equal(t, wantFoo, getContextFoo(ctx).Settings().GetMinLevel())
}

func TestAdjusterLevel(t *testing.T) {
	var getLoggerError = xop.LevelAdjuster(xop.WithDefault(xopnum.ErrorLevel))
	var getLoggerInfo = xop.LevelAdjuster(xop.WithDefault(xopnum.InfoLevel))
	var getLoggerFoo = xop.LevelAdjuster(xop.WithEnvironment("XOPLEVEL_foo"))

	want0 := xop.DefaultSettings.GetMinLevel()
	wantFoo := xop.DefaultSettings.GetMinLevel()
	wantInfo := xopnum.InfoLevel
	wantError := xopnum.ErrorLevel
	if lvl, ok := os.LookupEnv("XOPLEVEL_xoptestutil"); ok {
		want, err := xopnum.LevelString(lvl)
		require.NoError(t, err, "get level")
		t.Log("XOPLEVEL_xoptestutil set, now want", want)
		want0 = want
		wantInfo = want
		wantError = want
	}

	assert.Equal(t, want0, getLogger0(xop.Default).Settings().GetMinLevel())
	assert.Equal(t, wantInfo, getLoggerInfo(xop.Default).Settings().GetMinLevel())
	assert.Equal(t, wantError, getLoggerError(xop.Default).Settings().GetMinLevel())

	if lvl, ok := os.LookupEnv("XOPLEVEL_foo"); ok {
		want, err := xopnum.LevelString(lvl)
		require.NoError(t, err, "get level")
		t.Log("XOPLEVEL_foo set, now want", want)
		wantFoo = want
	}

	assert.Equal(t, wantFoo, getLoggerFoo(xop.Default).Settings().GetMinLevel())
}

type foo []xop.AdjusterOption

func (f foo) Context() func(context.Context) *xop.Log {
	return xop.ContextLevelAdjuster(xop.FromContextOrDefault, f...)
}

func (f foo) Logger() func(*xop.Log) *xop.Log {
	return xop.LevelAdjuster(f...)
}

func TestAdjusterMethod(t *testing.T) {
	var getLoggerError = foo{xop.WithDefault(xopnum.ErrorLevel)}.Logger()
	var getContextInfo = foo{xop.WithDefault(xopnum.InfoLevel)}.Context()
	wantInfo := xopnum.InfoLevel
	wantError := xopnum.ErrorLevel
	if lvl, ok := os.LookupEnv("XOPLEVEL_xoptestutil"); ok {
		want, err := xopnum.LevelString(lvl)
		require.NoError(t, err, "get level")
		wantInfo = want
		wantError = want
	}
	ctx := context.Background()
	assert.Equal(t, wantInfo, getContextInfo(ctx).Settings().GetMinLevel())
	assert.Equal(t, wantError, getLoggerError(xop.Default).Settings().GetMinLevel())
}
