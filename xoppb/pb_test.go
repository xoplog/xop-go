package xoppb_test

import (
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xoppb"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"
)

type testWriter struct {
	captured []writtenRequest
}

type writtenRequest struct {
	traceID xoptrace.HexBytes16
	request *xopproto.Request
}

func (tw *testWriter) SizeLimit() int32 { return 1<<25 }
func (tw *testWriter) Flush() error     { return nil }
func (tw *testWriter) Request(traceID xoptrace.HexBytes16, request *xopproto.Request) {
	tw.captured = append(tw.captured, writtenRequest{
		traceID: traceID,
		request: request,
	})
}

func TestReplay(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		t.Run(mc.Name, func(t *testing.T) {
			tWriter := &testWriter{}
			tLog := xoptest.New(t)
			pbLog := xoppb.New(tWriter)
			seed := xop.NewSeed(
				xop.WithBase(tLog),
				xop.WithBase(pbLog),
			)
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				seed = seed.Copy(mc.SeedMods...)
			}
			log := seed.Request(t.Name())
			t.Log("generate logs")
			mc.Do(t, log, tLog)

			/*
				rLog := xoptest.New(t)
				t.Log("replay from generated logs")
				err := tLog.LosslessReplay(context.Background(), tLog, rLog)
				require.NoError(t, err, "replay")
				t.Log("verify replay equals original")
				xoptestutil.VerifyReplay(t, tLog, rLog)
			*/
		})
	}
}
