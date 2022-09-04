//go:build !xoptesting

package xop

func SmallSleepForTesting() {}

func DebugPrint(...interface{}) {}

func Stack() string { return "" }
