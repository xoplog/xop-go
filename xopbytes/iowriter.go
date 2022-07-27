package xopbytes

import (
	"io"

	"github.com/muir/xop-go/trace"
)

var _ BytesWriter = IOWriter{}

type IOWriter struct {
	io.Writer
}

func WriteToIOWriter(w io.Writer) BytesWriter {
	return IOWriter{
		Writer: w,
	}
}

func (iow IOWriter) Buffered() bool                    { return false }
func (iow IOWriter) Flush() error                      { return nil }
func (iow IOWriter) Request(trace.Bundle) BytesRequest { return iow }

func (iow IOWriter) Close() {
	if wc, ok := iow.Writer.(io.WriteCloser); ok {
		_ = wc.Close()
	}
}
