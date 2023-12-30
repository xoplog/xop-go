package xop

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/xoplog/xop-go/xopnum"
	"testing"
	"time"
)

func TestLog_Settings(t *testing.T) {
	defaultLog := Default
	logSettings := defaultLog.Settings()
	assert.Equal(t, noFilenameFunc(DefaultSettings), noFilenameFunc(logSettings))
}

func TestSub_StackFrames(t *testing.T) {
	sub := Default.Sub()
	sub.StackFrames(xopnum.TraceLevel, 100)
	sub.StackFrames(xopnum.ErrorLevel, 100)
	sub.StackFrames(xopnum.DebugLevel, 2)
	assert.Equal(t, 100, sub.settings.stackFramesWanted[xopnum.ErrorLevel])
	assert.Equal(t, 100, sub.settings.stackFramesWanted[xopnum.AlertLevel])
	assert.Equal(t, 2, sub.settings.stackFramesWanted[xopnum.TraceLevel])
	assert.Equal(t, 2, sub.settings.stackFramesWanted[xopnum.DebugLevel])
}

func TestSub_SynchronousFlush(t *testing.T) {
	sub := Default.Sub()
	sub.SynchronousFlush(true)
	assert.True(t, sub.settings.synchronousFlushWhenDone)
	sub.SynchronousFlush(false)
	assert.False(t, sub.settings.synchronousFlushWhenDone)
}

func TestSub_MinLevel(t *testing.T) {
	sub := Default.Sub()
	sub.MinLevel(xopnum.ErrorLevel)
	assert.Equal(t, xopnum.ErrorLevel, sub.settings.minimumLogLevel)
}

func TestSub_TagLinesWithSpanSequence(t *testing.T) {
	sub := Default.Sub()
	sub.TagLinesWithSpanSequence(true)
	assert.True(t, sub.settings.tagLinesWithSpanSequence)
	sub.TagLinesWithSpanSequence(false)
	assert.False(t, sub.settings.tagLinesWithSpanSequence)
}

func TestSub_PrefillText(t *testing.T) {
	sub := Default.Sub()
	assert.Empty(t, sub.settings.prefillMsg)
	sub.PrefillText("text")
	assert.Equal(t, "text", sub.settings.prefillMsg)
}

func TestSub_NoPrefill(t *testing.T) {
	sub := Default.Sub()
	assert.Nil(t, sub.settings.prefillData)
	assert.Empty(t, sub.settings.prefillMsg)
	sub.PrefillAny(Key("foo"), "bar")
	sub.PrefillText("text")
	assert.NotNil(t, sub.settings.prefillData)
	assert.NotEmpty(t, sub.settings.prefillMsg)
	sub.NoPrefill()
	assert.Nil(t, sub.settings.prefillData)
	assert.Empty(t, sub.settings.prefillMsg)
}

func TestSub_PrefillBool(t *testing.T) {
	sub := Default.Sub()
	sub.PrefillBool(Key("key"), true)
	assert.Len(t, sub.settings.prefillData, 1)
}

func TestSub_PrefillDuration(t *testing.T) {
	sub := Default.Sub()
	sub.PrefillDuration(Key("key"), 1*time.Second)
	assert.Len(t, sub.settings.prefillData, 1)
}

func TestSub_PrefillError(t *testing.T) {
	sub := Default.Sub()
	sub.PrefillError(Key("key"), errors.New("error"))
	assert.Len(t, sub.settings.prefillData, 1)
}

func TestSub_PrefillFloat64(t *testing.T) {
	sub := Default.Sub()
	sub.PrefillFloat64(Key("key"), 64)
	assert.Len(t, sub.settings.prefillData, 1)
}
