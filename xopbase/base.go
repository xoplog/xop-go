// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopbase

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopat"
	"github.com/muir/xop-go/xopnum"
)

// Logger is the bottom half of a logger -- the part that actually
// outputs data somewhere.  There can be many Logger implementations.
type Logger interface {
	Request(ts time.Time, span trace.Bundle, description string) Request

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

	// Flush calls are single-threaded.  Flush can be triggered explicitly by
	// users and it can be triggered because all parts of a request have had
	// Done() called on them.
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
	Span(ts time.Time, span trace.Bundle, descriptionOrName string, spanSequenceCode string) Span

	// MetadataAny adds a key/value pair to describe the span.
	MetadataAny(*xopat.AnyAttribute, interface{})
	// MetadataBool adds a key/value pair to describe the span.
	MetadataBool(*xopat.BoolAttribute, bool)
	// MetadataEnum adds a key/value pair to describe the span.
	MetadataEnum(*xopat.EnumAttribute, xopat.Enum)
	// MetadataFloat64 adds a key/value pair to describe the span.
	MetadataFloat64(*xopat.Float64Attribute, float64)
	// MetadataInt64 adds a key/value pair to describe the span.
	MetadataInt64(*xopat.Int64Attribute, int64)
	// MetadataLink adds a key/value pair to describe the span.
	MetadataLink(*xopat.LinkAttribute, trace.Trace)
	// MetadataString adds a key/value pair to describe the span.
	MetadataString(*xopat.StringAttribute, string)
	// MetadataTime adds a key/value pair to describe the span.
	MetadataTime(*xopat.TimeAttribute, time.Time)

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

	// Done is called when (1) log.Done is called on the log corresponding
	// to this span; (2) log.Done is called on a parent log of the log
	// corresponding to this span, and the log is not Detach()ed; or
	// (3) preceeding Flush() if there has been logging activity since the
	// last call to Flush(), Done(), or the start of the span.
	Done(time.Time)
}

type Prefilling interface {
	ObjectParts

	PrefillComplete(msg string) Prefilled
}

type Prefilled interface {
	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().  The []uintptr slice are stack frames.
	Line(xopnum.Level, time.Time, []uintptr) Line
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
	// loggers may keep them a dictionary and send references.
	Static(string)
}

type ObjectParts interface {
	Enum(*xopat.EnumAttribute, xopat.Enum)
	// TODO: split the above off as "BasicObjectParts"
	// TODO: Table(string, table)
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
	// TODO: ExternalReference(name string, itemID string, storageID string)

	Float64(string, float64, DataType)
	Int64(string, int64, DataType)
	Uint64(string, uint64, DataType)

	Any(string, interface{})
	Bool(string, bool)
	Duration(string, time.Duration)
	Error(string, error)
	Link(string, trace.Trace)
	String(string, string)
	Time(string, time.Time)
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

//go:generate enumer -type=DataType -linecomment -json -sql

type DataType int

const (
	EnumDataType     DataType = iota
	AnyDataType      DataType = iota
	BoolDataType     DataType = iota
	DurationDataType DataType = iota
	ErrorDataType    DataType = iota
	Float32DataType  DataType = iota
	Float64DataType  DataType = iota
	IntDataType      DataType = iota
	Int16DataType    DataType = iota
	Int32DataType    DataType = iota
	Int64DataType    DataType = iota
	Int8DataType     DataType = iota
	LinkDataType     DataType = iota
	StringDataType   DataType = iota
	TimeDataType     DataType = iota
	UintDataType     DataType = iota
	Uint16DataType   DataType = iota
	Uint32DataType   DataType = iota
	Uint64DataType   DataType = iota
	Uint8DataType    DataType = iota
)
