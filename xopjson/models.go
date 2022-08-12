package xopjson

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"

	"github.com/google/uuid"
)

var _ xopbase.Logger = &Logger{}
var _ xopbase.Request = &request{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}

type Option func(*Logger)

type timeOption int

const (
	epochTime timeOption = iota
	epochQuoted
	strftimeTime
	timeTimeFormat
)

type Logger struct {
	writer                xopbytes.BytesWriter
	timeOption            timeOption
	timeFormat            string
	timeDivisor           time.Duration
	withGoroutine         bool
	fastKeys              bool
	durationFormat        DurationOption
	id                    uuid.UUID
	tagOption             TagOption
	requestCount          int64 // only incremented with tagOption == TraceSequenceNumberTagOption
	perRequestBufferLimit int
	attributesObject      bool // TODO: implement
	closeRequest          chan struct{}
	builderPool           sync.Pool // filled with *builder
	linePool              sync.Pool // filled with *line
	// prefilledPool	sync.Pool
}

type request struct {
	span
	errorFunc         func(error)
	writeBuffer       []byte
	completedLines    chan *line
	flushRequest      chan struct{}
	flushComplete     chan struct{}
	completedBuilders chan *builder
	allSpans          []*span
	allSpansLock      sync.Mutex
	idNum             int64
}

type span struct {
	endTime    int64 // XXX
	attributes xoputil.AttributeBuilder
	writer     xopbytes.BytesRequest
	trace      trace.Bundle
	logger     *Logger
	name       string
	request    *request
	startTime  time.Time
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
	level                xopconst.Level
	timestamp            time.Time
	prefillMsgPreEncoded []byte
}

type builder struct {
	dataBuffer        xoputil.JBuilder
	encoder           *json.Encoder
	span              *span
	attributesStarted bool
	attributesWanted  bool
}

type DurationOption int

const (
	AsNanos   DurationOption = iota // int64(duration)
	AsMillis                        // int64(duration / time.Milliscond)
	AsSeconds                       // int64(duration / time.Second)
	AsString                        // duration.String()
)

// WithDurtionFormat specifies the format used for durations.
// AsNanos is the default.
func WithDurationFormat(durationFormat DurationOption) Option {
	return func(l *Logger) {
		l.durationFormat = durationFormat
	}
}

type TagOption int

const (
	DefaultTagOption TagOption = iota // TODO: set
	SpanIDTagOption
	FullIDTagOption
	TraceIDTagOption
	TraceSequenceNumberTagOption // TODO emit trace sequence number
	OmitTagOption
)

// WithSpanTags specifies should reference the span that they're within.
// The default is OmitTagOption if WithBufferedLines(true) is used
// because in that sitatuion, there are other clues that can be used to
// figure out which span the line goes with.
//
// SpanIDTagOption indicates the the spanID should be included.  The key
// is "span_id".
//
// TraceIDTagOption indicates the traceID should be included.  If
// TagLinesWithSpanSequence(true) was used, then the span can be derrived
// that way.  The key is "trace_id".
//
// FullIDTagOption indicates that the traceID and the spanID should be
// included.  This is the default with WithBufferedLines(false).
// The key is "trace_header".
//
// TraceSequenceNumberTagOption indicates that that a trace sequence
// number should be included in each line.  This also means that each
// Request will emit a small record tying the traceID to a squence number.
// The key is "trace_num".
//
// OmitTagOption indicates that no Span information should be included with
// each Line object.
func WithSpanTags(tagOption TagOption) Option {
	return func(l *Logger) {
		l.tagOption = tagOption
	}
}

// WithBufferedLines indciates if line data should be buffered until
// Flush() is called.  If not, lines are emitted as they're completed.
// A value of zero (the default) indicates that lines are not buffered.
//
// A value less than 1024 will panic.  1MB is the suggested value.
func WithBufferedLines(bufferSize int) Option {
	if bufferSize < 1024 {
		panic("bufferSize too small")
	}
	return func(l *Logger) {
		l.perRequestBufferLimit = bufferSize
	}
}

func WithUncheckedKeys(b bool) Option {
	return func(l *Logger) {
		l.fastKeys = b
	}
}

// WithAttributesObject specifies if the user-defined
// attributes on lines, spans, and requests should be
// inside an "attributes" sub-object or part of the main
// object.
func WithAttributesObject(b bool) Option {
	return func(l *Logger) {
		l.attributesObject = b
	}
}

// TODO: allow custom error formats

// WithStrftime specifies how to format timestamps.
// See // https://github.com/phuslu/fasttime for the supported
// formats.
func WithStrftime(format string) Option {
	return func(l *Logger) {
		l.timeOption = strftimeTime
		l.timeFormat = format
	}
}

// WithTimeFormat specifies the use of the "time" package's
// Time.Format for formatting times.
func WithTimeFormat(format string) Option {
	return func(l *Logger) {
		l.timeOption = timeTimeFormat
		l.timeFormat = format
	}
}

// WithEpochTime specifies that time values are formatted as an
// integer time since Jan 1 1970.  If the units is seconds, then
// it is seconds since Jan 1 1970.  If the units is nanoseconds,
// then it is nanoseconds since Jan 1 1970.  Etc.
//
// Note that the JSON specification specifies int's are 32 bits,
// not 64 bits so a compliant JSON parser could fail for seconds
// since 1970 starting in year 2038.  For microseconds, and
// nanoseconds, a complicant parser alerady fails.
func WithEpochTime(units time.Duration) Option {
	return func(l *Logger) {
		l.timeOption = epochTime
		l.timeDivisor = units
	}
}

// WithQuotedEpochTime specifies that time values are formatted an
// integer string (integer with quotes around it) representing time
// since Jan 1 1970.  If the units is seconds, then
// it is seconds since Jan 1 1970.  If the units is nanoseconds,
// then it is nanoseconds since Jan 1 1970.  Etc.
//
// Note most JSON parsers can parse into an integer if given a
// a quoted integer.
func WithQuotedEpochTime(units time.Duration) Option {
	return func(l *Logger) {
		l.timeOption = epochQuoted
		l.timeDivisor = units
	}
}

// TODO
func WithGoroutineID(b bool) Option {
	return func(l *Logger) {
		l.withGoroutine = b
	}
}
