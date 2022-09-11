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

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func SeedModifier(ctx context.Context, tracer oteltrace.Tracer) xop.SeedModifier {
	return xop.CombineSeedModfiers(
		xop.WithBase(&logger{
			id: "otel-" + uuid.New().String(),
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
	return span.logger.Request(ctx, ts, span, description)
}

func (span *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		Builder: builder{
			span: span,
		},
	}
}

func (span *span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		Builder: Builder{
			span: span,
		},
	}
}

func (prefill *prefilling) PrefillComplete(msg string) xopbase.Prefilled {
	return &prefilled{
		Builder: prefill.Builder,
	}
}

func (prefilled *prefilled) Line(level xopnum.Level, _ time.Time, stack []uintptr) xopbase.Line {
	// TODO: get line from a pool
	line := &line{}
	line.span = prefilled.span
	line.attributes = line.prealloc[:0]
	line.attributes = append(line.attributes, span.spanPrefill...)
	line.attributes = append(line.attributes, prefilled.attributes...)
	line.attributes = append(line.attributes,
		logSeverityKey.String(level.String()),
	)
	// TODO: stack trace
	// semconv.ExceptionStacktraceKey.String
	return line
}

func (line *line) Static(msg string) { return line.Msg(msg) }
func (line *line) Msg(string) {
	line.span.span.AddEvent("log", trace.WithAttributes(line.attributes))
	// TODO: return line to pool
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (line *line) Template(template string) {
	kv := make(map[string]int)
	for i, a := range line.attributes {
		kv[string(a.Key)] = i
	}
	msg := templateRE.ReplaceAllStringFunc(line.Message, func(k string) string {
		k = k[1 : len(k)-1]
		if i, ok := kv[k]; ok {
			a := line.attributes[i]
			switch a.Value.Type() {
			case attributes.BOOL:
				return strconv.FormatBool(a.Value.AsBool())
			case attributes.INT64:
				return strconv.FormatInt(a.Value.AsInt64(), 10)
			case attributes.FLOAT64:
				return strconv.FormatFloat(a.Value.AsInt64(), 64, "e")
			case attributes.STRING:
				return a.Value.AsString()
			case attributes.BOOLSLICE:
				return fmt.Sprint(a.Value.AsBoolSlice())
			case attributes.INT64SLICE:
				return fmt.Sprint(a.Value.AsInt64Slice())
			case attributes.FLOAT64SLICE:
				return fmt.Sprint(a.Value.AsFloat64Slice())
			case attributes.STRINGSLICE:
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
	builder.attributes = append(builder.Attributes, attributes.Stringer(k.Key(), v))
}

func (builder *builder) Any(k string, v interface{}) {
	switch vt := v.(type) {
	case bool:
		builder.attributes = append(builder.Attributes, attributes.Bool(k, v))
	case []bool:
		builder.attributes = append(builder.Attributes, attributes.BoolSlice(k, v))
	case float64:
		builder.attributes = append(builder.Attributes, attributes.Float64(k, v))
	case []float64:
		builder.attributes = append(builder.Attributes, attributes.Float64Slice(k, v))
	case int64:
		builder.attributes = append(builder.Attributes, attributes.Int64(k, v))
	case []int64:
		builder.attributes = append(builder.Attributes, attributes.Int64Slice(k, v))
	case string:
		builder.attributes = append(builder.Attributes, attributes.String(k, v))
	case []string:
		builder.attributes = append(builder.Attributes, attributes.StringSlice(k, v))
	case fmt.Stringer:
		builder.attributes = append(builder.Attributes, attributes.Stringer(k, v))

	default:
		enc, err := json.Marshal(v)
		if err != nil {
			builder.attributes = append(builder.Attributes, attributes.String(k+"-error", err.Error()))
		} else {
			builder.attributes = append(builder.Attributes, attributes.String(k, string(enc)))
		}
	}
}

func (builder *builder) Time(k string, v time.Time) {
	builder.attributes = append(builder.Attributes, attributes.String(k, v.Format(time.RFC3339Nano)))
}

func (builder *builder) Duration(k string, v time.Duration) {
	builder.attributes = append(builder.Attributes, attributes.Stringer(k, v))
}

func (builder *builder) Error(k string, v error) {
	builder.attributes = append(builder.Attributes, attributes.String(k, v.Error()))
}

func (span *span) MetadataLink(k *xopat.LinkAttribute, v trace.Trace) {
	traceState, _ := oteltrace.ParseTraceState("")
	_, subspan := tracer.Start(span.ctx, k.Key(), oteltrace.WithLinks(
		oteltrace.Link{
			SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID:    v.TraceID().Array(),
				SpanID:     v.SpanID().Array(),
				TraceFlags: v.Flags().Array()[0],
				TraceState: traceState,
				Remote:     true, // information not available
			}),
		},
	))
	subspan.End()
}

func (builder *builder) Uint64(k string, v uint64, _ xopbase.DataType) {
	builder.attributes = append(builder.Attributes, attributes.String(k, strconv.FormatUint(v, 10)))
}

func (builder *builder) Bool(k string, v bool) {
	builder.attributes = append(builder.Attributes, attributes.Bool(k, v))
}

func (builder *builder) String(k string, v string) {
	builder.attributes = append(builder.Attributes, attributes.String(k, v))
}

func (builder *builder) Float64(k string, v float64, _ xopbase.DataType) {
	builder.attributes = append(builder.Attributes, attributes.Float64(k, v))
}

func (builder *builder) Int64(k string, v int64, _ xopbase.DataType) {
	builder.attributes = append(builder.Attributes, attributes.Int64(k, v))
}

func (span *span) MetadataAny(k *xopat.AnyAttribute, v interface{}) {
	key := k.Key()
	enc, err := json.Marshal(v)
	var value string
	if err != nil {
		value = fmt.Errorf("could not marshal %T value: %s", v, err)
	} else {
		value = string(enc)
	}
	if !k.Multiple() {
		if k.Locked() {
			span.lock.Lock()
			defer span.lock.Unlock()
			if span.priorString == nil {
				span.priorString = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorString[value]; ok {
					return
				}
				span.priorString[value] = struct{}{}
			}
			// CONDITIONAL ELSE
			if span.priorAny == nil {
				span.priorAny = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorAny[value]; ok {
					return
				}
				span.priorAny[value] = struct{}{}
			}
			// CONDITIONAL END
		}
		span.span.SetAttributes(attribute.String(key, value))
		// CONDITIONAL ELSE
		span.span.SetAttributes(attribute.Any(key, value))
		// CONDITIONAL END
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.priorDistinct[key]
		if !ok {
			seen := make(map[string]struct{})
			span.priorDistinct[key] = seen
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
		span.priorStringSlices = make(map[string][]interface{})
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
			// ELSE CONDITIONAL
			seen := make(map[bool]struct{})
			span.priorDistinct[key] = seen
			seen[value] = struct{}{}
		} else {
			// ELSE CONDITIONAL
			seen := seenRaw.(map[bool]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	// ELSE CONDITIONAL
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
			if span.priorString == nil {
				span.priorString = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorString[value]; ok {
					return
				}
				span.priorString[value] = struct{}{}
			}
			// CONDITIONAL ELSE
			if span.priorEnum == nil {
				span.priorEnum = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorEnum[value]; ok {
					return
				}
				span.priorEnum[value] = struct{}{}
			}
			// CONDITIONAL END
		}
		span.span.SetAttributes(attribute.String(key, value))
		// CONDITIONAL ELSE
		span.span.SetAttributes(attribute.Enum(key, value))
		// CONDITIONAL END
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.priorDistinct[key]
		if !ok {
			seen := make(map[string]struct{})
			span.priorDistinct[key] = seen
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
		span.priorStringSlices = make(map[string][]xopat.Enum)
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
			// ELSE CONDITIONAL
			seen := make(map[float64]struct{})
			span.priorDistinct[key] = seen
			seen[value] = struct{}{}
		} else {
			// ELSE CONDITIONAL
			seen := seenRaw.(map[float64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	// ELSE CONDITIONAL
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
			// ELSE CONDITIONAL
			seen := make(map[int64]struct{})
			span.priorDistinct[key] = seen
			seen[value] = struct{}{}
		} else {
			// ELSE CONDITIONAL
			seen := seenRaw.(map[int64]struct{})
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
		}
	}
	// ELSE CONDITIONAL
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
			if span.priorString == nil {
				span.priorString = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorString[value]; ok {
					return
				}
				span.priorString[value] = struct{}{}
			}
			// CONDITIONAL ELSE
			if span.priorString == nil {
				span.priorString = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorString[value]; ok {
					return
				}
				span.priorString[value] = struct{}{}
			}
			// CONDITIONAL END
		}
		span.span.SetAttributes(attribute.String(key, value))
		// CONDITIONAL ELSE
		span.span.SetAttributes(attribute.String(key, value))
		// CONDITIONAL END
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.priorDistinct[key]
		if !ok {
			seen := make(map[string]struct{})
			span.priorDistinct[key] = seen
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
			if span.priorString == nil {
				span.priorString = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorString[value]; ok {
					return
				}
				span.priorString[value] = struct{}{}
			}
			// CONDITIONAL ELSE
			if span.priorTime == nil {
				span.priorTime = map[string]struct{}{
					value: {},
				}
			} else {
				if _, ok := span.priorTime[value]; ok {
					return
				}
				span.priorTime[value] = struct{}{}
			}
			// CONDITIONAL END
		}
		span.span.SetAttributes(attribute.String(key, value))
		// CONDITIONAL ELSE
		span.span.SetAttributes(attribute.Time(key, value))
		// CONDITIONAL END
		return
	}
	span.lock.Lock()
	defer span.lock.Unlock()
	if k.Distinct() {
		seenRaw, ok := span.priorDistinct[key]
		if !ok {
			seen := make(map[string]struct{})
			span.priorDistinct[key] = seen
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
		span.priorStringSlices = make(map[string][]time.Time)
	}
	s := span.priorStringSlices[key]
	s = append(s, value)
	span.priorStringSlices[key] = s
	span.span.SetAttributes(attribute.StringSlice(key, s))
}
