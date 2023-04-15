package xopotel

import (
	"context"
	"crypto/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const attributeDefinitionPrefix = "xop.defineKey."
const xopSynthesizedForOTEL = "xopotel-shim type"

type logger struct {
	tracer          oteltrace.Tracer
	id              string
	doLogging       bool
	ignoreDone      oteltrace.Span
	spanFromContext bool
	bufferedRequest *bufferedRequest // only set when BufferedReplayLogger is used
}

type request struct {
	*span
	attributesDefined map[string]struct{}
	lineCount         int32
}

type span struct {
	otelSpan           oteltrace.Span
	logger             *logger
	request            *request
	ctx                context.Context
	lock               sync.Mutex
	priorBoolSlices    map[string][]bool
	priorFloat64Slices map[string][]float64
	priorStringSlices  map[string][]string
	priorInt64Slices   map[string][]int64
	hasPrior           map[string]struct{}
	metadataSeen       map[string]interface{}
	spanPrefill        []attribute.KeyValue // holds spanID & traceID
	isXOP              bool                 // true unless data is imported from OTEL
	XXX                xoptrace.Bundle
}

type prefilling struct {
	builder
}

type prefilled struct {
	builder
}

type line struct {
	builder
	prealloc  [15]attribute.KeyValue
	level     xopnum.Level
	timestamp time.Time
}

type builder struct {
	attributes []attribute.KeyValue
	span       *span
	prefillMsg string
	linkKey    string
	linkValue  xoptrace.Trace
}

type otelStuff struct {
	spanCounters
	Status               sdktrace.Status
	SpanKind             xopconst.SpanKindEnum
	Resource             bufferedResource
	InstrumentationScope instrumentation.Scope
}

type spanCounters struct {
	DroppedAttributes int
	DroppedLinks      int
	DroppedEvents     int
	ChildSpanCount    int
}

var _ xopbase.Logger = &logger{}
var _ xopbase.Request = &span{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}

var xopLevel = attribute.Key("xop.level")
var xopSpanSequence = attribute.Key("xop.xopSpanSequence")
var xopType = attribute.Key("xop.type")
var spanIsLinkAttributeKey = attribute.Key("xop.span.is-link-attribute")
var spanIsLinkEventKey = attribute.Key("xop.span.is-link-event")
var xopVersion = attribute.Key("xop.version")
var xopOTELVersion = attribute.Key("xop.otel-version")
var xopSource = attribute.Key("xop.source")
var xopNamespace = attribute.Key("xop.namespace")
var xopLinkData = attribute.Key("xop.link")
var xopModelType = attribute.Key("xop.modelType")
var xopEncoding = attribute.Key("xop.encoding")
var xopModel = attribute.Key("xop.model")
var xopLineFormat = attribute.Key("xop.format")
var xopTemplate = attribute.Key("xop.template")
var otelSpanKind = attribute.Key("span.kind")
var xopLineNumber = attribute.Key("xop.lineNumber")
var xopBaggage = attribute.Key("xop.baggage")
var xopStackTrace = attribute.Key("xop.stackTrace")

var replayFromOTEL = xopat.Make{Key: "span.replayedFromOTEL", Namespace: "XOP", Indexed: false, Prominence: 300,
	Description: "Data origin is OTEL, translated through xopotel.ExportToXOP, bundle of span config"}.AnyAttribute(&otelStuff{})

// TODO: find a better way to set this version string
const xopVersionValue = "0.3.0"
const xopotelVersionValue = xopVersionValue

const otelDataSource = "source-is-not-xop"

var xopPromotedMetadata = xopat.Make{Key: "xop.span-is-promoted", Namespace: "xopotel"}.BoolAttribute()

var emptyTraceState oteltrace.TraceState

type overrideContextKeyType struct{}

var overrideContextKey = overrideContextKeyType{}

func overrideIntoContext(ctx context.Context, bundle xoptrace.Bundle) context.Context {
	override := &idOverride{
		valid:   1,
		traceID: bundle.Trace.GetTraceID(),
		spanID:  bundle.Trace.GetSpanID(),
	}
	ctx = context.WithValue(ctx, overrideContextKey, override)
	if !bundle.Parent.IsZero() || !bundle.State.IsZero() {
		spanConfig := oteltrace.SpanContextConfig{
			TraceID:    bundle.Parent.TraceID().Array(),
			SpanID:     bundle.Parent.SpanID().Array(),
			TraceFlags: oteltrace.TraceFlags(bundle.Parent.Flags().Array()[0]),
		}
		if !bundle.State.IsZero() {
			state, err := oteltrace.ParseTraceState(bundle.State.String())
			if err == nil {
				spanConfig.TraceState = state
			}
		}
		spanContext := oteltrace.NewSpanContext(spanConfig)
		ctx = oteltrace.ContextWithSpanContext(ctx, spanContext)
	}
	return ctx
}

func overrideFromContext(ctx context.Context) *idOverride {
	v := ctx.Value(overrideContextKey)
	if v == nil {
		return nil
	}
	return v.(*idOverride)
}

type idGenerator struct{}

var _ sdktrace.IDGenerator = idGenerator{}

type idOverride struct {
	valid   int32
	traceID xoptrace.HexBytes16
	spanID  xoptrace.HexBytes8
}

func (o *idOverride) Get() (oteltrace.TraceID, oteltrace.SpanID, bool) {
	if atomic.CompareAndSwapInt32(&o.valid, 1, 0) {
		return o.traceID.Array(), o.spanID.Array(), true
	}
	return random16(), random8(), false
}

var zero8 = oteltrace.SpanID{}
var zero16 = oteltrace.TraceID{}

func random8() oteltrace.SpanID {
	var b oteltrace.SpanID
	for {
		_, _ = rand.Read(b[:])
		if b != zero8 {
			return b
		}
	}
}

func random16() oteltrace.TraceID {
	var b oteltrace.TraceID
	for {
		_, _ = rand.Read(b[:])
		if b != zero16 {
			return b
		}
	}
}

// IDGenerator generates an override that must be used with
// https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#NewTracerProvider
// when creating a TracerProvider.  This override causes the
// TracerProvider to respect the TraceIDs and SpanIDs generated by
// Xop. For accurate replay via xopotel, this is required. For general
// log generation, if this ID generator is not used, then the Span IDs
// and TraceIDs created by by the TracerProvider will be used for Xop
// logging.
func IDGenerator() sdktrace.TracerProviderOption {
	return sdktrace.WithIDGenerator(idGenerator{})
}
