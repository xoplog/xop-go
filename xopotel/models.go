package xopotel

import (
	"context"
	"crypto/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go"
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
	errorReporter     func(error)
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
}

type prefilling struct {
	builderWithSpan
}

type prefilled struct {
	builderWithSpan
}

type line struct {
	builderWithSpan
	prealloc  [15]attribute.KeyValue
	level     xopnum.Level
	timestamp time.Time
}

type builderWithSpan struct {
	span *span
	builder
}

type builder struct {
	attributes []attribute.KeyValue
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
	links                []oteltrace.Link // filled in by getStuff()
}

type spanCounters struct {
	DroppedAttributes int
	DroppedLinks      int
	DroppedEvents     int
	ChildSpanCount    int
}

var _ xopbase.Logger = &logger{}
var _ xopbase.Request = &request{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}

// Span-level
var otelSpanKind = attribute.Key("span.kind")
var spanIsLinkAttributeKey = attribute.Key("xop.span.is-link-attribute")
var spanIsLinkEventKey = attribute.Key("xop.span.is-link-event")
var xopBaggage = attribute.Key("xop.baggage")
var xopNamespace = attribute.Key("xop.namespace")
var xopOTELVersion = attribute.Key("xop.otel-version")
var xopSource = attribute.Key("xop.source")
var xopSpanSequence = attribute.Key("xop.xopSpanSequence")
var xopVersion = attribute.Key("xop.version")

// Line
var xopLevel = attribute.Key("xop.level")
var xopLineNumber = attribute.Key("xop.lineNumber")
var xopStackTrace = attribute.Key("xop.stackTrace")
var xopTemplate = attribute.Key("xop.template")
var xopType = attribute.Key("xop.type")

// Model
var xopEncoding = attribute.Key("xop.encoding")
var xopModel = attribute.Key("xop.model")
var xopModelType = attribute.Key("xop.modelType")

// Link
var xopLinkData = attribute.Key("xop.link")
var otelLink = xopat.Make{Key: "span.otelLinks", Namespace: "XOP", Indexed: false, Prominence: 300,
	Multiple: true, Distinct: true,
	Description: "Data origin is OTEL, span links w/o attributes; links also sent as Link()"}.LinkAttribute()
var xopLinkMetadataKey = attribute.Key("xop.linkMetadataKey")

var xopLinkTraceStateError = xop.Key("xop.linkTraceStateError")
var xopOTELLinkTranceState = xop.Key("xop.otelLinkTraceState")
var xopOTELLinkIsRemote = xop.Key("xop.otelLinkIsRemote")
var xopOTELLinkDetail = xop.Key("xop.otelLinkDetail")
var xopLinkRemoteError = xop.Key("xop.otelLinkRemoteError")
var xopOTELLinkDroppedAttributeCount = xop.Key("xop.otelLinkDroppedAttributeCount")
var xopLinkeDroppedError = xop.Key("xop.otelLinkDroppedError")

var otelReplayStuff = xopat.Make{Key: "span.replayedFromOTEL", Namespace: "XOP", Indexed: false, Prominence: 300,
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
