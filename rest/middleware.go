package rest

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/muir/xm"
	"github.com/muir/xm/propagate"
)

func makeChildSpan(parent xm.Log, r *http.Request) *xm.Log {
	route := mux.CurrentRoute(r)
	name := route.GetName()
	if name == "" {
		name = r.URL.String()
	}

	seed := parent.CopySeed()

	if b3 := r.Header.Get("b3"); b3 != "" {
		propagate.SetByB3Header(&seed, b3)
	} else if tp := r.Header.Get("traceparent"); tp != "" {
		propagate.SetByTraceParentHeader(&seed, tp)
	} else if b3TraceId := r.Header.Get("X-B3-TraceId"); b3TraceId != "" {
		seed.Trace().TraceId().SetString(b3TraceId)
		if b3ParentSpanId := r.Header.Get("X-B3-ParentSpanId"); b3ParentSpanId != "" {
			propagate.SetByB3ParentSpanId(&seed, b3ParentSpanId)
		} else {
			// Uh oh, no parent span id
			*seed.TraceParent() = xm.NewTrace()
		}
		if b3SpanId := r.Header.Get("X-B3-SpanId"); b3SpanId != "" {
			seed.Trace().SpanId().SetString(b3SpanId)
		} else {
			seed.Trace().SpanId().SetRandom()
		}

		if b3Sampling := r.Header.Get("X-B3-Sampled"); b3Sampling != "" {
			propagate.SetByB3Sampled(&seed, b3Sampling)
		}
	} else {
		seed.Trace().TraceId().SetRandom()
		seed.Trace().SpanId().SetRandom()
	}

	log := seed.Log(r.Method + " " + name)

	log.SpanIndex(
		"type", "http.endpoint",
		"endpoint", name,
		"url", r.URL.String(),
	)
	return log
}

func ParentLogMiddleware(parentLog xm.Log) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := makeChildSpan(parentLog, r)
			defer log.End()
			r = r.WithContext(log.IntoContext(ctx))
			startTime := time.Now()
			next(w, r)
			log.LocalSpanData(map[string]interface{}{
				"duration": time.Now().Sub(startTime),
			})
		}
	}
}

// MakeLogInjector is compatible with https://github.com/muir/nject/nvelope
func MakeLogInjector(parentLog xm.Log) func(func(*xm.Log), *http.Request) {
	return func(inner func(*xm.Log), r *http.Request) {
		log := makeChildSpan(parentLog, r)
		startTime := time.Now()
		defer log.End()
		inner(log)
		log.LocalSpanData(map[string]interface{}{
			"duration": time.Now().Sub(startTime),
		})
	}
}
