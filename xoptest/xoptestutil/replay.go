package xoptestutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go/xoptest"
)

func VerifyReplay(t *testing.T, origLog *xoptest.TestLogger, replayLog *xoptest.TestLogger) {
	require.Equal(t, len(origLog.Requests), len(replayLog.Requests), "count of requests")
	require.Equal(t, len(origLog.Spans), len(replayLog.Spans), "count of spans")
	require.Equal(t, len(origLog.Lines), len(replayLog.Lines), "count of lines")
}
