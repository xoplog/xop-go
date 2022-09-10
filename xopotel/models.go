package xopotel

import (
	"context"
	"sync"

	"github.com/muir/xop-go/xopbase"

	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Logger struct {
	tracer   oteltrace.Tracer
	id       string
}

type Span struct {
	span              oteltrace.Span
	logger            *Logger
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
	spanPrefill	[]attribute.KeyValue // holds spanID & traceID 
}

type Prefilling struct {
	Builder
}

type Prefilled struct {
	Builder
}

type Line struct {
	prealloc [15]attribute.KeyValue
	Builder
}

type Builder struct {
	prefill []attribute.KeyValue
	span    *Span
}

var _ xopbase.Logger = &BaseLogger{}
var _ xopbase.Request = &Span{}
var _ xopbase.Span = &Span{}
