package xopjson_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopjson"
	"github.com/muir/xop-go/xopnum"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoptest/xoptestutil"
	"github.com/muir/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	debugTlog     = true
	debugTspan    = true
	needsEscaping = `"\<'` + "\n\r\t\b\x00"
)

type supersetObject struct {
	// lines, spans, and requests

	Timestamp  xoptestutil.TS         `json:"ts"`
	Attributes map[string]interface{} `json:"attributes"`

	// lines

	Level  int      `json:"lvl"`
	SpanID string   `json:"span.id"`
	Stack  []string `json:"stack"`
	Msg    string   `json:"msg"`
	Format string   `json:"fmt"`

	// requests & spans

	Type     string `json:"type"`
	Name     string `json:"name"`
	Duration int64  `json:"dur"`

	// requests

	Implmentation  string `json:"impl"`
	TraceID        string `json:"trace.id"`
	ParentID       string `json:"parent.id"`
	RequestID      string `json:"request.id"`
	State          string `json:"trace.state"`
	Baggage        string `json:"trace.baggage"`
	RequestVersion int    `json:"request.ver"` // TODO: change to span.ver?

	// spans

	SpanVersion int `json:"span.ver"`
}

type checkConfig struct {
	minVersions         int
	maxVersions         int
	hasAttributesObject bool
}

type checker struct {
	tlog             *xoptest.TestLogger
	config           checkConfig
	spansSeen        []bool
	requestsSeen     []bool
	messagesNotSeen  map[string][]int
	spanIndex        map[string]int
	requestIndex     map[string]int
	accumulatedSpans map[string]map[string]interface{}
	sequencing       map[string]int
}

func TestASingleLine(t *testing.T) {
	var buffer xoputil.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithEpochTime(time.Microsecond),
		xopjson.WithDuration("dur", xopjson.AsString),
		xopjson.WithSpanTags(xopjson.SpanIDTagOption),
		xopjson.WithAttributesObject(true),
		xopjson.WithStackLineRewrite(func(s string) string {
			return "FOO-" + s
		}),
	)
	log := xop.NewSeed(xop.WithBase(jlog)).Request(t.Name())
	log.Alert().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Done()
	s := buffer.String()
	t.Log(s)
	lines := strings.Split(buffer.String(), "\n")
	require.Equal(t, 3, len(lines), "three lines")
	assert.Contains(t, lines[0], `"span.id":`)
	assert.Contains(t, lines[0], `"attributes":{`) // }
	assert.Contains(t, lines[0], `"foo":"bar"`)
	assert.Contains(t, lines[0], `"lvl":20`)
	assert.Contains(t, lines[0], `"ts":`)
	assert.Contains(t, lines[0], `"blast":99`)
	assert.Contains(t, lines[0], `"stack":["FOO-`)
	assert.NotContains(t, lines[0], `"trace.id":`)
	assert.NotContains(t, lines[1], `"stack":[`)
	assert.Contains(t, lines[1], `"span.id":`)
	assert.Contains(t, lines[1], `"dur":"`)
	assert.Contains(t, lines[1], `"request.ver":0`)
	assert.Contains(t, lines[1], `"type":"request"`)
	assert.Contains(t, lines[1], `"span.name":"TestASingleLine"`)
}

func TestParameters(t *testing.T) {
	jsonCases := []struct {
		name         string
		joptions     []xopjson.Option
		settings     func(settings *xop.LogSettings)
		waitForFlush bool
		checkConfig  checkConfig
		extraFlushes int
	}{
		{
			name: "buffered",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(true),
				xopjson.WithBufferedLines(8 * 1024 * 1024),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			},
			checkConfig: checkConfig{
				minVersions:         2,
				hasAttributesObject: true,
			},
		},
		{
			name: "unbuffered/no-attributes",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(true),
				xopjson.WithBufferedLines(8 * 1024 * 1024),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
				xopjson.WithAttributesObject(false),
			},
			checkConfig: checkConfig{
				minVersions:         2,
				hasAttributesObject: false,
			},
		},
		{
			name: "unsynced",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(false),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			},
			settings: func(settings *xop.LogSettings) {
				settings.SynchronousFlush(false)
			},
			// with sync=false, we don't know when .Done will trigger a flush.
			waitForFlush: true,
			checkConfig: checkConfig{
				minVersions:         1,
				hasAttributesObject: true,
			},
		},
	}

	messageCases := []struct {
		name         string
		extraFlushes int
		do           func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger)
	}{
		{
			name: "one span",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				log.Info().Msg("basic info message")
				log.Error().Msg("basic error message")
				log.Alert().Msg("basic alert message")
				log.Debug().Msg("basic debug message")
				log.Trace().Msg("basic trace message")
				log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")

				ss := log.Sub().Detach().Fork("a fork")
				xoptestutil.MicroNap()
				ss.Alert().String("frightening", "stuff").Static("like a rock" + needsEscaping)
				ss.Span().String(xopconst.EndpointRoute, "/some/thing")

				xoptestutil.MicroNap()
				tlog.CustomEvent("before log.Done")
				log.Done()
				tlog.CustomEvent("after log.Done")
				ss.Debug().Msg("sub-span debug message")
				xoptestutil.MicroNap()
				tlog.CustomEvent("before ss.Done")
				ss.Done()
				tlog.CustomEvent("after ss.Done")
			},
		},
		{
			name: "metadata singles in request",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				log.Span().Bool(ExampleMetadataSingleBool, false)
				log.Span().Bool(ExampleMetadataSingleBool, true)
				log.Span().Bool(ExampleMetadataLockedBool, true)
				log.Span().Bool(ExampleMetadataLockedBool, false)
				log.Span().String(ExampleMetadataLockedString, "loki"+needsEscaping)
				log.Span().String(ExampleMetadataLockedString, "thor"+needsEscaping)
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata traces",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata float64",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata time",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata singles in span",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				ss.Span().String(ExampleMetadataSingleString, "athena"+needsEscaping)
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata any",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata iota enum",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata embedded enum",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata enum",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata multiples",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				// ss := log.Sub().Fork("a fork")
				log.Span().Bool(ExampleMetadataMultipleBool, true)
				log.Span().Bool(ExampleMetadataMultipleBool, true)
				log.Span().Int(ExampleMetadataMultipleInt, 3)
				log.Span().Int(ExampleMetadataMultipleInt, 5)
				log.Span().Int(ExampleMetadataMultipleInt, 7)
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "metadata distinct",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				// ss := log.Sub().Fork("a fork")
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
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "one done",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				_ = log.Sub().Fork("a fork")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "prefill",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillFloat64("f", 23).PrefillText("pre!").Log()
				p.Error().Int16("i16", int16(7)).Msg("pf")
				log.Alert().Int32("i32", int32(77)).Msgf("pf %s", "bar")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name:         "add/remove loggers with a seed",
			extraFlushes: 2,
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				tlog2 := xoptest.New(t)
				r2 := log.Span().Seed(xop.WithBase(tlog2)).Request("R2")
				r3 := r2.Span().Seed(xop.WithoutBase(tlog2)).Request("R3")
				r2.Info().Static("log to both test loggers")
				r3.Info().Static("log to just the original set")
				xoptestutil.MicroNap()
				log.Done()
				r2.Done()
				r3.Done()
			},
		},
		{
			name: "add/remove loggers with a span",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				tlog2 := xoptest.New(t)
				s2 := log.Sub().Step("S2", xop.WithBase(tlog2))
				s3 := s2.Sub().Detach().Fork("S3", xop.WithoutBase(tlog2))
				s2.Info().Static("log to both test loggers")
				s3.Info().Static("log to just the original set")
				xoptestutil.MicroNap()
				s2.Done()
				s3.Done()
				log.Done()
			},
		},
		{
			name:         "log after done",
			extraFlushes: 1,
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				s2 := log.Sub().Step("S2")
				s2.Info().Int8("i8", 9).Msg("a line before done")
				xoptestutil.MicroNap()
				s2.Done()
				assert.Empty(t, tlog.FindLines(xoptest.TextContains("XOP: log was already done, but was used again")), "no err")
				s2.Info().Int16("i16", 940).Msg("a post-done line, should trigger an error log")
				assert.NotEmpty(t, tlog.FindLines(xoptest.TextContains("XOP: log was already done, but was used again")), "no err")
				assert.Empty(t, tlog.FindLines(xoptest.TextContains("called on log object when it was already Done")), "no err")
				xoptestutil.MicroNap()
				s2.Done()
				assert.NotEmpty(t, tlog.FindLines(xoptest.TextContains("called on log object when it was already Done")), "now err")
				log.Flush()
				s2.Warn().Int32("i32", 940940).Msg("another post-done line, should trigger an error log")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "lots of types",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillInt("int", 439).PrefillInt8("int8", 82).PrefillInt16("int16", 829).
					PrefillInt32("int32", 4328).PrefillInt64("int64", -2382).
					PrefillUint("uint", 439).PrefillUint8("uint8", 82).PrefillUint16("uint16", 829).
					PrefillUint32("uint32", 4328).PrefillUint64("uint64", 2382).
					PrefillString("foo", "bar").PrefillBool("on/off", true).
					PrefillString("needsEscaping", needsEscaping).
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
					String("needsEscaping2", needsEscaping).
					Float32("f32", 92.2).
					Float64("f64", 292.1).
					Any("any", map[string]interface{}{"x": "y", "z": 19}).
					AnyImmutable("anyim", map[string]interface{}{"x": "y", "z": 19}).
					Enum(ExampleMetadataMultipleXEnum, xopconst.SpanKindClient).
					EmbeddedEnum(LockedEEnumTwo).
					Msgs("ha", true)
				p.Error().Static("prefilled!")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "type time",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillTime("-1m", time.Now().Add(-time.Minute).Round(time.Millisecond)).Log()
				p.Warn().Time("now", time.Now().Round(time.Millisecond)).Msgs("time!")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "type duration",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillDuration("1m", time.Minute).Log()
				p.Warn().Duration("hour", time.Hour).Msg("duration")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "type trace",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillLink("me", log.Span().Bundle().Trace).Log()
				p.Warn().Link("me, again", log.Span().Bundle().Trace).Static("trace")
				xoptestutil.MicroNap()
				log.Done()
			},
		},
		{
			name: "type error",
			do: func(t *testing.T, log *xop.Log, tlog *xoptest.TestLogger) {
				p := log.Sub().PrefillError("question", fmt.Errorf("why would you pre-fill an error?")).Log()
				p.Warn().Error("answer", fmt.Errorf("I don't know, why would you prefill an error")).Msgs(time.Now())
				xoptestutil.MicroNap()
				log.Done()
			},
		},
	}

	for _, tc := range jsonCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, mc := range messageCases {
				mc := mc
				t.Run(mc.name, func(t *testing.T) {
					var buffer xoputil.Buffer
					joptions := []xopjson.Option{
						xopjson.WithTimeFormat(time.RFC3339Nano),
						xopjson.WithDuration("dur", xopjson.AsNanos),
						xopjson.WithSpanStarts(true),
						xopjson.WithBufferedLines(8 * 1024 * 1024),
						xopjson.WithAttributesObject(true),
					}
					joptions = append(joptions, tc.joptions...)

					jlog := xopjson.New(
						xopbytes.WriteToIOWriter(&buffer),
						joptions...)
					tlog := xoptest.New(t)
					settings := func(settings *xop.LogSettings) {
						settings.SynchronousFlush(true)
					}
					if tc.settings != nil {
						settings = tc.settings
					}
					log := xop.NewSeed(
						xop.WithBase(jlog),
						xop.WithBase(tlog),
						xop.WithSettings(settings),
					).Request(t.Name())

					mc.do(t, log, tlog)

					expectedFlushes := 1 + tc.extraFlushes + mc.extraFlushes
					if tc.waitForFlush {
						assert.Eventually(t, func() bool {
							return xoptestutil.EventCount(tlog, xoptest.FlushEvent) >= expectedFlushes
						}, time.Second, time.Millisecond*3)
					}
					t.Log("\n", buffer.String())
					xoptestutil.DumpEvents(t, tlog)
					assert.Equal(t, expectedFlushes, xoptestutil.EventCount(tlog, xoptest.FlushEvent), "count of flush")
					newChecker(t, tlog, tc.checkConfig).check(t, buffer.String())
				})
			}
		})
	}
}

func newChecker(t *testing.T, tlog *xoptest.TestLogger, config checkConfig) *checker {
	if config.maxVersions < config.minVersions {
		config.maxVersions = config.minVersions
	}
	c := &checker{
		tlog:             tlog,
		config:           config,
		spansSeen:        make([]bool, len(tlog.Spans)),
		requestsSeen:     make([]bool, len(tlog.Requests)),
		messagesNotSeen:  make(map[string][]int),
		spanIndex:        make(map[string]int),
		requestIndex:     make(map[string]int),
		accumulatedSpans: make(map[string]map[string]interface{}),
		sequencing:       make(map[string]int),
	}
	for i, line := range tlog.Lines {
		if debugTlog {
			t.Logf("recorded line: '%s'", line.Message)
		}
		c.messagesNotSeen[line.Message] = append(c.messagesNotSeen[line.Message], i)
	}
	for i, span := range tlog.Spans {
		if debugTspan {
			t.Logf("recorded span: %s - %s", span.Trace.Trace.SpanIDString(), span.Name)
		}
		_, ok := c.spanIndex[span.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate span id %s", span.Trace.Trace.SpanIDString())
		c.spanIndex[span.Trace.Trace.SpanIDString()] = i
	}
	for i, request := range tlog.Requests {
		if debugTspan {
			t.Logf("recorded request: %s - %s", request.Trace.Trace.SpanIDString(), request.Name)
		}
		_, ok := c.spanIndex[request.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate span/request id %s", request.Trace.Trace.SpanIDString())
		_, ok = c.requestIndex[request.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate request id %s", request.Trace.Trace.SpanIDString())
		c.requestIndex[request.Trace.Trace.SpanIDString()] = i
	}
	for spanID, versions := range c.sequencing {
		if c.config.minVersions == c.config.maxVersions {
			assert.Equal(t, versions, c.config.minVersions, "version count for span %s", spanID)
		} else {
			assert.GreaterOrEqualf(t, versions, c.config.minVersions, "version count for span %s", spanID)
			assert.LessOrEqualf(t, versions, c.config.maxVersions, "version count for span %s", spanID)
		}
	}
	return c
}

func (c *checker) check(t *testing.T, data string) {
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		var generic map[string]interface{}
		err := json.Unmarshal([]byte(line), &generic)
		require.NoErrorf(t, err, "decode to generic '%s'", line)

		var super supersetObject
		err = json.Unmarshal([]byte(line), &super)
		require.NoErrorf(t, err, "decode to super: %s", line)

		switch super.Type {
		case "", "line":
			t.Logf("check line: %s", line)
			c.line(t, super, generic)
		case "span":
			t.Logf("check span: %s", line)
			c.span(t, super, generic)
		case "request":
			t.Logf("check request: %s", line)
			c.request(t, super, generic)
		}
	}
	for _, ia := range c.messagesNotSeen {
		for _, li := range ia {
			line := c.tlog.Lines[li]
			t.Errorf("line '%s' not found in JSON output", line.Text)
		}
	}
	for _, span := range c.tlog.Spans {
		spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
			if c.config.hasAttributesObject {
				t.Logf("comparing metadata: %+v vs %+v", span.Metadata, spanAttributes)
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.attributes", false)
			} else {
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
	}
	for _, span := range c.tlog.Requests {
		spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
			if c.config.hasAttributesObject {
				t.Logf("comparing metadata: %+v vs %+v", span.Metadata, spanAttributes)
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.request.attributes", false)
			} else {
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
	}
}

func (c *checker) line(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.NotEqual(t, xopnum.Level(0), super.Level, "level")
	assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
	assert.NotEmpty(t, super.Msg, "message")
	mns := c.messagesNotSeen[super.Msg]
	if !assert.NotNilf(t, mns, "test line with message '%s'", super.Msg) {
		return
	}
	line := c.tlog.Lines[mns[0]]
	c.messagesNotSeen[super.Msg] = c.messagesNotSeen[super.Msg][1:]
	assert.Truef(t, super.Timestamp.Round(time.Millisecond).Equal(line.Timestamp.Round(time.Millisecond)), "timestamps %s vs %s", line.Timestamp, super.Timestamp)
	assert.Equal(t, int(line.Level), super.Level, "level")
	if c.config.hasAttributesObject {
		compareData(t, line.Data, line.DataType, "xoptest.Data", super.Attributes, "xopjson.Attributes", false)
	} else {
		assert.Empty(t, super.Attributes)
		compareData(t, line.Data, line.DataType, "xoptest.Data", generic, "xopjson.Generic", true)
	}
}

func (c *checker) span(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.Empty(t, super.Level, "no level expected")
	var prior int
	var ok bool
	if assert.NotEmpty(t, super.SpanID, "has span id") {
		prior, ok = c.sequencing[super.SpanID]
	}
	if super.SpanVersion > 0 {
		if assert.True(t, ok, "has prior version") {
			assert.Equal(t, prior+1, super.SpanVersion, "version is in sequence")
		}
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
	} else {
		assert.False(t, ok, "no prior version expected")
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	if c.config.hasAttributesObject {
		combineAttributes(super.Attributes, c.accumulatedSpans[super.SpanID])
	} else {
		combineAttributes(generic, c.accumulatedSpans[super.SpanID])
	}
	c.sequencing[super.SpanID] = super.SpanVersion
	assert.Less(t, super.Duration, int64(time.Second*10), "duration")
}

func (c *checker) request(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.Empty(t, super.Level, "no level expected")
	var prior int
	var ok bool
	if assert.NotEmpty(t, super.SpanID, "has span id") {
		prior, ok = c.sequencing[super.SpanID]
	}
	if super.RequestVersion > 0 {
		if assert.True(t, ok, "has prior version") {
			assert.Equal(t, prior+1, super.RequestVersion, "version is in sequence")
		}
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
	} else {
		assert.False(t, ok, "no prior version expected")
		assert.NotEmpty(t, super.TraceID, "has trace id")
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	if c.config.hasAttributesObject {
		combineAttributes(super.Attributes, c.accumulatedSpans[super.SpanID])
	} else {
		combineAttributes(generic, c.accumulatedSpans[super.SpanID])
	}
	c.sequencing[super.SpanID] = super.RequestVersion
	assert.Less(t, super.Duration, int64(time.Second*10), "duration")
}

func combineAttributes(from map[string]interface{}, attributes map[string]interface{}) {
	for k, v := range from {
		attributes[k] = v
	}
}

var xoptestConvert map[xopbase.DataType]func(interface{}) interface{}

func init() {
	xoptestConvert = map[xopbase.DataType]func(interface{}) interface{}{
		xopbase.ErrorDataType: func(generic interface{}) interface{} {
			return generic.(error).Error()
		},
		xopbase.LinkDataType: func(generic interface{}) interface{} {
			return map[string]interface{}{
				"xop.link": generic.(trace.Trace).String(),
			}
		},
		xopbase.LinkArrayDataType:  genArrayConvert(xopbase.LinkDataType),
		xopbase.ErrorArrayDataType: genArrayConvert(xopbase.ErrorDataType),
	}
}

func genArrayConvert(edt xopbase.DataType) func(interface{}) interface{} {
	return func(generic interface{}) interface{} {
		var sl []interface{}
		for _, element := range generic.([]interface{}) {
			sl = append(sl, xoptestConvert[edt](element))
		}
		return sl
	}
}

func compareData(t *testing.T, aOrig map[string]interface{}, types map[string]xopbase.DataType, aDesc string, b map[string]interface{}, bDesc string, ignoreExtra bool) {
	if len(aOrig) == 0 && len(b) == 0 {
		return
	}
	a := make(map[string]interface{})
	for k, v := range aOrig {
		if f, ok := xoptestConvert[types[k]]; ok {
			a[k] = f(v)
		} else {
			a[k] = v
		}
	}
	if ignoreExtra {
		tmp := make(map[string]interface{})
		for k := range a {
			if v, ok := b[k]; ok {
				tmp[k] = v
			}
		}
		b = tmp
	}
	if len(a) == 0 && len(b) == 0 {
		return
	}
	aEnc, err := json.Marshal(a)
	if !assert.NoErrorf(t, err, "marshal %s", aDesc) {
		return
	}
	var aRedone map[string]interface{}
	if !assert.NoErrorf(t, json.Unmarshal(aEnc, &aRedone), "remarshal %s", aDesc) {
		return
	}
	bEnc, err := json.Marshal(b)
	if !assert.NoErrorf(t, err, "marshal %s", bDesc) {
		return
	}
	var bRedone map[string]interface{}
	if !assert.NoErrorf(t, json.Unmarshal(bEnc, &bRedone), "remarshal %s", bDesc) {
		return
	}
	assert.Equalf(t, aRedone, bRedone, "%s vs %s", aDesc, bDesc)
}
