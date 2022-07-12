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
	Request(span trace.Bundle, description string) Request

	// ReferencesKept should return true if Any() objects are not immediately
	// serialized (the object is kept around and serilized later).  If copies
	// are kept, then xoplog.Log will make copies.
	ReferencesKept() bool

	Close()
}

type Request interface {
	// Calls to Flush are single-threaded along with calls to SpanInfo
	Flush()
	Span
}

type Span interface {
	// Span creates a new Span that should inherit prefil but not data
	Span(span trace.Bundle, descriptionOrName string) Span

	// SpanInfo replaces the span type and span data.
	// SpanInfo calls are only made while holding the Flush() lock
	SpanInfo(xopconst.SpanType, []xop.Thing)

	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().
	Line(xopconst.Level, time.Time) Line
}

type Line interface {
	// TODO: ExternalReference(name string, itemId string, storageId string)
	// TODO: Guage()
	// TODO: Event()

	ObjectParts

	// Msg may only be called once.  After calling Msg, the line
	// may not be used for anything else unless Recycle is called.
	Msg(string)

	// Recycle starts the line ready to use again.
	Recycle(xopconst.Level, time.Time)
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
