// xopbytes wrapes io.Writer with additional methods useful for wrting logs
package xopbytes

import (
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
)

type BytesWriter interface {
	Request(request Request) BytesRequest
	Buffered() bool
	Close()                                      // no point in returning an error
	DefineAttribute(*xopat.Attribute)            // duplicate calls should be ignored
	DefineEnum(*xopat.EnumAttribute, xopat.Enum) // duplicate calls should be ignored
}

type BytesRequest interface {
	Flush() error
	ReclaimMemory() // called when we know BytesRequest will never be used again
	Span(Span, Buffer) error
	Line(Line) error
	AttributeReferenced(*xopat.Attribute) error
}

type Buffer interface {
	AsBytes() []byte
	ReclaimMemory() // call ReclaimMemory after the output of AsBytes() is copied.
}

type Line interface {
	Buffer
	GetSpanID() xoptrace.HexBytes8
	GetLevel() xopnum.Level
	GetTime() time.Time
}

type Request interface {
	Span
	GetErrorCount() int32
	GetAlertCount() int32
}

type Span interface {
	GetBundle() xoptrace.Bundle
	GetStartTime() time.Time
	GetEndTimeNano() int64
	IsRequest() bool
}
