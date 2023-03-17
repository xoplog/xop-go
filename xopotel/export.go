// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"

	"github.com/muir/list"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	_ sdktrace.SpanExporter = &spanExporter{}
	_ sdktrace.SpanExporter = &unhack{}
)

type spanExporter struct {
	base xopbase.Logger
}

func NewExporter(base xopbase.Logger) sdktrace.SpanExporter {
	return &spanExporter{base: base}
}

func (e *spanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	id2Index := makeIndex(spans)
	subSpans, todo := makeSubspans(id2Index, spans)
	_ = subSpans // XXX
	baseSpans := make([]xopbase.Span, len(spans))
	for _, i := range todo {
		span := spans[i]
		parentIndex, ok := lookupParent(id2Index, span)
		// attributeMap := mapAttributes(span)
		switch span.SpanKind() {
		case oteltrace.SpanKindUnspecified, oteltrace.SpanKindInternal:
			if ok {
				parentContext := spans[parentIndex].SpanContext()
				xopParent := baseSpans[parentIndex]
				var bundle xoptrace.Bundle
				if parentContext.HasTraceID() {
					bundle.Parent.TraceID().SetArray(parentContext.TraceID())
				}
				if parentContext.HasSpanID() {
					bundle.Parent.SpanID().SetArray(parentContext.SpanID())
				}
				if parentContext.IsSampled() {
					bundle.Parent.Flags().SetArray([1]byte{1})
				}
				bundle.Parent.Version().SetArray([1]byte{1})
				spanContext := span.SpanContext()
				if spanContext.HasTraceID() {
					bundle.Trace.TraceID().SetArray(spanContext.TraceID())
				} else {
					bundle.Trace.TraceID().Set(bundle.Parent.GetTraceID())
				}
				if spanContext.HasSpanID() {
					bundle.Trace.SpanID().SetArray(spanContext.SpanID())
				}
				if spanContext.IsSampled() {
					bundle.Trace.Flags().SetArray([1]byte{1})
				}
				bundle.Trace.Version().SetArray([1]byte{1})
				baseSpan := xopParent.Span(ctx, span.StartTime(), bundle, span.Name(), defaulted(attributeMap.Get(logSpanSequence), ""))
				baseSpans[i] = baseSpan
			}
		default:
		}
	}
	return nil
}

func (e *spanExporter) Shutdown(ctx context.Context) error {
	// XXX
	return nil
}

type unhack struct {
	next sdktrace.SpanExporter
}

// NewUnhacker wraps a SpanExporter and if the input is from BaseLogger or SpanLog,
// then it "fixes" the data hack in the output that puts inter-span links in sub-spans
// rather than in the span that defined them.
func NewUnhacker(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return &unhack{next: exporter}
}

func (u *unhack) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// TODO: fix up SpanKind if spanKind is one of the attributes
	id2Index := makeIndex(spans)
	subLinks := make([][]sdktrace.Link, len(spans))
	for i, span := range spans {
		parentIndex, ok := lookupParent(id2Index, span)
		if !ok {
			continue
		}
		var addToParent bool
		for _, attribute := range span.Attributes() {
			switch attribute.Key {
			case spanIsLinkAttributeKey, spanIsLinkEventKey:
				spans[i] = nil
				addToParent = true
			}
		}
		if !addToParent {
			continue
		}
		subLinks[parentIndex] = append(subLinks[parentIndex], span.Links()...)
	}
	n := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for i, span := range spans {
		span := span
		switch {
		case len(subLinks[i]) > 0:
			n = append(n, wrappedReadOnlySpan{
				ReadOnlySpan: span,
				links:        append(list.Copy(span.Links()), subLinks[i]...),
			})
		case span == nil:
			// skip
		default:
			n = append(n, span)
		}
	}
	return u.next.ExportSpans(ctx, n)
}

func (u *unhack) Shutdown(ctx context.Context) error {
	return u.next.Shutdown(ctx)
}

type wrappedReadOnlySpan struct {
	sdktrace.ReadOnlySpan
	links []sdktrace.Link
}

var _ sdktrace.ReadOnlySpan = wrappedReadOnlySpan{}

func (w wrappedReadOnlySpan) Links() []sdktrace.Link {
	return w.links
}

func makeIndex(spans []sdktrace.ReadOnlySpan) map[oteltrace.SpanID]int {
	id2Index := make(map[oteltrace.SpanID]int)
	for i, span := range spans {
		spanContext := span.SpanContext()
		if spanContext.HasSpanID() {
			id2Index[spanContext.SpanID()] = i
		}
	}
	return id2Index
}

func lookupParent(id2Index map[oteltrace.SpanID]int, span sdktrace.ReadOnlySpan) (int, bool) {
	parent := span.Parent()
	if !parent.HasSpanID() {
		return 0, false
	}
	parentIndex, ok := id2Index[parent.SpanID()]
	if !ok {
		return 0, false
	}
	return parentIndex, true
}

func makeSubspans(id2Index map[oteltrace.SpanID]int, spans []sdktrace.ReadOnlySpan) ([][]oteltrace.SpanID, []int) {
	ss := make([][]oteltrace.SpanID, len(spans))
	noParent := make([]int, 0, len(spans))
	for i, span := range spans {
		parentIndex, ok := lookupParent(id2Index, span)
		if !ok {
			noParent = append(noParent, i)
		}
		ss[parentIndex] = append(ss[parentIndex], i)
	}
	return ss, noParent
}
