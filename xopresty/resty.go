/*
package xopresty adds to the resty package to
propagate xop context to through an HTTP request.

As of March 29th, 2023, the released resty package does not provide
a way to have a logger that knows which request it is logging about.
The resty package does not provide a way to know when requests are
complete.

Pull requests to fix these issues have been merged but not
made part of a release.

In the meantime, this package depends upon https://github.com/muir/resty.

The agumented resty Client requires that a context that
has the parent log span be provided:

	client.R().SetContext(log.IntoContext(context.Background()))

If there is no logger in the context, the request will fail.

If you use resty's Client.SetDebug(true), note that the output
will be logged at Debug level which is below the default
minimum level for xop.
*/
package xopresty

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/muir/resty"
	"github.com/pkg/errors"
)

var _ resty.Logger = restyLogger{}

type restyLogger struct {
	log *xop.Logger
}

func (rl restyLogger) Errorf(format string, v ...interface{}) { rl.log.Error().Msgf(format, v...) }
func (rl restyLogger) Warnf(format string, v ...interface{})  { rl.log.Warn().Msgf(format, v...) }
func (rl restyLogger) Debugf(format string, v ...interface{}) { rl.log.Debug().Msgf(format, v...) }

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextNameType struct{}

var contextNameKey = contextNameType{}

type contextValue struct {
	b3Sent            bool
	b3Trace           xoptrace.Trace
	response          bool
	log               *xop.Logger
	retryCount        int
	originalStartTime time.Time
}

type config struct {
	requestToName func(r *resty.Request) string
	extraLogging  ExtraLogging
}

type ClientOpt func(*config)

var traceResponseHeaderKey = xop.Key("header")
var requestTimeKey = xop.Key("request_time.total")
var requestTimeServerKey = xop.Key("request_time.server")
var requestTimeDNSKey = xop.Key("request_time.dns")

// WithNameGenerate provides a function to convert a request into
// a description for the span.
func WithNameGenerate(f func(*resty.Request) string) ClientOpt {
	return func(config *config) {
		config.requestToName = f
	}
}

// ExtraLogging provides a hook for extra logging to be done.
// It is possible that the response parameter will be null.
// If error is not null, then the request has failed.
// ExtraLogging should only be called once but if another resty
// callback panic's, it is possible ExtraLogging will be called
// twice.
type ExtraLogging func(log *xop.Logger, originalStartTime time.Time, retryCount int, request *resty.Request, response *resty.Response, err error)

func WithExtraLogging(f ExtraLogging) ClientOpt {
	return func(config *config) {
		config.extraLogging = f
	}
}

// WithNameInDescription adds a span name to a context.  If present,
// a name in context overrides WithNameGenerate.
func WithNameInContext(ctx context.Context, nameOrDescription string) context.Context {
	return context.WithValue(ctx, contextNameKey, nameOrDescription)
}

func Client(client *resty.Client, opts ...ClientOpt) *resty.Client {
	config := &config{
		requestToName: func(r *resty.Request) string {
			url := r.URL
			i := strings.IndexByte(url, '?')
			if i != -1 {
				url = url[:i]
			}
			return r.Method + " " + url
		},
		extraLogging: func(log *xop.Logger, originalStartTime time.Time, retryCount int, request *resty.Request, response *resty.Response, err error) {
		},
	}
	for _, f := range opts {
		f(config)
	}

	// c := *client
	// c.Header = client.Header.Clone()
	// clinet = &c
	return client.
		OnBeforeRequest(func(_ *resty.Client, r *resty.Request) error {
			// OnBeforeRequest can execute multiple times for the same attempt if there
			// are retries.  It won't execute at all of the request is invalid.
			ctx := r.Context()
			cvRaw := ctx.Value(contextKey)
			var cv *contextValue
			if cvRaw != nil {
				cv = cvRaw.(*contextValue)
				cv.retryCount++
				return nil
			}
			log, ok := xop.FromContext(r.Context())
			if !ok {
				return errors.Errorf("context is missing logger, use Request.SetContext(Log.IntoContext(request.Context()))")
			}
			nameRaw := ctx.Value(contextNameKey)
			var name string
			if nameRaw != nil {
				name = nameRaw.(string)
			} else {
				name = config.requestToName(r)
			}
			log = log.Sub().Step(name)
			cv = &contextValue{
				originalStartTime: time.Now(),
				log:               log,
			}
			r.SetContext(context.WithValue(ctx, contextKey, cv))
			r.SetLogger(restyLogger{log: log})

			if r.Body != nil {
				log.Trace().Model(r.Body, "request")
			}

			log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
			log.Span().String(xopconst.URL, r.URL)
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
				cv.b3Trace = b3Trace
				cv.b3Sent = true
			}
			return nil
		}).
		OnAfterResponse(func(_ *resty.Client, resp *resty.Response) error {
			// OnAfterRequest is run for each individual request attempt
			r := resp.Request
			ctx := r.Context()
			cvRaw := ctx.Value(contextKey)
			var cv *contextValue
			if cvRaw == nil {
				return fmt.Errorf("xopresty: internal error, context missing in response")
			}
			cv = cvRaw.(*contextValue)
			log := cv.log

			tr := resp.Header().Get("traceresponse")
			if tr != "" {
				trace, ok := xoptrace.TraceFromString(tr)
				if ok {
					cv.response = true
					log.Info().Link(trace, xopconst.RemoteTrace.Key().String())
					log.Span().Link(xopconst.RemoteTrace, trace)
				} else {
					log.Warn().String(traceResponseHeaderKey, tr).Msg("invalid traceresponse received")
				}
			}
			if r.Result != nil {
				log.Info().Model(resp.Result(), "response")
			}
			ti := r.TraceInfo()
			if ti.TotalTime != 0 {
				log.Info().
					Duration(requestTimeKey, ti.TotalTime).
					Duration(requestTimeServerKey, ti.ServerTime).
					Duration(requestTimeDNSKey, ti.DNSLookup).
					Msg("timings")
			}
			return nil
		}).
		OnError(func(r *resty.Request, err error) {
			ctx := r.Context()
			cv := ctx.Value(contextKey).(*contextValue)
			log := cv.log
			var re *resty.ResponseError
			if errors.As(err, &re) {
				config.extraLogging(log, cv.originalStartTime, cv.retryCount, r, re.Response, re.Err)
			} else {
				config.extraLogging(log, cv.originalStartTime, cv.retryCount, r, nil, err)
			}
		}).
		OnPanic(func(r *resty.Request, err error) {
			ctx := r.Context()
			cv := ctx.Value(contextKey).(*contextValue)
			log := cv.log
			config.extraLogging(log, cv.originalStartTime, cv.retryCount, r, nil, err)
		}).
		OnSuccess(func(c *resty.Client, resp *resty.Response) {
			ctx := resp.Request.Context()
			cv := ctx.Value(contextKey).(*contextValue)
			log := cv.log
			config.extraLogging(log, cv.originalStartTime, cv.retryCount, resp.Request, resp, nil)
		})
}
