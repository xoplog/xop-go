package xoptest_test

import (
	"testing"
	"time"

	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoptest"

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
	log.Info().Str("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	lines := tlog.FindLines(xoptest.MessageEquals("basic trace message"))
	if assert.NotEmpty(t, lines, "found some") {
		assert.True(t, !lines[0].Timestamp.Before(start), "time seq")
		assert.Equal(t, "basic trace message", lines[0].Message, "message")
		assert.Equal(t, xopconst.TraceLevel, lines[0].Level, "level")
	}
	assert.Empty(t, tlog.FindLines(xoptest.MessageEquals("basic debug message")), "debug filtered out by log level")
	assert.Equal(t, 1, tlog.CountLines(xoptest.MessageEquals("basic alert message")), "count alert")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test")), "count a test")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test bar")), "count a test foo")
	assert.Equal(t, 1, tlog.CountLines(xoptest.TextContains("a test bar with 38")), "count a test foo with 38")
}
