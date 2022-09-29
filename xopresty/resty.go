/*
package xopresty adds to the resty package to
propagate xop context to through an HTTP request.

The resty package does not provide a clean way to
pass in a logger or context.  To get around this 
we'll need a fresh resty client for each request.

Another challenge with resty is that the resty Logger
is per-Client, not per-Requet.

Thiw would all be simpler if there was a Copy method
for resty Clients, but there isn't.
*/
package xopresty

import (
	"github.com/muir/xop-go"

	"github.com/go-resty/resty/v2"
)

var _ resty.Logger = restyLogger{}

type restyLogger struct {
	log *xop.Log
}

func (rl restyLogger) Errorf(format string, v ...interface{}) { rl.log.Error().Msgf(format, v...) }
func (rl restyLogger) Warnf(format string, v ...interface{})  { rl.log.Warn().Msgf(format, v...) }
func (rl restyLogger) Debugf(format string, v ...interface{}) { rl.log.Debug().Msgf(format, v...) }

func wrap(log *xop.Log, client *resty.Client, description string) *resty.Client {
	log = log.Sub().Step(description)
	var b3Sent bool
	var b3Trace trace.Trace
	var response bool
	defer func() {
		if b3Sent && !response {
			log.Info().Link("span.far_side_id", b3Trace).Static("span id set with B3")
			log.Span().Link("span.far_side_id", b3Trace)
		}
		log.Done()
	}

	// c := *client
	// c.Header = client.Header.Clone()
	// clinet = &c
	return client.
		SetLogger(restyLogger{log: log}).
		OnBeforeRequest(func(_ *Client, r *Request) error {
			log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
			log.Span().String(xopconst.URL, r.URL.String())
			log.Span().String(xopconst.HTTPMethod, r.Method)
			r.Header.Set("traceparent", log.Span().Bundle().Trace.String())
			if !log.Span().TraceBaggage().IsZero() {
				r.Header.Set("baggage", log.Span().TraceBaggage().String())
			}
			if !log.Span().TraceState().IsZero() {
				r.Header.Set("state", log.Span().TraceState().String())
			}
			if log.Config().UseB3 {	
				b3Trace := log.Span().Bundle().Trace
				b3Trace.SpanID().SetRandom()
				r.Header.Set("b3",
					b3Trace.GetTraceID().String()+"-"+
					b3Trace.TraceID().String()+"-"+
					b3Trace.GetFlags().String()[1:2]+"-"+
					log.Span().Trace().GetSpanID().String())
			}
			return nil
		}).
		OnAfterResponse(func(_ *Client, r *Response) error {
			tr := r.Header().Get("traceresponse") 
			if tr != "" {
				trace, ok := trace.TraceFromString(tr)
				if ok {	
					response = true
					log.Info().Link("span.far_side_id", trace).Static("traceresponse")
					log.Span().Link("span.far_side_id", trace)
				} else {
					log.Warn().String("header", tr).Static("invalid traceresponse received")
				}
			}
			if res != nil {
				log.Info().Any("response", r.Result())
			}
			ti := r.Request.TraceInfo()
			log.Info().
				Duration("request_time.total", ti.TotalTime).
				Duration("request_time.server", ti.ServerTime).
				Duration("request_time.dns", ti.DNSLookup).
				Static("timings")


			return nil
		}).
		EnableTrace().
				
			


}
