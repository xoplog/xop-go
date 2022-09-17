package xopotel_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopotel"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoptest/xoptestutil"
	"github.com/muir/xop-go/xoputil"

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
			ctx := context.Background()
			defer func() {
				err := tracerProvider.Shutdown(ctx)
				assert.NoError(t, err, "shutdown")
			}()

			tracer := tracerProvider.Tracer("")

			ctx, span := tracer.Start(ctx, mc.Name)
			log := xopotel.SpanLog(ctx, mc.Name)
			tlog := xoptest.New(t)
			mc.Do(t, log, tlog)

			span.End()
			tracerProvider.ForceFlush(context.Background())
			t.Log("logged:", buffer.String())
			assert.NotEmpty(t, buffer.String())
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
	Name        string
	SpanContext *OTELSpanContext
	Parent      *OTELSpanContext
	SpanKind    int
	StartTime   time.Time
	EndTime     time.Time
	Attributes  []OTELAttribute
	Links       []OTELLink
	// TODO: Links
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
	accumulatedSpans map[string]map[string]interface{}
	sequencing       map[string]int
}

const debugTlog = true
const debugTspan = true

func newChecker(t *testing.T, tlog *xoptest.TestLogger) *checker {
	c := &checker{
		tlog:             tlog,
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
	/*
		for _, ia := range c.messagesNotSeen {
			for _, li := range ia {
				line := c.tlog.Lines[li]
				t.Errorf("line '%s' not found in JSON output", line.Text)
			}
		}
		for _, span := range c.tlog.Spans {
			spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
			if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
				// compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
		for _, span := range c.tlog.Requests {
			spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
			if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
					// compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
	*/
}

func (c *checker) span(t *testing.T, otel OTELSpan) {
	assert.NotEmpty(t, otel.Name, "span name")
	assert.False(t, otel.StartTime.IsZero(), "start time set")
	assert.False(t, otel.EndTime.IsZero(), "end time set")
	assert.NotEmpty(t, otel.SpanContext.TraceID, "trace id")
	assert.NotEmpty(t, otel.SpanContext.SpanID, "span id")
	c.accumulatedSpans[otel.SpanContext.SpanID] = toData(otel.Attributes)
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
		case "STRING":
			dt = xopbase.StringDataType
		default:
			dt = xopbase.AnyDataType
		}
		td.data[a.Key] = a.Value.Value
		td.types[a.Key] = dt
	}
}

/*
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
*/
