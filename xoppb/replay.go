package xoppb

import (
	"bytes"
	"context"
	"time"

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
			logger: logger,
			requestInput: request,
			traceID: xoptrace.NewHexBytes16FromSlice(trace.TraceID),
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type replayRequest struct {
	logger xopbase.Logger
	requestInput *xopproto.Request
	traceID xoptrace.HexBytes16
	request xopbase.Request
	spansSeen map[xoptrace.HexBytes8]int32
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
		spanInput: requestSpan,
	}.Replay(ctx)
	if err != nil {
		return err
	}
	for i := len(input.Spans) - 1; i > 0; i-- { // 0 is processed above
		err = replaySpan{
			replayRequest: x,
			spanInput: input.Spans[i],
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
		time.Unix(0, x.spanInput.StartTime)
		bundle,
		x.spanInput.Name,
		x.spanInput.SequenceCode)
	for _, attribute := range x.spanInput.Attributes {
		// attribute.AttributeDefinitionSequenceNumber
	// repeated AttributeValue values = 2; // at least one is required

	}
	if x.spanInput.endTime != nil {
		span.Done(time.Unix(0, *x.spanInput.EndTime), false)
	}
}
