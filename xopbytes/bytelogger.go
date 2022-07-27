package xopbytes

import (
	"github.com/muir/xoplog/trace"
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
