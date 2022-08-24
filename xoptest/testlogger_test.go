package xoptest_test

import (
	"sync"
	"testing"
	"time"

	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoputil"
	"github.com/stretchr/testify/assert"
)

func TestLogMethods(t *testing.T) {
	start := time.Now()
	tLog := xoptest.New(t)
	defer tLog.Close()
	log := tLog.Log()
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	lines := tLog.FindLines(xoptest.MessageEquals("basic trace message"))
	if assert.NotEmpty(t, lines, "found some") {
		assert.True(t, !lines[0].Timestamp.Before(start), "time seq")
		assert.Equal(t, "basic trace message", lines[0].Message, "message")
		assert.Equal(t, xopconst.TraceLevel, lines[0].Level, "level")
	}
	f := log.Sub().Fork("forkie")
	f.Span().Int(xopconst.HTTPStatusCode, 204)
	assert.Empty(t, tLog.FindLines(xoptest.MessageEquals("basic debug message")), "debug filtered out by log level")
	assert.Equal(t, 1, tLog.CountLines(xoptest.MessageEquals("basic alert message")), "count alert")
	assert.Equal(t, 1, tLog.CountLines(xoptest.TextContains("a test")), "count a test")
	assert.Equal(t, 1, tLog.CountLines(xoptest.TextContains("a test bar")), "count a test foo")
	assert.Equal(t, 1, tLog.CountLines(xoptest.TextContains("a test bar with 38")), "count a test foo with 38")
	if assert.NotEmpty(t, tLog.Spans, "have a sub-span") {
		assert.Equal(t, ".A", tLog.Spans[0].SequenceCode, "span sequence for fork")
		assert.Equal(t, int64(204), tLog.Spans[0].Metadata["http.status_code"], "an explicit attribute")
		assert.Equal(t, xoputil.BaseAttributeTypeInt64, tLog.Spans[0].MetadataTypes["http.status_code"], "http status code attribute type")
	}
}

func TestWithLock(t *testing.T) {
	message := make(chan int, 2)
	defer close(message)
	wg := sync.WaitGroup{}
	invocations := 0
	f := func(l *xoptest.TestLogger) error {
		time.Sleep(100 * time.Millisecond)
		invocations++
		message <- invocations
		return nil
	}
	tLog := xoptest.New(t)
	defer tLog.Close()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tLog.WithLock(f)
		assert.NoError(t, err)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tLog.WithLock(f)
		assert.NoError(t, err)
	}()

	wg.Wait()
	assert.Equal(t, 1, <-message)
	assert.Equal(t, 2, <-message)
}

func TestCustomEvent(t *testing.T) {
	tLog := xoptest.New(t)
	defer tLog.Close()
	f := func(_ error) {}
	tLog.SetErrorReporter(f)
	tLog.CustomEvent("custom message (%v-%v)", "foo", "bar")
	assert.Len(t, tLog.Events, 1)
	assert.Equal(t, "custom message (foo-bar)", tLog.Events[0].Msg)
	assert.Equal(t, xoptest.CustomEvent, tLog.Events[0].Type)
}
