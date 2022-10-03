// xopconst provides pre-registered standard attributes for describing spans
package xopconst

import (
	"github.com/xoplog/xop-go/xopat"
)

// ParentLink is added automatically by xop in all situations where the information is present
var ParentLink = xopat.Make{Key: "span.parent", Namespace: "xop", Indexed: true, Description: "Parent span"}.LinkAttribute()

var EndpointRoute = xopat.Make{Key: "http.route", Namespace: "xop", Indexed: true, Prominence: 10,
	Description: "HTTP handler route used to handle the request." +
		" If there are path parameters in the route their generic names should be used," +
		" eg '/invoice/{number}' or '/invoice/:number' depending on the router used"}.StringAttribute()

var Boring = xopat.Make{Key: "boring", Namespace: "xop", Indexed: false, Prominence: 200,
	Description: "spans are boring if they're an internal span (created by log.Fork() or" +
		" log.Step()) or they're a request and log.Boring() has been called, and if" +
		" there have has been nothing logged at the Error or Alert level"}.BoolAttribute()

var SpanSequenceCode = xopat.Make{Key: "span.seq", Namespace: "xop", Indexed: false, Prominence: 500,
	Description: "sub-spans only: an indicator of how the sub-span relates to it's parent" +
		" span.  A .n number indicates a sequential setp.  A .l letter indicates one fork of" +
		" several things happening in parallel.  Automatically added to all sub-spans"}.StringAttribute()

var SpanType = xopat.Make{Key: "span.type", Namespace: "xop", Indexed: true, Prominence: 11,
	Description: "what kind of span this is.  Often added automatically.  eg: SpanTypeHTTPClientRequest"}.
	IotaEnumAttribute()

var (
	SpanTypeHTTPServerEndpoint = SpanType.Iota("endpoint")
	SpanTypeHTTPClientRequest  = SpanType.Iota("REST")
	SpanTypeCronJob            = SpanType.Iota("cron_job")
)

var RemoteTrace = xopat.Make{Key: "http.remote_trace", Namespace: "xop", Indexed: true, Prominence: 40,
	Description: "The traceID and spanID for for the remote side of a outgoing HTTP request, if known"}.
	LinkAttribute()
