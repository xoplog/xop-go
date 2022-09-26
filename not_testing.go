//go:build !xoptesting

package xop

// SmallSleepForTesting is for debugging xop only
func SmallSleepForTesting() {}

// DebugPrint is for debugging xop only
func DebugPrint(...interface{}) {}

// Stack is for debugging xop only
func Stack() string { return "" }
