package xoppb

import (
	"sync"

	"github.com/xoplog/xop-go/xopbase"
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

type Writer interface {
	SizeLimit() int32
	Request(traceID xoptrace.HexBytes16, request *xopproto.Request)
	Flush() error
}

type Logger struct {
	writer Writer
	id     uuid.UUID
	//builderPool sync.Pool // filled with *builder
	//linePool    sync.Pool // filled with *line
}

type request struct {
	span
	errorCount           int32
	errorFunc            func(error)
	alertCount           int32
	sourceInfo           xopbase.SourceInfo
	lines                []*xopproto.Line
	lineLock             sync.Mutex
	requestLock          sync.Mutex
	priorLines           int
	attributeDefinitions []*xopproto.AttributeDefinition
	attributeIndex       map[int32]uint32
	flushGeneration      int
}

type span struct {
	protoSpan    xopproto.Span
	endTime      int64
	bundle       xoptrace.Bundle
	logger       *Logger
	request      *request
	attributeMap map[string]*xopproto.SpanAttribute // TODO: combine with distinction?
	distinctMaps map[string]*distinction
	mu           sync.Mutex
	spanLock     sync.Mutex
	parent       *span
	needFlushing []*span
	lastFlush    int // not atomic, no locking because only used inside Flush()
}

type distinction struct {
	mu         sync.Mutex
	seenString map[string]struct{}
	seenInt    map[int64]struct{}
	seenFloat  map[float64]struct{}
}

type prefilling struct {
	*builder
}

type prefilled struct {
	*builder
	prefillMsg string
}

type line struct {
	*builder
	prefillMsg string
	protoLine  *xopproto.Line
}

type builder struct {
	attributes []*xopproto.Attribute
	span       *span
}
