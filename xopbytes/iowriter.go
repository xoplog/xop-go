package xopbytes

import (
	"io"

	"github.com/xoplog/xop-go/xopat"
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

func (iow IOWriter) Buffered() bool                              { return false }
func (iow IOWriter) Flush() error                                { return nil }
func (iow IOWriter) ReclaimMemory()                              {}
func (iow IOWriter) Request(_ Request) BytesRequest              { return iow }
func (iow IOWriter) DefineEnum(*xopat.EnumAttribute, xopat.Enum) {}
func (iow IOWriter) DefineAttribute(*xopat.Attribute)            {}
func (iow IOWriter) Line(line Line) error {
	_, err := iow.Write(line.AsBytes())
	line.ReclaimMemory()
	return err
}
func (iow IOWriter) Span(_ Span, buffer Buffer) error {
	_, err := iow.Write(buffer.AsBytes())
	buffer.ReclaimMemory()
	return err
}
func (iow IOWriter) Close() {
	if wc, ok := iow.Writer.(io.WriteCloser); ok {
		_ = wc.Close()
	}
}
