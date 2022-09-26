package xopmiddle

import (
	"context"
	"net/http"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopprop"
)

type inbound struct {
	requestToName func(*http.Request) string
	seed          xop.Seed
}

func New(seed xop.Seed, requestToName func(*http.Request) string) inbound {
	return inbound{
		seed:          seed,
		requestToName: requestToName,
	}
}

func (i inbound) HandlerFuncMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log, ctx := i.makeChildSpan(r)
			defer log.Done()
			r = r.WithContext(log.IntoContext(ctx))
			next(w, r)
		}
	}
}

func (i inbound) HandlerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log, ctx := i.makeChildSpan(r)
			defer log.Done()
			r = r.WithContext(log.IntoContext(ctx))
			next.ServeHTTP(w, r)
		})
	}
}

// InjectorWithContext is compatible with https://github.com/muir/nject/nvelope and
// provides a *xop.Log to the injection chain.  It also puts the log in
// the request context.
func (i inbound) InjectorWithContext() func(inner func(*xop.Log, *http.Request), r *http.Request) {
	return func(inner func(*xop.Log, *http.Request), r *http.Request) {
		log, ctx := i.makeChildSpan(r)
		defer log.Done()
		r = r.WithContext(log.IntoContext(ctx))
		inner(log, r)
	}
}

// InjectorWithContext is compatible with https://github.com/muir/nject/nvelope and
// provides a *xop.Log to the injection chain.
func (i inbound) Injector() func(inner func(*xop.Log), r *http.Request) {
	return func(inner func(*xop.Log), r *http.Request) {
		log, _ := i.makeChildSpan(r)
		defer log.Done()
		inner(log)
	}
}

func (i inbound) makeChildSpan(r *http.Request) (*xop.Log, context.Context) {
	name := i.requestToName(r)
	if name == "" {
		name = r.URL.String()
	}

	bundle := i.seed.Bundle()

	if b3 := r.Header.Get("b3"); b3 != "" {
		xopprop.SetByB3Header(&bundle, b3)
	} else if tp := r.Header.Get("traceparent"); tp != "" {
		xopprop.SetByTraceParentHeader(&bundle, tp)
	} else if b3TraceID := r.Header.Get("X-B3-TraceId"); b3TraceID != "" {
		bundle.Trace.TraceID().SetString(b3TraceID)
		if b3ParentSpanID := r.Header.Get("X-B3-ParentSpanId"); b3ParentSpanID != "" {
			xopprop.SetByB3ParentSpanID(&bundle, b3ParentSpanID)
		} else {
			// Uh oh, no parent span id
			bundle.TraceParent = trace.NewTrace()
		}
		if b3SpanID := r.Header.Get("X-B3-SpanId"); b3SpanID != "" {
			bundle.Trace.SpanID().SetString(b3SpanID)
		} else {
			bundle.Trace.SpanID().SetRandom()
		}

		if b3Sampling := r.Header.Get("X-B3-Sampled"); b3Sampling != "" {
			xopprop.SetByB3Sampled(&bundle, b3Sampling)
		}
	} else {
		bundle.Trace.TraceID().SetRandom()
		bundle.Trace.SpanID().SetRandom()
	}

	ctx := r.Context()
	log := i.seed.Copy(
		xop.WithContext(ctx),
		xop.WithBundle(bundle),
	).Request(r.Method + " " + name)
	log.Span().Enum(xopconst.SpanKind, xopconst.SpanKindClient)
	log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
	log.Span().String(xopconst.URL, r.URL.String())
	return log, ctx
}
