package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopprop"
)

func makeChildSpan(parent xop.Log, r *http.Request) *xop.Log {
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

	log := parent.Span().Seed(xop.WithBundle(bundle)).Request(r.Method + " " + name)
	log.Span().Enum(xopconst.SpanKind, xopconst.SpanKindClient)
	log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)
	log.Span().Str(xopconst.URL, r.URL.String())
	return log
}

func ParentLogMiddleware(parentLog xop.Log) func(http.HandlerFunc) http.HandlerFunc {
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
func MakeLogInjector(parentLog xop.Log) func(func(*xop.Log), *http.Request) {
	return func(inner func(*xop.Log), r *http.Request) {
		log := makeChildSpan(parentLog, r)
		defer log.Done()
		inner(log)
	}
}
