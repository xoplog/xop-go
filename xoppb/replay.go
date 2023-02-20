// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"context"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

func (log *Logger) Replay(ctx context.Context, input any, logger xopbase.Logger) error {
	return log.LosslessReplay(ctx, input, logger)
}

type replayTrace struct {
	logger    xopbase.Logger
	traceID   xoptrace.HexBytes16
	spansSeen map[xoptrace.HexBytes8]spanData
}

func (_ *Logger) LosslessReplay(ctx context.Context, input any, logger xopbase.Logger) error {
	trace, ok := input.(*xopproto.Trace)
	if !ok {
		return errors.Errorf("expected *xopproto.Trace for xoppb.Replay, got %T", input)
	}
	x := replayTrace{
		logger:    logger,
		traceID:   xoptrace.NewHexBytes16FromSlice(trace.TraceID),
		spansSeen: make(map[xoptrace.HexBytes8]spanData),
	}
	for _, request := range trace.Requests {
		err := replayRequest{
			replayTrace:  &x,
			requestInput: request,
			registry:     xopat.NewRegistry(false),
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type spanData struct {
	version int32
	span    xopbase.Span
}

type replayRequest struct {
	*replayTrace
	requestInput *xopproto.Request
	request      xopbase.Request
	registry     *xopat.Registry
}

func (x replayRequest) Replay(ctx context.Context) error {
	requestID := xoptrace.NewHexBytes8FromSlice(x.requestInput.Span.SpanID)
	previous, ok := x.spansSeen[requestID]
	if ok && previous.version > x.requestInput.Span.Version {
		return nil
	}
	if ok {
		x.request = previous.span.(xopbase.Request)
	} else {
		var bundle xoptrace.Bundle
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.SpanID().Set(requestID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		bundle.Parent.SpanID().SetBytes(x.requestInput.Span.ParentID)
		bundle.Parent.Flags().SetBytes([]byte{1})
		if len(x.requestInput.ParentTraceID) != 0 {
			bundle.Parent.TraceID().SetBytes(x.requestInput.ParentTraceID)
		} else {
			bundle.Parent.TraceID().Set(x.traceID)
		}
		bundle.State.SetString(x.requestInput.TraceState)
		bundle.Baggage.SetString(x.requestInput.Baggage)

		sourceInfo := xopbase.SourceInfo{
			Source:    x.requestInput.SourceID,
			Namespace: x.requestInput.SourceNamespace,
		}
		var err error
		sourceInfo.SourceVersion, err = semver.StrictNewVersion(x.requestInput.SourceVersion)
		if err != nil {
			return errors.Wrapf(err, "invalid source version in request (%s-%s)", x.traceID, requestID)
		}
		sourceInfo.NamespaceVersion, err = semver.StrictNewVersion(x.requestInput.SourceNamespaceVersion)
		if err != nil {
			return errors.Wrapf(err, "invalid namespace version in request (%s-%s)", x.traceID, requestID)
		}

		x.request = x.logger.Request(ctx,
			time.Unix(0, x.requestInput.Span.StartTime),
			bundle,
			x.requestInput.Span.Name,
			sourceInfo)
	}
	x.spansSeen[requestID] = spanData{
		version: x.requestInput.Span.Version,
		span:    x.request,
	}
	err := replaySpan{
		parentSpan:    x.request,
		replayRequest: x,
		spanInput:     x.requestInput.Span,
	}.Replay(ctx, false)
	if err != nil {
		return err
	}

	for _, line := range x.requestInput.Lines {
		spanID := xoptrace.NewHexBytes8FromSlice(line.SpanID)
		span, ok := x.spansSeen[spanID]
		if !ok {
			return errors.Errorf("line references spanID (%s) that does not exist", spanID)
		}
		err := replayLine{
			replayRequest: x,
			span:          span.span,
			lineInput:     line,
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}

	if x.requestInput.Span.EndTime != nil {
		x.request.Done(time.Unix(0, *x.requestInput.Span.EndTime), false)
	}
	return nil
}

type replaySpan struct {
	replayRequest
	parentSpan xopbase.Span
	spanInput  *xopproto.Span
	span       xopbase.Span
}

func (x replaySpan) Replay(ctx context.Context, doDone bool) error {
	spanID := xoptrace.NewHexBytes8FromSlice(x.spanInput.SpanID)
	previous, ok := x.spansSeen[spanID]
	if ok && previous.version > x.spanInput.Version {
		return nil
	}
	if ok {
		x.span = previous.span
	} else {
		var bundle xoptrace.Bundle
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		bundle.Trace.SpanID().Set(spanID)
		bundle.Parent.SpanID().SetBytes(x.spanInput.ParentID)
		bundle.Parent.TraceID().Set(x.traceID)
		bundle.Parent.Flags().SetBytes([]byte{1})
		x.span = x.parentSpan.Span(ctx,
			time.Unix(0, x.spanInput.StartTime),
			bundle,
			x.spanInput.Name,
			x.spanInput.SequenceCode)
	}
	x.spansSeen[spanID] = spanData{
		version: x.spanInput.Version,
		span:    x.span,
	}

	for i := len(x.spanInput.Spans) - 1; i >= 0; i-- {
		err := replaySpan{
			replayRequest: x.replayRequest,
			parentSpan:    x.span,
			spanInput:     x.spanInput.Spans[i],
		}.Replay(ctx, true)
		if err != nil {
			return err
		}
	}
	for _, attribute := range x.spanInput.Attributes {
		err := x.replayAttribute(attribute)
		if err != nil {
			return err
		}
	}
	if x.spanInput.EndTime != nil && doDone {
		x.span.Done(time.Unix(0, *x.spanInput.EndTime), false)
	}
	return nil
}

func (x replaySpan) replayAttribute(attribute *xopproto.SpanAttribute) error {
	def := x.requestInput.AttributeDefinitions[attribute.AttributeDefinitionSequenceNumber]
	m := xopat.Make{
		Key:         def.Key,
		Description: def.Description,
		Namespace:   def.Namespace + " " + def.NamespaceSemver,
		Indexed:     def.ShouldIndex,
		Prominence:  int(def.Prominence),
		Multiple:    def.Multiple,
		Distinct:    def.Distinct,
		// XXX Ranged      :
		Locked: def.Locked,
	}
	switch xopat.AttributeType(def.Type).SpanAttributeType() {
	case xopat.AttributeTypeAny:
		registeredAttribute, err := x.registry.ConstructAnyAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataAny(registeredAttribute, xopbase.ModelArg{
				TypeName: v.StringValue,
				Encoded:  v.BytesValue,
				Encoding: xopproto.Encoding(v.IntValue),
			})
		}
		return nil
	case xopat.AttributeTypeBool:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			var b bool
			if v.IntValue != 0 {
				b = true
			}
			x.span.MetadataBool(registeredAttribute, b)
		}
		return nil
	case xopat.AttributeTypeEnum:
		registeredAttribute, err := x.registry.ConstructEnumAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			enum := registeredAttribute.Add64(v.IntValue, v.StringValue)
			x.span.MetadataEnum(&registeredAttribute.EnumAttribute, enum)
		}
		return nil
	case xopat.AttributeTypeFloat64:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataFloat64(registeredAttribute, v.FloatValue)
		}
		return nil
	case xopat.AttributeTypeInt64:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataInt64(registeredAttribute, v.IntValue)
		}
		return nil
	case xopat.AttributeTypeLink:
		registeredAttribute, err := x.registry.ConstructLinkAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			t, ok := xoptrace.TraceFromString(v.StringValue)
			if !ok {
				return errors.Errorf("invalid trace attribute '%s'", v.StringValue)
			}
			x.span.MetadataLink(registeredAttribute, t)
		}
		return nil
	case xopat.AttributeTypeString:
		registeredAttribute, err := x.registry.ConstructStringAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataString(registeredAttribute, v.StringValue)
		}
		return nil
	case xopat.AttributeTypeTime:
		registeredAttribute, err := x.registry.ConstructTimeAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			x.span.MetadataTime(registeredAttribute, time.Unix(0, v.IntValue))
		}
		return nil

	default:
		return errors.Errorf("unexpected attribute type %s", def.Type)
	}
}

type replayLine struct {
	replayRequest
	lineInput *xopproto.Line
	span      xopbase.Span
}

func (x replayLine) Replay(ctx context.Context) error {
	line := x.span.NoPrefill().Line(
		xopnum.Level(x.lineInput.LogLevel),
		time.Unix(0, x.lineInput.Timestamp),
		nil, // XXX todo
	)
	for _, attribute := range x.lineInput.Attributes {
		switch attribute.Type {
		case xopproto.AttributeType_Enum:
			m := xopat.Make{
				Key: attribute.Key,
			}
			ea, err := x.registry.ConstructEnumAttribute(m)
			if err != nil {
				return err
			}
			enum := ea.Add64(attribute.Value.IntValue, attribute.Value.StringValue)
			line.Enum(&ea.EnumAttribute, enum)
		case xopproto.AttributeType_Float64, xopproto.AttributeType_Float32:
			line.Float64(attribute.Key, attribute.Value.FloatValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Int64, xopproto.AttributeType_Int32, xopproto.AttributeType_Int16,
			xopproto.AttributeType_Int8, xopproto.AttributeType_Int:
			line.Int64(attribute.Key, attribute.Value.IntValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_String, xopproto.AttributeType_Error, xopproto.AttributeType_Stringer:
			line.String(attribute.Key, attribute.Value.StringValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Uint64, xopproto.AttributeType_Uint32, xopproto.AttributeType_Uint16,
			xopproto.AttributeType_Uint8, xopproto.AttributeType_Uintptr, xopproto.AttributeType_Uint:
			line.Uint64(attribute.Key, attribute.Value.UintValue, xopbase.DataType(attribute.Type))
		case xopproto.AttributeType_Any:
			line.Any(attribute.Key, xopbase.ModelArg{
				TypeName: attribute.Value.StringValue,
				Encoded:  attribute.Value.BytesValue,
				Encoding: xopproto.Encoding(attribute.Value.IntValue),
			})
		case xopproto.AttributeType_Bool:
			var b bool
			if attribute.Value.IntValue != 0 {
				b = true
			}
			line.Bool(attribute.Key, b)
		case xopproto.AttributeType_Duration:
			line.Duration(attribute.Key, time.Duration(attribute.Value.IntValue))
		case xopproto.AttributeType_Time:
			line.Time(attribute.Key, time.Unix(0, attribute.Value.IntValue))
		default:
			return errors.Errorf("unknown data type %s", attribute.Type)
		}
	}
	switch {
	case x.lineInput.Model != nil:
		line.Model(x.lineInput.Message, xopbase.ModelArg{
			TypeName: x.lineInput.Model.Type,
			Encoded:  x.lineInput.Model.Encoded,
			Encoding: x.lineInput.Model.Encoding,
		})
	case x.lineInput.Link != "":
		trace, ok := xoptrace.TraceFromString(x.lineInput.Link)
		if !ok {
			return errors.Errorf("invalid trace (%s)", x.lineInput.Link)
		}
		line.Link(x.lineInput.Message, trace)
	case x.lineInput.MessageTemplate != "":
		line.Template(x.lineInput.MessageTemplate)
	default:
		line.Msg(x.lineInput.Message)
	}
	return nil
}
