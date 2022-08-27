//go:build xoptesting

package xop

import (
	"fmt"
	"time"
)

func SmallSleepForTesting() {
	time.Sleep(10 * time.Millisecond)
}

func DebugPrint(v ...interface{}) {
	fmt.Println(v...)
}
