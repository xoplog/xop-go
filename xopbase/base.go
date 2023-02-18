// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

/*
Package xopbase defines the base-level loggers for xop.

In xop, the logger is divided into a top-level logger and a lower-level
logger.  The top-level logger is what users use to add logs to their
programs.  The lower-level logger is used to send those logs somewhere
useful.  Xopjson, xoptest, and xopotel are all examples.

In OpenTelemetry, these are called "exporters".
In logr, they are called "LogSinks".
*/
package xopbase

import (
	"context"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
)

// Logger is the bottom half of a logger -- the part that actually
// outputs data somewhere.  There can be many Logger implementations.
//
// Loggers must support concurrent requests for all methods.
type Logger interface {
	// Request beings a new span that represents the start of an
	// operation: either a call to a server, a cron-job, or an event
	// being processed.  The provided Context is a pass-through from
	// the Seed and if the seed does not provide a context, the context
	// can be nil.
	Request(ctx context.Context, ts time.Time, span xoptrace.Bundle, description string, source SourceInfo) Request

	// ID returns a unique id for this instance of a logger.  This
	// is used to prevent duplicate Requests from being created when
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

	// Replay is how a Logger will feed it's output to another Logger.  The
	// input param should be the collected output from the first Logger.  If
	// it isn't in the right format, or type, it should throw an error.  This
	// capability can be used to transform from one format to another, it can
	// also be used to during testing to make sure that a Logger can round-trip
	// without data loss.  Loggers are not required to round-trip without
	// data loss.
	Replay(ctx context.Context, input any, logger Logger) error
}

// SourceInfo records both the running program (log source) and the logging
// namespace that it uses. Note that atttributes are separately namespaced, but
// that does not cover request and span naming. Source will be called once
// for each base Logger instance.
type SourceInfo struct {
	Source           string
	SourceVersion    *semver.Version
	Namespace        string
	NamespaceVersion *semver.Version
}

func (si SourceInfo) Size() int32 {
	return int32(len(si.Source) + len(si.Namespace) + len(si.SourceVersion.String()) + len(si.NamespaceVersion.String()))
}

type RoundTripLogger interface {
	Logger

	// LosslessReplay is just like Replay but guarantees that no data
	// is lost.
	LosslessReplay(ctx context.Context, input any, logger Logger) error
}

type Request interface {
	Span

	// Flush calls are single-threaded.  Flush can be triggered explicitly by
	// users and it can be triggered because all parts of a request have had
	// Done() called on them.  Flush can be callled more than once on the same
	// request.
	Flush()

	// SetErrorReported will always be called before any other method on the
	// Request.
	//
	// If a base logger encounters an error, it may use the provided function to
	// report it.  The base logger cannot assume that execution will stop.
	// The base logger may not panic.
	SetErrorReporter(func(error))

	// Final is called when there is no possibility of any further calls to
	// this Request of any sort.  There is no guarantee that Final will be
	// called in a timely fashion or even at all before program exit.
	Final()
}

type Span interface {
	// Span creates a new Span that should inherit prefil but not data.  Calls
	// to Span can be made in parallel.
	Span(ctx context.Context, ts time.Time, span xoptrace.Bundle, descriptionOrName string, spanSequenceCode string) Span

	// MetadataAny adds a key/value pair to describe the span.  Calls to
	// MetadataAny are can be concurrent with other calls to set Metadata.
	MetadataAny(*xopat.AnyAttribute, ModelArg)
	// MetadataBool adds a key/value pair to describe the span.  Calls to
	// MetadataBool are can be concurrent with other calls to set Metadata.
	MetadataBool(*xopat.BoolAttribute, bool)
	// MetadataEnum adds a key/value pair to describe the span.  Calls to
	// MetadataEnum are can be concurrent with other calls to set Metadata.
	MetadataEnum(*xopat.EnumAttribute, xopat.Enum)
	// MetadataFloat64 adds a key/value pair to describe the span.  Calls to
	// MetadataFloat64 are can be concurrent with other calls to set Metadata.
	MetadataFloat64(*xopat.Float64Attribute, float64)
	// MetadataInt64 adds a key/value pair to describe the span.  Calls to
	// MetadataInt64 are can be concurrent with other calls to set Metadata.
	MetadataInt64(*xopat.Int64Attribute, int64)
	// MetadataLink adds a key/value pair to describe the span.  Calls to
	// MetadataLink are can be concurrent with other calls to set Metadata.
	MetadataLink(*xopat.LinkAttribute, xoptrace.Trace)
	// MetadataString adds a key/value pair to describe the span.  Calls to
	// MetadataString are can be concurrent with other calls to set Metadata.
	MetadataString(*xopat.StringAttribute, string)
	// MetadataTime adds a key/value pair to describe the span.  Calls to
	// MetadataTime are can be concurrent with other calls to set Metadata.
	MetadataTime(*xopat.TimeAttribute, time.Time)

	// Boring true indicates that a span (or request) is boring.  The
	// suggested meaning for this is that a boring request that is buffered
	// can ignore Flush() and never get sent to output.  A boring span
	// can be un-indexed. Boring requests that do get sent to output can
	// be marked as boring so that they're dropped at the indexing stage.
	//
	// Calls to Boring are single-threaded with respect to other calls to
	// Boring.  XXX make true.
	Boring(bool)

	// ID must return the same string as the Logger it came from
	ID() string

	// TODO: Gauge()
	// TODO: Event()

	// NoPrefill must work in parallel with other calls to NoPrefill, Span,
	// and StartPrefill.
	NoPrefill() Prefilled

	// StartPrefill must work in parallel with other calls to NoPrefill, Span,
	// and StartPrefill.
	StartPrefill() Prefilling

	// Done is called when (1) log.Done is called on the log corresponding
	// to this span; (2) log.Done is called on a parent log of the log
	// corresponding to this span, and the log is not Detach()ed; or
	// (3) preceding Flush() if there has been logging activity since the
	// last call to Flush(), Done(), or the start of the span.
	//
	// final is true when the log is done, it is false when Done is called
	// prior to a Flush().  Just because Done was called with final does not
	// mean that Done won't be called again.  Any further calls would only
	// happen due a bug in the application: logging things after calling
	// log.Done.
	//
	// If the application never calls log.Done(), then final will never
	// be true.
	//
	// Done can be called in parallel to calls to Done on other Spans
	// and other activity including Flush(). There is no particular order
	// to the calls to Done().
	Done(endTime time.Time, final bool)
}

type Prefilling interface {
	AttributeParts

	PrefillComplete(msg string) Prefilled
}

type Prefilled interface {
	// Line starts another line of log output.  Span implementations
	// can expect multiple calls simultaneously and even during a call
	// to SpanInfo() or Flush().  The []uintptr slice are stack frames.
	Line(xopnum.Level, time.Time, []uintptr) Line
}

type Line interface {
	AttributeParts

	LineDone
}

// LineDone are methods that complete the line.  No additional methods may
// be invoked on the line after one of these is called.
type LineDone interface {
	Msg(string)
	Template(string)
	// Object may change in the future to also take an un-redaction string,
	Model(string, ModelArg)
	Link(string, xoptrace.Trace)
}

type AttributeParts interface {
	// Enum adds a key/value pair.  Calls to Enum are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Enum(*xopat.EnumAttribute, xopat.Enum)

	// Float64 adds a key/value pair.  Calls to Float64 are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Float64(string, float64, DataType)
	// Int64 adds a key/value pair.  Calls to Int64 are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Int64(string, int64, DataType)
	// String adds a key/value pair.  Calls to String are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	String(string, string, DataType)
	// Uint64 adds a key/value pair.  Calls to Uint64 are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Uint64(string, uint64, DataType)

	// Any adds a key/value pair.  Calls to Any are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Any(string, ModelArg)
	// Bool adds a key/value pair.  Calls to Bool are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Bool(string, bool)
	// Duration adds a key/value pair.  Calls to Duration are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Duration(string, time.Duration)
	// Time adds a key/value pair.  Calls to Time are expected to be
	// sequenced with other calls to add attributes to prefill and/or
	// lines.
	Time(string, time.Time)

	// TODO: split the above off as "BasicAttributeParts"
	// TODO: Table(string, table)
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
	// TODO: ExternalReference(name string, itemID string, storageID string)
	// TODO: RedactedString(name string, value string, unredaction string)
}

// TODO
type Encoder interface {
	MimeType() string
	ProducesText() bool
	Encode(elementName string, data interface{}) ([]byte, error)
}

type DataType int

const (
	EnumDataType          = DataType(xopproto.AttributeType_Enum)
	EnumArrayDataType     = DataType(xopproto.AttributeType_ArrayEnum)
	AnyDataType           = DataType(xopproto.AttributeType_Any)
	AnyArrayDataType      = DataType(xopproto.AttributeType_ArrayAny)
	BoolDataType          = DataType(xopproto.AttributeType_Bool)
	BoolArrayDataType     = DataType(xopproto.AttributeType_ArrayBool)
	DurationDataType      = DataType(xopproto.AttributeType_Duration)
	DurationArrayDataType = DataType(xopproto.AttributeType_ArrayDuration)
	ErrorDataType         = DataType(xopproto.AttributeType_Error)
	ErrorArrayDataType    = DataType(xopproto.AttributeType_ArrayError)
	Float32DataType       = DataType(xopproto.AttributeType_Float32)
	Float32ArrayDataType  = DataType(xopproto.AttributeType_ArrayFloat32)
	Float64DataType       = DataType(xopproto.AttributeType_Float64)
	Float64ArrayDataType  = DataType(xopproto.AttributeType_ArrayFloat64)
	IntDataType           = DataType(xopproto.AttributeType_Int)
	IntArrayDataType      = DataType(xopproto.AttributeType_ArrayInt)
	Int16DataType         = DataType(xopproto.AttributeType_Int16)
	Int16ArrayDataType    = DataType(xopproto.AttributeType_ArrayInt16)
	Int32DataType         = DataType(xopproto.AttributeType_Int32)
	Int32ArrayDataType    = DataType(xopproto.AttributeType_ArrayInt32)
	Int64DataType         = DataType(xopproto.AttributeType_Int64)
	Int64ArrayDataType    = DataType(xopproto.AttributeType_ArrayInt64)
	Int8DataType          = DataType(xopproto.AttributeType_Int8)
	Int8ArrayDataType     = DataType(xopproto.AttributeType_ArrayInt8)
	LinkDataType          = DataType(xopproto.AttributeType_Link)
	LinkArrayDataType     = DataType(xopproto.AttributeType_ArrayLink)
	StringDataType        = DataType(xopproto.AttributeType_String)
	StringArrayDataType   = DataType(xopproto.AttributeType_ArrayString)
	StringerDataType      = DataType(xopproto.AttributeType_Stringer)
	StringerArrayDataType = DataType(xopproto.AttributeType_ArrayStringer)
	TimeDataType          = DataType(xopproto.AttributeType_Time)
	TimeArrayDataType     = DataType(xopproto.AttributeType_ArrayTime)
	UintDataType          = DataType(xopproto.AttributeType_Uint)
	UintArrayDataType     = DataType(xopproto.AttributeType_ArrayUint)
	Uint16DataType        = DataType(xopproto.AttributeType_Uint16)
	Uint16ArrayDataType   = DataType(xopproto.AttributeType_ArrayUint16)
	Uint32DataType        = DataType(xopproto.AttributeType_Uint32)
	Uint32ArrayDataType   = DataType(xopproto.AttributeType_ArrayUint32)
	Uint64DataType        = DataType(xopproto.AttributeType_Uint64)
	Uint64ArrayDataType   = DataType(xopproto.AttributeType_ArrayUint64)
	Uint8DataType         = DataType(xopproto.AttributeType_Uint8)
	Uint8ArrayDataType    = DataType(xopproto.AttributeType_ArrayUint8)
	UintptrDataType       = DataType(xopproto.AttributeType_Uintptr)
	UintptrArrayDataType  = DataType(xopproto.AttributeType_ArrayUintptr)
)

func (dt DataType) String() string {
	return xopproto.AttributeType(dt).String()
}
