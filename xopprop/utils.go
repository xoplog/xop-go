package xopprop

import (
	"strings"

	"github.com/muir/xoplog/trace"
)

// "traceparent" header
// Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
func SetByTraceParentHeader(b *trace.Bundle, h string) {
	splits := strings.SplitN(h, "-", 5)
	if len(splits) != 4 {
		return
	}
	b.TraceParent.Version().SetString(splits[0])
	b.Trace.Version().SetBytes(b.TraceParent.Version().Bytes())

	b.TraceParent.TraceId().SetString(splits[1])
	b.Trace.TraceId().SetBytes(b.TraceParent.TraceId().Bytes())

	b.TraceParent.SpanId().SetString(splits[2])
	b.Trace.SpanId().SetRandom()

	b.TraceParent.Flags().SetString(splits[3])
	b.Trace.Flags().SetBytes(b.TraceParent.Flags().Bytes())
}

// https://github.com/openzipkin/b3-propagation
// b3: traceid-spanid-sampled-parentspanid
func SetByB3Header(b *trace.Bundle, h string) {
	switch h {
	case "0", "1":
		SetByB3Sampled(b, h)
		return
	}
	splits := strings.SplitN(h, "-", 5)
	if len(splits) != 4 {
		return
	}
	b.TraceParent.TraceId().SetString(splits[0])
	b.Trace.Version().SetBytes(b.TraceParent.Version().Bytes())

	b.Trace.SpanId().SetString(splits[1])

	b.TraceParent.Flags().SetString(splits[2])
	b.Trace.Flags().SetBytes(b.TraceParent.Flags().Bytes())

	b.TraceParent.SpanId().SetString(splits[3])
}

// X-B3-Sampled
func SetByB3Sampled(b *trace.Bundle, h string) {
	switch h {
	case "1":
		b.Trace.Flags().SetBytes([]byte{1})
	case "0":
		b.Trace.Flags().SetBytes([]byte{0})
	}
	return
}

// X-B3-ParentSpanId
// Implies parent trace id is the same as my trace id
func SetByB3ParentSpanId(b *trace.Bundle, h string) {
	b.TraceParent.SpanId().SetString(h)
	if b.TraceParent.SpanId().IsZero() {
		b.TraceParent.TraceId().SetZero()
	} else {
		b.TraceParent.TraceId().SetBytes(b.Trace.TraceId().Bytes())
	}
	if b.Trace.TraceId().IsZero() {
		b.Trace.TraceId().SetRandom()
	}
	if b.Trace.SpanId().IsZero() {
		b.Trace.SpanId().SetRandom()
	}
}
