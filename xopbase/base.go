// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

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

	// StackFramesWanted returns the number of stack frames wanted for
	// each logging level.  Base loggers may be provided with more or
	// fewer frames than they requested.  It is presumed that higher logging
	// levels want more frames than lower levels, so only increases are
	// allowed as the severity increases.
	StackFramesWanted() map[xopconst.Level]int // XXX implement

	// SetErrorReported will always be called before any other method. If a
	// base logger encounters an error, it may use the provided function to
	// report it.  The base logger cannot assume that execution will stop.
	// The base logger may not panic.
	SetErrorReporter(func(error)) // XXX implement

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

	// MetadataAny adds a key/value pair to describe the span.
	MetadataAny(*xopconst.AnyAttribute, interface{})
	// MetadataBool adds a key/value pair to describe the span.
	MetadataBool(*xopconst.BoolAttribute, bool)
	// MetadataEnum adds a key/value pair to describe the span.
	MetadataEnum(*xopconst.EnumAttribute, xopconst.Enum)
	// MetadataInt64 adds a key/value pair to describe the span.
	MetadataInt64(*xopconst.Int64Attribute, int64)
	// MetadataLink adds a key/value pair to describe the span.
	MetadataLink(*xopconst.LinkAttribute, trace.Trace)
	// MetadataNumber adds a key/value pair to describe the span.
	MetadataNumber(*xopconst.NumberAttribute, float64)
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

	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().
	Line(xopconst.Level, time.Time) Line
}

type Line interface {
	// TODO: ExternalReference(name string, itemID string, storageID string)
	// TODO: Guage()
	// TODO: Event()

	ObjectParts

	// Msg may only be called once.  After calling Msg, the line
	// may not be used for anything else unless Recycle is called.
	Msg(string)

	// Template may only be called once.  It is an alternative to Msg.
	Template(string)

	// SetAsPrefill may only be called once.  It is an alternative to Msg.
	// Whatever is in the line becomes part of every following line.
	// This will only be used right when a Span is created and thus
	// locking should not be required.
	SetAsPrefill(string)

	// Static is the same as Msg, but it hints that the supplied string is
	// constant rather than something generated.  Since it's static, base
	// loggers may keep them a dictionary and send entry numbers.
	Static(string)

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
	Enum(*xopconst.EnumAttribute, xopconst.Enum)
	Link(string, trace.Trace)
	Duration(string, time.Duration)
	// TODO: Table(string, table)
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
}

type Buffer interface {
	Context()
	Flush()
}
