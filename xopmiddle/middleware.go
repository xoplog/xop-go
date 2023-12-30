package xopmiddle

import (
	"context"
	"net/http"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopconst"
)

type Inbound struct {
	requestToName func(*http.Request) string
	seed          xop.Seed
}

func New(seed xop.Seed, requestToName func(*http.Request) string) Inbound {
	return Inbound{
		seed:          seed,
		requestToName: requestToName,
	}
}

func (i Inbound) HandlerFuncMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log, ctx := i.makeChildSpan(w, r)
			defer log.Done()
			r = r.WithContext(log.IntoContext(ctx))
			next(w, r)
		}
	}
}

func (i Inbound) HandlerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log, ctx := i.makeChildSpan(w, r)
			defer log.Done()
			r = r.WithContext(log.IntoContext(ctx))
			next.ServeHTTP(w, r)
		})
	}
}

// InjectorWithContext is compatible with https://github.com/muir/nject/nvelope and
// provides a *xop.Logger to the injection chain.  It also puts the log in
// the request context.
func (i Inbound) InjectorWithContext() func(inner func(*xop.Logger, *http.Request), w http.ResponseWriter, r *http.Request) {
	return func(inner func(*xop.Logger, *http.Request), w http.ResponseWriter, r *http.Request) {
		log, ctx := i.makeChildSpan(w, r)
		defer log.Done()
		r = r.WithContext(log.IntoContext(ctx))
		inner(log, r)
	}
}

// InjectorWithContext is compatible with https://github.com/muir/nject/nvelope and
// provides a *xop.Logger to the injection chain.
func (i Inbound) Injector() func(inner func(*xop.Logger), w http.ResponseWriter, r *http.Request) {
	return func(inner func(*xop.Logger), w http.ResponseWriter, r *http.Request) {
		log, _ := i.makeChildSpan(w, r)
		defer log.Done()
		inner(log)
	}
}

func (i Inbound) makeChildSpan(w http.ResponseWriter, r *http.Request) (*xop.Logger, context.Context) {
	name := i.requestToName(r)
	if name == "" {
		name = r.URL.String()
	}

	bundle := i.seed.Bundle()

	if b3 := r.Header.Get("b3"); b3 != "" {
		SetByB3Header(&bundle, b3)
	} else if tp := r.Header.Get("traceparent"); tp != "" {
		SetByParentTraceHeader(&bundle, tp)
	} else if b3TraceID := r.Header.Get("X-B3-TraceId"); b3TraceID != "" {
		bundle.Trace.TraceID().SetString(b3TraceID)
		if b3Sampling := r.Header.Get("X-B3-Sampled"); b3Sampling != "" {
			SetByB3Sampled(&bundle.Trace, b3Sampling)
		}
		bundle.Parent = bundle.Trace
		if b3ParentSpanID := r.Header.Get("X-B3-ParentSpanId"); b3ParentSpanID != "" {
			bundle.Parent.SpanID().SetString(b3ParentSpanID)
		} else {
			// Uh oh, no parent span id
			bundle.Parent.SpanID().SetZero()
		}
		if b3SpanID := r.Header.Get("X-B3-SpanId"); b3SpanID != "" {
			bundle.Trace.SpanID().SetString(b3SpanID)
		} else {
			bundle.Trace.SpanID().SetRandom()
		}
	}
	if bundle.Trace.TraceID().IsZero() {
		bundle.Trace.TraceID().SetRandom()
	}
	if bundle.Trace.SpanID().IsZero() {
		bundle.Trace.SpanID().SetRandom()
	}

	ctx := r.Context()
	log := i.seed.Copy(
		xop.WithContext(ctx),
		xop.WithBundle(bundle),
	).Request(r.Method + " " + name)

	w.Header().Set("traceresponse", log.Span().Trace().String())
	log.Span().Enum(xopconst.SpanKind, xopconst.SpanKindClient)
	log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
	log.Span().String(xopconst.URL, r.URL.String())
	return log, ctx
}
