package xoptestutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/trace"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptest"

	"github.com/stretchr/testify/assert"
)

const NeedsEscaping = `"\<'` + "\n\r\t\b\x00"

var MessageCases = []struct {
	Name         string
	ExtraFlushes int
	Do           func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger)
	SkipOTEL     bool
	SeedMods     []xop.SeedModifier
}{
	{
		Name: "one span",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			log.Info().Msg("basic info message")
			log.Error().Msg("basic error message")
			log.Alert().Msg("basic alert message")
			log.Debug().Msg("basic debug message")
			log.Trace().Msg("basic trace message")
			log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")

			ss := log.Sub().Detach().Fork("a fork one span")
			MicroNap()
			ss.Alert().String("frightening", "stuff").Static("like a rock" + NeedsEscaping)
			ss.Span().String(xopconst.EndpointRoute, "/some/thing")

			MicroNap()
			tlog.CustomEvent("before log.Done")
			log.Done()
			tlog.CustomEvent("after log.Done")
			ss.Debug().Msg("sub-span debug message")
			MicroNap()
			tlog.CustomEvent("before ss.Done")
			ss.Done()
			tlog.CustomEvent("after ss.Done")
		},
	},
	{
		Name: "metadata singles in request",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			log.Span().Bool(ExampleMetadataSingleBool, false)
			log.Span().Bool(ExampleMetadataSingleBool, true)
			log.Span().Bool(ExampleMetadataLockedBool, true)
			log.Span().Bool(ExampleMetadataLockedBool, false)
			log.Span().String(ExampleMetadataLockedString, "loki"+NeedsEscaping)
			log.Span().String(ExampleMetadataLockedString, "thor"+NeedsEscaping)
			log.Span().Int(ExampleMetadataLockedInt, 38)
			log.Span().Int(ExampleMetadataLockedInt, -38)
			log.Span().Int8(ExampleMetadataLockedInt8, 39)
			log.Span().Int8(ExampleMetadataLockedInt8, -39)
			log.Span().Int16(ExampleMetadataLockedInt16, 329)
			log.Span().Int16(ExampleMetadataLockedInt16, -329)
			log.Span().Int32(ExampleMetadataLockedInt32, -932)
			log.Span().Int32(ExampleMetadataLockedInt32, 932)
			log.Span().Int64(ExampleMetadataLockedInt64, -93232)
			log.Span().Int64(ExampleMetadataLockedInt64, 93232)
			log.Span().String(ExampleMetadataSingleString, "athena")
			log.Span().Int(ExampleMetadataSingleInt, 3)
			log.Span().Int8(ExampleMetadataSingleInt8, 9)
			log.Span().Int16(ExampleMetadataSingleInt16, 29)
			log.Span().Int32(ExampleMetadataSingleInt32, -32)
			log.Span().Int64(ExampleMetadataSingleInt64, -3232)
			MicroNap()
			log.Done()
		},
	},
	{
		Name:     "metadata traces",
		SkipOTEL: true,
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			s2 := log.Sub().Fork("S2")
			s3 := s2.Sub().Fork("S3")
			log.Span().Link(ExampleMetadataSingleLink, s2.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataSingleLink, s3.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataLockedLink, s2.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataLockedLink, s3.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataMultipleLink, s2.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataMultipleLink, s3.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataMultipleLink, s3.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataDistinctLink, s2.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataDistinctLink, s3.Span().Bundle().Trace)
			log.Span().Link(ExampleMetadataDistinctLink, s3.Span().Bundle().Trace)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata float64",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			log.Span().Float64(ExampleMetadataSingleFloat64, 40.3)
			log.Span().Float64(ExampleMetadataSingleFloat64, 40.4)
			log.Span().Float64(ExampleMetadataLockedFloat64, 30.5)
			log.Span().Float64(ExampleMetadataLockedFloat64, 30.6)
			log.Span().Float64(ExampleMetadataMultipleFloat64, 10.7)
			log.Span().Float64(ExampleMetadataMultipleFloat64, 10.8)
			log.Span().Float64(ExampleMetadataMultipleFloat64, 10.7)
			log.Span().Float64(ExampleMetadataDistinctFloat64, 20.8)
			log.Span().Float64(ExampleMetadataDistinctFloat64, 20.7)
			log.Span().Float64(ExampleMetadataDistinctFloat64, 20.8)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata time",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			t1 := time.Now().Round(time.Second)
			t2 := t1.Add(time.Minute)
			log.Span().Time(ExampleMetadataSingleTime, t1)
			log.Span().Time(ExampleMetadataSingleTime, t2)
			log.Span().Time(ExampleMetadataLockedTime, t1)
			log.Span().Time(ExampleMetadataLockedTime, t2)
			log.Span().Time(ExampleMetadataMultipleTime, t1)
			log.Span().Time(ExampleMetadataMultipleTime, t2)
			log.Span().Time(ExampleMetadataMultipleTime, t2)
			log.Span().Time(ExampleMetadataDistinctTime, t1)
			log.Span().Time(ExampleMetadataDistinctTime, t2)
			log.Span().Time(ExampleMetadataDistinctTime, t2)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata singles in span",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			ss := log.Sub().Fork("spoon")
			ss.Span().Bool(ExampleMetadataSingleBool, false)
			ss.Span().Bool(ExampleMetadataSingleBool, true)
			ss.Span().Bool(ExampleMetadataLockedBool, true)
			ss.Span().Bool(ExampleMetadataLockedBool, false)
			ss.Span().String(ExampleMetadataLockedString, "loki")
			ss.Span().String(ExampleMetadataLockedString, "thor")
			ss.Span().Int(ExampleMetadataLockedInt, 38)
			ss.Span().Int(ExampleMetadataLockedInt, -38)
			ss.Span().Int8(ExampleMetadataLockedInt8, 39)
			ss.Span().Int8(ExampleMetadataLockedInt8, -39)
			ss.Span().Int16(ExampleMetadataLockedInt16, 329)
			ss.Span().Int16(ExampleMetadataLockedInt16, -329)
			ss.Span().Int32(ExampleMetadataLockedInt32, -932)
			ss.Span().Int32(ExampleMetadataLockedInt32, 932)
			ss.Span().Int64(ExampleMetadataLockedInt64, -93232)
			ss.Span().Int64(ExampleMetadataLockedInt64, 93232)
			ss.Span().String(ExampleMetadataSingleString, "athena"+NeedsEscaping)
			ss.Span().Int(ExampleMetadataSingleInt, 3)
			ss.Span().Int8(ExampleMetadataSingleInt8, 9)
			ss.Span().Int16(ExampleMetadataSingleInt16, 29)
			ss.Span().Int32(ExampleMetadataSingleInt32, -32)
			ss.Span().Int64(ExampleMetadataSingleInt64, -3232)
			ss.Span().Bool(ExampleMetadataSingleBool, false)
			ss.Span().Bool(ExampleMetadataSingleBool, true)
			ss.Span().Bool(ExampleMetadataLockedBool, true)
			ss.Span().Bool(ExampleMetadataLockedBool, false)
			ss.Span().String(ExampleMetadataLockedString, "loki")
			ss.Span().String(ExampleMetadataLockedString, "thor")
			ss.Span().Int(ExampleMetadataLockedInt, 38)
			ss.Span().Int(ExampleMetadataLockedInt, -38)
			ss.Span().Int8(ExampleMetadataLockedInt8, 39)
			ss.Span().Int8(ExampleMetadataLockedInt8, -39)
			ss.Span().Int16(ExampleMetadataLockedInt16, 329)
			ss.Span().Int16(ExampleMetadataLockedInt16, -329)
			ss.Span().Int32(ExampleMetadataLockedInt32, -932)
			ss.Span().Int32(ExampleMetadataLockedInt32, 932)
			ss.Span().Int64(ExampleMetadataLockedInt64, -93232)
			ss.Span().Int64(ExampleMetadataLockedInt64, 93232)
			ss.Span().String(ExampleMetadataSingleString, "athena")
			ss.Span().Int(ExampleMetadataSingleInt, 3)
			ss.Span().Int8(ExampleMetadataSingleInt8, 9)
			ss.Span().Int16(ExampleMetadataSingleInt16, 29)
			ss.Span().Int32(ExampleMetadataSingleInt32, -32)
			ss.Span().Int64(ExampleMetadataSingleInt64, -3232)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata any",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			ss := log.Sub().Fork("knife")
			a := map[string]interface{}{
				"foo":   "bar",
				"count": 329,
				"array": []int{8, 22},
			}
			b := map[string]interface{}{
				"foo":   "baz",
				"count": 10,
				"array": []int{33, 39},
			}
			ss.Span().Any(ExampleMetadataSingleAny, a)
			ss.Span().Any(ExampleMetadataSingleAny, b)
			ss.Span().Any(ExampleMetadataLockedAny, a)
			ss.Span().Any(ExampleMetadataLockedAny, b)
			ss.Span().Any(ExampleMetadataMultipleAny, a)
			ss.Span().Any(ExampleMetadataMultipleAny, b)
			ss.Span().Any(ExampleMetadataDistinctAny, a)
			ss.Span().Any(ExampleMetadataDistinctAny, a)
			ss.Span().Any(ExampleMetadataDistinctAny, b)
			log.Span().Any(ExampleMetadataSingleAny, a)
			log.Span().Any(ExampleMetadataSingleAny, b)
			log.Span().Any(ExampleMetadataLockedAny, a)
			log.Span().Any(ExampleMetadataLockedAny, b)
			log.Span().Any(ExampleMetadataMultipleAny, a)
			log.Span().Any(ExampleMetadataMultipleAny, b)
			log.Span().Any(ExampleMetadataDistinctAny, a)
			log.Span().Any(ExampleMetadataDistinctAny, a)
			log.Span().Any(ExampleMetadataDistinctAny, b)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata iota enum",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			ss := log.Sub().Step("stool")
			ss.Span().EmbeddedEnum(SingleEnumTwo)
			ss.Span().EmbeddedEnum(SingleEnumTwo)
			ss.Span().EmbeddedEnum(SingleEnumThree)
			ss.Span().EmbeddedEnum(LockedEnumTwo)
			ss.Span().EmbeddedEnum(LockedEnumTwo)
			ss.Span().EmbeddedEnum(LockedEnumThree)
			log.Span().EmbeddedEnum(MultipleEnumTwo)
			log.Span().EmbeddedEnum(MultipleEnumTwo)
			log.Span().EmbeddedEnum(MultipleEnumThree)
			log.Span().EmbeddedEnum(DistinctEnumTwo)
			log.Span().EmbeddedEnum(DistinctEnumTwo)
			log.Span().EmbeddedEnum(DistinctEnumThree)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata embedded enum",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			ss := log.Sub().Step("stool")
			ss.Span().EmbeddedEnum(SingleEEnumTwo)
			ss.Span().EmbeddedEnum(SingleEEnumTwo)
			ss.Span().EmbeddedEnum(SingleEEnumThree)
			ss.Span().EmbeddedEnum(LockedEEnumTwo)
			ss.Span().EmbeddedEnum(LockedEEnumTwo)
			ss.Span().EmbeddedEnum(LockedEEnumThree)
			log.Span().EmbeddedEnum(MultipleEEnumTwo)
			log.Span().EmbeddedEnum(MultipleEEnumTwo)
			log.Span().EmbeddedEnum(MultipleEEnumThree)
			log.Span().EmbeddedEnum(DistinctEEnumTwo)
			log.Span().EmbeddedEnum(DistinctEEnumTwo)
			log.Span().EmbeddedEnum(DistinctEEnumThree)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata enum",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			ss := log.Sub().Step("stool")
			ss.Span().Enum(ExampleMetadataSingleXEnum, xopconst.SpanKindServer)
			ss.Span().Enum(ExampleMetadataSingleXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataSingleXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataLockedXEnum, xopconst.SpanKindServer)
			ss.Span().Enum(ExampleMetadataLockedXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataLockedXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataMultipleXEnum, xopconst.SpanKindServer)
			ss.Span().Enum(ExampleMetadataMultipleXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataMultipleXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataDistinctXEnum, xopconst.SpanKindServer)
			ss.Span().Enum(ExampleMetadataDistinctXEnum, xopconst.SpanKindClient)
			ss.Span().Enum(ExampleMetadataDistinctXEnum, xopconst.SpanKindClient)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata multiples",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			// ss := log.Sub().Fork("a fork metadata multiples")
			log.Span().Bool(ExampleMetadataMultipleBool, true)
			log.Span().Bool(ExampleMetadataMultipleBool, true)
			log.Span().Int(ExampleMetadataMultipleInt, 3)
			log.Span().Int(ExampleMetadataMultipleInt, 5)
			log.Span().Int(ExampleMetadataMultipleInt, 7)
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "metadata distinct",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			// ss := log.Sub().Fork("a fork metadata distinct")
			log.Span().Bool(ExampleMetadataDistinctBool, true)
			log.Span().Bool(ExampleMetadataDistinctBool, true)
			log.Span().Bool(ExampleMetadataDistinctBool, false)
			log.Span().Int(ExampleMetadataDistinctInt, 3)
			log.Span().Int(ExampleMetadataDistinctInt, 5)
			log.Span().Int(ExampleMetadataDistinctInt, 3)
			log.Span().Int(ExampleMetadataDistinctInt, 7)
			log.Span().Int64(ExampleMetadataDistinctInt64, 73)
			log.Span().Int64(ExampleMetadataDistinctInt64, 75)
			log.Span().Int64(ExampleMetadataDistinctInt64, 73)
			log.Span().Int64(ExampleMetadataDistinctInt64, 77)
			log.Span().String(ExampleMetadataDistinctString, "abc")
			log.Span().String(ExampleMetadataDistinctString, "abc")
			log.Span().String(ExampleMetadataDistinctString, "def")
			log.Span().String(ExampleMetadataDistinctString, "abc")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "one done",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			_ = log.Sub().Fork("a fork one done")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "prefill",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillFloat64("f", 23).PrefillText("pre!").Log()
			p.Error().Int16("i16", int16(7)).Msg("pf")
			log.Alert().Int32("i32", int32(77)).Msgf("pf %s", "bar")
			MicroNap()
			log.Done()
		},
	},
	{
		Name:         "add/remove loggers with a seed",
		ExtraFlushes: 2,
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			tlog2 := xoptest.New(t)
			r2 := log.Span().Seed(xop.WithBase(tlog2)).Request("R2")
			r3 := r2.Span().Seed(xop.WithoutBase(tlog2)).Request("R3")
			r2.Info().Static("log to both test loggers")
			r3.Info().Static("log to just the original set")
			MicroNap()
			log.Done()
			r2.Done()
			r3.Done()
		},
	},
	{
		Name: "add/remove loggers with a span",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			tlog2 := xoptest.New(t)
			s2 := log.Sub().Step("S2", xop.WithBase(tlog2))
			s3 := s2.Sub().Detach().Fork("S3", xop.WithoutBase(tlog2))
			s2.Info().Static("log to both test loggers")
			s3.Info().Static("log to just the original set")
			MicroNap()
			s2.Done()
			s3.Done()
			log.Done()
		},
	},
	{
		Name:         "log after done",
		ExtraFlushes: 1,
		SkipOTEL:     true,
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			s2 := log.Sub().Step("S2")
			s2.Info().Int8("i8", 9).Msg("a line before done")
			MicroNap()
			s2.Done()
			assert.Empty(t, tlog.FindLines(xoptest.TextContains("XOP: log was already done, but was used again")), "no err")
			s2.Info().Int16("i16", 940).Msg("a post-done line, should trigger an error log")
			assert.NotEmpty(t, tlog.FindLines(xoptest.TextContains("XOP: log was already done, but was used again")), "no err")
			assert.Empty(t, tlog.FindLines(xoptest.TextContains("called on log object when it was already Done")), "no err")
			MicroNap()
			s2.Done()
			assert.NotEmpty(t, tlog.FindLines(xoptest.TextContains("called on log object when it was already Done")), "now err")
			log.Flush()
			s2.Warn().Int32("i32", 940940).Msg("another post-done line, should trigger an error log")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "lots of types",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillInt("int", 439).PrefillInt8("int8", 82).PrefillInt16("int16", 829).
				PrefillInt32("int32", 4328).PrefillInt64("int64", -2382).
				PrefillUint("uint", 439).PrefillUint8("uint8", 82).PrefillUint16("uint16", 829).
				PrefillUint32("uint32", 4328).PrefillUint64("uint64", 2382).
				PrefillString("foo", "bar").PrefillBool("on/off", true).
				PrefillString("needsEscaping", NeedsEscaping).
				PrefillFloat32("f32", 92.2).
				PrefillFloat64("f64", 292.1).
				PrefillAny("any", map[string]interface{}{"x": "y", "z": 19}).
				PrefillEnum(ExampleMetadataMultipleXEnum, xopconst.SpanKindClient).
				PrefillEmbeddedEnum(LockedEEnumTwo).
				Log()
			log.Warn().Int("int", 439).Int8("int8", 82).Int16("int16", 829).
				Int32("int32", 4328).Int64("int64", -2382).
				Uint("uint", 439).Uint8("uint8", 82).Uint16("uint16", 829).
				Uint32("uint32", 4328).Uint64("uint64", 2382).
				String("foo", "bar").Bool("on/off", true).
				String("needsEscaping2", NeedsEscaping).
				Float32("f32", 92.2).
				Float64("f64", 292.1).
				Any("any", map[string]interface{}{"x": "y", "z": 19}).
				AnyImmutable("anyim", map[string]interface{}{"x": "y", "z": 19}).
				Enum(ExampleMetadataMultipleXEnum, xopconst.SpanKindClient).
				EmbeddedEnum(LockedEEnumTwo).
				Msgs("ha", true)
			p.Error().Static("prefilled!")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "type time",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillTime("-1m", time.Now().Add(-time.Minute).Round(time.Millisecond)).Log()
			p.Warn().Time("now", time.Now().Round(time.Millisecond)).Msgs("time!")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "type duration",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillDuration("1m", time.Minute).Log()
			p.Warn().Duration("hour", time.Hour).Msg("duration")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "type trace",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillLink("me", log.Span().Bundle().Trace).Log()
			p.Warn().Link("me, again", log.Span().Bundle().Trace).Static("trace")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "type error",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			p := log.Sub().PrefillError("question", fmt.Errorf("why would you pre-fill an error?")).Log()
			p.Warn().Error("answer", fmt.Errorf("I don't know, why would you prefill an error")).Msgs(time.Now())
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "log levels",
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			var callCount int
			sc := newStringCounter(&callCount, "foobar")
			skipper := log.Sub().MinLevel(xopnum.InfoLevel).Log()
			skipper.Debug().
				Stringer("avoid", sc).
				String("avoid", "blaf").
				Any("null", nil).
				AnyImmutable("null", nil).
				Error("no", fmt.Errorf("bar")).
				Msg("no foobar")
			log.Trace().Stringer("do", sc).Msg("yes, foobar")
			assert.Equal(t, 1, callCount, "stringer called once")
			MicroNap()
			log.Done()
		},
	},
	{
		Name: "simulate inbound propagation",
		SeedMods: []xop.SeedModifier{
			xop.WithBundle(func() trace.Bundle {
				var bundle trace.Bundle
				bundle.Parent.Flags().SetString("01")
				bundle.Parent.TraceID().SetString("a60a3cc0123a043fee48839c9d52a645")
				bundle.Parent.SpanID().SetString("c63f9d81e2285f34")
				bundle.Trace = bundle.Parent
				bundle.Trace.SpanID().SetRandom()
				bundle.State.SetString("congo=t61rcWkgMzE")
				bundle.Baggage.SetString("userId=alice,serverNode=DF%2028,isProduction=false")
				return bundle
			}()),
		},
		Do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
			assert.Equal(t, "00-a60a3cc0123a043fee48839c9d52a645-c63f9d81e2285f34-01", log.Span().Bundle().Parent.String(), "trace parent")
			assert.Equal(t, "a60a3cc0123a043fee48839c9d52a645", log.Span().Bundle().Trace.GetTraceID().String(), "trace trace")
			assert.NotEqual(t, "c63f9d81e2285f34", log.Span().Bundle().Trace.GetSpanID().String(), "trace trace")
			assert.Equal(t, "congo=t61rcWkgMzE", log.Span().Bundle().State.String(), "trace state")
			assert.Equal(t, "userId=alice,serverNode=DF%2028,isProduction=false", log.Span().Bundle().Baggage.String())
			MicroNap()
			log.Done()
		},
	},
}

type stringCounter struct {
	cp *int
	s  string
}

func (s *stringCounter) String() string {
	*s.cp++
	return s.s
}

func newStringCounter(cp *int, s string) *stringCounter {
	return &stringCounter{
		cp: cp,
		s:  s,
	}
}
