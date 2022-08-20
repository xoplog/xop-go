package xoptestutil

import (
	"testing"

	"github.com/muir/xop-go/xoptest"

	"github.com/stretchr/testify/assert"
)

func ExpectEventCount(t testing.TB, tlog *xoptest.TestLogger, typ xoptest.EventType, want int) {
	_ = tlog.WithLock(func(tlog *xoptest.TestLogger) error {
		var got int
		for _, event := range tlog.Events {
			if event.Type == typ {
				got++
			}
		}
		assert.Equalf(t, want, got, "event count %s", typ)
		return nil
	})
}
