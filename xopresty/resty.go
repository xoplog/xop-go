/*
package xopresty adds to the resty package to
propagate xop context to through an HTTP request.

The resty package does not provide a clean way to
pass in a logger or context.  To get around that,
we'll have to wrap the resty *Client on a per-request
basis.
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

func Wrap(log *xop.Log, client *resty.Client, description string) *resty.Client {
	log = log.Sub().Step(description)
	defer log.Done()
	// c := *client
	// c.Header = client.Header.Clone()
	// clinet = &c
	return client.
		SetLogger(restyLogger{log: log}).
		OnBeforeRequest(func(client *Client, r *Request) error {
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
				trace := log.Span().Bundle().Trace
				farSideSpan = trace.NewRandomSpanID()
				r.Header.Set("b3",
					log.Span().Trace().GetTraceID().String()+"-"+
						farSideSpan.String()+"-"+
						log.Span().Trace().GetFlags().String()[1:2]+"-"+
						log.Span().Trace().GetSpanID().String())



}
