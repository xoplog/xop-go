// xopbytes wrapes io.Writer with additional methods useful for wrting logs
package xopbytes

import (
	"github.com/xoplog/xop-go/trace"
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
