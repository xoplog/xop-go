package xopprop

import (
	"strings"

	"github.com/muir/xop"
)

// "traceparent" header
// Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
func SetByTraceParentHeader(s *xop.Seed, h string) {
	splits := strings.SplitN(h, "-", 5)
	if len(splits) != 4 {
		return
	}
	s.TraceParent().Version().SetString(splits[0])
	s.Trace().Version().SetBytes(s.TraceParent().Version().Bytes())

	s.TraceParent().TraceId().SetString(splits[1])
	s.Trace().TraceId().SetBytes(s.TraceParent().TraceId().Bytes())

	s.TraceParent().SpanId().SetString(splits[2])
	s.Trace().SpanId().SetRandom()

	s.TraceParent().Flags().SetString(splits[3])
	s.Trace().Flags().SetBytes(s.TraceParent().Flags().Bytes())
}

// https://github.com/openzipkin/b3-propagation
// b3: traceid-spanid-sampled-parentspanid
func SetByB3Header(s *xop.Seed, h string) {
	switch h {
	case "0", "1":
		SetByB3Sampled(s, h)
		return
	}
	splits := strings.SplitN(h, "-", 5)
	if len(splits) != 4 {
		return
	}
	s.TraceParent().TraceId().SetString(splits[0])
	s.Trace().Version().SetBytes(s.TraceParent().Version().Bytes())

	s.Trace().SpanId().SetString(splits[1])

	s.TraceParent().Flags().SetString(splits[2])
	s.Trace().Flags().SetBytes(s.TraceParent().Flags().Bytes())

	s.TraceParent().SpanId().SetString(splits[3])
}

// X-B3-Sampled
func SetByB3Sampled(s *xop.Seed, h string) {
	switch h {
	case "1":
		s.Trace().Flags().SetBytes([]byte{1})
	case "0":
		s.Trace().Flags().SetBytes([]byte{0})
	}
	return
}

// X-B3-ParentSpanId
// Implies parent trace id is the same as my trace id
func SetByB3ParentSpanId(s *xop.Seed, h string) {
	s.TraceParent().SpanId().SetString(h)
	if s.TraceParent().SpanId().IsZero() {
		s.TraceParent().TraceId().SetZero()
	} else {
		s.TraceParent().TraceId().SetBytes(s.Trace().TraceId().Bytes())
	}
	if s.Trace().TraceId().IsZero() {
		s.Trace().TraceId().SetRandom()
	}
	if s.Trace().SpanId().IsZero() {
		s.Trace().SpanId().SetRandom()
	}
}
