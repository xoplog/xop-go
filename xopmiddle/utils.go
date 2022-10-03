package xopmiddle

import (
	"regexp"

	"github.com/xoplog/xop-go/trace"
)

// SetByParentTraceHeader sets bundle.ParentTrace.TraceID and
// then copies bundle.ParentTrace to bundle.Trace.  It then sets
// the bundle.Trace.SpanID to random.
//
// "traceparent" header
// Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
func SetByParentTraceHeader(b *trace.Bundle, h string) {
	parent, ok := trace.TraceFromString(h)
	if !ok {
		b.Trace = trace.NewTrace()
		b.Trace.TraceID().SetRandom()
		b.Trace.SpanID().SetRandom()
		return
	}
	b.ParentTrace = parent
	b.Trace = parent
	b.Trace.SpanID().SetRandom()
}

var b3RE = regexp.MustCompile(`^([a-fA-F0-9]{32})-([a-fA-F0-9]{16})-(0|1|true|false|d)(?:-([a-fA-F0-9]{16}))?$`)

// https://github.com/openzipkin/b3-propagation
// b3: traceid-spanid-sampled-parentspanid
func SetByB3Header(b *trace.Bundle, h string) {
	switch h {
	case "0", "1", "true", "false", "d":
		b.ParentTrace = trace.NewTrace()
		SetByB3Sampled(&b.ParentTrace, h)
		b.Trace = b.ParentTrace
		b.Trace.TraceID().SetRandom()
		b.Trace.SpanID().SetRandom()
		return
	}
	m := b3RE.FindStringSubmatch(h)
	if m == nil {
		return
	}
	b.ParentTrace.TraceID().SetString(m[1])
	SetByB3Sampled(&b.ParentTrace, m[3])
	if m[4] == "" {
		b.ParentTrace.SpanID().SetZero()
	} else {
		b.ParentTrace.SpanID().SetString(m[4])
	}
	b.Trace = b.ParentTrace
	b.Trace.SpanID().SetString(m[2])
}

// SetByB3Sampled process the "X-B3-Sampled" header or
// the sampled portion of a combined "b3" header
// Potentially the "d" value could be used to decrease
// the minimum logging level.
func SetByB3Sampled(t *trace.Trace, h string) {
	switch h {
	case "1", "true", "d":
		t.Flags().SetBytes([]byte{1})
	case "0", "false":
		t.Flags().SetBytes([]byte{0})
	}
}
