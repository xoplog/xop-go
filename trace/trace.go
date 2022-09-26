// trace is the data structures to describe trace_id and span_id
package trace

import (
	"encoding/hex"
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
// why a row was updated, what should you use?  TraceID+RequestID as hex?  That's
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
// These headers can be used with gRPC too
// (https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)
//
type Trace struct {
	version HexBytes1

	// This is an identifier that should flow through all aspects of a
	// request.  It's mandated by W3C and is useful when there are logs
	// missing and the parent ids can't be tied together
	traceID HexBytes16

	// This is the "parent-id" field in the header.  The W3C spec name for this
	// causes confusion.  It's also considered the "span-id".
	spanID HexBytes8

	flags HexBytes1

	headerString string // version + traceID + spanID + flags

	initialized bool
}

func NewTrace() Trace {
	var trace Trace
	trace.initialize()
	return trace
}

func (t *Trace) Version() WrappedHexBytes1 {
	t.initialize()
	return WrappedHexBytes1{
		offset:    startOfVersion,
		trace:     t,
		HexBytes1: &t.version,
	}
}

func (t *Trace) TraceID() WrappedHexBytes16 {
	t.initialize()
	return WrappedHexBytes16{
		offset:     startOfTraceID,
		trace:      t,
		HexBytes16: &t.traceID,
	}
}

func (t *Trace) SpanID() WrappedHexBytes8 {
	t.initialize()
	return WrappedHexBytes8{
		offset:    startOfSpanID,
		trace:     t,
		HexBytes8: &t.spanID,
	}
}

func (t *Trace) Flags() WrappedHexBytes1 {
	t.initialize()
	return WrappedHexBytes1{
		offset:    startOfFlags,
		trace:     t,
		HexBytes1: &t.flags,
	}
}

func (t Trace) GetVersion() HexBytes1  { return t.version.initialized(t) }
func (t Trace) GetTraceID() HexBytes16 { return t.traceID.initialized(t) }
func (t Trace) GetSpanID() HexBytes8   { return t.spanID.initialized(t) }
func (t Trace) GetFlags() HexBytes1    { return t.flags.initialized(t) }
func (t Trace) Copy() Trace            { return t }
func NewSpanID() HexBytes8             { return newHexBytes8() }
func NewTraceID() HexBytes16           { return newHexBytes16() }

func (t Trace) IsZero() bool {
	return t.String() == "00-00000000000000000000000000000000-0000000000000000-00"
}

func (t Trace) String() string {
	if !t.initialized {
		return "00-00000000000000000000000000000000-0000000000000000-00"
	}
	return t.headerString
}

func NewRandomSpanID() HexBytes8 {
	spanID := newHexBytes8()
	spanID.setRandom()
	return spanID
}

var zeroBytes = make([]byte, 16)

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

func (t *Trace) RebuildSetNonZero() {
	if t.traceID.IsZero() {
		t.traceID.setRandom()
	}
	if t.spanID.IsZero() {
		t.spanID.setRandom()
	}
	t.rebuild()
}

func (t *Trace) initialize() {
	if !t.initialized {
		t.initialized = true
		t.version.initialize()
		t.traceID.initialize()
		t.spanID.initialize()
		t.flags.initialize()
		t.rebuild()
	}
}

const startOfVersion = 0
const startOfTraceID = startOfVersion + 2 + 1
const startOfSpanID = startOfTraceID + 32 + 1
const startOfFlags = startOfSpanID + 16 + 1

func (t *Trace) rebuild() {
	// 0         3         36       53
	// version + traceID + spanID + flags
	b := make([]byte, 0, 55)
	b = append(b, t.version.h[:]...)
	b = append(b, '-')
	b = append(b, t.traceID.h[:]...)
	b = append(b, '-')
	b = append(b, t.spanID.h[:]...)
	b = append(b, '-')
	b = append(b, t.flags.h[:]...)
	t.headerString = string(b)
}
