package xopbytes

import (
	"io"
	"sync"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"
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

var defineAttributeStarter = []byte(`{"type":"defineKey","key":"`) // }

func (iow IOWriter) DefineAttribute(k *xopat.Attribute, requestTrace *xoptrace.Trace) error {
	rawBuilder := iow.builderPool.Get()
	var b *xoputil.JBuilder
	if rawBuilder == nil {
		b = &xoputil.JBuilder{}
		b.B = make([]byte, len(defineAttributeStarter), len(defineAttributeStarter)+100)
		copy(b.B, defineAttributeStarter)
	} else {
		b = rawBuilder.(*xoputil.JBuilder)
		b.B = b.B[:len(defineAttributeStarter)]
	}
	b.AddStringBody(k.Key())
	b.AppendBytes([]byte(`","desc":"`))
	b.AddStringBody(k.Description())
	b.AppendBytes([]byte(`","ns":"`))
	b.AddStringBody(k.Namespace())
	b.AppendByte(' ')
	b.AddStringBody(k.SemverString())
	if k.Indexed() {
		b.AppendBytes([]byte(`","indexed":true,"prom":`))
	} else {
		b.AppendBytes([]byte(`","prom":`))
	}
	b.AddInt32(int32(k.Prominence()))
	if k.Multiple() {
		if k.Distinct() {
			b.AppendBytes([]byte(`,"mult":true,"distinct":true,"vtype":`))
		} else {
			b.AppendBytes([]byte(`,"mult":true,"vtype":`))
		}
	} else {
		if k.Locked() {
			b.AppendBytes([]byte(`,"locked":true,"vtype":`))
		} else {
			b.AppendBytes([]byte(`,"vtype":`))
		}
	}
	b.AddSafeString(k.ProtoType().String())
	if k.Ranged() {
		b.AppendBytes([]byte(`,"ranged":true`))
	}
	if requestTrace != nil {
		b.AppendBytes([]byte(`,"span.id":"`))
		b.AppendString(requestTrace.SpanID().String())
		b.AppendByte('"')
	}
	// {
	b.AppendBytes([]byte("}\n"))
	_, err := iow.Write(b.B)
	if err != nil {
		return err
	}
	iow.builderPool.Put(b)
	return nil
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
