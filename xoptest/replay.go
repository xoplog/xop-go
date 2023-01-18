package xoptest

import (
	"context"

	"github.com/pkg/errors"
)

func (log *TestLogger) Replay(ctx context.Context, input any, logger Logger) error {
	return log.LosslessReplay(ctx, input, logger)
}

func (log *TestLogger) LosslessReplay(ctx context.Context, input any, logger Logger) error {
	switch v := input.(type) {
	case *TestLogger:
		for _, request := range v.Requests {
			err := log.LosslessReplay(ctx, request, logger)
			if err != nil {
				return err
			}
		}
		return nil
	case *Span:
		if !v.IsRequest {
			return errors.Errorf("replay must start with a *xoptest.Span that is a request or a *xoptest.TestLogger")
		}
		request := logger.Request(ctx, v.StartTime, v.Bundle, v.Name)
		for _, inputSpan := range request.Spans {
			span := request.Span
func (span *Span) Span(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, name string, spanSequenceCode string) xopbase.Span {

		request.Done(v.EndTime, time.Unix(0, v.EndTime))
		request.Flush()
	default:
		return errors.Errorf("invalid input type for xoptest.Replay: %T", input)
	}
}

