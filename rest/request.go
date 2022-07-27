package rest

import (
	"fmt"
	"net/http"

	"github.com/muir/rest"
	"github.com/muir/xop"
	"github.com/muir/xop/trace"
	"github.com/muir/xop/xopconst"
)

var (
	errorTemplate = fmt.Sprintf("{description} FAILED {%s}: {error}", xopconst.HTTPMethod.Key())
	template      = fmt.Sprintf("{%s} {description} {%s} {%s}",
		xopconst.HTTPStatusCode.Key(),
		xopconst.HTTPMethod.Key(),
		xopconst.URL.Key())
)

func Log(log xop.Log) *rest.RequestOpts {
	var step *xop.Log
	var farSideSpan trace.HexBytes8
	return rest.Make().
		DoBefore(func(o *rest.RequestOpts, r *http.Request) error {
			step = log.Step(o.Description)
			step.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
			step.Span().Enum(xopconst.SpanKind, xopconst.SpanKindClient)
			step.Span().Str(xopconst.URL, r.URL.String())
			step.Span().Str(xopconst.HTTPMethod, r.Method)
			r.Header.Set("traceparent", step.Span().Trace().HeaderString())
			if !step.Span().TraceBaggage().IsZero() {
				r.Header.Set("baggage", step.Span().TraceBaggage().String())
			}
			if !step.Span().TraceState().IsZero() {
				r.Header.Set("state", step.Span().TraceState().String())
			}
			if step.Config().UseB3 {
				farSideSpan = trace.NewSpanID()
				r.Header.Set("b3",
					step.Span().Trace().GetTraceID().String()+"-"+
						farSideSpan.String()+"-"+
						step.Span().Trace().GetFlags().String()[1:2]+"-"+
						step.Span().Trace().GetSpanID().String())
			}
			return nil
		}).
		DoAfter(func(result rest.Result) rest.Result {
			var line *xop.LogLine
			if result.Error != nil {
				line = step.Error()
			} else {
				line = step.Info()
			}

			line = line.Str(xopconst.HTTPMethod.Key(), result.Request.Method).
				Str(xopconst.URL.Key(), result.Request.URL.String()).
				Str("description", result.Options.Description)

			if result.Error != nil {
				line = line.Error("error", result.Error)
				line.Template(errorTemplate)
			} else {
				line = line.Int(xopconst.HTTPStatusCode.Key(), result.Response.StatusCode)
				tr := result.Response.Header.Get("traceresponse")
				if tr != "" {
					line = line.Str(xopconst.TraceResponse.Key(), tr)
				}
				if !farSideSpan.IsZero() {
					// TODO: standard name?
					// TODO: use Link()
					line = line.Str("b3.spanid", farSideSpan.String())
				}
				if result.DecodeTarget != nil {
					line = line.Any("response.data", result.DecodeTarget)
				}
				if result.Options.HasData {
					line = line.Any("request.data", result.Options.Data)
				}
				line.Template(template)
			}
			return result
		})
}
