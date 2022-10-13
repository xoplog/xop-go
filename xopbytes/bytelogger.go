// xopbytes wrapes io.Writer with additional methods useful for wrting logs
package xopbytes

import (
	"time"

	"github.com/xoplog/xop-go/trace"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopnum"
)

type BytesWriter interface {
	Request(span trace.Bundle) BytesRequest
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
	ReclaimMemory()
}

type Line interface {
	Buffer
	GetSpanID() trace.HexBytes8
	GetLevel() xopnum.Level
	GetTime() time.Time
}

type Span interface {
	GetSpanID() trace.HexBytes8
}
