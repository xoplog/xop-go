// xopnum provides constants used across the xop ecosystem
package xopnum

//go:generate enumer -type=Level -linecomment -json -sql

type Level int32

const (
	// Open Telemetry puts tracing as lower level than debugging.  Why?
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	// Most of the levels correspond to OTEl's levels, but
	//   TraceLevel is OTEL's "Trace2"
	//   AlertLevel is OTEL's "Error4"
	// There is no fatal level.
	TraceLevel Level = 2  // trace
	DebugLevel Level = 5  // debug
	InfoLevel  Level = 9  // info
	WarnLevel  Level = 13 // warn
	ErrorLevel Level = 17 // error
	AlertLevel Level = 20 // alert
)

const MaxLevel = AlertLevel
