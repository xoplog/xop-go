//go:build xoptesting

package xop

import "time"

func SmallSleepForTesting() {
	time.Sleep(10 * time.Millisecond)
}
