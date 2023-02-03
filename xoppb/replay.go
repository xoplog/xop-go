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
		traceID := xoptrace.NewHexBytes16FromSlice(trace.TraceID)
		err := replayRequest(ctx, traceID, request, logger)
		if err != nil {
			return err
		}
	}
	return nil
}

func replayRequest(ctx context.Context, traceID xoptrace.HexBytes16, input *xopproto.Request, logger xopbase.Logger) error {
	requestID := xoptrace.NewHexBytes8FromSlice(input.RequestID)
	if len(input.Spans) == 0 {
		return errors.Errorf("expected >0 spans in request (%s-%s)", traceID, requestID)
	}
	if !bytes.Equal(input.RequestID, input.Spans[0].SpanID) {
		return errors.Errorf("expected first span of request (%s-%s) to be the request", traceID, requestID)
	}
	requestSpan := input.Spans[0]
	var bundle xoptrace.Bundle
	bundle.Trace.TraceID().Set(traceID)
	bundle.Trace.SpanID().Set(requestID)
	bundle.Parent.SpanID().SetBytes(requestSpan.ParentID)
	if len(input.ParentTraceID) != 0 {
		bundle.Parent.TraceID().SetBytes(input.ParentTraceID)
	} else {
		bundle.Parent.TraceID().Set(traceID)
	}
	bundle.State.SetString(requestSpan.TraceState)
	bundle.Baggage.SetString(requestSpan.Baggage)

	sourceInfo := xopbase.SourceInfo{
		Source:    input.SourceID,
		Namespace: input.SourceNamespace,
	}
	var err error
	sourceInfo.SourceVersion, err = semver.StrictNewVersion(input.SourceVersion)
	if err != nil {
		return errors.Errorf("invalid source version in request (%s-%s): %w", traceID, requestID, err)
	}
	sourceInfo.NamespaceVersion, err = semver.StrictNewVersion(input.SourceNamespaceVersion)
	if err != nil {
		return errors.Errorf("invalid namespace version in request (%s-%s): %w", traceID, requestID, err)
	}

	request := logger.Request(ctx,
		time.Unix(0, requestSpan.StartTime),
		bundle,
		requestSpan.Name,
		sourceInfo)
	spansSeen := make(map[xoptrace.HexBytes8]int32)
	err = replaySpan(ctx, spansSeen, requestSpan)
	if err != nil {
		return err
	}
	for i := len(input.Spans) - 1; i > 0; i-- { // 0 is processed above
		err := replaySpan(ctx, spansSeen, input.Spans[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func replaySpan(ctx context.Context, spansSeen map[xoptrace.HexBytes8]struct{}, span *xopproto.Span) error {
	spanID := xoptrace.NewHexBytes8FromSlice(span.SpanID)
	if version, ok := spansSeen[spanID]; ok && version > span.Version {
		return nil
	}
	spansSeen[spanID] = span.Version

}
