package xopoteltest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// SpanStub from https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v1.14.0/trace/tracetest#SpanStub because
// it doesn't implement UnmarshalJSON. Why not?
type SpanStub struct {
	Name              string
	SpanContext       SpanContext
	Parent            SpanContext
	SpanKind          oteltrace.SpanKind
	StartTime         time.Time
	EndTime           time.Time
	Attributes        []attribute.KeyValue
	Events            []sdktrace.Event
	Links             []Link
	Status            sdktrace.Status
	DroppedAttributes int
	DroppedEvents     int
	DroppedLinks      int
	ChildSpanCount    int
	Resource          any
	Scope             instrumentation.Scope `json:"InstrumentationLibrary"`
}

func (s SpanStub) String() string { return fmt.Sprintf("span %s - %s", s.Name, s.SpanContext) }

// SpanContext copied from https://github.com/open-telemetry/opentelemetry-go/blob/2e54fbb3fede5b54f316b3a08eab236febd854e0/trace/trace.go#L290
// because it doesn't implement UnmarshalJSON. Why not?
type SpanContext struct {
	oteltrace.SpanContext
}

func (sc SpanContext) String() string {
	return fmt.Sprintf("00-%s-%s-%s", sc.TraceID(), sc.SpanID(), sc.TraceFlags())
}

func (sc *SpanContext) UnmarshalJSON(i []byte) error {
	var tmp struct {
		TraceID    TraceID
		SpanID     SpanID
		TraceFlags TraceFlags
		TraceState TraceState
		Remote     bool
	}
	err := json.Unmarshal(i, &tmp)
	if err != nil {
		return err
	}
	scc := oteltrace.SpanContextConfig{
		TraceID:    tmp.TraceID.TraceID,
		SpanID:     tmp.SpanID.SpanID,
		TraceFlags: tmp.TraceFlags.TraceFlags,
		TraceState: tmp.TraceState.TraceState,
		Remote:     tmp.Remote,
	}
	sc.SpanContext = oteltrace.NewSpanContext(scc)
	return nil
}

// Link copied from https://pkg.go.dev/go.opentelemetry.io/otel/trace#Link because
// it doesn't implement UnmarshalJSON. Why not?
type Link struct {
	SpanContext           SpanContext
	Attributes            []attribute.KeyValue
	DroppedAttributeCount int
}

// SpanID exists because oteltrace.SpanID doesn't implement UnmarshalJSON
type SpanID struct {
	oteltrace.SpanID
}

func (s *SpanID) UnmarshalText(h []byte) error {
	return decode(s.SpanID[:], h)
}

// TraceID exists because oteltrace.TraceID doesn't implement UnmarshalJSON
type TraceID struct {
	oteltrace.TraceID
}

func (t *TraceID) UnmarshalText(h []byte) error { return decode(t.TraceID[:], h) }

func decode(s []byte, h []byte) error {
	b, err := hex.DecodeString(string(h))
	if err != nil {
		return err
	}
	if len(b) != len(s) {
		return fmt.Errorf("wrong length")
	}
	copy(s[:], b)
	return nil
}

type TraceFlags struct {
	oteltrace.TraceFlags
}

func (tf *TraceFlags) UnmarshalText(h []byte) error {
	var a [1]byte
	err := decode(a[:], h)
	if err != nil {
		return err
	}
	tf.TraceFlags = oteltrace.TraceFlags(a[0])
	return nil
}

type TraceState struct {
	oteltrace.TraceState
}

func (ts *TraceState) UnmarshalText(i []byte) error {
	s, err := oteltrace.ParseTraceState(string(i))
	if err != nil {
		return err
	}
	ts.TraceState = s
	return nil
}
