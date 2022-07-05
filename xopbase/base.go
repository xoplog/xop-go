package xopbase

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopconst"
)

// Logger is the bottom half of a logger -- the part that actually
// outputs data somewhere.  There can be many Logger implementations.
type Logger interface {
	Request() Request

	// ReferencesKept should return true if Any() objects are not immediately
	// serialized (the object is kept around and serilized later).  If copies
	// are kept, then xoplog.Log will make copies.
	ReferencesKept() bool

	Close()
}

type Request interface {
	// Calls to Flush are single-threaded along with calls to SpanInfo
	Flush()
	Span(span trace.Bundle) Span
}

type Span interface {
	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().
	Line(xopconst.Level, time.Time) Line

	// SpanInfo replaces the span type and span data.
	// Calls to flush are single-threaded along with calls to SpanInfo
	SpanInfo(xopconst.SpanType, []xop.Thing)

	// AddPrefill adds to what has already been provided for this span.
	// Calls to AddPrefill will not overlap other calls to AddPrefil or
	// to ResetLinePrefil.
	AddPrefill([]xop.Thing)
	ResetLinePrefill()

	// Span creates a new Span that should inherit prefil but not data
	Span(span trace.Bundle) Span
}

type Line interface {
	ObjectParts
	// TODO: ExternalReference(name string, itemId string, storageId string)
	Msg(string)
	// TODO: Guage()
	// TODO: Event()
}

type SubObject interface {
	ObjectParts
	Complete()
}

type Encoder interface {
	MimeType() string
	ProducesText() bool
	Encode(elementName string, data interface{}) ([]byte, error)
}

type ObjectParts interface {
	Int(string, int64)
	Uint(string, uint64)
	Bool(string, bool)
	Str(string, string)
	Time(string, time.Time)
	Error(string, error)
	Any(string, interface{}) // generally serialized with JSON
	// TODO: TraceReference(string, trace.Trace)
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
}

type Buffer interface {
	Context()
	Flush()
}
