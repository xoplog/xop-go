package xoptag

import (
	"net/http"
	"time"
)

// SpanKind comes from https://opentelemetry.io/docs/reference/specification/trace/api/#spankind
type SpanKind string

const (
	ServerSpanKind   SpanKind = "SERVER"   // server receiving a request
	ClientSpanKind   SpanKind = "CLIENT"   // making a request to a server
	ProducerSpanKind SpanKind = "PRODUCER" // initiates asynchronous request
	ConsumerSpanKind SpanKind = "CONSUMER" // handles asynchronous request
	InternalSpanKind SpanKind = ""         // child of one of the above, INTERNAL
)

type Span struct {
	Name      string    `json:"name"                     update:"replace" index:"true"`
	Kind      SpanKind  `json:"otel.span_kind,omitempty" update:"replace" index:"false"`
	StartTime time.Time `json:"startTime"                update:"ignore"  index:"true,ranged"`
	Duration  int64     `json:"duration"                 update:"replace" index:"true,ranged"`
}

type RequestType string

const (
	HTTP    RequestType = "http"
	CronJob RequestType = "cron"
)

type Request struct {
	Span
	Type RequestType `json:"type" update:"override" index:"true"`
}

type HTTPRequest struct {
	Request
	Method         string      `json:"http.method"          update:"replace" index:"true"`  // OTEL
	URL            string      `json:"http.url"             update:"replace" index:"true"`  // OTEL
	Host           string      `json:"http.host,omitempty"  update:"replace" index:"true"`  // OTEL
	StatusCode     int         `json:"http.status_code"     update:"replace" index:"true"`  // OTEL
	RequestHeaders http.Header `json:"http.request_headers" update:"replace" index:"false"` // violates OTEL
}
