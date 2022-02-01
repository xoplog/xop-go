package xm

import (
	"sync/atomic"
)

const (
	// Open Telemetry puts tracing as lower level than debugging.  Why?
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	// Aside from that, we'll map to their numbers
	DebugLevel Level = 5
	TraceLevel       = 8 // OTEL "Debug4"
	InfoLevel        = 9
	WarnLevel        = 13
	ErrorLevel       = 17
	AlertLevel       = 20 // OTEL "Error4"
)

type Level int32

func (level *Level) AtomicLoad() Level {
	return Level(atomic.LoadInt32((*int32)(level)))
}

func (level *Level) AtomicStore(newLevel Level) {
	atomic.StoreInt32((*int32)(level), int32(newLevel))
}
