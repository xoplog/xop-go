package xoptestutil

import (
	"github.com/xoplog/xop-go/xoptest"
)

func EventCount(tlog *xoptest.TestLogger, typ xoptest.EventType) int {
	var got int
	_ = tlog.WithLock(func(tlog *xoptest.TestLogger) error {
		for _, event := range tlog.Events {
			if event.Type == typ {
				got++
			}
		}
		return nil
	})
	return got
}
