package xopconst

// The descriptions are lifted from https://opentelemetry.io/ and are thus Copyright(c)
// the Open Telementry authors.

var SpanKind = Make{Key: "span.kind", Namespace: "OTAP", Indexed: true, Prominence: 30,
	Description: "https://opentelemetry.io/docs/reference/specification/trace/api/#spankind" +
		" Use one of SpanKindServer, SpanKindClient, SpanKindProducer, SpanKindConsumer, SpanKindInternal"}.
	EnumAttribute(SpanKindServer)

//go:generate enumer -type=SpanKindEnum -linecomment -json -sql
type SpanKindEnum int

const (
	SpanKindServer   SpanKindEnum = iota // SERVER
	SpanKindClient                       // CLIENT
	SpanKindProducer                     // PRODUCER
	SpanKindConsumer                     // CONSUMER
	SpanKindInternal                     // INTERNAL
)

func (i SpanKindEnum) Int64() int64 { return int64(i) }

var HTTPMethod = Make{Key: "http.method", Namespace: "OTAP", Indexed: true, Prominence: 10,
	Description: "HTTP request method"}.StrAttribute()

var URL = Make{Key: "http.url", Namespace: "OTAP", Indexed: true, Prominence: 12,
	Description: "Full HTTP request URL in the form scheme://host[:port]/path?query[#fragment]." +
		" Usually the fragment is not transmitted over HTTP, but if it is known," +
		" it should be included nevertheless"}.StrAttribute()

var HTTPTarget = Make{Key: "http.target", Namespace: "OTAP", Indexed: true, Prominence: 25,
	Description: "The full request target as passed in a HTTP request line or equivalent"}.StrAttribute()

var HTTPHost = Make{Key: "http.host", Namespace: "OTAP", Indexed: true, Prominence: 45,
	Description: "The value of the HTTP host header. An empty Host header should also be reported"}.StrAttribute()

var HTTPStatusCode = Make{Key: "http.status_code", Namespace: "OTAP", Indexed: true, Prominence: 5,
	Description: "HTTP response status code"}.IntAttribute()

var TraceResponse = Make{Key: "http.response.header.traceresponse", Namespace: "OTAP", Indexed: true, Prominence: 50,
	Description: "Response 'traceresponse' heeader received"}.StrAttribute()
