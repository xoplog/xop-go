package xopjson

import (
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopjson/xopjsonutil"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/google/uuid"
)

var _ xopbase.Logger = &Logger{}
var _ xopbase.Request = &request{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}
var _ xopbytes.Buffer = &builder{}
var _ xopbytes.Line = &line{}
var _ xopbytes.Span = &span{}
var _ xopbytes.Request = &request{}

type Option func(*Logger, *xoputil.Prealloc)

// TimeFormatter is the function signature for custom time formatters
// if anything other than time.RFC3339Nano is desired.  The value must
// be appended to the byte slice (which must be returned).
//
// For example:
//
//	func timeFormatter(b []byte, t time.Time) []byte {
//		b = append(b, '"')
//		b = append(b, []byte(t.Format(time.RFC3339))...)
//		b = append(b, '"')
//		return b
//	}
//
// The slice may not be safely accessed outside of the duration of the
// call.  The only acceptable operation on the slice is to append.
type TimeFormatter func(b []byte, t time.Time) []byte

type Logger struct {
	requestCount             int64 // only incremented with tagOption == TraceSequenceNumberTagOption
	writer                   xopbytes.BytesWriter
	fastKeys                 bool
	durationFormat           DurationOption
	spanStarts               bool
	spanChangesOnly          bool
	id                       uuid.UUID
	tagOption                TagOption
	attributesObject         bool
	builderPool              sync.Pool // filled with *builder
	linePool                 sync.Pool // filled with *line
	preallocatedKeys         [100]byte
	durationKey              []byte
	timeFormatter            TimeFormatter
	activeRequests           sync.WaitGroup
	attributeOption          AttributeOption
	attributesTrackingLogger sync.Map
}

type request struct {
	idNum      int64
	errorCount int32
	span
	errorFunc                 func(error)
	alertCount                int32
	sourceInfo                xopbase.SourceInfo
	attributesTrackingRequest sync.Map
	attributesDefined         *sync.Map
}

type span struct {
	endTime            int64
	writer             xopbytes.BytesRequest
	bundle             xoptrace.Bundle
	logger             *Logger
	name               string
	request            *request
	startTime          time.Time
	serializationCount int32
	attributes         xopjsonutil.AttributeBuilder
	sequenceCode       string
	spanIDBuffer       [len(`"trace.header":`) + 55 + 2]byte
	spanIDPrebuilt     xoputil.JBuilder
	isRequest          bool
}

type prefilling struct {
	*builder
}

type prefilled struct {
	data          []byte
	preEncodedMsg []byte
	span          *span
}

type line struct {
	*builder
	level                xopnum.Level
	timestamp            time.Time
	prefillMsgPreEncoded []byte
}

type builder struct {
	xopjsonutil.Builder
	span              *span
	attributesStarted bool
	attributesWanted  bool
}

type DurationOption int

const (
	AsNanos   DurationOption = iota // int64(duration)
	AsMicros                        // int64(duration / time.Milliscond)
	AsMillis                        // int64(duration / time.Milliscond)
	AsSeconds                       // int64(duration / time.Second)
	AsString                        // duration.String()
)

/* TODO: add back for xopjs
// WithDuration specifies the format used for durations. If
// set, durations will be recorded for spans and requests.  If not
// set, durations explicitly recorded will be recoreded as nanosecond
// numbers.
func WithDuration(key string, durationFormat DurationOption) Option {
	return func(l *Logger, p *xoputil.Prealloc) {
		l.durationKey = p.Pack(xoputil.BuildKey(key))
		l.durationFormat = durationFormat
	}
}
*/

type TagOption int

const (
	SpanIDTagOption       TagOption = 1 << iota // 16 bytes hex
	TraceIDTagOption                = 1 << iota // 32 bytes hex
	TraceHeaderTagOption            = 1 << iota // 2+1+32+1+16+1+2 = 55 bytes/
	TraceNumberTagOption            = 1 << iota // integer trace count
	SpanSequenceTagOption           = 1 << iota // eg ".1.A"
)

/* TODO: add back for xopjs
// WithSpanTags specifies how lines should reference the span that they're within.
// The default is SpanSequenceTagOption if WithBufferedLines(true) is used
// because in that sitatuion, there are other clues that can be used to
// figure out the spanID and traceID.  WithSpanTags() also modifies how spans
// (but not requests) are logged: both TraceNumberTagOption, TraceNumberTagOption
// apply to spans also.
//
// SpanIDTagOption indicates the the spanID should be included.  The key
// is "span.id".
//
// TraceIDTagOption indicates the traceID should be included.  If
// TagLinesWithSpanSequence(true) was used, then the span can be derrived
// that way.  The key is "trace.id".
//
// TraceNumberTagOption indicates that that a trace sequence
// number should be included in each line.  This also means that each
// Request will emit a small record tying the traceID to a squence number.
// The key is "trace.num".
//
// SpanSequenceTagOption indicates that the dot-notation span context
// string should be included in each line.  The key is "span.ctx".

func WithSpanTags(tagOption TagOption) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.tagOption = tagOption
	}
}
*/

// WithSpanStarts controls logging of the start of spans and requests.
// When false, span-level data is output only when when Done() is called.
// Done() can be called more than once. The default is that span starts
// are logged.
func WithSpanStarts(b bool) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.spanStarts = b
	}
}

/* TODO add back for xopjs
// WithSpanChangesOnly controls the data included when span-level and
// request-level data is logged.  When true, only changed fields will
// be output. When false, all data will be output at each call to Done().
func WithSpanChangesOnly(b bool) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.spanChangesOnly = b
	}
}
*/

func WithUncheckedKeys(b bool) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.fastKeys = b
	}
}

type AttributeOption int

const (
	AttributesDefinedAlways AttributeOption = iota
	AttributesDefinedOnce
	AttributesDefinedEachRequest
)

// WithAttributeDefinitions specifies how attribute definitions
// should be handled.  With AttributesDefinedAlways, every time
// an attribute is used, IOWriter.DefineAttribute will be called.
// With AttributesDefinedOnce, IOWriter.DefineAttribute will only
// be called once per attribute.  With AttributesDefinedEachRequest,
// IOWriter.DefineAttribute will be called once per attribute used
// per Request.
//
// The default is AttributesDefinedAlways
func WithAttributeDefinitions(ao AttributeOption) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.attributeOption = ao
	}
}

/* TODO: add back for xopjs
// WithAttributesObject specifies if the user-defined
// attributes on lines, spans, and requests should be
// inside an "attributes" sub-object or part of the main
// object.
func WithAttributesObject(b bool) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.attributesObject = b
	}
}
*/

/* TODO: add back for xopjs
// WithTimeFormatter specifies how time.Time should be
// serialized to JSON.  The default is time.RFC3339Nano.
//
// Note: if serializing as a number, integers beyond 2^50
// may lose precision because they're actually read as
// float64s.
func WithTimeFormatter(formatter TimeFormatter) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.timeFormatter = formatter
	}
}
*/

/*
func WithGoroutineID(b bool) Option {
	return func(l *Logger, _ *xoputil.Prealloc) {
		l.withGoroutine = b
	}
}
*/

// WithRoundedIntegersAsStrings changes the encoding of int64 values
// tha are outside the range that can be exactly represented by
// a float64.  JSON treats "numbers" as floats.  Integers that are
// in the range [-2**53, 2**53] can be exactly represented as
// float64s, but integers outside that range cannot.  When this is
// true, then integers outside this range will be converted to strings.
//
// For Go, decoding a quoted integer into an int64 works just fine.
//
// TODO

// WithErrorEncoder changes the encoding of error values.
// The default encoding is simply a string: error.Error().
//
// TODO
