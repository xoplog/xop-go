package xopotel

import (
	"context"
	"sync"

	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopnum"

	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type logger struct {
	tracer    oteltrace.Tracer
	id        string
	doLogging bool
}

type span struct {
	span              oteltrace.Span
	logger            *logger
	ctx               context.Context
	lock              sync.Mutex
	priorBoolSlices   map[string][]bool
	priorFloat64Slics map[string][]float64
	priorStringSlices map[string][]string
	priorInt64Slices  map[string][]int64
	priorBool         map[string]struct{}
	priorFloat        map[string]struct{}
	priorString       map[string]struct{}
	priorInt          map[string]struct{}
	metadataSeen      map[string]interface{}
	spanPrefill       []attribute.KeyValue // holds spanID & traceID
}

type prefilling struct {
	builder
}

type prefilled struct {
	builder
}

type line struct {
	builder
	prealloc [15]attribute.KeyValue
	level    xopnum.Level
}

type builder struct {
	attributes []attribute.KeyValue
	span       *span
}

var _ xopbase.Logger = &logger{}
var _ xopbase.Request = &span{}
var _ xopbase.Span = &span{}
var _ xopbase.Line = &line{}
var _ xopbase.Prefilling = &prefilling{}
var _ xopbase.Prefilled = &prefilled{}

// This block copied from https://github.com/uptrace/opentelemetry-go-extra/blob/main/otelzap/otelzap.go
var (
	logSeverityKey = attribute.Key("log.severity")
	logMessageKey  = attribute.Key("log.message")
	logTemplateKey = attribute.Key("log.template")
)
