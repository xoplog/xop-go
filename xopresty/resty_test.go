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
)

func TestXopResty(t *testing.T) {
	tLog := xoptest.New(t)
	seed := xop.Seed(xop.WithBase(tLog))
	log := seed.Request("client")

	client := resty.NewClient()
	client = xopresty.Client(client)
	ctx := log.IntoContext(context.Background())

	inbound := xopmiddle.New(seed, func(r *http.Request) string {
		return r.Method
	})

	ts := httptest.NewServer(inbound.HandlerMiddlewareFunc()(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "generall broken", 500)
	}))
	defer ts.Close()
}
