package xopjs

import (
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"

	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type JSONWriter struct {
	receiver   func(Bundle) error
	attributes map[attributeKey]*AttributeDefinition
	mu         sync.Mutex
}

type attributeKey struct {
	Namespace        string
	NamespaceVersion string
	Key              string
}

type WriteRequest struct {
	Request
	Bundle Bundle
	Writer *JSONWriter
}

// NewJSONWriter hands over data to the recevier each time there
// is a sync.  The data includes many byte slices (json.RawMessage).
// Those all become invalid when the recevier returns so if they
// are to be kept, Bundle.Copy() should be called.
func NewJSONWriter(receiver func(Bundle) error) *JSONWriter {
	return &JSONWriter{
		receiver:   receiver,
		attributes: make(map[attributeKey]*AttributeDefinition),
	}
}

var _ xopbytes.BytesWriter = &JSONWriter{}

func (w *JSONWriter) Buffered() bool { return true }
func (w *JSONWriter) Close()         {}

func (w *JSONWriter) DefineAttribute(attribute *xopat.Attribute) {
	key := attributeKey{
		Namespace:        attribute.Namespace(),
		NamespaceVersion: attribute.SemverString(),
		Key:              attribute.Key(),
	}
	a := &AttributeDefinition{
		Make: xopat.Make{
			Key:         key.Key,
			Description: attribute.Description(),
			Namespace:   key.Namespace,
			Indexed:     attribute.Indexed(),
			Prominence:  attribute.Prominence(),
			Multiple:    attribute.Multiple(),
			Distinct:    attribute.Distinct(),
			Ranged:      attribute.Ranged(),
			Locked:      attribute.Locked(),
		},
		Type:            attribute.ProtoType(),
		RecordType:      "atdef",
		NamespaceSemver: key.NamespaceVersion,
		// EnumValues      map[string]int64       `json:"enum,omitempty"`
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.attributes[key] = a
}

func (w *JSONWriter) DefineEnum(*xopat.EnumAttribute, xopat.Enum) {} // XXX

func (w *JSONWriter) Request(request xopbytes.Request) xopbytes.BytesRequest {
	bundle := request.GetBundle()
	req := Request{
		Span: Span{
			Timestamp: request.GetStartTime(),
			Type:      "request",
			SpanID:    bundle.Trace.GetSpanID(),
		},
		IsRequest: true,
		Baggage:   bundle.Baggage.String(),
		State:     bundle.State.String(),
	}
	trace := Trace{
		TraceID:  bundle.Trace.GetTraceID(),
		Requests: []Request{req},
	}

	if ptid := bundle.Parent.GetTraceID(); !ptid.IsZero() && ptid != trace.TraceID {
		req.ParentTraceID = &ptid
	}
	return &WriteRequest{
		Bundle: Bundle{
			Traces: []Trace{trace},
		},
		Request: req,
	}
}

func (r *WriteRequest) Flush() error {
	err := r.Writer.receiver(r.Bundle)
	for _, span := range r.Spans {
		span.buffer.ReclaimMemory()
	}
	for _, line := range r.Lines {
		line.line.ReclaimMemory()
	}
	return err
}

func (r *WriteRequest) ReclaimMemory() {}

func (r *WriteRequest) Span(span xopbytes.Span, buffer xopbytes.Buffer) error {
	var s *Span
	if span.IsRequest() {
		s = &r.Request.Span
	} else {
		s = &Span{
			Type: "span",
		}
		r.Spans = append(r.Spans, s)
	}
	s.Timestamp = span.GetStartTime()
	bundle := span.GetBundle()
	s.SpanID = bundle.Trace.GetSpanID()
	s.ParentSpanID = bundle.Parent.GetSpanID()
	s.Duration = time.Unix(0, span.GetEndTimeNano()).Sub(s.Timestamp)
	fj, err := fastjson.ParseBytes(buffer.AsBytes())
	if err != nil {
		return errors.Wrap(err, "xopjs.Span buffer invalid")
	}
	s.Name = string(fj.GetStringBytes("span.name"))
	attributes := fj.GetObject("attributes")
	s.Attributes = attributes.MarshalTo(nil)
	s.buffer = buffer
	// XXX walk attributes to pull in attribute definitions
	return nil
}

func (r *WriteRequest) Line(line xopbytes.Line) error {
	l := Line{
		SpanID:    line.GetSpanID(),
		Level:     line.GetLevel(),
		Timestamp: line.GetTime(),
		line:      line,
	}
	fj, err := fastjson.ParseBytes(line.AsBytes())
	if err != nil {
		return errors.Wrap(err, "xopjs.Line buffer invalid")
	}
	l.Msg = string(fj.GetStringBytes("msg"))
	l.Format = string(fj.GetStringBytes("fmt"))
	attributes := fj.GetObject("attributes")
	l.Attributes = attributes.MarshalTo(nil)
	stack := fj.Get("stack")
	if stack != nil {
		l.Stack = stack.MarshalTo(nil)
	}
	r.Lines = append(r.Lines, l)
	return nil
}
