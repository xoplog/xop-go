package xoptest_test

import (
	"testing"
	"time"

	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
)

func TestLogMethods(t *testing.T) {
	start := time.Now()
	tlog := xoptest.New(t)
	log := tlog.Log()
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	lines := tlog.FindLines(xoptest.MessageEquals("basic trace message"))
	if assert.NotEmpty(t, lines, "found some") {
		assert.True(t, !lines[0].Timestamp.Before(start), "time seq")
		assert.Equal(t, "basic trace message", lines[0].Message, "message")
		assert.Equal(t, xopconst.TraceLevel, lines[0].Level, "level")
	}
	f := log.Sub().Fork("forkie")
	f.Span().Int(xopconst.HTTPStatusCode, 204)
	assert.Empty(t, tlog.FindLines(xoptest.MessageEquals("basic debug message")), "debug filtered out by log level")
	assert.Equal(t, 1, tlog.CountLines(xoptest.MessageEquals("basic alert message")), "count alert")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test")), "count a test")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test bar")), "count a test foo")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test bar with 38")), "count a test foo with 38")
	if assert.NotEmpty(t, tlog.Spans, "have a sub-span") {
		assert.Equal(t, tlog.Spans[0].Metadata["span.seq"], ".A", "span sequence for fork")
		assert.Equal(t, tlog.Spans[0].Metadata["http.status_code"], int64(204), "an explicit attribute")
		assert.Equal(t, xoputil.BaseAttributeTypeString, tlog.Spans[0].MetadataTypes["span.seq"], "span.seq attribute type")
		assert.Equal(t, xoputil.BaseAttributeTypeInt64, tlog.Spans[0].MetadataTypes["http.status_code"], "http status code attribute type")
	}
}
