package xm

import (
	"encoding/hex"
	"math/rand"
	"strings"
)

// TraceState represents W3C tracing headers.
// See https://www.w3.org/TR/trace-context/ and
// https://github.com/w3c/trace-context/blob/main/spec/30-processing-model.md
//
// The spec has some unfortuante side-effects.  The spec doesn't require the
// "traceresponse" header except when the trace-id is discarded.  That means that
// parent requests do not have enough information to directly link to child
// requests.
//
// The "parent-id" is misnamed and thus the spec is harder to understand that it
// should be which could easily lead to incorrect implementations.
//
// The overall trace header includes things that aren't part of the id and thus
// it's not clear what part should be stored in a database and what part should be
// searchable.  Proper log/trace database structure clearly needs at least two
// fields, but if you're adding a trace column to other tables to use to reference
// why a row was updated, what should you use?  TraceId+RequestId as hex?  That's
// 49 bytes if you include a dash and a terminator.  That's often going to be more than
// the rest of the row.  Or store two fields.  In hex or as bytes?
//
// If you want to tag log lines with the trace id, do you add 49 bytes to each line?
//
// If you have lightweight spans within a request, do you create a new
// "parent-id"/"span-id" for each?  If you want to play nicely with Jaeger and
// Zipkin, then lightweight spans need full span-ids.
//
// So when you are looking at the trace for one request and traverse to its parent,
// you can land at one of these
//
// Zipkin has "b3" tracing headers which are an alternative to the W3C headers.
// Both can be supported.  See [b3](https://github.com/openzipkin/b3-propagation)
//
// These headers can be used with
// [gRPC too](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)
//
type traceState struct {
	parentTrace Trace
	myTrace     Trace
	state       State
	baggage     Baggage
}

func (t traceState) Copy() traceState {
	return traceState{
		parentTrace: t.parentTrace.Copy(),
		myTrace:     t.myTrace.Copy(),
		state:       t.state,
		baggage:     t.baggage,
	}
}

type Trace struct {
	version HexBytes // 1 byte

	// This is an identifier that should flow through all aspects of a
	// request.  It's mandated by W3C and is useful when there are logs
	// missing and the parent ids can't be tied together
	traceId HexBytes // 16 bytes

	// This is the "parent-id" field in the header.  The W3C spec name for this
	// causes confusion.  It's also considered the "span-id".
	spanId HexBytes // 8 bytes

	flags HexBytes // 1 byte

	headerString  string
	traceIdString string
}

func (t Trace) Copy() Trace {
	return Trace{
		version:       t.version.Copy(),
		traceId:       t.traceId.Copy(),
		spanId:        t.spanId.Copy(),
		flags:         t.flags.Copy(),
		headerString:  t.headerString,
		traceIdString: t.traceIdString,
	}
}

type HexBytes struct {
	b []byte
	s string
}

func (s Seed) Trace() *Trace {
	return &s.myTrace
}
func (s Seed) TraceParent() *Trace {
	return &s.parentTrace
}

func NewTrace() Trace {
	return Trace{
		version: NewHexBytes(1),
		traceId: NewHexBytes(16),
		spanId:  NewHexBytes(8),
		flags:   NewHexBytes(1),
	}
}

func (t *Trace) Version() *HexBytes  { return &t.version }
func (t *Trace) TraceId() *HexBytes  { return &t.traceId }
func (t *Trace) SpanId() *HexBytes   { return &t.spanId }
func (t *Trace) Flags() *HexBytes    { return &t.flags }
func (t Trace) GetVersion() HexBytes { return t.version }
func (t Trace) GetTraceId() HexBytes { return t.traceId }
func (t Trace) GetSpanId() HexBytes  { return t.traceId }
func (t Trace) GetFlags() HexBytes   { return t.flags }

func NewHexBytes(length int) HexBytes {
	return HexBytes{
		b: make([]byte, length),
		s: strings.Repeat("0", length*2),
	}
}
func NewSpanId() HexBytes {
	return NewHexBytes(8)
}
func NewTraceId() HexBytes {
	return NewHexBytes(8)
}

func (x HexBytes) IsZero() bool   { return allZero(x.b) }
func (x HexBytes) String() string { return x.s }
func (x HexBytes) Bytes() []byte  { return x.b }
func (x *HexBytes) SetBytes(b []byte) {
	setBytes(x.b, b)
	x.s = hex.EncodeToString(x.b)
}
func (x *HexBytes) SetString(s string) {
	setBytesFromString(x.b, s)
	x.s = hex.EncodeToString(x.b)
}
func (x *HexBytes) SetZero() {
	setBytes(x.b, zeroBytes)
	x.s = hex.EncodeToString(x.b)
}
func (x *HexBytes) SetRandom() {
	randomBytesNotAllZero(x.b)
	x.s = hex.EncodeToString(x.b)
}
func (x HexBytes) Copy() HexBytes {
	b := make([]byte, len(x.b))
	copy(b, x.b)
	return HexBytes{
		b: b,
		s: x.s,
	}
}

var zeroBytes = make([]byte, 16)

func randomBytesNotAllZero(byts []byte) {
	for {
		_, _ = rand.Read(byts)
		if !allZero(byts) {
			return
		}
	}
}

func allZero(byts []byte) bool {
	for _, b := range byts {
		if b != 0 {
			return false
		}
	}
	return true
}

func setBytesFromString(dest []byte, h string) {
	b, err := hex.DecodeString(h)
	if err != nil {
		copy(dest, zeroBytes[:len(dest)])
		return
	}
	setBytes(dest, b)
}

func setBytes(dest []byte, b []byte) {
	if len(b) >= len(dest) {
		copy(dest, b[0:len(dest)])
	} else {
		copy(dest, b[0:len(dest)])
		copy(dest[len(b):], zeroBytes[:len(dest)-len(b)])
	}
}

func (t *Trace) rebuildSetNonZero() {
	if t.traceId.IsZero() {
		t.traceId.SetRandom()
	}
	if t.spanId.IsZero() {
		t.spanId.SetRandom()
	}
	t.rebuild()
}

func (t *Trace) rebuild() {
	t.headerString = t.version.String() +
		"-" + t.traceId.String() +
		"-" + t.spanId.String() +
		"-" + t.flags.String()
	t.traceIdString = t.traceId.String() + "/" + t.spanId.String()
}

func (s Seed) SubSpan() Seed {
	s.parentTrace = s.myTrace.Copy()
	s.myTrace.spanId.SetRandom()
	return s
}

func (l *Logger) TracingState() State     { return l.seed.state }
func (l *Logger) TracingBaggage() Baggage { return l.seed.baggage }
func (l *Logger) TracingParent() Trace    { return l.seed.parentTrace }
func (l *Logger) Tracing() Trace          { return l.seed.myTrace }
func (l *Logger) TracingId() string       { return l.seed.myTrace.traceIdString }
func (l *Logger) TracingHeader() string   { return l.seed.myTrace.headerString }
