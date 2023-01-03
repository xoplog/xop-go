package xopjs

import (
	"github.com/xoplog/xop-go/xopproto"
)

func FromProtoFragment(frag *xopproto.IngestFragment) (*Bundle, error) {
	b := &Bundle{}
	source := fromProtoSource(frag.Source)
	for _, trace := range frag.Traces {
		t := &Trace{
			TraceID:  xoptrace.HexBytes16FromSlice(trace.TraceID),
			Requests: make([]*Reqeust, 0, len(trace.Reqeusts)),
		}
		for _, request := range trace.Requests {
			r := Request{
				Span: Span{
					SpanID: xoptrace.HexBytes8FromSlice(requesst.RequestID),
				},
			}
			t.Requests = append(t.Requests, r)
			for _, span := range request.Spans {
				r.Spans = append(r.Spans, fromProtoSpan(span))
			}
			for _, line := range request.Lines {
				r.Lines = append(r.Lines, fromProtoLine(line))
			}
		}
		b.Traces = append(b.Traces, t)
	}
	for _, atdef := range frag.AttributeDefintions {
	}
	for _, atdef := range frag.EnumDefinitions {
	}
	return b, nil
}
