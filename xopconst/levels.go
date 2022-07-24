package xopconst

import (
	"sync/atomic"
)

//go:generate enumer -type=Level -linecomment -json -sql

const (
	// Open Telemetry puts tracing as lower level than debugging.  Why?
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	// Aside from that, we'll mostly map to their numbers.
	// TraceLevel is OTEL's "Debug4"
	// AlertLevel is OTEL's "Error4"
	DebugLevel Level = 5  // debug
	TraceLevel Level = 8  // trace
	InfoLevel  Level = 9  // info
	WarnLevel  Level = 13 // warn
	ErrorLevel Level = 17 // error
	AlertLevel Level = 20 // alert
)

type Level int32

func (level *Level) AtomicLoad() Level {
	return Level(atomic.LoadInt32((*int32)(level)))
}

func (level *Level) AtomicStore(newLevel Level) {
	atomic.StoreInt32((*int32)(level), int32(newLevel))
}
