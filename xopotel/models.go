package xopotel

import (
	"context"
	"crypto/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"

	"go.opentelemetry.io/otel/attribute"
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
	bypass          bool
}

type request struct {
	*span
	attributesDefined map[string]struct{}
	lineCount         int32
}

type span struct {
	otelSpan           otelSpanWrap
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

var _ xopbase.Logger = &logger{}
var _ xopbase.Request = &span{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}

var logMessageKey = attribute.Key("xop.message")
var xopSpanSequence = attribute.Key("xop.xopSpanSequence")
var typeKey = attribute.Key("xop.type")
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

// TODO: find a better way to set this version string
const xopVersionValue = "0.3.0"
const xopotelVersionValue = xopVersionValue

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
		spanConfig := spanConfigFromTrace(bundle.Parent)
		addStateToSpanConfig(&spanConfig, bundle.State)
		spanContext := oteltrace.NewSpanContext(spanConfig)
		ctx = oteltrace.ContextWithSpanContext(ctx, spanContext)
	}
	return ctx
}

func addStateToSpanConfig(spanConfig *oteltrace.SpanContextConfig, state xoptrace.State) {
	if !state.IsZero() {
		state, err := oteltrace.ParseTraceState(state.String())
		if err == nil {
			spanConfig.TraceState = state
		}
	}
}
func spanConfigFromTrace(trace xoptrace.Trace) oteltrace.SpanContextConfig {
	return oteltrace.SpanContextConfig{
		TraceID:    trace.TraceID().Array(),
		SpanID:     trace.SpanID().Array(),
		TraceFlags: oteltrace.TraceFlags(trace.Flags().Array()[0]),
	}
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
