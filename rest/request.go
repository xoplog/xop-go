package rest

import (
	"net/http"

	"github.com/muir/rest"
	"github.com/muir/xoplog"
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopconst"
)

func Log(log xoplog.Log) *rest.RequestOpts {
	var step *xoplog.Log
	var farSideSpan trace.HexBytes
	return rest.Make().
		DoBefore(func(o *rest.RequestOpts, r *http.Request) error {
			step = log.Step(o.Description,
				xoplog.WithData(xop.NewBuilder().
					Str("xop:type", "http.request").
					Str("xop:url", r.URL.String()).
					Str("xop:method", r.Method).
					Bool("xop:is-request", false).
					Things()...),
			)
			r.Header.Set("traceparent", step.Span().Trace().HeaderString())
			if !step.Span().TraceBaggage().IsZero() {
				r.Header.Set("baggage", step.Span().TraceBaggage().String())
			}
			if !step.Span().TraceState().IsZero() {
				r.Header.Set("state", step.Span().TraceState().String())
			}
			if step.Config().UseB3 {
				farSideSpan = trace.NewSpanId()
				r.Header.Set("b3",
					step.Span().Trace().GetTraceId().String()+"-"+
						farSideSpan.String()+"-"+
						step.Span().Trace().GetFlags().String()[1:2]+"-"+
						step.Span().Trace().GetSpanId().String())
			}
			return nil
		}).
		DoAfter(func(result rest.Result) rest.Result {
			fields := make([]xop.Thing, 0, 20)
			fields = append(fields,
				xop.Str("type", "http.request"),
				xop.Str("url", result.Request.URL.String()),
				xop.Str("method", result.Request.Method))
			// TODO: xop.Duration("duration", time.Now().Sub(startTime)))

			if result.Error != nil {
				fields = append(fields, xop.Error("error", result.Error))
			} else {
				fields = append(fields, xop.Int("http.status", result.Response.StatusCode))
				tr := result.Response.Header.Get("traceresponse")
				if tr != "" {
					fields = append(fields, xop.Str("traceresponse", tr))
				}
				if !farSideSpan.IsZero() {
					fields = append(fields, xop.Str("b3.spanid", farSideSpan.String()))
				}
				if result.DecodeTarget != nil {
					fields = append(fields, xop.Any("response", result.DecodeTarget))
				}
				if result.Options.HasData {
					fields = append(fields, xop.Any("request", result.Options.Data))
				}
			}
			step.LogThings(xopconst.InfoLevel, result.Options.Description, fields...)
			return result
		})
}
