// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/xoplog/xop-go/internal/util/version"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/muir/list"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	_ sdktrace.SpanExporter = &spanExporter{}
	_ sdktrace.SpanExporter = &unhack{}
)

type spanExporter struct {
	base xopbase.Logger
}

func NewExporter(base xopbase.Logger) sdktrace.SpanExporter {
	return &spanExporter{base: base}
}

type spanReplay struct {
	id2index map[oteltrace.SpanID]int
	spans    []sdktrace.ReadOnlySpan
	subSpans [][]oteltrace.SpanID
	data     []*datum
}

type datum struct {
	baseSpan             xopbase.Span
	requestIndex         int
	attributeDefinitions map[string]*decodeAttributeDefinition
}

func (e *spanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	id2Index := makeIndex(spans)
	subSpans, todo := makeSubspans(id2Index, spans)
	x := spanReplay{
		id2Index: id2Index,
		spans:    spans,
		subSpans: subSpans,
		data:     make([]*datum, len(spans)),
	}
	for _, i := range todo {
		x.data[i] = &datum{}
		err := x.Replay(ctx, spans[i], x.data[i], i)
		if err != nil {
			return err
		}
	}
}

func (x spanReplay) Replay(ctx context.Context, span sdktrace.ReadOnlySpan, data *datum, myIndex int) error {
	attributeMap := mapAttributes(span)
	var bundle xoptrace.Bundle
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		bundle.Trace.TraceID().SetArray(spanContext.TraceID())
	}
	if spanContext.HasSpanID() {
		bundle.Trace.SpanID().SetArray(spanContext.SpanID())
	}
	if spanContext.IsSampled() {
		bundle.Trace.Flags().SetArray([1]byte{1})
	}
	if spanContext.TraceState().Len() != 0 {
		bundle.State.SetString(spanContext.TraceState().String())
	}
	parentIndex, ok := lookupParent(x.id2Index, span)
	if ok {
		parentContext := spans[parentIndex].SpanContext()
		xopParent := x.data[parentIndex]
		if parentContext.HasTraceID() {
			bundle.Parent.TraceID().SetArray(parentContext.TraceID())
			if bundle.Trace.TraceID().IsZero() {
				bundle.Trace.TraceID().Set(bundle.Parent.GetTraceID())
			}
		}
		if parentContext.HasSpanID() {
			bundle.Parent.SpanID().SetArray(parentContext.SpanID())
		}
		if parentContext.IsSampled() {
			bundle.Parent.Flags().SetArray([1]byte{1})
		}
		bundle.Parent.Version().SetArray([1]byte{1})
	}
	bundle.Trace.Version().SetArray([1]byte{1})
	spanKind := span.SpanKind()
	if spanKind == oteltrace.SpanKindUnspecified {
		spanKind = oteltrace.SpanKind(defaulted(attributeMap.GetInt(otelSpanKind), int(oteltrace.SpanKindUnspecified)))
	}
	var baseSpan xopbase.Span
	switch spanKind {
	case oteltrace.SpanKindUnspecified, oteltrace.SpanKindInternal:
		if ok {
			data.baseSpan = xopParent.baseSpan.Span(ctx, span.StartTime(), bundle, span.Name(), defaulted(attributeMap.GetString(logSpanSequence), ""))
			data.requestIndex = xopParent.requestIndex
		} else {
			// This is a difficult sitatuion. We have an internal/unspecified span
			// that does not have a parent present. There is no right answer for what
			// to do. In the Xop world, such a span isn't allowed to exist. We'll treat
			// this span as a request, but mark it as promoted.
			data.baseSpan = e.base.Request(ctx, span.StartTime(), bundle, span.Name(), buildSourceInfo(span, attributeMap))
			data.baseSpan.MetadataBool(xopPromotedMetadata, true)
			data.requestIndex = myIndex
			data.attributeDefinitions = make(map[string]*decodeAttributeDefinition)
		}
	default:
		data.baseSpan = e.base.Request(ctx, span.StartTime(), bundle, span.Name(), buildSourceInfo(span, attributeMap))
		data.requestIndex = myIndex
		data.attributeDefinitions = make(map[string]*decodeAttributeDefinition)
	}
	y := baseSpanReplay{
		spanReplay: x,
		span:       span,
		base:       baseSpan,
		datam:      data,
	}
	for _, attribute := range span.Attributes() {
		err := y.AddAttribute(ctx, attribute)
		if err != nil {
			return nil, err
		}
	}
	err := y.AddEvents(ctx)
	if err != nil {
		return nil, err
	}
	return baseSpan, attributeDefIndex, nil
}

type baseSpanReplay struct {
	spanReplay
	span  sdktrace.ReadOnlySpan
	base  xopbase.Span
	datam *datum
}

type decodeAttributeDefinition struct {
	xopat.Make
	AttributeType xopproto.AttributeType `json:"vtype"`
}

func (x baseSpanReplay) AddAttribute(ctx context.Context, a attribute.KeyValue) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "add span attribute %s with type %s", a.Key, a.Value.Type())
		}
	}()
	if strings.HasPrefix(a.Key, attributeDefintionPrefix) {
		key := strings.TrimPrefix(a.Key, attributeDefintionPrefix)
		if _, ok := x.data[x.data.requestIndex].attributeDefintions[key]; ok {
			return nil
		}
		if a.Value.Type() != attribute.STRING {
			return errors.Errorf("expected type to be string")
		}
		var aDef decodeAttributeDefinition
		aDef, err := json.Unmarshal([]byte(a.Value.AsString()), &aDef)
		if err != nil {
			return errors.Wrapf(err, "could not unmarshal attribute defintion")
		}
		x.data[x.data.requestIndex].attributeDefintions[key] = &aDef
		return nil
	}

	if aDef, ok := x.data[x.data.requestIndex].attributeDefintions[key]; ok {
		return x.AddXopAttribute(ctx, a, aDef)
	}

	mkMake := func(key string, multiple bool) xopat.Make {
		return xopat.Make{
			Description: xopSynthesizedForOTEL,
			Key:         key,
			Multiple:    multiple,
		}
	}
	switch a.Value.Type() {
	case attribute.BOOL:
		registeredAttribute, err := x.attributeRegistry.ConstructBoolAttribute(mkMake(a.Key, false), xopat.AttributeTypeBool)
		if err != nil {
			return err
		}
		x.datum.baseSpan.MetadataBool(registeredAttribute, a.Value.AsBool())
	case attribute.BOOLSLICE:
		registeredAttribute, err := x.attributeRegistry.ConstructBoolAttribute(mkMake(a.Key, true), xopat.AttributeTypeBool)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsBoolSlice() {
			x.datum.baseSpan.MetadataBool(registeredAttribute, v)
		}
	case attribute.FLOAT64:
		registeredAttribute, err := x.attributeRegistry.ConstructFloat64Attribute(mkMake(a.Key, false), xopat.AttributeTypeFloat64)
		if err != nil {
			return err
		}
		x.datum.baseSpan.MetadataFloat64(registeredAttribute, a.Value.AsFloat64())
	case attribute.FLOAT64SLICE:
		registeredAttribute, err := x.attributeRegistry.ConstructFloat64Attribute(mkMake(a.Key, true), xopat.AttributeTypeFloat64)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsFloat64Slice() {
			x.datum.baseSpan.MetadataFloat64(registeredAttribute, v)
		}
	case attribute.INT64:
		registeredAttribute, err := x.attributeRegistry.ConstructInt64Attribute(mkMake(a.Key, false), xopat.AttributeTypeInt64)
		if err != nil {
			return err
		}
		x.datum.baseSpan.MetadataInt64(registeredAttribute, a.Value.AsInt64())
	case attribute.INT64SLICE:
		registeredAttribute, err := x.attributeRegistry.ConstructInt64Attribute(mkMake(a.Key, true), xopat.AttributeTypeInt64)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsInt64Slice() {
			x.datum.baseSpan.MetadataInt64(registeredAttribute, v)
		}
	case attribute.STRING:
		registeredAttribute, err := x.attributeRegistry.ConstructStringAttribute(mkMake(a.Key, false), xopat.AttributeTypeString)
		if err != nil {
			return err
		}
		x.datum.baseSpan.MetadataString(registeredAttribute, a.Value.AsString())
	case attribute.STRINGSLICE:
		registeredAttribute, err := x.attributeRegistry.ConstructStringAttribute(mkMake(a.Key, true), xopat.AttributeTypeString)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsStringSlice() {
			x.datum.baseSpan.MetadataString(registeredAttribute, v)
		}

	case attribute.INVALID:
		fallthrough
	default:
		return errors.Errorf("span attribute key (%s) has value type (%s) that is not expected", a.Key, a.Value.Type())
	}
}

func (x baseSpanReplay) AddXopAttribute(ctx context.Context, a attribute.KeyValue, aDef *decodeAttributeDefinition) error {
	switch aDef.AttributeType {
	case xopproto.AttributeType_Any:
		registeredAttribute, err := x.attributeRegistry.ConstructAnyAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v string) (xopbase.ModelArg, error) {
			var ma xopbase.ModelArg
			return ma, ma.UnmarshalJSON([]byte(v))
		}
		if k.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataAny(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsString()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataAny(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Bool:
		registeredAttribute, err := x.attributeRegistry.ConstructBoolAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.BOOL, attribute.BOOLSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v bool) (bool, error) { return v, nil }
		if k.Multiple() {
			values := a.Value.AsBoolSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataBool(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsBool()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataBool(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Duration:
		registeredAttribute, err := x.attributeRegistry.ConstructDurationAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		if k.Multiple() {
			values := a.Value.AsDurationSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataDuration(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsDuration()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataDuration(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Enum:
		registeredAttribute, err := x.attributeRegistry.ConstructEnumAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		// XXX
		if k.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataEnum(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsString()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataEnum(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Float64:
		registeredAttribute, err := x.attributeRegistry.ConstructFloat64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.FLOAT64, attribute.FLOAT64SLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v float64) (float64, error) { return v, nil }
		if k.Multiple() {
			values := a.Value.AsFloat64Slice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataFloat64(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsFloat64()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataFloat64(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Int:
		registeredAttribute, err := x.attributeRegistry.ConstructIntAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		if k.Multiple() {
			values := a.Value.AsIntSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataInt(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsInt()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataInt(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Int16:
		registeredAttribute, err := x.attributeRegistry.ConstructInt16Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		if k.Multiple() {
			values := a.Value.AsInt16Slice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataInt16(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsInt16()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataInt16(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Int32:
		registeredAttribute, err := x.attributeRegistry.ConstructInt32Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		if k.Multiple() {
			values := a.Value.AsInt32Slice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataInt32(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsInt32()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataInt32(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Int64:
		registeredAttribute, err := x.attributeRegistry.ConstructInt64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v int64) (int64, error) { return v, nil }
		if k.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataInt64(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsInt64()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataInt64(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Int8:
		registeredAttribute, err := x.attributeRegistry.ConstructInt8Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		if k.Multiple() {
			values := a.Value.AsInt8Slice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataInt8(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsInt8()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataInt8(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Link:
		registeredAttribute, err := x.attributeRegistry.ConstructLinkAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v string) (xoptrace.Trace, error) {
			t, ok := xoptrace.TraceFromString(v)
			if !ok {
				return xoptrace.Trace{}, errors.Errorf("invalid trace string %s", v)
			}
		}
		if k.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataLink(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsString()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataLink(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_String:
		registeredAttribute, err := x.attributeRegistry.ConstructStringAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v string) (string, error) { return v, nil }
		if k.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataString(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsString()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataString(registeredAttribute, decoded)
		}
	case xopproto.AttributeType_Time:
		registeredAttribute, err := x.attributeRegistry.ConstructTimeAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if k.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		decoder := func(v string) (time.Time, error) { return time.Parse(time.RFC3339Nano, v) }
		if k.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				decoded, err := decoder(value)
				if err != nil {
					return err
				}
				x.datum.baseSpan.MetadataTime(registeredAttribute, decoded)
			}
		} else {
			value := a.Value.AsString()
			decoded, err := decoder(value)
			if err != nil {
				return err
			}
			x.datum.baseSpan.MetadataTime(registeredAttribute, decoded)
		}

	default:
		return errors.Errorf("unexpected attribute type %s", aDef.AttributeType)
	}
	return nil
}

func (e *spanExporter) Shutdown(ctx context.Context) error {
	// XXX
	return nil
}

type unhack struct {
	next sdktrace.SpanExporter
}

// NewUnhacker wraps a SpanExporter and if the input is from BaseLogger or SpanLog,
// then it "fixes" the data hack in the output that puts inter-span links in sub-spans
// rather than in the span that defined them.
func NewUnhacker(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return &unhack{next: exporter}
}

func (u *unhack) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// TODO: fix up SpanKind if spanKind is one of the attributes
	id2Index := makeIndex(spans)
	subLinks := make([][]sdktrace.Link, len(spans))
	for i, span := range spans {
		parentIndex, ok := lookupParent(id2Index, span)
		if !ok {
			continue
		}
		var addToParent bool
		for _, attribute := range span.Attributes() {
			switch attribute.Key {
			case spanIsLinkAttributeKey, spanIsLinkEventKey:
				spans[i] = nil
				addToParent = true
			}
		}
		if !addToParent {
			continue
		}
		subLinks[parentIndex] = append(subLinks[parentIndex], span.Links()...)
	}
	n := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for i, span := range spans {
		span := span
		switch {
		case len(subLinks[i]) > 0:
			n = append(n, wrappedReadOnlySpan{
				ReadOnlySpan: span,
				links:        append(list.Copy(span.Links()), subLinks[i]...),
			})
		case span == nil:
			// skip
		default:
			n = append(n, span)
		}
	}
	return u.next.ExportSpans(ctx, n)
}

func (u *unhack) Shutdown(ctx context.Context) error {
	return u.next.Shutdown(ctx)
}

type wrappedReadOnlySpan struct {
	sdktrace.ReadOnlySpan
	links []sdktrace.Link
}

var _ sdktrace.ReadOnlySpan = wrappedReadOnlySpan{}

func (w wrappedReadOnlySpan) Links() []sdktrace.Link {
	return w.links
}

func makeIndex(spans []sdktrace.ReadOnlySpan) map[oteltrace.SpanID]int {
	id2Index := make(map[oteltrace.SpanID]int)
	for i, span := range spans {
		spanContext := span.SpanContext()
		if spanContext.HasSpanID() {
			id2Index[spanContext.SpanID()] = i
		}
	}
	return id2Index
}

func lookupParent(id2Index map[oteltrace.SpanID]int, span sdktrace.ReadOnlySpan) (int, bool) {
	parent := span.Parent()
	if !parent.HasSpanID() {
		return 0, false
	}
	parentIndex, ok := id2Index[parent.SpanID()]
	if !ok {
		return 0, false
	}
	return parentIndex, true
}

func makeSubspans(id2Index map[oteltrace.SpanID]int, spans []sdktrace.ReadOnlySpan) ([][]oteltrace.SpanID, []int) {
	ss := make([][]oteltrace.SpanID, len(spans))
	noParent := make([]int, 0, len(spans))
	for i, span := range spans {
		parentIndex, ok := lookupParent(id2Index, span)
		if !ok {
			noParent = append(noParent, i)
		}
		ss[parentIndex] = append(ss[parentIndex], i)
	}
	return ss, noParent
}

func buildSourceInfo(span sdktrace.ReadOnlySpan, attributeMap AttributeMap) {
	var si xopbase.SourceInfo
	var source string
	// XXX grab namespace from scope instead
	if s := attributeMap.GetString(xopSource); s != "" {
		source = s
	} else if n := span.InstrumentationScope().Name; n != "" {
		if v := span.InstrumentationScope().Version; v != "" {
			source = n + " " + v
		} else {
			source = n
		}
	} else {
		source = "OTEL"
	}
	namespace := defaulted(attributeMap.GetString(xopNamespace), source)
	si.Source, si.SourceVersion = version.SplitVersion(source)
	si.Namespace, si.NamespaceVersion = version.SplitVersion(namespace)
	return si
}
