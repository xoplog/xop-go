//go:build !xoptesting

package xop

// smallSleepForTesting is for debugging xop only
func smallSleepForTesting() {}

// debugPrint is for debugging xop only
func debugPrint(...interface{}) {}

// Stack is for debugging xop only
func stack() string { return "" }
