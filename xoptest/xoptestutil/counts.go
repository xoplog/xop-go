package xoptestutil

import (
	"github.com/xoplog/xop-go/xoprecorder"
)

// XXX move to xoprecorder
func EventCount(tlog *xoprecorder.Logger, typ xoprecorder.EventType) int {
	var got int
	_ = tlog.WithLock(func(tlog *xoprecorder.Logger) error {
		for _, event := range tlog.Events {
			if event.Type == typ {
				got++
			}
		}
		return nil
	})
	return got
}
