//go:build xoptesting

package xop

import (
	"fmt"
	"runtime/debug"
	"time"
)

func SmallSleepForTesting() {
	time.Sleep(10 * time.Millisecond)
}

func DebugPrint(v ...interface{}) {
	fmt.Println(v...)
}

func Stack() string {
	return string(debug.Stack())
}
