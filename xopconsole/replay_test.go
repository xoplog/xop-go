package xopconsole_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopconsole"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"

	"github.com/stretchr/testify/require"
)

type buffer struct {
	t   *testing.T
	buf []byte
}

func (buf *buffer) Write(b []byte) (int, error) {
	buf.t.Log(string(b))
	buf.buf = append(buf.buf, b...)
	return len(b), nil
}

func TestReplayConsole(t *testing.T) {
	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		t.Run(mc.Name, func(t *testing.T) {
			tLog := xoptest.New(t)
			buf := &buffer{t: t}
			cLog := xopconsole.New(xopconsole.WithWriter(buf))
			seed := xop.NewSeed(
				xop.WithBase(cLog),
				xop.WithBase(tLog),
			)
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				seed = seed.Copy(mc.SeedMods...)
			}
			log := seed.Request(t.Name())
			mc.Do(t, log, tLog)
			t.Log("Replay")
			rLog := xoptest.New(t)
			err := xopconsole.Replay(context.Background(), bytes.NewBuffer(buf.buf), rLog)
			require.NoError(t, err, "replay")

			t.Log("verify replay equals original")
			xoptestutil.VerifyTestReplay(t, tLog, rLog)
		})
	}
}
