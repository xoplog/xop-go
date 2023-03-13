// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"

	"github.com/xoplog/xop-go/xopbase"

	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type spanExporter struct {
	base xopbase.Logger
}

func NewExporter(base xopbase.Logger) sdktrace.SpanExporter {
	return &spanExporter{base: base}
}

func (e *spanExporter) ExportSpans(ctx context.Context, spans []oteltrace.ReadOnlySpan) error {
	// XXX
	return nil
}

func (e *spanExporter) Shutdown(ctx context.Context) error {
	// XXX
	return nil
}
