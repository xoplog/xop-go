package xm

type Level int

const (
	// Open Telemetry puts tracing as lower level than debugging.  Why?
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	// Aside from that, we'll map to their numbers
	DebugLevel  Level = 5
	TraceLevel        = 8 // OTEL "Debug4"
	InfoLevel         = 9
	WarnLevel         = 13
	ErrorLevel        = 17
	AlertLevel        = 20 // OTEL "Error4"
	MetricLevel       = 14 // OTEL "Warn2"
)
