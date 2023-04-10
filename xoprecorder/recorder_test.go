package xoprecorder_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
)

func TestRecorderLogMethods(t *testing.T) {
	start := time.Now()
	rLog := xoprecorder.New()
	log := xop.NewSeed(xop.WithBase(rLog)).Request(t.Name())
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	lines := rLog.FindLines(xoprecorder.MessageEquals("basic trace message"))
	if assert.NotEmpty(t, lines, "found some") {
		assert.True(t, !lines[0].Timestamp.Before(start), "time seq")
		assert.Equal(t, "basic trace message", lines[0].Message, "message")
		assert.Equal(t, xopnum.TraceLevel, lines[0].Level, "level")
	}
	f := log.Sub().Fork("forkie")
	f.Span().Int(xopconst.HTTPStatusCode, 204)
	assert.Empty(t, rLog.FindLines(xoprecorder.MessageEquals("basic debug message")), "debug filtered out by log level")
	assert.Equal(t, 1, rLog.CountLines(xoprecorder.MessageEquals("basic alert message")), "count alert")
	assert.Equal(t, 1, rLog.CountLines(xoprecorder.TextContains("a test")), "count a test")
	assert.Equal(t, 1, rLog.CountLines(xoprecorder.TextContains("a test bar")), "count a test foo")
	assert.Equal(t, 1, rLog.CountLines(xoprecorder.TextContains("a test bar with 38")), "count a test foo with 38")
	if assert.NotEmpty(t, rLog.Spans, "have a sub-span") {
		assert.Equal(t, ".A", rLog.Spans[0].SpanSequenceCode, "span sequence for fork")
		mv := rLog.Spans[0].SpanMetadata.Get("http.status_code")
		require.NotNil(t, mv, "has http.status_code metadata")
		assert.Equal(t, int64(204), mv.Value, "an explicit attribute")
		assert.Equal(t, xopbase.Int64DataType, mv.Type, "http status code attribute type")
	}
}

func TestRecorderWithLock(t *testing.T) {
	message := make(chan int, 2)
	defer close(message)
	wg := sync.WaitGroup{}
	invocations := 0
	f := func(l *xoprecorder.Logger) error {
		time.Sleep(100 * time.Millisecond)
		invocations++
		message <- invocations
		return nil
	}
	tLog := xoptest.New(t)
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := tLog.Recorder().WithLock(f)
		assert.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		err := tLog.Recorder().WithLock(f)
		assert.NoError(t, err)
	}()

	wg.Wait()
	assert.Equal(t, 1, <-message)
	assert.Equal(t, 2, <-message)
}

func TestRecorderCustomEvent(t *testing.T) {
	tLog := xoptest.New(t)
	f := func(_ error) {}
	tLog.SetErrorReporter(f)
	tLog.Recorder().CustomEvent("custom message (%v-%v)", "foo", "bar")
	assert.Len(t, tLog.Recorder().Events, 1)
	assert.Equal(t, "custom message (foo-bar)", tLog.Recorder().Events[0].Msg)
	assert.Equal(t, xoprecorder.CustomEvent, tLog.Recorder().Events[0].Type)
}

func TestReplayRecorder(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		t.Run(mc.Name, func(t *testing.T) {
			tLog := xoptest.New(t)
			rLog := xoprecorder.New()
			log := xop.NewSeed(
				xop.WithBase(rLog),
				xop.WithBase(tLog)).
				Request(t.Name())
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				log = log.Span().SubSeed(mc.SeedMods...).Request(t.Name())
			}
			t.Log("generate logs")
			mc.Do(t, log, tLog)
			replayLog := xoptest.New(t)
			t.Log("replay from generated logs")
			err := rLog.Replay(context.Background(), replayLog)
			require.NoError(t, err, "replay")
			t.Log("verify replay equals original")
			xoptestutil.VerifyTestReplay(t, tLog, replayLog)
		})
	}
}
