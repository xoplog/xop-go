package xopconst

import (
	"github.com/xoplog/xop-go/xopat"
)

// The descriptions are lifted from https://opentelemetry.io/ and are thus Copyright(c)
// the Open Telementry authors.

var SpanKind = xopat.Make{Key: "span.kind", Namespace: "OTEL", Indexed: true, Prominence: 30,
	Locked: true,
	Description: "https://opentelemetry.io/docs/reference/specification/trace/api/#spankind" +
		" Use one of SpanKindServer, SpanKindClient, SpanKindProducer, SpanKindConsumer, SpanKindInternal"}.
	EnumAttribute(SpanKindServer)

//go:generate enumer -type=SpanKindEnum -linecomment -json -sql
type SpanKindEnum int

// These values are identical to the OTEL values.
// see https://pkg.go.dev/go.opentelemetry.io/otel/trace#SpanKind
// They can be cast to oteltrace.SpanKind

const (
	SpanKindUnspecified SpanKindEnum = iota // UNSPECIFIED
	SpanKindInternal                        // INTERNAL
	SpanKindServer                          // SERVER
	SpanKindClient                          // CLIENT
	SpanKindProducer                        // PRODUCER
	SpanKindConsumer                        // CONSUMER
)

func (i SpanKindEnum) Int64() int64 { return int64(i) }

var HTTPMethod = xopat.Make{Key: "http.method", Namespace: "OTEL", Indexed: true, Prominence: 10,
	Description: "HTTP request method"}.StringAttribute()

var URL = xopat.Make{Key: "http.url", Namespace: "OTEL", Indexed: true, Prominence: 12,
	Description: "Full HTTP request URL in the form scheme://host[:port]/path?query[#fragment]." +
		" Usually the fragment is not transmitted over HTTP, but if it is known," +
		" it should be included nevertheless"}.StringAttribute()

var HTTPTarget = xopat.Make{Key: "http.target", Namespace: "OTEL", Indexed: true, Prominence: 25,
	Description: "The full request target as passed in a HTTP request line or equivalent"}.StringAttribute()

var HTTPHost = xopat.Make{Key: "http.host", Namespace: "OTEL", Indexed: true, Prominence: 45,
	Description: "The value of the HTTP host header. An empty Host header should also be reported"}.StringAttribute()

var HTTPStatusCode = xopat.Make{Key: "http.status_code", Namespace: "OTEL", Indexed: true, Prominence: 5,
	Description: "HTTP response status code"}.IntAttribute()

var TraceResponse = xopat.Make{Key: "http.response.header.traceresponse", Namespace: "OTEL", Indexed: true, Prominence: 50,
	Description: "Response 'traceresponse' heeader received"}.StringAttribute()
