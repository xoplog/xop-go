package xopotel

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// this files contains copies that are required for compatability reasons

// this interface defintion copied from OTEL
type ReadOnlySpan interface {
	Name() string
	SpanContext() oteltrace.SpanContext
	Parent() oteltrace.SpanContext
	SpanKind() oteltrace.SpanKind
	StartTime() time.Time
	EndTime() time.Time
	Attributes() []attribute.KeyValue
	Links() []sdktrace.Link
	Events() []sdktrace.Event
	Status() sdktrace.Status
	InstrumentationScope() instrumentation.Scope
	InstrumentationLibrary() instrumentation.Library
	Resource() *resource.Resource
	DroppedAttributes() int
	DroppedLinks() int
	DroppedEvents() int
	ChildSpanCount() int
}
