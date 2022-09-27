package xopmiddle

import (
	"strings"

	"github.com/muir/xop-go/trace"
)

// "traceparent" header
// Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
func SetByTraceParentHeader(b *trace.Bundle, h string) {
	parent := trace.TraceFromString(h)
	if parent.IsZero() {
		b.Trace = trace.NewTrace()
		b.Trace.TraceID().SetRandom()
		b.Trace.SpanID().SetRandom()
		return
	}
	b.TraceParent = parent
	b.Trace = parent
	b.Trace.SpanID().SetRandom()
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
	b.TraceParent.TraceID().SetString(splits[0])
	b.Trace.Version().SetBytes(b.TraceParent.Version().Bytes())

	b.Trace.SpanID().SetString(splits[1])

	b.TraceParent.Flags().SetString(splits[2])
	b.Trace.Flags().SetBytes(b.TraceParent.Flags().Bytes())

	b.TraceParent.SpanID().SetString(splits[3])
}

// X-B3-Sampled
func SetByB3Sampled(b *trace.Bundle, h string) {
	switch h {
	case "1", "true", "d":
		b.Trace.Flags().SetBytes([]byte{1})
	case "0", "false":
		b.Trace.Flags().SetBytes([]byte{0})
	}
}

// X-B3-ParentSpanID
// Implies parent trace id is the same as my trace id
func SetByB3ParentSpanID(b *trace.Bundle, h string) {
	b.TraceParent.SpanID().SetString(h)
	if b.TraceParent.SpanID().IsZero() {
		b.TraceParent.TraceID().SetZero()
	} else {
		b.TraceParent.TraceID().SetBytes(b.Trace.TraceID().Bytes())
	}
	if b.Trace.TraceID().IsZero() {
		b.Trace.TraceID().SetRandom()
	}
	if b.Trace.SpanID().IsZero() {
		b.Trace.SpanID().SetRandom()
	}
}
