package xopotel

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type otelSpan struct {
	name                 string
	spanContext          oteltrace.SpanContext
	parentContext        oteltrace.SpanContext
	spanKind             oteltrace.SpanKind
	startTime            time.Time
	endTime              time.Time
	attributes           []attribute.KeyValue
	attributeMap         map[attribute.Key]int
	links                oteltrace.Link
	events               []oteltrace.Event
	status               oteltrace.Status
	instrumentationScope instrumentation.Scope
	resource             *resource.Resource
	droppedAttributes    int
	droppedLinks         int
	droppedEvents        int
	childSpanCount       int
}

var _ sdktrace.ReadOnlySpan = otelSpan{}

func (s otelSpan) Name() string                                    { return s.name }
func (s otelSpan) SpanContext() oteltrace.SpanContext              { return s.spanContext }
func (s otelSpan) Parent() oteltrace.SpanContext                   { return s.parentContext }
func (s otelSpan) SpanKind() oteltrace.SpanKind                    { return s.spanKind }
func (s otelSpan) StartTime() time.Time                            { return s.startTime }
func (s otelSpan) EndTime() time.Time                              { return s.endTime }
func (s otelSpan) Attributes() []attribute.KeyValue                { return s.attributes }
func (s otelSpan) Links() oteltrace.Link                           { return s.links }
func (s otelSpan) Events() []oteltrace.Event                       { return s.events }
func (s otelSpan) Status() oteltrace.Status                        { return s.status }
func (s otelSpan) Resource() *resource.Resource                    { return s.resource }
func (s otelSpan) DroppedAttributes() int                          { return s.droppedAttributes }
func (s otelSpan) DroppedLinks() int                               { return s.droppedLinks }
func (s otelSpan) DroppedEvents() int                              { return s.droppedEvents }
func (s otelSpan) ChildSpanCount() int                             { return s.childSpanCount }
func (s otelSpan) InstrumentationScope() instrumentation.Scope     { return s.instrumentationScope }
func (s otelSpan) InstrumentationLibrary() instrumentation.Library { return s.instrumentationScope }

func (s *otelSpan) SetAttributes(attributes []attribute.KeyValue) {
	for _, a := range attributes {
		if i, ok := s.attributeMap[a.Key]; ok {
			s.attributes[i] = a
			continue
		}
		attributeMap[a.Key] = len(s.attributes)
		s.attributes = append(s.attributes, a)
	}
}

func (s *otelSpan) end(ts time.Time) {
	s.endTime = ts
	// XXX send it
}

func (s *otelSpan) addEvent(name string, ts time.Time, attributes []attribute.KeyValue) {
	s.events = append(s.events, stktrace.Event{
		Name:                  name,
		Time:                  ts,
		Attributes:            attributes,
		DroppedAttributeCount: 0,
	})
}

type otelSpanWrap interface {
	IsRecording() bool
	SpanContext() oteltrace.SpanContext
	SetAttributes([]attribute.KeyValue)
	// ignoreDone()
	end(time.Time)
	addEvent(string, time.Time, []attribute.KeyValue)
}

var _ otelSpanWrap = &otelSpan{}
var _ otelSpanWrap = wrappedSpan{}

type wrappedSpan struct {
	oteltrace.Span
}

func (w wrappedSpan) end(ts time.Time) { span.otelSpan.End(oteltrace.WithTimestamp(ts)) }
func (w wrappedSpan) addEvent(name string, ts time.Time, attributes []attribute.KeyValue) {
	w.AddEvent(line.level.String(),
		oteltrace.WithTimestamp(line.timestamp),
		oteltrace.WithAttributes(line.attributes...))
}
