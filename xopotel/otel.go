// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

/*
Package xopotel provides a gateway from xop into open telemetry
using OTEL's top-level APIs.

This gateway can be used either as a base layer for xop allowing
xop to output through OTEL; or it can be used to bridge the gap
between an application that is otherwise using OTEL and a library
that is expects to be provided with a xop logger.
*/
package xopotel

import (
	"context"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"

	"github.com/google/uuid"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type BaseLogger struct {
	tracer oteltrace.Tracer
	id     string
}

// var _ xopbase.Logger = &BaseLogger{}

func OTELSeed(topCtx context.Context, tracer oteltrace.Tracer) xop.SeedModifier {
	return xop.WithReactive(func(ctx context.Context, seed xop.Seed, selfIndex int, nameOrDescription string, isChildSpan bool) xop.Seed {
		if ctx == nil {
			ctx = topCtx
		}
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
	})
	// ctx, span := logger.Start(context.Background(), description, opts ...SpanStartOption)
}

func NewLogger(tracer oteltrace.Tracer) *BaseLogger {
	return &BaseLogger{
		tracer: tracer,
		id:     "otel-" + uuid.New().String(),
	}
}

func (logger *BaseLogger) ID() string           { return logger.id }
func (logger *BaseLogger) ReferencesKept() bool { return true }
func (logger *BaseLogger) Buffered() bool       { return false }

type Request interface{}

func (logger *BaseLogger) Request(ts time.Time, span trace.Bundle, description string) Request {
	// ctx, span := logger.Start(context.Background(), description, opts ...SpanStartOption)
	// WithSpanKind (internal, etc)
	return nil
}
