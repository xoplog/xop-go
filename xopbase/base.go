// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopbase

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"
)

// Logger is the bottom half of a logger -- the part that actually
// outputs data somewhere.  There can be many Logger implementations.
type Logger interface {
	Request(span trace.Bundle, description string) Request

	// ID returns a unique id for this instance of a logger.  This
	// is used to prevent duplicate Requets from being created when
	// additional loggers to added to an ongoing Request.
	ID() string

	// ReferencesKept should return true if Any() objects are not immediately
	// serialized (the object is kept around and serilized later).  If copies
	// are kept, then xop.Log will make copies.
	ReferencesKept() bool

	// Buffered should return true if the logger buffers output and sends
	// it when Flush() is called. Even if Buffered() returns false,
	// Flush() may still be invoked but it doesn't have to do anything.
	Buffered() bool

	Close()
}

type Request interface {
	Span

	// Flush calls are single-threaded
	Flush()

	// SetErrorReported will always be called before any other method on the
	// Request.
	//
	// If a base logger encounters an error, it may use the provided function to
	// report it.  The base logger cannot assume that execution will stop.
	// The base logger may not panic.
	SetErrorReporter(func(error))
}

type Span interface {
	// Span creates a new Span that should inherit prefil but not data
	Span(span trace.Bundle, descriptionOrName string) Span

	// MetadataAny adds a key/value pair to describe the span.
	MetadataAny(*xopconst.AnyAttribute, interface{})
	// MetadataBool adds a key/value pair to describe the span.
	MetadataBool(*xopconst.BoolAttribute, bool)
	// MetadataEnum adds a key/value pair to describe the span.
	MetadataEnum(*xopconst.EnumAttribute, xopconst.Enum)
	// MetadataFloat64 adds a key/value pair to describe the span.
	MetadataFloat64(*xopconst.Float64Attribute, float64)
	// MetadataInt64 adds a key/value pair to describe the span.
	MetadataInt64(*xopconst.Int64Attribute, int64)
	// MetadataLink adds a key/value pair to describe the span.
	MetadataLink(*xopconst.LinkAttribute, trace.Trace)
	// MetadataStr adds a key/value pair to describe the span.
	MetadataStr(*xopconst.StrAttribute, string)
	// MetadataTime adds a key/value pair to describe the span.
	MetadataTime(*xopconst.TimeAttribute, time.Time)

	// Boring true indicates that a span (or request) is boring.  The
	// suggested meaning for this is that a boring request that is buffered
	// can ignore Flush() and never get sent to output.  A boring span
	// can be un-indexed. Boring requests that do get sent to output can
	// be marked as boring so that they're dropped at the indexing stage.
	Boring(bool)

	// ID must return the same string as the Logger it came from
	ID() string

	// TODO: Guage()
	// TODO: Event()

	NoPrefill() Prefilled

	StartPrefill() Prefilling
}

type Prefilling interface {
	ObjectParts

	PrefillComplete(msg string) Prefilled
}

type Prefilled interface {
	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().  The []uintptr slice are stack frames.
	Line(xopconst.Level, time.Time, []uintptr) Line
}

type Line interface {
	ObjectParts

	// Msg may only be called once.  After calling Msg, the line
	// may not be used for anything else unless Recycle is called.
	Msg(string)

	// Template may only be called once.  It is an alternative to Msg.
	Template(string)

	// Static is the same as Msg, but it hints that the supplied string is
	// constant rather than something generated.  Since it's static, base
	// loggers may keep them a dictionary and send entry numbers.
	Static(string)
}

type ObjectParts interface {
	Enum(*xopconst.EnumAttribute, xopconst.Enum)
	// TODO: split the above off as "BasicObjectParts"
	// TODO: Table(string, table)
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
	// TODO: ExternalReference(name string, itemID string, storageID string)

	Any(string, interface{})
	Bool(string, bool)
	Duration(string, time.Duration)
	Error(string, error)
	Float64(string, float64)
	Int(string, int64)
	Link(string, trace.Trace)
	Str(string, string)
	Time(string, time.Time)
	Uint(string, uint64)
}

// TODO
type SubObject interface {
	ObjectParts
	Complete()
}

// TODO
type Encoder interface {
	MimeType() string
	ProducesText() bool
	Encode(elementName string, data interface{}) ([]byte, error)
}
