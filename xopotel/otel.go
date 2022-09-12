// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

/*
Package xopotel provides a gateway from xop into open telemetry
using OTEL's top-level APIs.

This gateway can be used either as a base layer for xop allowing
xop to output through OTEL; or it can be used to bridge the gap
between an application that is otherwise using OTEL and a library
that is expects to be provided with a xop logger.

OTEL supports far fewer data types than xop.  Mostly, xop types
can be converted cleanly, but links are a special case: links can
only be added to OTEL spans when the span is created.  Since xop
allows links to be made at any time, MetadataLink()s will be added as
ephemeral sub-spans.  Distinct, Multiple, and Locked attributes will
be ignored for links.

OTEL does not support unsigned ints so they get formatted as strings.
*/
package xopotel

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopat"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopnum"
	"github.com/muir/xop-go/xoputil"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func SeedModifier(ctx context.Context, tracer oteltrace.Tracer, doLogging bool) xop.SeedModifier {
	return xop.CombineSeedModfiers(
		xop.WithBase(&logger{
			id:        "otel-" + uuid.New().String(),
			doLogging: doLogging,
		}),
		xop.WithContext(ctx),
		xop.WithReactive(func(ctx context.Context, seed xop.Seed, selfIndex int, nameOrDescription string, isChildSpan bool) xop.Seed {
			if isChildSpan {
				ctx, span := tracer.Start(ctx, nameOrDescription)
				return seed.Copy(
					xop.WithContext(ctx),
					xop.WithSpan(span.SpanContext().SpanID()),
				)
			}
			ctx, span := tracer.Start(ctx, nameOrDescription, oteltrace.WithNewRoot())
			bundle := seed.Bundle()
			if bundle.TraceParent.IsZero() {
				bundle.State.SetString(span.SpanContext().TraceState().String())
				bundle.Trace.Flags().Set([1]byte{byte(span.SpanContext().TraceFlags())})
				bundle.Trace.Version().Set([1]byte{1})
				bundle.Trace.TraceID().Set(span.SpanContext().TraceID())
			}
			bundle.Trace.SpanID().Set(span.SpanContext().SpanID())
			return seed.Copy(
				xop.WithContext(ctx),
				xop.WithBundle(bundle),
			)
		}),
	)
}

func (logger *logger) ID() string           { return logger.id }
func (logger *logger) ReferencesKept() bool { return true }
func (logger *logger) Buffered() bool       { return false }
func (logger *logger) Close()               {}

func (logger *logger) Request(ctx context.Context, ts time.Time, _ trace.Bundle, description string) xopbase.Request {
	otelspan := oteltrace.SpanFromContext(ctx)
	return &span{
		logger: logger,
		span:   otelspan,
		ctx:    ctx,
	}
}

func (span *span) Flush()                         {}
func (span *span) SetErrorReporter(f func(error)) {}
func (span *span) Boring(_ bool)                  {}
func (span *span) ID() string                     { return span.logger.id }
func (span *span) Done(endTime time.Time, final bool) {
	if !final {
		return
	}
}

// TODO: store span sequence code
func (span *span) Span(ctx context.Context, ts time.Time, bundle trace.Bundle, description string, spanSequenceCode string) xopbase.Span {
	return span.logger.Request(ctx, ts, bundle, description)
}

func (span *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		builder: builder{
			span: span,
		},
	}
}

func (span *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: builder{
			span: span,
		},
	}
}

func (prefill *prefilling) PrefillComplete(msg string) xopbase.Prefilled {
	return &prefilled{
		builder: prefill.builder,
	}
}

func (prefilled *prefilled) Line(level xopnum.Level, _ time.Time, stack []uintptr) xopbase.Line {
	if !prefilled.span.logger.doLogging || !prefilled.span.span.IsRecording() {
		return xoputil.SkipLine
	}
	// TODO: get line from a pool
	line := &line{}
	line.span = prefilled.span
	line.attributes = line.prealloc[:0]
	line.attributes = append(line.attributes, prefilled.span.spanPrefill...)
	line.attributes = append(line.attributes, prefilled.attributes...)
	line.spanStartOptions = nil
	line.spanStartOptions = append(line.spanStartOptions, prefilled.spanStartOptions...)
	// TODO: stack trace
	// semconv.ExceptionStacktraceKey.String
	return line
}

func (line *line) Static(msg string) { line.Msg(msg) }

func (line *line) Msg(msg string) {
	line.attributes = append(line.attributes, logMessageKey.String(msg))
	if len(line.spanStartOptions) == 0 {
		line.span.span.AddEvent(line.level.String(), oteltrace.WithAttributes(line.attributes...))
		return
		// TODO: return line to pool
	}
	_, tmpSpan := line.span.logger.tracer.Start(line.span.ctx, line.level.String(), line.spanStartOptions...)
	tmpSpan.AddEvent(line.level.String(), oteltrace.WithAttributes(line.attributes...))
	tmpSpan.End()
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
	line.Msg(msg)
}

func (builder *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	builder.attributes = append(builder.attributes, attribute.Stringer(k.Key(), v))
}

func (builder *builder) Any(k string, v interface{}) {
	switch typed := v.(type) {
	case bool:
		builder.attributes = append(builder.attributes, attribute.Bool(k, typed))
	case []bool:
		builder.attributes = append(builder.attributes, attribute.BoolSlice(k, typed))
	case float64:
		builder.attributes = append(builder.attributes, attribute.Float64(k, typed))
	case []float64:
		builder.attributes = append(builder.attributes, attribute.Float64Slice(k, typed))
	case int64:
		builder.attributes = append(builder.attributes, attribute.Int64(k, typed))
	case []int64:
		builder.attributes = append(builder.attributes, attribute.Int64Slice(k, typed))
	case string:
		builder.attributes = append(builder.attributes, attribute.String(k, typed))
	case []string:
		builder.attributes = append(builder.attributes, attribute.StringSlice(k, typed))
	case fmt.Stringer:
		builder.attributes = append(builder.attributes, attribute.Stringer(k, typed))

	default:
		enc, err := json.Marshal(v)
		if err != nil {
			builder.attributes = append(builder.attributes, attribute.String(k+"-error", err.Error()))
		} else {
			builder.attributes = append(builder.attributes, attribute.String(k, string(enc)))
		}
	}
}

func (builder *builder) Time(k string, v time.Time) {
	builder.attributes = append(builder.attributes, attribute.String(k, v.Format(time.RFC3339Nano)))
}

func (builder *builder) Duration(k string, v time.Duration) {
	builder.attributes = append(builder.attributes, attribute.Stringer(k, v))
}

func (builder *builder) Error(k string, v error) {
	builder.attributes = append(builder.attributes, attribute.String(k, v.Error()))
}

func (span *span) MetadataLink(k *xopat.LinkAttribute, v trace.Trace) {
	_, tmpSpan := span.logger.tracer.Start(span.ctx, k.Key(), oteltrace.WithLinks(
		oteltrace.Link{
			SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID:    v.TraceID().Array(),
				SpanID:     v.SpanID().Array(),
				TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
				TraceState: emptyTraceState, // TODO: is this right?
				Remote:     true,            // information not available
			}),
		},
	))
	tmpSpan.End()
}

func (builder *builder) Uint64(k string, v uint64, _ xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.String(k, strconv.FormatUint(v, 10)))
}

func (builder *builder) Link(k string, v trace.Trace) {
	builder.spanStartOptions = append(builder.spanStartOptions, oteltrace.WithLinks(
		oteltrace.Link{
			SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID:    v.TraceID().Array(),
				SpanID:     v.SpanID().Array(),
				TraceFlags: oteltrace.TraceFlags(v.Flags().Array()[0]),
				TraceState: emptyTraceState,
				Remote:     true, // information not available
			}),
		},
	))
}

func (builder *builder) Bool(k string, v bool) {
	builder.attributes = append(builder.attributes, attribute.Bool(k, v))
}

func (builder *builder) String(k string, v string) {
	builder.attributes = append(builder.attributes, attribute.String(k, v))
}

func (builder *builder) Float64(k string, v float64, _ xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.Float64(k, v))
}

func (builder *builder) Int64(k string, v int64, _ xopbase.DataType) {
	builder.attributes = append(builder.attributes, attribute.Int64(k, v))
}

func (span *span) MetadataAny(k *xopat.AnyAttribute, v interface{}) {
	key := k.Key()
	enc, err := json.Marshal(v)
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
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	key := k.Key()
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Bool(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[bool]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorBoolSlices[key]
	s = append(s, value)
	span.priorBoolSlices[key] = s
	span.span.SetAttributes(attribute.BoolSlice(key, s))
}

func (span *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	key := k.Key()
	value := v.String()
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	key := k.Key()
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Float64(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[float64]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorFloat64Slices[key]
	s = append(s, value)
	span.priorFloat64Slices[key] = s
	span.span.SetAttributes(attribute.Float64Slice(key, s))
}

func (span *span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	key := k.Key()
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.Int64(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[int64]struct{})
			span.metadataSeen[key] = seen
			seen[value] = struct{}{}
		} else {
			seen := seenRaw.(map[int64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	if span.priorInt64Slices == nil {
		span.priorInt64Slices = make(map[string][]int64)
	}
	s := span.priorInt64Slices[key]
	s = append(s, value)
	span.priorInt64Slices[key] = s
	span.span.SetAttributes(attribute.Int64Slice(key, s))
}

func (span *span) MetadataString(k *xopat.StringAttribute, v string) {
	key := k.Key()
	value := v
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}

func (span *span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	key := k.Key()
	value := v.Format(time.RFC3339Nano)
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.hasPrior == nil {
				span.hasPrior = make(map[string]struct{})
			}
			if _, ok := span.hasPrior[key]; ok {
				return
			}
			span.hasPrior[key] = struct{}{}
		}
		span.span.SetAttributes(attribute.String(key, value))
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.metadataSeen[key]
		if !ok {
			seen := make(map[string]struct{})
			span.metadataSeen[key] = seen
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
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}
