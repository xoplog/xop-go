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

	headerString  string // version + traceID + spanID + flags
	traceIDString string // traceID + spanID
}

func (t Trace) Copy() Trace {
	return Trace{
		version:       t.version.Copy(),
		traceID:       t.traceID.Copy(),
		spanID:        t.spanID.Copy(),
		flags:         t.flags.Copy(),
		headerString:  t.headerString,
		traceIDString: t.traceIDString,
	}
}

func NewTrace() Trace {
	return Trace{
		version: NewHexBytes1(),
		traceID: NewHexBytes16(),
		spanID:  NewHexBytes8(),
		flags:   NewHexBytes1(),
	}
}

func (t *Trace) Version() *HexBytes1   { return &t.version }
func (t *Trace) TraceID() *HexBytes16  { return &t.traceID }
func (t *Trace) SpanID() *HexBytes8    { return &t.spanID }
func (t *Trace) Flags() *HexBytes1     { return &t.flags }
func (t *Trace) RandomizeSpanID()      { t.spanID.SetRandom(); t.rebuild() }
func (t Trace) GetVersion() HexBytes1  { return t.version }
func (t Trace) GetTraceID() HexBytes16 { return t.traceID }
func (t Trace) GetSpanID() HexBytes8   { return t.spanID }
func (t Trace) GetFlags() HexBytes1    { return t.flags }
func (t Trace) IsZero() bool           { return t.traceID.IsZero() }
func (t Trace) IDString() string       { return t.traceIDString }
func (t Trace) String() string         { return t.headerString }
func (t Trace) TraceIDString() string  { return t.headerString[3:35] }
func (t Trace) SpanIDString() string   { return t.headerString[36:52] }

func NewRandomSpanID() HexBytes8 {
	spanID := NewHexBytes8()
	spanID.SetRandom()
	return spanID
}

func NewSpanID() HexBytes8 {
	return NewHexBytes8()
}

func NewTraceID() HexBytes16 {
	return NewHexBytes16()
}

func allZero(byts []byte) bool {
	for _, b := range byts {
		if b != 0 {
			return false
		}
	}
	return true
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
		t.traceID.SetRandom()
	}
	if t.spanID.IsZero() {
		t.spanID.SetRandom()
	}
	t.rebuild()
}

func (t *Trace) rebuild() {
	t.headerString = t.version.String() +
		"-" + t.traceID.String() +
		"-" + t.spanID.String() +
		"-" + t.flags.String()
	t.traceIDString = t.traceID.String() + "/" + t.spanID.String()
}
