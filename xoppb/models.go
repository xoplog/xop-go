package xoppb

import (
	"sync"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

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

type Logger struct {
	writer xopbytes.BytesWriter
	// fastKeys         bool
	// durationFormat   DurationOption
	// spanStarts       bool
	// spanChangesOnly  bool
	id uuid.UUID
	// tagOption        TagOption
	// requestCount     int64 // only incremented with tagOption == TraceSequenceNumberTagOption
	// attributesObject bool
	builderPool sync.Pool // filled with *builder
	linePool    sync.Pool // filled with *line
	// preallocatedKeys [100]byte
	// durationKey      []byte
	// stackLineRewrite func(string) string
	// timeFormatter    TimeFormatter
	// activeRequests   sync.WaitGroup
}

type request struct {
	span
	errorCount int32
	errorFunc  func(error)
	idNum      int64
	alertCount int32
}

type span struct {
	protoSpan    xopproto.Span
	endTime      int64
	writer       xopbytes.BytesRequest
	bundle       xoptrace.Bundle
	logger       *Logger
	request      *request
	attributeMap map[string]*xopproto.SpanAttribute
	distinctMaps map[string]*distinction
	mu           sync.Mutex
}

type distinction struct {
	mu         sync.Mutex
	seenString map[string]struct{}
	seenInt    map[int64]struct{}
	seenUint   map[uint64]struct{}
	seenFloat  map[float64]struct{}
}

type prefilling struct {
	*builder
}

type prefilled struct {
	data       []*xopproto.Attribute
	prefillMsg string
	span       *span
}

type line struct {
	*builder
	protoLine xopproto.Line
	//XXX spanID?
}

type builder struct {
	attributes []*xopproto.Attribute
	span       *span
}
