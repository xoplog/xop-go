package xopbytes

import (
	"io"
	"sync"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xoptrace"
)

var _ BytesWriter = IOWriter{}

type IOWriter struct {
	io.Writer
	builderPool sync.Pool
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

// DefineAttribute writes a JSON defintion for the attribute to the writer.
// The requestTrace is optional. If not nil, the span.id is added to the definition.
func (iow IOWriter) DefineAttribute(k *xopat.Attribute, requestTrace *xoptrace.Trace) error {
	if requestTrace != nil {
		jdef := k.DefinitionJSONBytes()
		b := make([]byte, len(jdef)-2, len(jdef)+len(`,"span.id":""{`)+16)
		copy(b, jdef[0:len(jdef)-2])
		b = append(b, []byte(`,"span.id":"`)...)
		b = append(b, requestTrace.SpanID().HexBytes()...)
		b = append(b, '"', '}', '\n')
		_, err := iow.Write(b)
		return err
	}
	_, err := iow.Write(k.DefinitionJSONBytes())
	return err
}

func (iow IOWriter) AttributeReferenced(_ *xopat.Attribute) error { return nil }
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
