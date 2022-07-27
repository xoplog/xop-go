package xopbytes

import (
	"github.com/muir/xop/trace"
)

type BytesWriter interface {
	Request(span trace.Bundle) BytesRequest
	Buffered() bool
	Close() // no point in returning an error
}

type BytesRequest interface {
	Flush() error
	Write([]byte) (int, error)
}
