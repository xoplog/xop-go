// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func xopTraceFromSpan(span oteltrace.Span) xoptrace.Trace {
	var xoptrace xoptrace.Trace
	sc := span.SpanContext()
	xoptrace.TraceID().SetArray(sc.TraceID())
	xoptrace.SpanID().SetArray(sc.SpanID())
	xoptrace.Flags().SetArray([1]byte{byte(sc.TraceFlags())})
	// xoptrace.Version().SetArray([1]byte{0})
	return xoptrace
}

// SpanToLog allows xop to add logs to an existing OTEL span.  log.Done() will be
// ignored for this span.
func SpanToLog(ctx context.Context, name string, extraModifiers ...xop.SeedModifier) *xop.Logger {
	span := oteltrace.SpanFromContext(ctx)
	xoptrace := xopTraceFromSpan(span)
	tracer := span.TracerProvider().Tracer("xoputil")
	log := xop.NewSeed(makeSeedModifier(ctx, tracer,
		xop.WithTrace(xoptrace),
		xop.WithBase(&logger{
			id:         "otel-" + uuid.New().String(),
			doLogging:  true,
			ignoreDone: span,
			tracer:     tracer,
		}),
	)).SubSpan(name)
	go func() {
		<-ctx.Done()
		log.Done()
	}()
	return log
}

func (_ idGenerator) NewIDs(ctx context.Context) (oteltrace.TraceID, oteltrace.SpanID) {
	override := overrideFromContext(ctx)
	traceID, spanID, _ := override.Get()
	return traceID, spanID
}

func (_ idGenerator) NewSpanID(ctx context.Context, _ oteltrace.TraceID) oteltrace.SpanID {
	override := overrideFromContext(ctx)
	_, spanID, _ := override.Get()
	return spanID
}

// SeedModifier provides a xop.SeedModifier to set up an OTEL Tracer as a xopbase.Logger
// so that xop logs are output through the OTEL Tracer.
//
// As of the writing of this comment, the Open Telemetry Go library does not support
// logging so to use it for logging purposes, log lines are sent as span "Events".
//
// The recommended way to create a TracerProvider includes using WithBatcher to
// control the flow of data to SpanExporters.  The default configuration for the Batcher
// limits spans to 128 Events each. It imposes other limits too but the default event
// limit is the one that is likely to be hit with even modest usage.
//
// Using SeedModifier, the TraceProvider does not have to have been created using IDGenerator().
func SeedModifier(ctx context.Context, traceProvider oteltrace.TracerProvider) xop.SeedModifier {
	tracer := traceProvider.Tracer("xopotel",
		oteltrace.WithInstrumentationAttributes(
			xopOTELVersion.String(xopotelVersionValue),
			xopVersion.String(xopVersionValue),
		),
		oteltrace.WithInstrumentationVersion(xopotelVersionValue),
	)
	return makeSeedModifier(ctx, tracer)
}

func makeSeedModifier(ctx context.Context, tracer oteltrace.Tracer, extraModifiers ...xop.SeedModifier) xop.SeedModifier {
	modifiers := []xop.SeedModifier{
		xop.WithBase(&logger{
			id:              "otel-" + uuid.New().String(),
			doLogging:       true,
			tracer:          tracer,
			spanFromContext: true,
		}),
		xop.WithContext(ctx),
		xop.WithReactive(func(ctx context.Context, seed xop.Seed, nameOrDescription string, isChildSpan bool, ts time.Time) []xop.SeedModifier {
			if isChildSpan {
				ctx, span := buildSpan(ctx, ts, seed.Bundle(), nameOrDescription, tracer, nil)
				return []xop.SeedModifier{
					xop.WithContext(ctx),
					xop.WithSpan(span.SpanContext().SpanID()),
				}
			}
			parentCtx := ctx
			bundle := seed.Bundle()
			ctx, _, otelSpan := buildRequestSpan(ctx, ts, bundle, nameOrDescription, seed.SourceInfo(), tracer, nil)
			if bundle.Parent.IsZero() {
				parentSpan := oteltrace.SpanFromContext(parentCtx)
				if parentSpan.SpanContext().HasTraceID() {
					bundle.Parent.Flags().SetArray([1]byte{byte(parentSpan.SpanContext().TraceFlags())})
					bundle.Parent.TraceID().SetArray(parentSpan.SpanContext().TraceID())
					bundle.Parent.SpanID().SetArray(parentSpan.SpanContext().SpanID())
				}
				bundle.State.SetString(otelSpan.SpanContext().TraceState().String())
				bundle.Trace.Flags().SetArray([1]byte{byte(otelSpan.SpanContext().TraceFlags())})
				bundle.Trace.TraceID().SetArray(otelSpan.SpanContext().TraceID())
				bundle.Trace.SpanID().SetArray(otelSpan.SpanContext().SpanID())
			}
			bundle.Trace.SpanID().SetArray(otelSpan.SpanContext().SpanID())
			return []xop.SeedModifier{
				xop.WithContext(ctx),
				xop.WithBundle(bundle),
			}
		}),
	}
	return xop.CombineSeedModifiers(append(modifiers, extraModifiers...)...)
}

// BaseLogger provides SeedModifiers to set up an OTEL Tracer as a xopbase.Logger
// so that xop logs are output through the OTEL Tracer.
//
// As of the writing of this comment, the Open Telemetry Go library does not support
// logging so to use it for logging purposes, log lines are sent as span "Events".
//
// The recommended way to create a TracerProvider includes using WithBatcher to
// control the flow of data to SpanExporters.  The default configuration for the Batcher
// limits spans to 128 Events each. It imposes other limits too but the default event
// limit is the one that is likely to be hit with even modest usage.
//
// The TracerProvider MUST be created with IDGenerator().  Without that, the SpanID
// created by Xop will be ignored and that will cause problems with propagation.
func BaseLogger(traceProvider oteltrace.TracerProvider) xopbase.Logger {
	tracer := traceProvider.Tracer("xopotel",
		oteltrace.WithInstrumentationAttributes(
			xopOTELVersion.String(xopotelVersionValue),
			xopVersion.String(xopVersionValue),
		),
		oteltrace.WithInstrumentationVersion(xopotelVersionValue),
	)
	return &logger{
		id:              "otel-" + uuid.New().String(),
		doLogging:       true,
		tracer:          tracer,
		spanFromContext: false,
	}
}

func (logger *logger) ID() string           { return logger.id }
func (logger *logger) ReferencesKept() bool { return true }
func (logger *logger) Buffered() bool       { return false }

func buildRequestSpan(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo, tracer oteltrace.Tracer, bufferedRequest *bufferedRequest) (context.Context, bool, oteltrace.Span) {
	spanKind := oteltrace.SpanKindServer
	isXOP := sourceInfo.Source != otelDataSource
	if !isXOP {
		// replaying data originally coming from OTEL.
		// The spanKind is encoded as the namespace.
		if sk, ok := spanKindFromString[sourceInfo.Namespace]; ok {
			spanKind = sk
		}
	}
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(spanKind),
		oteltrace.WithTimestamp(ts),
	}

	otelStuff := bufferedRequest.getStuff(bundle, true)
	opts = append(opts, otelStuff.SpanOptions()...)

	if bundle.Parent.TraceID().IsZero() {
		opts = append(opts, oteltrace.WithNewRoot())
	}
	ctx, otelSpan := tracer.Start(overrideIntoContext(ctx, bundle), description, opts...)
	if isXOP {
		if !bundle.Baggage.IsZero() {
			otelSpan.SetAttributes(xopBaggage.String(bundle.Baggage.String()))
		}
		otelSpan.SetAttributes(
			xopVersion.String(xopVersionValue),
			xopOTELVersion.String(xopotelVersionValue),
			xopSource.String(sourceInfo.Source+" "+sourceInfo.SourceVersion.String()),
			xopNamespace.String(sourceInfo.Namespace+" "+sourceInfo.NamespaceVersion.String()),
		)
	}
	otelStuff.Set(otelSpan)
	return ctx, isXOP, otelSpan
}

func (logger *logger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	var otelSpan oteltrace.Span
	var isXOP bool
	if logger.spanFromContext {
		otelSpan = oteltrace.SpanFromContext(ctx)
		isXOP = sourceInfo.Source != otelDataSource
	} else {
		ctx, isXOP, otelSpan = buildRequestSpan(ctx, ts, bundle, description, sourceInfo, logger.tracer, logger.bufferedRequest)
	}
	r := &request{
		span: &span{
			logger:   logger,
			otelSpan: otelSpan,
			ctx:      ctx,
			isXOP:    isXOP,
		},
		attributesDefined: make(map[string]struct{}),
	}
	r.span.request = r
	return r
}

func (request *request) SetErrorReporter(f func(error)) { request.errorReporter = f }
func (request *request) Flush()                         {}
func (request *request) Final()                         {}

func (span *span) Boring(_ bool) {}
func (span *span) ID() string    { return span.logger.id }
func (span *span) Done(endTime time.Time, final bool) {
	if !final {
		return
	}
	if span.logger.ignoreDone == span.otelSpan {
		// skip Done for spans passed in to SpanLog()
		return
	}
	span.otelSpan.End(oteltrace.WithTimestamp(endTime))
}

func buildSpan(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, tracer oteltrace.Tracer, bufferedRequest *bufferedRequest) (context.Context, oteltrace.Span) {
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithTimestamp(ts),
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
	}
	otelStuff := bufferedRequest.getStuff(bundle, true)
	opts = append(opts, otelStuff.SpanOptions()...)
	ctx, otelSpan := tracer.Start(overrideIntoContext(ctx, bundle), description, opts...)
	otelStuff.Set(otelSpan)
	return ctx, otelSpan
}

func (parentSpan *span) Span(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, spanSequenceCode string) xopbase.Span {
	var otelSpan oteltrace.Span
	if parentSpan.logger.spanFromContext {
		otelSpan = oteltrace.SpanFromContext(ctx)
	} else {
		if parentSpan.logger.bufferedRequest != nil {
			ctx = parentSpan.ctx
		}
		ctx, otelSpan = buildSpan(ctx, ts, bundle, description, parentSpan.logger.tracer, parentSpan.logger.bufferedRequest)
	}
	s := &span{
		logger:   parentSpan.logger,
		otelSpan: otelSpan,
		ctx:      ctx,
		request:  parentSpan.request,
		isXOP:    parentSpan.isXOP,
	}
	if parentSpan.isXOP && spanSequenceCode != "" {
		otelSpan.SetAttributes(xopSpanSequence.String(spanSequenceCode))
	}
	return s
}

func (span *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		builderWithSpan: builderWithSpan{
			span: span,
		},
	}
}

func (span *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builderWithSpan: builderWithSpan{
			span: span,
		},
	}
}

func (prefill *prefilling) PrefillComplete(msg string) xopbase.Prefilled {
	prefill.builder.prefillMsg = msg
	return &prefilled{
		builderWithSpan: prefill.builderWithSpan,
	}
}

func (prefilled *prefilled) Line(level xopnum.Level, ts time.Time, frames []runtime.Frame) xopbase.Line {
	if !prefilled.span.logger.doLogging || !prefilled.span.otelSpan.IsRecording() {
		return xopbase.SkipLine
	}
	// PERFORMANCE: get line from a pool
	line := &line{}
	line.level = level
	line.span = prefilled.span
	line.attributes = line.prealloc[:2] // reserving two spots at the beginnging
	line.attributes = append(line.attributes, prefilled.span.spanPrefill...)
	line.attributes = append(line.attributes, prefilled.attributes...)
	line.prefillMsg = prefilled.prefillMsg
	line.linkKey = prefilled.linkKey
	line.linkValue = prefilled.linkValue
	line.timestamp = ts
	if len(frames) > 0 {
		fs := make([]string, len(frames))
		for i, frame := range frames {
			fs[i] = frame.File + ":" + strconv.Itoa(frame.Line)
		}
		line.attributes = append(line.attributes, xopStackTrace.StringSlice(fs))
	}
	return line
}

func (line *line) Link(k string, v xoptrace.Trace) {
	if k == xopOTELLinkDetail.String() {
		// Link will not be called with OTEL->XOP->OTEL so no need to
		// suppress anything
		return
	}
	line.attributes = append(line.attributes,
		xopType.String("link"),
		xopLinkData.String(v.String()),
	)
	line.done(line.prefillMsg + k)
	if line.span.logger.bufferedRequest != nil {
		// add span.Links() is handled in buffered.zzzgo so
		// a sub-span is not needed here.
		return
	}
	_, tmpSpan := line.span.logger.tracer.Start(line.span.ctx, k,
		oteltrace.WithLinks(
			oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
					TraceID:    v.TraceID().Array(),
					SpanID:     v.SpanID().Array(),
					TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
					TraceState: emptyTraceState, // TODO: is this right?
					Remote:     true,            // information not available
				}),
			}),
		oteltrace.WithAttributes(
			spanIsLinkEventKey.Bool(true),
		),
	)
	tmpSpan.AddEvent(line.level.String(),
		oteltrace.WithTimestamp(line.timestamp),
		oteltrace.WithAttributes(line.attributes...))
	tmpSpan.SetAttributes(xopType.String("link-event"))
	tmpSpan.End()
}

func (line *line) Model(msg string, v xopbase.ModelArg) {
	v.Encode()
	line.attributes = append(line.attributes,
		xopType.String("model"),
		xopModelType.String(v.ModelType),
		xopEncoding.String(v.Encoding.String()),
		xopModel.String(string(v.Encoded)),
	)
	line.done(line.prefillMsg + msg)
}

func (line *line) Msg(msg string) {
	if line.span.isXOP {
		line.attributes = append(line.attributes, xopType.String("line"))
	}
	line.done(line.prefillMsg + msg)
	// PERFORMANCE: return line to pool
}

func (line *line) done(msg string) {
	if line.span.isXOP {
		line.attributes[0] = xopLineNumber.Int64(int64(atomic.AddInt32(&line.span.request.lineCount, 1)))
		line.attributes[1] = xopLevel.String(line.level.String())
	} else {
		line.attributes = line.attributes[2:]
	}
	if line.timestamp.IsZero() {
		line.span.otelSpan.AddEvent(msg,
			oteltrace.WithAttributes(line.attributes...))
	} else {
		line.span.otelSpan.AddEvent(msg,
			oteltrace.WithTimestamp(line.timestamp),
			oteltrace.WithAttributes(line.attributes...))
	}
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (line *line) Template(template string) {
	kv := make(map[string]int)
	for i, a := range line.attributes {
		kv[string(a.Key)] = i
	}
	msg := templateRE.ReplaceAllStringFunc(template, func(k string) string {
		k = k[1 : len(k)-1]
		if i, ok := kv[k]; ok {
			a := line.attributes[i]
			switch a.Value.Type() {
			case attribute.BOOL:
				return strconv.FormatBool(a.Value.AsBool())
			case attribute.INT64:
				return strconv.FormatInt(a.Value.AsInt64(), 10)
			case attribute.FLOAT64:
				return strconv.FormatFloat(a.Value.AsFloat64(), 'g', -1, 64)
			case attribute.STRING:
				return a.Value.AsString()
			case attribute.BOOLSLICE:
				return fmt.Sprint(a.Value.AsBoolSlice())
			case attribute.INT64SLICE:
				return fmt.Sprint(a.Value.AsInt64Slice())
			case attribute.FLOAT64SLICE:
				return fmt.Sprint(a.Value.AsFloat64Slice())
			case attribute.STRINGSLICE:
				return fmt.Sprint(a.Value.AsStringSlice())
			default:
				return "{" + k + "}"
			}
		}
		return "''"
	})
	line.attributes = append(line.attributes,
		xopType.String("line"),
		xopTemplate.String(template),
	)
	line.done(line.prefillMsg + msg)
}

func (builder *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.Key().String(), []string{v.String(), "enum", strconv.FormatInt(v.Int64(), 10)}))
}

func (builder *builder) Any(k xopat.K, v xopbase.ModelArg) {
	v.Encode()
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{string(v.Encoded), "any", v.Encoding.String(), v.ModelType}))
}

func (builder *builder) Time(k xopat.K, v time.Time) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{v.Format(time.RFC3339Nano), "time"}))
}

func (builder *builder) Duration(k xopat.K, v time.Duration) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{v.String(), "dur"}))
}

func (builder *builder) Uint64(k xopat.K, v uint64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{strconv.FormatUint(v, 10), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Int64(k xopat.K, v int64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{strconv.FormatInt(v, 10), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Float64(k xopat.K, v float64, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{strconv.FormatFloat(v, 'g', -1, 64), xopbase.DataTypeToString[dt]}))
}

func (builder *builder) String(k xopat.K, v string, dt xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.StringSlice(k.String(), []string{v, xopbase.DataTypeToString[dt]}))
}

func (builder *builder) Bool(k xopat.K, v bool) {
	builder.attributes = append(builder.attributes, attribute.Bool(k.String(), v))
}

var skipIfOTEL = map[string]struct{}{
	otelReplayStuff.Key().String(): {},
}

func (span *span) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	if k.Key().String() == otelReplayStuff.Key().String() {
		return
	}
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	enc, err := v.MarshalJSON()
	var value string
	if err != nil {
		value = fmt.Sprintf("[zopotel] could not marshal %T value: %s", v, err)
	} else {
		value = string(enc)
	}
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.String(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key.String()]
	s = append(s, value)
	span.priorStringSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
}

func (span *span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.Bool(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[bool]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[bool]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorBoolSlices == nil {
		span.priorBoolSlices = make(map[string][]bool)
	}
	s := span.priorBoolSlices[key.String()]
	s = append(s, value)
	span.priorBoolSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.BoolSlice(key.String(), s))
}

func (span *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v.String() + "/" + strconv.FormatInt(v.Int64(), 10)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.String(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key.String()]
	s = append(s, value)
	span.priorStringSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
}

func (span *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.Float64(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[float64]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[float64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorFloat64Slices == nil {
		span.priorFloat64Slices = make(map[string][]float64)
	}
	s := span.priorFloat64Slices[key.String()]
	s = append(s, value)
	span.priorFloat64Slices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.Float64Slice(key.String(), s))
}

func (span *span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		if k.SubType() == xopat.AttributeTypeDuration {
			span.otelSpan.SetAttributes(attribute.String(key.String(), time.Duration(value).String()))
		} else {
			span.otelSpan.SetAttributes(attribute.Int64(key.String(), value))
		}
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[int64]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[int64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if k.SubType() == xopat.AttributeTypeDuration {
		if span.priorStringSlices == nil {
			span.priorStringSlices = make(map[string][]string)
		}
		s := span.priorStringSlices[key.String()]
		s = append(s, time.Duration(value).String())
		span.priorStringSlices[key.String()] = s
		span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
	} else {
		if span.priorInt64Slices == nil {
			span.priorInt64Slices = make(map[string][]int64)
		}
		s := span.priorInt64Slices[key.String()]
		s = append(s, value)
		span.priorInt64Slices[key.String()] = s
		span.otelSpan.SetAttributes(attribute.Int64Slice(key.String(), s))
	}
}

func (span *span) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	if k.Key().String() == otelLink.Key().String() {
		return
	}
	value := v.String()
	if span.logger.bufferedRequest == nil {
		_, tmpSpan := span.logger.tracer.Start(span.ctx, k.Key().String(),
			oteltrace.WithLinks(
				oteltrace.Link{
					SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
						TraceID:    v.TraceID().Array(),
						SpanID:     v.SpanID().Array(),
						TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
						TraceState: emptyTraceState, // TODO: is this right?
						Remote:     true,            // information not available
					}),
					Attributes: []attribute.KeyValue{
						xopLinkMetadataKey.String(key.String()),
					},
				}),
			oteltrace.WithAttributes(
				spanIsLinkAttributeKey.Bool(true),
			),
		)
		tmpSpan.SetAttributes(spanIsLinkAttributeKey.Bool(true))
		tmpSpan.End()
	}
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.String(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key.String()]
	s = append(s, value)
	span.priorStringSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
}

func (span *span) MetadataString(k *xopat.StringAttribute, v string) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.String(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key.String()]
	s = append(s, value)
	span.priorStringSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
}

func (span *span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	key := k.Key()
	if span.isXOP {
		if _, ok := span.request.attributesDefined[key.String()]; !ok {
			if k.Description() != xopSynthesizedForOTEL {
				span.request.otelSpan.SetAttributes(attribute.String(attributeDefinitionPrefix+key.String(), k.DefinitionJSONString()))
				span.request.attributesDefined[key.String()] = struct{}{}
			}
		}
	}
	value := v.Format(time.RFC3339Nano)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key.String()]; ok {
				return
			}
			span.hasPrior[key.String()] = struct{}{}
		}
		span.otelSpan.SetAttributes(attribute.String(key.String(), value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		if span.metadataSeen == nil {
			span.metadataSeen = make(map[string]interface{})
		}
		seenRaw, ok := span.metadataSeen[key.String()]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key.String()] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[string]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorStringSlices == nil {
		span.priorStringSlices = make(map[string][]string)
	}
	s := span.priorStringSlices[key.String()]
	s = append(s, value)
	span.priorStringSlices[key.String()] = s
	span.otelSpan.SetAttributes(attribute.StringSlice(key.String(), s))
}

var spanKindFromString = map[string]oteltrace.SpanKind{
	oteltrace.SpanKindClient.String():   oteltrace.SpanKindClient,
	oteltrace.SpanKindConsumer.String(): oteltrace.SpanKindConsumer,
	oteltrace.SpanKindInternal.String(): oteltrace.SpanKindInternal,
	oteltrace.SpanKindProducer.String(): oteltrace.SpanKindProducer,
	oteltrace.SpanKindServer.String():   oteltrace.SpanKindServer,
}
