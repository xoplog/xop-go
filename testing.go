//go:build xoptesting

package xop

import (
	"fmt"
	"runtime/debug"
	"time"
)

func smallSleepForTesting() {
	time.Sleep(10 * time.Millisecond)
}

func debugPrint(v ...interface{}) {
	fmt.Println(v...)
}

func stack() string {
	return string(debug.Stack())
}
