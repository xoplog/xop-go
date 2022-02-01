package rest

import (
	"net/http"
	"time"

	"github.com/muir/rest"
	"github.com/muir/xm"
	"github.com/muir/xm/trace"
	"github.com/muir/xm/zap"
)

func Log(log xm.Log) *rest.RequestOpts {
	var startTime time.Time
	var step *xm.Log
	var farSideSpan trace.HexBytes
	return rest.Make().
		DoBefore(func(o *rest.RequestOpts, r *http.Request) error {
			startTime = time.Now()
			step := log.Step(o.Description, xm.Data(map[string]interface{}{
				"type":   "http.request",
				"url":    r.URL.String(),
				"method": r.Method,
			}))
			r.Header.Set("traceparent", step.TracingHeader())
			if !step.TracingBaggage().IsZero() {
				r.Header.Set("baggage", step.TracingBaggage().String())
			}
			if !step.TracingState().IsZero() {
				r.Header.Set("state", step.TracingState().String())
			}
			if step.Config().UseB3 {
				farSideSpan = trace.NewSpanId()
				r.Header.Set("b3",
					step.Tracing().GetTraceId().String()+"-"+
						farSideSpan.String()+"-"+
						step.Tracing().GetFlags().String()[1:2]+"-"+
						step.Tracing().GetSpanId().String())
			}
			return nil
		}).
		DoAfter(func(result rest.Result) rest.Result {
			fields := make([]zap.Field, 0, 20)
			fields = append(fields,
				zap.String("type", "http.request"),
				zap.String("url", result.Request.URL.String()),
				zap.String("method", result.Request.Method),
				zap.Duration("duration", time.Now().Sub(startTime)))

			if result.Error != nil {
				fields = append(fields, zap.NamedError("error", result.Error))
			} else {
				fields = append(fields, zap.Int("http.status", result.Response.StatusCode))
				tr := result.Response.Header.Get("traceresponse")
				if tr != "" {
					fields = append(fields, zap.String("traceresponse", tr))
				}
				if !farSideSpan.IsZero() {
					fields = append(fields, zap.String("b3.spanid", farSideSpan.String()))
				}
				if result.DecodeTarget != nil {
					fields = append(fields, zap.Any("response", result.DecodeTarget))
				}
				if result.Options.HasData {
					fields = append(fields, zap.Any("request", result.Options.Data))
				}
			}
			step.Info(result.Options.Description, fields...)
			return result
		})
}
