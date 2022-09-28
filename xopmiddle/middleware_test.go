package xopmiddle_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopmiddle"
	"github.com/muir/xop-go/xoptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var headerCases = []struct {
	name              string
	headers           []string
	expectParentTrace string
	expectParentSpan  string
	expectParentFlags string
	expectTrace       string // defaults to expectParentTrace
	expectSpan        string // defaults to random
	expectFlags       string
}{
	{
		name:              "traceparent set",
		headers:           []string{"traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"},
		expectParentTrace: "0af7651916cd43dd8448eb211c80319c",
		expectParentSpan:  "b7ad6b7169203331",
		expectParentFlags: "01",
		expectFlags:       "01",
	},
	{
		name:              "no header",
		headers:           nil,
		expectParentTrace: "00000000000000000000000000000000",
		expectParentSpan:  "0000000000000000",
		expectParentFlags: "01",
		expectTrace:       "random",
		expectFlags:       "01",
	},
	{
		name: "b3 with everything",
		headers: []string{
			"X-B3-TraceId", "0af7651916cd43dd8448eb211c80319c",
			"X-B3-ParentSpanId", "b7ad6b7169203331",
			"X-B3-SpanId", "91e961630d5d22de",
			"X-B3-Sampled", "1",
		},
		expectParentTrace: "0af7651916cd43dd8448eb211c80319c",
		expectParentSpan:  "b7ad6b7169203331",
		expectParentFlags: "01",
		expectSpan:        "91e961630d5d22de",
		expectFlags:       "01",
	},
	{
		name: "b3 without span",
		headers: []string{
			"X-B3-TraceId", "0af7651916cd43dd8448eb211c80319c",
			"X-B3-ParentSpanId", "b7ad6b7169203331",
		},
		expectParentTrace: "0af7651916cd43dd8448eb211c80319c",
		expectParentSpan:  "b7ad6b7169203331",
		expectParentFlags: "01",
		expectFlags:       "01",
	},
	{
		name: "b3 single line",
		headers: []string{
			"b3", "80f198ee56343ba864fe8b2a57d3eff7-e457b5a2e4d86bd1-1",
		},
		expectParentTrace: "80f198ee56343ba864fe8b2a57d3eff7",
		expectParentSpan:  "0000000000000000",
		expectSpan:        "e457b5a2e4d86bd1",
		expectParentFlags: "01",
		expectFlags:       "01",
	},
	{
		name: "b3 single line with parent",
		headers: []string{
			"b3", "80f198ee56343ba864fe8b2a57d3eff7-e457b5a2e4d86bd1-1-05e3ac9a4f6e3b90",
		},
		expectParentTrace: "80f198ee56343ba864fe8b2a57d3eff7",
		expectParentSpan:  "05e3ac9a4f6e3b90",
		expectSpan:        "e457b5a2e4d86bd1",
		expectParentFlags: "01",
		expectFlags:       "01",
	},
	{
		name: "b3 sampled",
		headers: []string{
			"X-B3-TraceId", "0af7651916cd43dd8448eb211c80319c",
			"X-B3-ParentSpanId", "b7ad6b7169203331",
			"X-B3-Sampled", "0",
		},
		expectParentTrace: "0af7651916cd43dd8448eb211c80319c",
		expectParentSpan:  "b7ad6b7169203331",
		expectParentFlags: "00",
		expectFlags:       "00",
	},
	{
		name:              "no headers",
		headers:           []string{},
		expectParentTrace: "00000000000000000000000000000000",
		expectParentSpan:  "0000000000000000",
		expectTrace:       "random",
		expectParentFlags: "01",
		expectFlags:       "01",
	},
}

var injectMethods = []struct {
	name string
	f    func(t *testing.T, inbound xopmiddle.Inbound, w http.ResponseWriter, r *http.Request)
}{
	{
		name: "handlerFunc",
		f: func(t *testing.T, inbound xopmiddle.Inbound, w http.ResponseWriter, r *http.Request) {
			var called bool
			handler := func(w http.ResponseWriter, r *http.Request) {
				_ = xop.FromContextOrPanic(r.Context())
				called = true
			}
			inbound.HandlerFuncMiddleware()(handler)(w, r)
			assert.True(t, called, "called")
		},
	},
	{
		name: "handler",
		f: func(t *testing.T, inbound xopmiddle.Inbound, w http.ResponseWriter, r *http.Request) {
			var called bool
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = xop.FromContextOrPanic(r.Context())
				called = true
			})
			inbound.HandlerMiddleware()(handler).ServeHTTP(w, r)
			assert.True(t, called, "called")
		},
	},
	{
		name: "injector",
		f: func(t *testing.T, inbound xopmiddle.Inbound, w http.ResponseWriter, r *http.Request) {
			var called bool
			inbound.Injector()(func(log *xop.Log) {
				assert.NotNil(t, log, "log set")
				called = true
			}, w, r)
			assert.True(t, called, "called")
		},
	},
	{
		name: "injector with context",
		f: func(t *testing.T, inbound xopmiddle.Inbound, w http.ResponseWriter, r *http.Request) {
			var called bool
			inbound.InjectorWithContext()(func(log *xop.Log, r *http.Request) {
				_ = xop.FromContextOrPanic(r.Context())
				assert.NotNil(t, log, "log set")
				called = true
			}, w, r)
			assert.True(t, called, "called")
		},
	},
}

func TestHandlerFuncMiddleware(t *testing.T) {
	for _, hc := range headerCases {
		hc := hc
		t.Run(hc.name, func(t *testing.T) {
			for _, im := range injectMethods {
				im := im
				t.Run(im.name, func(t *testing.T) {

					tLog := xoptest.New(t)
					seed := xop.NewSeed(xop.WithBase(tLog))
					inbound := xopmiddle.New(seed, func(r *http.Request) string {
						return r.URL.String()
					})
					r, err := http.NewRequest("GET", "/foo", nil)
					require.NoError(t, err, "new request")
					w := httptest.NewRecorder()
					for i := 0; i < len(hc.headers); i += 2 {
						r.Header.Set(hc.headers[i], hc.headers[i+1])
					}

					im.f(t, inbound, w, r)

					require.Equal(t, 1, len(tLog.Requests), "one request")
					request := tLog.Requests[0]
					assert.Equal(t, hc.expectParentTrace, request.Trace.TraceParent.TraceID().String(), "parent traceID")
					assert.Equal(t, hc.expectParentSpan, request.Trace.TraceParent.SpanID().String(), "parent spanID")
					assert.Equal(t, "00", request.Trace.TraceParent.Version().String(), "parent version")
					assert.Equal(t, hc.expectParentFlags, request.Trace.TraceParent.Flags().String(), "parent flags")

					if hc.expectTrace == "" {
						hc.expectTrace = hc.expectParentTrace
					}

					if hc.expectTrace == "random" {
						assert.False(t, request.Trace.Trace.TraceID().IsZero(), "trace traceID zero")
					} else {
						assert.Equal(t, hc.expectTrace, request.Trace.Trace.TraceID().String(), "trace traceID")
					}
					if hc.expectSpan != "" {
						assert.Equal(t, hc.expectSpan, request.Trace.Trace.SpanID().String(), "trace spanID")
					} else {
						assert.False(t, request.Trace.Trace.SpanID().IsZero(), "trace spanID is zero")
						assert.NotEqual(t, hc.expectParentSpan, request.Trace.Trace.SpanID().String(), "trace spanID")
					}
					assert.Equal(t, "00", request.Trace.Trace.Version().String(), "trace version")
					assert.Equal(t, hc.expectFlags, request.Trace.Trace.Flags().String(), "trace flags")

					assert.Equal(t, request.Trace.Trace.String(), w.Header().Get("traceresponse"), "trace response header")
				})
			}
		})
	}
}
