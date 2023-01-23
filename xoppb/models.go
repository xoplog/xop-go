package xopjson

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbytes"
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
	writer           xopbytes.BytesWriter
	// fastKeys         bool
	// durationFormat   DurationOption
	// spanStarts       bool
	// spanChangesOnly  bool
	id               uuid.UUID
	// tagOption        TagOption
	// requestCount     int64 // only incremented with tagOption == TraceSequenceNumberTagOption
	// attributesObject bool
	builderPool      sync.Pool // filled with *builder
	linePool         sync.Pool // filled with *line
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
	xopproto.Span
	endTime            int64
	writer             xopbytes.BytesRequest
	bundle             xoptrace.Bundle
	logger             *Logger
	request            *request
	attributeMap map[string]int 
	distinctMap map[string]int
}

type prefilling struct {
	*builder
}

type prefilled struct {
	data          []xopyproto.Attribute
	preEncodedMsg string
	span          *span
}

type line struct {
	*builder
	xopproto.Line
	//XXX spanID?
}

type builder struct {
	attributes []xopproto.Attribute
	span              *span
}
