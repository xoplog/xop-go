package xoptestutil

import "time"

// Nap makes sure that the time advances by at least a microsecond.
// We're guessing that lack of clock resolution is the cause of test
// flakiness on Windows.
func MicroNap() {
	start := time.Now()
	for {
		if time.Since(start) >= time.Microsecond {
			return
		}
		time.Sleep(time.Microsecond)
	}
}
