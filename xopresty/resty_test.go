package xopresty_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopmiddle"
	"github.com/muir/xop-go/xopnum"
	"github.com/muir/xop-go/xopresty"
	"github.com/muir/xop-go/xoptest"

	"github.com/muir/resty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exampleRequest struct {
	Name  string
	Count int
}

type exampleResult struct {
	Score   float64
	Comment string
}

var cases = []struct {
	name         string
	clientMod    func(*resty.Client) *resty.Client
	requestMod   func(*resty.Request) *resty.Request
	handler      func(t *testing.T, log *xop.Log, w http.ResponseWriter, r *http.Request)
	restyOpts    []xopresty.ClientOpt
	expectError  bool
	expectedText []string
}{
	{
		name: "with debugging and tracing",
		clientMod: func(c *resty.Client) *resty.Client {
			return c.SetDebug(true)
		},
		requestMod: func(r *resty.Request) *resty.Request {
			return r.EnableTrace()
		},
	},
	{
		name: "without debugging, without tracing",
	},
	{
		name: "with model",
		requestMod: func(r *resty.Request) *resty.Request {
			var res exampleResult
			return r.
				SetBody(exampleRequest{
					Name:  "Joe",
					Count: 38,
				}).SetResult(&res).
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json")
		},
		handler: func(t *testing.T, log *xop.Log, w http.ResponseWriter, r *http.Request) {
			enc, err := json.Marshal(exampleResult{
				Score:   3.8,
				Comment: "good progress",
			})
			assert.NoError(t, err, "marshal")
			w.Header().Set("Content-Type", "application/json")
			w.Write(enc)
			log.Trace().Msg("sent response")
		},
		expectedText: []string{
			"T1.1.1: request body={Name:Joe Count:38}",
			"T1.1.1: received response=&{Score:3.8 Comment:good progress}",
		},
	},
}

func TestXopResty(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tLog := xoptest.New(t)
			seed := xop.NewSeed(xop.WithBase(tLog))
			log := seed.Request("client")
			log.Info().Msg("i am the base log")
			ctx := log.Sub().MinLevel(xopnum.DebugLevel).Log().IntoContext(context.Background())

			var called bool
			inbound := xopmiddle.New(seed, func(r *http.Request) string {
				return "handler:" + r.Method
			})
			ts := httptest.NewServer(inbound.HandlerFuncMiddleware()(func(w http.ResponseWriter, r *http.Request) {
				called = true
				log := xop.FromContextOrPanic(r.Context())
				log.Info().Msg("in request handler")
				if tc.handler == nil {
					http.Error(w, "no handler", 500)
					return
				}
				tc.handler(t, log, w, r)
			}))
			defer ts.Close()

			log.Done()
			c := xopresty.Client(resty.New())
			if tc.clientMod != nil {
				c = tc.clientMod(c)
			}
			r := c.R().SetContext(ctx)
			if tc.requestMod != nil {
				r = tc.requestMod(r)
			}

			_, err := r.Get(ts.URL)

			requestSpan := tLog.FindSpan(xoptest.ShortEquals("T1.1.1"))

			require.NotNil(t, requestSpan, "requestSpan")
			assert.NotEmpty(t, requestSpan.EndTime, "client request span completed")

			if tc.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			farSideSpan := tLog.FindSpan(xoptest.ShortEquals("T1.2"))
			require.NotNil(t, farSideSpan, "farSideSpan")
			assert.NotEmpty(t, farSideSpan.EndTime, "server endpoint span completed")
			assert.NoError(t, err, "Get")
			assert.True(t, called, "handler called")

			text := "T1.1.1: traceresponse http.remote_trace=" + farSideSpan.Trace.Trace.String()
			assert.Equalf(t, 1, tLog.CountLines(xoptest.TextContains(text)), "count lines with '%s'", text)

			for _, text := range tc.expectedText {
				assert.Equalf(t, 1, tLog.CountLines(xoptest.TextContains(text)), "count lines with '%s'", text)
			}
		})
	}
}
