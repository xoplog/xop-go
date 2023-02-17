package xoputil

import (
	"sync/atomic"
)

func AtomicMaxInt64(target *int64, value int64) {
	for {
		old := atomic.LoadInt64(target)
		if old >= value {
			return
		}
		if atomic.CompareAndSwapInt64(target, old, value) {
			return
		}
	}
}
