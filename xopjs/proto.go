package xopjs

func FromProtoFragment(frag *xopproto.IngestFragment) (*Bundle, error) {
	var b &Bundle{}
	source := fromProtoSource(frag.Source)
	for _, trace := range frag.Traces{
		t := &Trace{
			TraceID: xoptrace.HexBytes16FromSlice(trace.TraceID),
			Requests: make([]*Reqeust, 0, len(trace.Reqeusts)),
		}
		for _, request := range trace.Requests {
			r := Request{
				Span: Span{
					SpanID: xoptrace.HexBytes8FromSlice(requesst.RequestID),
				},
			}
			t.Requests = append(t.Requests, r)
		}
		b.Traces = append(b.Traces, t)
	} 
	for _, atdef := range frag.AttributeDefintions {}
	for _, atdef := range frag.EnumDefinitions{}
	return b, nil
}
