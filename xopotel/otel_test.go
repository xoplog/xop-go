package xopotel_test

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopotel"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestASingleLine(t *testing.T) {
	var buffer xoputil.Buffer

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&buffer),
		stdouttrace.WithPrettyPrint(),
	)
	require.NoError(t, err, "exporter")

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	ctx := context.Background()
	defer func() {
		err := tracerProvider.Shutdown(ctx)
		assert.NoError(t, err, "shutdown")
	}()

	tracer := tracerProvider.Tracer("")

	ctx, span := tracer.Start(ctx, "test-span")
	log := xopotel.SpanLog(ctx, "test-span")
	log.Alert().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Done()
	span.End()
	tracerProvider.ForceFlush(context.Background())
	t.Log("logged:", buffer.String())
	assert.NotEmpty(t, buffer.String())
}

func TestSpanLog(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		if mc.SkipOTEL {
			continue
		}
		t.Run(mc.Name, func(t *testing.T) {
			var buffer xoputil.Buffer

			exporter, err := stdouttrace.New(
				stdouttrace.WithWriter(&buffer),
			)
			require.NoError(t, err, "exporter")

			tracerProvider := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exporter),
			)
			ctx, cancel := context.WithCancel(context.Background())
			defer func() {
				err := tracerProvider.Shutdown(context.Background())
				assert.NoError(t, err, "shutdown")
			}()

			tracer := tracerProvider.Tracer("")

			tlog := xoptest.New(t)
			ctx, span := tracer.Start(ctx, mc.Name)
			seed := xopotel.SpanLog(ctx, mc.Name).Span().Seed(xop.WithBase(tlog))
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				seed = seed.Copy(mc.SeedMods...)
			}
			log := seed.SubSpan("adding-tlog")
			mc.Do(t, log, tlog)

			span.End()
			cancel()
			tracerProvider.ForceFlush(context.Background())
			t.Log("logged:", buffer.String())
			assert.NotEmpty(t, buffer.String())

			newChecker(t, tlog, []string{
				span.SpanContext().SpanID().String(),
			}).check(t, buffer.String())
		})
	}
}

func TestBaseLogger(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		if mc.SkipOTEL {
			continue
		}
		t.Run(mc.Name, func(t *testing.T) {
			var buffer xoputil.Buffer

			exporter, err := stdouttrace.New(
				stdouttrace.WithWriter(&buffer),
			)
			require.NoError(t, err, "exporter")

			tracerProvider := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exporter),
			)
			ctx, cancel := context.WithCancel(context.Background())
			defer func() {
				err := tracerProvider.Shutdown(context.Background())
				assert.NoError(t, err, "shutdown")
			}()

			tracer := tracerProvider.Tracer("")

			tlog := xoptest.New(t)
			seed := xop.NewSeed(
				xop.WithBase(tlog),
				xopotel.BaseLogger(ctx, tracer, true),
			)
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				seed = seed.Copy(mc.SeedMods...)
			}
			log := seed.Request(t.Name())
			mc.Do(t, log, tlog)

			cancel()
			tracerProvider.ForceFlush(context.Background())
			t.Log("logged:", buffer.String())
			assert.NotEmpty(t, buffer.String())

			newChecker(t, tlog, nil).check(t, buffer.String())
		})
	}
}

type OTELSpanContext struct {
	TraceID    string
	SpanID     string
	TraceFlags string
	TraceState string
	Remote     bool
}

type OTELEvent struct {
	Name                  string
	Attributes            []OTELAttribute
	DroppedAttributeCount int
	Time                  time.Time
}

type OTELValue struct {
	Type  string
	Value interface{}
}

type OTELAttribute struct {
	Key   string
	Value OTELValue
}

type OTELStatus struct {
	Code        string
	Description string
}

type OTELInstrumentationLibrary struct {
	Name      string
	Version   string
	SchemaURL string
}

type OTELLink struct {
	SpanContext           OTELSpanContext
	Attributes            []OTELAttribute
	DroppedAttributeCount int
}

type OTELSpan struct {
	Name                   string
	SpanContext            *OTELSpanContext
	Parent                 *OTELSpanContext
	SpanKind               int
	StartTime              time.Time
	EndTime                time.Time
	Attributes             []OTELAttribute
	Links                  []OTELLink
	Events                 []OTELEvent
	Status                 OTELStatus
	DroppedAttributes      int
	DroppedEvents          int
	DroppedLinks           int
	ChildSpanCount         int
	Resource               []OTELAttribute
	InstrumentationLibrary OTELInstrumentationLibrary
}

type checker struct {
	tlog             *xoptest.TestLogger
	spansSeen        []bool
	requestsSeen     []bool
	messagesNotSeen  map[string][]int
	spanIndex        map[string]int
	requestIndex     map[string]int
	accumulatedSpans map[string]typedData
	sequencing       map[string]int
	later            []func()
	notInTest        map[string]struct{}
}

const debugTlog = true
const debugTspan = true

func newChecker(t *testing.T, tlog *xoptest.TestLogger, spansNotIntTest []string) *checker {
	c := &checker{
		tlog:             tlog,
		spansSeen:        make([]bool, len(tlog.Spans)),
		requestsSeen:     make([]bool, len(tlog.Requests)),
		messagesNotSeen:  make(map[string][]int),
		spanIndex:        make(map[string]int),
		requestIndex:     make(map[string]int),
		accumulatedSpans: make(map[string]typedData),
		sequencing:       make(map[string]int),
		notInTest:        make(map[string]struct{}),
	}
	for _, spanID := range spansNotIntTest {
		c.notInTest[spanID] = struct{}{}
	}
	for i, line := range tlog.Lines {
		if debugTlog {
			t.Logf("recorded line: '%s'", line.Message)
		}
		c.messagesNotSeen[line.Message] = append(c.messagesNotSeen[line.Message], i)
	}
	for i, span := range tlog.Spans {
		if debugTspan {
			t.Logf("recorded span: %s - %s", span.Bundle.Trace.SpanID().String(), span.Name)
		}
		_, ok := c.spanIndex[span.Bundle.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate span id %s", span.Bundle.Trace.SpanID().String())
		c.spanIndex[span.Bundle.Trace.SpanID().String()] = i
	}
	for i, request := range tlog.Requests {
		if debugTspan {
			t.Logf("recorded request: %s - %s", request.Bundle.Trace.SpanID().String(), request.Name)
		}
		_, ok := c.spanIndex[request.Bundle.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate span/request id %s", request.Bundle.Trace.SpanID().String())
		_, ok = c.requestIndex[request.Bundle.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate request id %s", request.Bundle.Trace.SpanID().String())
		c.requestIndex[request.Bundle.Trace.SpanID().String()] = i
	}
	return c
}

func (c *checker) check(t *testing.T, data string) {
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}

		var otel OTELSpan
		err := json.Unmarshal([]byte(line), &otel)
		require.NoError(t, err, "unmarshal to otelspan")

		c.span(t, otel)
	}
	for _, f := range c.later {
		f()
	}
	for _, ia := range c.messagesNotSeen {
		for _, li := range ia {
			line := c.tlog.Lines[li]
			t.Errorf("line '%s' not found in OTEL output", line.Message)
		}
	}
	for _, span := range c.tlog.Spans {
		spanAttributes := c.accumulatedSpans[span.Bundle.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes.data) != 0 {
			compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes.data, "xoptel.span.generic", true)
		}
	}
	for _, span := range c.tlog.Requests {
		spanAttributes := c.accumulatedSpans[span.Bundle.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes.data) != 0 {
			compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes.data, "xopotel.span.generic", true)
		}
	}
}

func (c *checker) span(t *testing.T, span OTELSpan) {
	t.Logf("checking span %s (%s) with %d links and %d events",
		span.Name, span.SpanContext.SpanID, len(span.Links), len(span.Events))
	assert.NotEmpty(t, span.Name, "span name")
	assert.False(t, span.StartTime.IsZero(), "start time set")
	assert.False(t, span.EndTime.IsZero(), "end time set")
	assert.NotEmpty(t, span.SpanContext.TraceID, "trace id")
	assert.NotEmpty(t, span.SpanContext.SpanID, "span id")
	c.accumulatedSpans[span.SpanContext.SpanID] = toData(span.Attributes)

	mustFind := true
	if len(span.Attributes) == 1 {
		switch span.Attributes[0].Key {
		case "span.is-link-event":
			// span is just a fake link attribute and the entire span
			// should be considered an event.
			if assert.Equal(t, 1, len(span.Events), "link-event span event count") {
				c.line(t, span.Events[0], &span)
			}
			return
		case "span.is-link-attribute":
			// span is just additional metadata on it's parent span
			c.later = append(c.later, func() {
				addLink(t, c.accumulatedSpans[span.Parent.SpanID], &span)
			})
			mustFind = false
		}
	}

	if mustFind {
		_, ok1 := c.spanIndex[span.SpanContext.SpanID]
		_, ok2 := c.requestIndex[span.SpanContext.SpanID]
		_, ok3 := c.notInTest[span.SpanContext.SpanID]
		assert.Truef(t, ok1 || ok2 || ok3, "span %s (%s) also exists in tlog", span.Name, span.SpanContext.SpanID)
	}

	for _, line := range span.Events {
		c.line(t, line, nil)
	}
}

func (c *checker) line(t *testing.T, line OTELEvent, linkSpan *OTELSpan) {
	ld := toData(line.Attributes)
	addLink(t, ld, linkSpan)
	msgI, ok := ld.data["log.message"]
	if !assert.True(t, ok, "line has log.message attribute") {
		return
	}
	msg := msgI.(string)
	delete(ld.data, "log.message")
	lineIndicies := c.messagesNotSeen[msg]
	if !assert.Equalf(t, 1, len(lineIndicies), "count lines with msg '%s'", msg) {
		return
	}
	delete(c.messagesNotSeen, msg)
	testLine := c.tlog.Lines[lineIndicies[0]]

	_, err := xopnum.LevelString(line.Name)
	assert.NoErrorf(t, err, "event 'name' is a valid log level: %s", line.Name)

	compareData(t, testLine.Data, testLine.DataType, "xoptest", ld.data, "xopotel", false)
}

func addLink(t *testing.T, td typedData, span *OTELSpan) {
	if span == nil {
		return
	}
	if assert.Equalf(t, 1, len(span.Links), "count of metadata link in span %s", span.SpanContext.SpanID) {
		td.data[span.Name] = []string{
			"link",
			span.Links[0].SpanContext.TraceID,
			span.Links[0].SpanContext.SpanID,
		}
		td.types[span.Name] = xopbase.LinkDataType
	}
}

type typedData struct {
	data  map[string]interface{}
	types map[string]xopbase.DataType
}

func toData(attributes []OTELAttribute) typedData {
	td := typedData{
		data:  make(map[string]interface{}),
		types: make(map[string]xopbase.DataType),
	}
	for _, a := range attributes {
		var dt xopbase.DataType
		switch a.Value.Type {
		case "BOOL":
			dt = xopbase.BoolDataType
		case "INT64":
			dt = xopbase.Int64DataType
		case "FLOAT64":
			dt = xopbase.Float64DataType
		case "STRING":
			dt = xopbase.StringDataType
		case "BOOLSLICE", "INT64SLICE", "FLOAT64SLICE", "STRINGSLICE":
			// TODO add these as Xop base types?
			fallthrough
		default:
			dt = xopbase.AnyDataType
		}
		td.data[a.Key] = a.Value.Value
		td.types[a.Key] = dt
	}
	return td
}

/*
func combineAttributes(from map[string]interface{}, attributes map[string]interface{}) {
	for k, v := range from {
		attributes[k] = v
	}
}
*/

var xoptestConvert map[xopbase.DataType]func(interface{}) interface{}

func init() {
	xoptestConvert = map[xopbase.DataType]func(interface{}) interface{}{
		xopbase.DurationDataType: func(generic interface{}) interface{} {
			return generic.(time.Duration).String()
		},
		xopbase.LinkDataType: func(generic interface{}) interface{} {
			link := generic.(xoptrace.Trace)
			return []string{"link", link.TraceID().String(), link.SpanID().String()}
		},
		xopbase.Uint64DataType: func(generic interface{}) interface{} {
			return strconv.FormatUint(generic.(uint64), 10)
		},
		xopbase.AnyDataType: func(generic interface{}) interface{} {
			enc, err := json.Marshal(generic)
			if err != nil {
				return err.Error()
			} else {
				return string(enc)
			}
		},
		xopbase.DurationArrayDataType: genArrayConvert(xopbase.DurationDataType),
		xopbase.AnyArrayDataType:      genArrayConvert(xopbase.AnyDataType),
		xopbase.LinkArrayDataType:     genArrayConvert(xopbase.LinkDataType),
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
	delete(b, "exception.stacktrace")
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
