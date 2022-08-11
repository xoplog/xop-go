package xop_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopjson"
)

var msg = "The quick brown fox jumps over the lazy dog"
var obj = struct {
	Rate string
	Low  int
	High float32
}{"15", 16, 123.2}

func BenchmarkDisableXop(b *testing.B) {
	logger := xop.NewSeed(xop.WithBase(
		xopjson.New(
			xopbytes.WriteToIOWriter(ioutil.Discard),
			xopjson.WithEpochTime(time.Nanosecond),
			xopjson.WithDurationFormat(xopjson.AsNanos),
			xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			xopjson.WithAttributesObject(false)))).
		Request("disable")
	for i := 0; i < b.N; i++ {
		logger.Debug().String("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
	logger.Done()
}

func BenchmarkNormalXop(b *testing.B) {
	logger := xop.NewSeed(xop.WithBase(
		xopjson.New(
			xopbytes.WriteToIOWriter(ioutil.Discard),
			xopjson.WithEpochTime(time.Nanosecond),
			xopjson.WithDurationFormat(xopjson.AsNanos),
			xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			xopjson.WithAttributesObject(false)))).
		Request("disable")
	for i := 0; i < b.N; i++ {
		logger.Info().String("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
	logger.Done()
}

func BenchmarkPrintfXop(b *testing.B) {
	logger := xop.NewSeed(xop.WithBase(
		xopjson.New(
			xopbytes.WriteToIOWriter(ioutil.Discard),
			xopjson.WithEpochTime(time.Nanosecond),
			xopjson.WithDurationFormat(xopjson.AsNanos),
			xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			xopjson.WithAttributesObject(false)))).
		Request("disable")
	for i := 0; i < b.N; i++ {
		logger.Info().Msgf("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
	logger.Done()
}

func BenchmarkCallerXop(b *testing.B) {
	logger := xop.NewSeed(xop.WithBase(
		xopjson.New(
			xopbytes.WriteToIOWriter(ioutil.Discard),
			xopjson.WithEpochTime(time.Nanosecond),
			xopjson.WithDurationFormat(xopjson.AsNanos),
			xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			xopjson.WithAttributesObject(false)))).
		Request("disable").
		Sub().
		StackFrames(xopconst.InfoLevel, 1).
		Log()
	for i := 0; i < b.N; i++ {
		logger.Info().String("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
	logger.Done()
}
