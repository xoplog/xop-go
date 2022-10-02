package xopresty_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopmiddle"
	"github.com/muir/xop-go/xopresty"
	"github.com/muir/xop-go/xoptest"

	"github.com/muir/resty"
)

func TestXopResty(t *testing.T) {
	tLog := xoptest.New(t)
	seed := xop.NewSeed(xop.WithBase(tLog))
	log := seed.Request("client")
	log.Info().Msg("i am the client")
	ctx := log.IntoContext(context.Background())

	inbound := xopmiddle.New(seed, func(r *http.Request) string {
		return r.Method
	})
	ts := httptest.NewServer(inbound.HandlerFuncMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		log := xop.FromContextOrPanic(r.Context())
		log.Info().Msg("in request handler")
		http.Error(w, "generall broken", 500)
	}))
	defer ts.Close()

	xopresty.Client(resty.New()).SetDebug(true).R().SetContext(ctx).EnableTrace().Get(ts.URL)
}
