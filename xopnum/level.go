// xopnum provides constants used across the xop ecosystem
package xopnum

//go:generate enumer -type=Level -linecomment -json -sql

type Level int32

const (
	// Open Telemetry puts tracing as lower level than debugging.  Why?
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	// Most of the levels correspond to OTEl's levels, but
	//   TraceLevel is OTEL's "Trace2"
	//   DebugLevel is OTEL's "Debug3"
	//   LogLevel is OTEL's "Info"
	//   InfoLevel is OTEL's "Info3"
	//   WarnLevel is OTEL's "Warn"
	//   ErrorLevel is OTEL's "Error"
	//   AlertLevel is OTEL's "Error4"
	// At this time, there is no fatal level in XOP.
	//
	// LogLevel is the expected default level to use for most logs.
	TraceLevel Level = 2  // trace
	DebugLevel Level = 7  // debug
	LogLevel   Level = 9  // log
	InfoLevel  Level = 11 // info
	WarnLevel  Level = 13 // warn
	ErrorLevel Level = 17 // error
	AlertLevel Level = 20 // alert
)

const MaxLevel = AlertLevel
