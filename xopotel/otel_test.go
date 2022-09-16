package xopotel_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

			var otel OTELSpan
			err = json.Unmarshal([]byte(buffer.String()), &otel)
			require.NoError(t, err, "unmarshal to otelspan")
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

func newChecker(t *testing.T, tlog *xoptest.TestLogger) *checker {
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
