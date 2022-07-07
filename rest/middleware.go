package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/muir/xoplog"
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopconst"
	"github.com/muir/xoplog/xopprop"
)

var HTTPRequestSpanType = xopconst.RegisterSpanType("xop", "http-request",
	[]string{"xop:endpoint", "xop:url"},
	nil,
	xopconst.AllSpans)

func makeChildSpan(parent xoplog.Log, r *http.Request) *xoplog.Log {
	route := mux.CurrentRoute(r)
	name := route.GetName()
	if name == "" {
		name = r.URL.String()
	}

	bundle := parent.Span().Bundle()

	if b3 := r.Header.Get("b3"); b3 != "" {
		xopprop.SetByB3Header(&bundle, b3)
	} else if tp := r.Header.Get("traceparent"); tp != "" {
		xopprop.SetByTraceParentHeader(&bundle, tp)
	} else if b3TraceId := r.Header.Get("X-B3-TraceId"); b3TraceId != "" {
		bundle.Trace.TraceId().SetString(b3TraceId)
		if b3ParentSpanId := r.Header.Get("X-B3-ParentSpanId"); b3ParentSpanId != "" {
			xopprop.SetByB3ParentSpanId(&bundle, b3ParentSpanId)
		} else {
			// Uh oh, no parent span id
			bundle.TraceParent = trace.NewTrace()
		}
		if b3SpanId := r.Header.Get("X-B3-SpanId"); b3SpanId != "" {
			bundle.Trace.SpanId().SetString(b3SpanId)
		} else {
			bundle.Trace.SpanId().SetRandom()
		}

		if b3Sampling := r.Header.Get("X-B3-Sampled"); b3Sampling != "" {
			xopprop.SetByB3Sampled(&bundle, b3Sampling)
		}
	} else {
		bundle.Trace.TraceId().SetRandom()
		bundle.Trace.SpanId().SetRandom()
	}

	log := parent.Span().Seed(xoplog.WithBundle(bundle)).Request(r.Method + " " + name)
	log.Span().SetType(HTTPRequestSpanType)
	log.Span().AddData(
		xop.NewBuilder().
			Str("xop:type", "http:endpoint").
			Str("xop:endpoint", name).
			Str("xop:url", r.URL.String()).
			Bool("xop:is-request", true).
			Things()...)
	return log
}

func ParentLogMiddleware(parentLog xoplog.Log) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := makeChildSpan(parentLog, r)
			defer log.Done()
			r = r.WithContext(log.IntoContext(ctx))
			next(w, r)
		}
	}
}

// MakeLogInjector is compatible with https://github.com/muir/nject/nvelope
func MakeLogInjector(parentLog xoplog.Log) func(func(*xoplog.Log), *http.Request) {
	return func(inner func(*xoplog.Log), r *http.Request) {
		log := makeChildSpan(parentLog, r)
		defer log.Done()
		inner(log)
	}
}
