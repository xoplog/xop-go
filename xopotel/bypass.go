package xopotel

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type outputSpan struct {
	name                 string
	spanContext          oteltrace.SpanContext
	parentContext        oteltrace.SpanContext
	spanKind             oteltrace.SpanKind
	startTime            time.Time
	endTime              time.Time
	attributes           []attribute.KeyValue
	attributeMap         map[attribute.Key]int
	links                []sdktrace.Link
	events               []sdktrace.Event
	status               sdktrace.Status
	instrumentationScope instrumentation.Scope
	resource             *resource.Resource
	droppedAttributes    int
	droppedLinks         int
	droppedEvents        int
	childSpanCount       int
}

var _ ReadOnlySpan = outputSpan{}

func (s outputSpan) Name() string                                    { return s.name }
func (s outputSpan) SpanContext() oteltrace.SpanContext              { return s.spanContext }
func (s outputSpan) Parent() oteltrace.SpanContext                   { return s.parentContext }
func (s outputSpan) SpanKind() oteltrace.SpanKind                    { return s.spanKind }
func (s outputSpan) StartTime() time.Time                            { return s.startTime }
func (s outputSpan) EndTime() time.Time                              { return s.endTime }
func (s outputSpan) Attributes() []attribute.KeyValue                { return s.attributes }
func (s outputSpan) Links() []sdktrace.Link                          { return s.links }
func (s outputSpan) Events() []sdktrace.Event                        { return s.events }
func (s outputSpan) Status() sdktrace.Status                         { return s.status }
func (s outputSpan) Resource() *resource.Resource                    { return s.resource }
func (s outputSpan) DroppedAttributes() int                          { return s.droppedAttributes }
func (s outputSpan) DroppedLinks() int                               { return s.droppedLinks }
func (s outputSpan) DroppedEvents() int                              { return s.droppedEvents }
func (s outputSpan) ChildSpanCount() int                             { return s.childSpanCount }
func (s outputSpan) InstrumentationScope() instrumentation.Scope     { return s.instrumentationScope }
func (s outputSpan) InstrumentationLibrary() instrumentation.Library { return s.instrumentationScope }

func (s *outputSpan) IsRecording() bool { return true }
func (s *outputSpan) SetAttributes(attributes ...attribute.KeyValue) {
	for _, a := range attributes {
		if i, ok := s.attributeMap[a.Key]; ok {
			s.attributes[i] = a
			continue
		}
		s.attributeMap[a.Key] = len(s.attributes)
		s.attributes = append(s.attributes, a)
	}
}

func (s *outputSpan) end(ts time.Time) {
	s.endTime = ts
	// XXX send it
}

func (s *outputSpan) addEvent(name string, ts time.Time, attributes []attribute.KeyValue) {
	s.events = append(s.events, sdktrace.Event{
		Name:                  name,
		Time:                  ts,
		Attributes:            attributes,
		DroppedAttributeCount: 0,
	})
}

type canSetAttributes interface {
	SetAttributes(...attribute.KeyValue)
}
type otelSpanWrap interface {
	canSetAttributes
	IsRecording() bool
	SpanContext() oteltrace.SpanContext
	// ignoreDone()
	end(time.Time)
	addEvent(string, time.Time, []attribute.KeyValue)
}

var _ otelSpanWrap = &outputSpan{}
var _ otelSpanWrap = wrappedSpan{}

type wrappedSpan struct {
	oteltrace.Span
}

func (w wrappedSpan) end(ts time.Time) { w.End(oteltrace.WithTimestamp(ts)) }
func (w wrappedSpan) addEvent(name string, ts time.Time, attributes []attribute.KeyValue) {
	w.AddEvent(name, oteltrace.WithTimestamp(ts), oteltrace.WithAttributes(attributes...))
}
