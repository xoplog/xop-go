// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"bytes"
	"context"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

func (log *Logger) Replay(ctx context.Context, input any, logger xopbase.Logger) error {
	return log.LosslessReplay(ctx, input, logger)
}

func (_ *Logger) LosslessReplay(ctx context.Context, input any, logger xopbase.Logger) error {
	trace, ok := input.(*xopproto.Trace)
	if !ok {
		return errors.Errorf("expected *xopproto.Trace for xoppb.Replay, got %T", input)
	}
	for _, request := range trace.Requests {
		err := replayRequest{
			logger:       logger,
			requestInput: request,
			traceID:      xoptrace.NewHexBytes16FromSlice(trace.TraceID),
			registry:     xopat.NewRegistry(false),
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type replayRequest struct {
	logger       xopbase.Logger
	requestInput *xopproto.Request
	traceID      xoptrace.HexBytes16
	request      xopbase.Request
	spansSeen    map[xoptrace.HexBytes8]int32
	registry     *xopat.Registry
}

func (x replayRequest) Replay(ctx context.Context) error {
	requestID := xoptrace.NewHexBytes8FromSlice(x.requestInput.RequestID)
	if len(input.Spans) == 0 {
		return errors.Errorf("expected >0 spans in request (%s-%s)", x.traceID, requestID)
	}
	if !bytes.Equal(x.requestInput.RequestID, x.requestInput.Spans[0].SpanID) {
		return errors.Errorf("expected first span of request (%s-%s) to be the request", x.traceID, requestID)
	}
	requestSpan := x.requestInput.Spans[0]
	var bundle xoptrace.Bundle
	bundle.Trace.TraceID().Set(traceID)
	bundle.Trace.SpanID().Set(requestID)
	bundle.Parent.SpanID().SetBytes(requestSpan.ParentID)
	if len(input.ParentTraceID) != 0 {
		bundle.Parent.TraceID().SetBytes(x.requestInput.ParentTraceID)
	} else {
		bundle.Parent.TraceID().Set(traceID)
	}
	bundle.State.SetString(requestSpan.TraceState)
	bundle.Baggage.SetString(requestSpan.Baggage)

	sourceInfo := xopbase.SourceInfo{
		Source:    x.requestInput.SourceID,
		Namespace: x.requestInput.SourceNamespace,
	}
	var err error
	sourceInfo.SourceVersion, err = semver.StrictNewVersion(x.requestInput.SourceVersion)
	if err != nil {
		return errors.Errorf("invalid source version in request (%s-%s): %w", x.traceID, requestID, err)
	}
	sourceInfo.NamespaceVersion, err = semver.StrictNewVersion(input.SourceNamespaceVersion)
	if err != nil {
		return errors.Errorf("invalid namespace version in request (%s-%s): %w", x.traceID, requestID, err)
	}

	x.request = logger.Request(ctx,
		time.Unix(0, requestSpan.StartTime),
		bundle,
		requestSpan.Name,
		sourceInfo)
	x.spansSeen = make(map[xoptrace.HexBytes8]int32)
	err = replaySpan{
		replayRequest: x,
		spanInput:     requestSpan,
	}.Replay(ctx)
	if err != nil {
		return err
	}
	for i := len(input.Spans) - 1; i > 0; i-- { // 0 is processed above
		err = replaySpan{
			replayRequest: x,
			spanInput:     input.Spans[i],
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type replaySpan struct {
	replayRequest
	spanInput *xopproto.Span
}

func (x replaySpan) Replay(ctx context.Context) error {
	spanID := xoptrace.NewHexBytes8FromSlice(x.spanInput.SpanID)
	if version, ok := spansSeen[spanID]; ok && version > x.spanInput.Version {
		return nil
	}
	spansSeen[spanID] = x.spanInput.Version
	var bundle xoptrace.Bundle
	bundle.Trace.TraceID().Set(x.traceID)
	bundle.Trace.SpanID().Set(spanID)
	bundle.Parent.SpanID().SetBytes(x.spanInput.ParentID)
	bundle.Parent.TraceID().Set(traceID)
	span := request.Span(ctx,
		time.Unix(0, x.spanInput.StartTime),
		bundle,
		x.spanInput.Name,
		x.spanInput.SequenceCode)
	for _, attribute := range x.spanInput.Attributes {
		attr := x.replayAttribute(attribute)
	}
	if x.spanInput.endTime != nil {
		span.Done(time.Unix(0, *x.spanInput.EndTime), false)
	}
}

func (x replaySpan) getAttribute(attribute *xopproto.SpanAttribute) error {
	def := requestInput.AttributeDefinitions[attribute.AttributeDefinitionSequenceNumber]
	m := Make{
		Key:         def.Key,
		Description: def.Description,
		Namespace:   def.Namespace + " " + def.NamespaceSemver,
		Indexed:     def.ShouldIndex,
		Prominence:  def.Prominence,
		Multiple:    def.Multiple,
		Distinct:    def.Distinct,
		// XXX Ranged      :
		Locked: def.Locked,
	}
	switch xopat.AttributeType(def.AttributeType) {
	case xopat.AttributeTypeAny:
		registeredAttribute, err := x.registry.ConstructAnyAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataAny(&registeredAttribute, xopbase.Model{
				TypeName: v.StringValue,
				Encoded:  v.BytesValue,
				Encoding: xopproto.Encoding(v.IntValue),
			})
		}
		return &a, nil
	case xopat.AttributeTypeBool:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			var b bool
			if v.intValue != 0 {
				b = true
			}
			span.MetadataBool(&registeredAttribute, b)
		}
		return &a, nil
	case xopat.AttributeTypeDuration:
		registeredAttribute, err := x.registry.ConstructDurationAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataDuration(&registeredAttribute, time.Duration(v.IntValue))
		}
		return &a, nil
	case xopat.AttributeTypeEnum:
		registeredAttribute, err := x.registry.ConstructEnumAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
		}
		return &a, nil
	case xopat.AttributeTypeFloat32:
		registeredAttribute, err := x.registry.ConstructFloat32Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataFloat32(&registeredAttribute, float32(v.FloatValue))
		}
		return &a, nil
	case xopat.AttributeTypeFloat64:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataFloat64(&registeredAttribute, v.FloatValue)
		}
		return &a, nil
	case xopat.AttributeTypeInt16:
		registeredAttribute, err := x.registry.ConstructInt16Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataInt16(&registeredAttribute, int16(v.IntValue))
		}
		return &a, nil
	case xopat.AttributeTypeInt32:
		registeredAttribute, err := x.registry.ConstructInt32Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataInt32(&registeredAttribute, int32(v.IntValue))
		}
		return &a, nil
	case xopat.AttributeTypeInt64:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataInt64(&registeredAttribute, v.IntValue)
		}
		return &a, nil
	case xopat.AttributeTypeInt8:
		registeredAttribute, err := x.registry.ConstructInt8Attribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataInt8(&registeredAttribute, int8(v.IntValue))
		}
		return &a, nil
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
			span.MetadataLink(&registeredAttribute, t)
		}
		return &a, nil
	case xopat.AttributeTypeString:
		registeredAttribute, err := x.registry.ConstructStringAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataString(&registeredAttribute, v.StringValue)
		}
		return &a, nil
	case xopat.AttributeTypeTime:
		registeredAttribute, err := x.registry.ConstructTimeAttribute(m)
		if err != nil {
			return err
		}
		for _, v := range attribute.Values {
			span.MetadataTime(&registeredAttribute, time.Unix(0, v.IntValue))
		}
		return &a, nil

	default:
		return nil, errors.Errorf("unexpected attribute type %s", def.AttributeType)
	}
}
