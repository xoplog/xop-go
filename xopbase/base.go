package xopbase

import (
	"time"

	"github.com/muir/xoplog/trace"
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

	// Buffered should return true if the logger buffers output and sends
	// it when Flush() is called. Even if Buffered() returns false,
	// Flush() may still be invoked but it doesn't have to do anything.
	Buffered() bool

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

	// These all add key/value pairs to describe a span
	MetadataInt(*xopconst.IntAttribute, int64)
	MetadataStr(*xopconst.StrAttribute, string)
	MetadataLink(*xopconst.LinkAttribute, trace.Trace)
	MetadataTime(*xopconst.TimeAttribute, time.Time)
	MetadataDuration(*xopconst.DurationAttribute, time.Duration)
	MetadataAny(*xopconst.Attribute, interface{})
	MetadataBool(*xopconst.BoolAttribute, bool)

	// Boring true indicates that a span (or request) is boring.  The
	// suggested meaning for this is that a boring request that is buffered
	// can ignore Flush() and never get sent to output.  A boring span
	// can be un-indexed. Boring requests that do get sent to output can
	// be marked as boring so that they're dropped at the indexing stage.
	Boring(bool)

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
	// TODO: Table(string, table)
	// TODO: URI(string, string)
	// TODO: Link(string, trace.Trace)
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
}

type Buffer interface {
	Context()
	Flush()
}
