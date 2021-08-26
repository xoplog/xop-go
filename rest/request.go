package rest

import (
	"net/http"
	"time"

	"github.com/muir/rest"
	"github.com/muir/xm"
)

func Log(log xm.Log) *rest.RequestOpts {
	var startTime time.Time
	var step *xm.Log
	var farSideSpan xm.HexBytes
	return rest.Make().
		DoBefore(func(o *rest.RequestOpts, r *http.Request) error {
			startTime = time.Now()
			step := log.Step(o.Description, xm.Data(
				xm.String("type", "http.request"),
				xm.String("url", r.URL.String()),
				xm.String("method", r.Method)))
			r.Header.Set("traceparent", step.TracingHeader())
			if !step.TracingBaggage().IsZero() {
				r.Header.Set("baggage", step.TracingBaggage().String())
			}
			if !step.TracingState().IsZero() {
				r.Header.Set("state", step.TracingState().String())
			}
			if step.Config().UseB3 {
				farSideSpan = xm.NewSpanId()
				r.Header.Set("b3",
					step.Tracing().GetTraceId().String()+"-"+
						farSideSpan.String()+"-"+
						step.Tracing().GetFlags().String()[1:2]+"-"+
						step.Tracing().GetSpanId().String())
			}
			return nil
		}).
		DoAfter(func(result rest.Result) rest.Result {
			fields := make([]xm.Field, 0, 20)
			fields = append(fields,
				xm.String("type", "http.request"),
				xm.String("url", result.Request.URL.String()),
				xm.String("method", result.Request.Method),
				xm.Duration("duration", time.Now().Sub(startTime)))

			if result.Error != nil {
				fields = append(fields, xm.NamedError("error", result.Error))
			} else {
				fields = append(fields, xm.Int("http.status", result.Response.StatusCode))
				tr := result.Response.Header.Get("traceresponse")
				if tr != "" {
					fields = append(fields, xm.String("traceresponse", tr))
				}
				if !farSideSpan.IsZero() {
					fields = append(fields, xm.String("b3.spanid", farSideSpan.String()))
				}
				if result.DecodeTarget != nil {
					fields = append(fields, xm.Any("response", result.DecodeTarget))
				}
				if result.Options.HasData {
					fields = append(fields, xm.Any("request", result.Options.Data))
				}
			}
			step.Info(result.Options.Description, fields...)
			return result
		})
}
