package xoppb_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xoppb"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"
)

type testWriter struct {
	captured []*xopproto.Trace
}

func (tw *testWriter) SizeLimit() int32 { return 1 << 25 }
func (tw *testWriter) Flush() error     { return nil }
func (tw *testWriter) Request(traceID xoptrace.HexBytes16, request *xopproto.Request) {
	if len(tw.captured) != 0 && bytes.Equal(tw.captured[len(tw.captured)-1].TraceID, traceID.Bytes()) {
		tw.captured[len(tw.captured)-1].Requests = append(tw.captured[len(tw.captured)-1].Requests, request)
	} else {
		tw.captured = append(tw.captured, &xopproto.Trace{
			TraceID:  traceID.Bytes(),
			Requests: []*xopproto.Request{request},
		})
	}
}

func TestReplayPB(t *testing.T) {
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

			t.Log("replay from generated logs")
			rLog := xoptest.New(t)
			for _, trace := range tWriter.captured {
				err := pbLog.LosslessReplay(context.Background(), trace, rLog)
				require.NoError(t, err, "replay")
			}
			t.Log("verify replay equals original")
			xoptestutil.VerifyTestReplay(t, tLog, rLog)
		})
	}
}
