package xoptest_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
)

func TestLogMethods(t *testing.T) {
	start := time.Now()
	tLog := xoptest.New(t)
	log := tLog.Logger()
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Log().Msg("basic log message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	lines := tLog.Recorder().FindLines(xoprecorder.MessageEquals("basic log message"))
	if assert.NotEmpty(t, lines, "found some") {
		assert.True(t, !lines[0].Timestamp.Before(start), "time seq")
		assert.Equal(t, "basic log message", lines[0].Message, "message")
		assert.Equal(t, xopnum.LogLevel, lines[0].Level, "level")
	}
	f := log.Sub().Fork("forkie")
	f.Span().Int(xopconst.HTTPStatusCode, 204)
	assert.Empty(t, tLog.Recorder().FindLines(xoprecorder.MessageEquals("basic trace message")), "debug filtered out by log level")
	assert.Equal(t, 1, tLog.Recorder().CountLines(xoprecorder.MessageEquals("basic alert message")), "count alert")
	assert.Equal(t, 1, tLog.Recorder().CountLines(xoprecorder.TextContains("a test")), "count a test")
	assert.Equal(t, 1, tLog.Recorder().CountLines(xoprecorder.TextContains("a test bar")), "count a test foo")
	assert.Equal(t, 1, tLog.Recorder().CountLines(xoprecorder.TextContains("a test bar with 38")), "count a test foo with 38")
	if assert.NotEmpty(t, tLog.Recorder().Spans, "have a sub-span") {
		assert.Equal(t, ".A", tLog.Recorder().Spans[0].SpanSequenceCode, "span sequence for fork")
		md := tLog.Recorder().Spans[0].SpanMetadata.Get("http.status_code")
		if assert.NotNil(t, md, "has http.status_code metadata") {
			assert.Equal(t, int64(204), md.Value, "an explicit attribute")
			assert.Equal(t, xopbase.Int64DataType, md.Type, "http status code attribute type")
		}
	}
}

func TestReplayTextLogger(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		t.Run(mc.Name, func(t *testing.T) {
			tLog := xoptest.New(t)
			log := tLog.Logger()
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				log = log.Span().SubSeed(mc.SeedMods...).Request(t.Name())
			}
			t.Log("generate logs")
			mc.Do(t, log, tLog)
			rLog := xoptest.New(t)
			t.Log("replay from generated logs")
			err := tLog.Recorder().Replay(context.Background(), rLog)
			require.NoError(t, err, "replay")
			t.Log("verify replay equals original")
			xoptestutil.VerifyTestReplay(t, tLog, rLog)
		})
	}
}
