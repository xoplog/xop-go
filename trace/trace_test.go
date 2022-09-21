package trace_test

import (
	"testing"

	"github.com/muir/xop-go/trace"

	"github.com/stretchr/testify/assert"
)

func TestSetRandom(t *testing.T) {
	trace := trace.NewTrace()
	trace.RebuildSetNonZero()
	trace.SpanID().SetRandom()
	a := trace.SpanID().String()
	trace.SpanID().SetRandom()
	b := trace.SpanID().String()
	assert.NotEqual(t, a, b)
}

func TestTraceZero(t *testing.T) {
	var trace trace.Trace
	assert.True(t, trace.IsZero())
	assert.Equal(t, "00", trace.GetFlags().String())
	assert.Equal(t, "0000000000000000", trace.GetSpanID().String())
	assert.Equal(t, "00-00000000000000000000000000000000-0000000000000000-00", trace.String())
	assert.Equal(t, "00000000000000000000000000000000", trace.GetTraceID().String())
	assert.Equal(t, "00", trace.GetVersion().String())
}

func TestTracePartial(t *testing.T) {
	var trace trace.Trace
	trace.SpanID().SetRandom()
	assert.NotEqual(t, "0000000000000000", trace.SpanID().String())
	assert.Equal(t, "00-00000000000000000000000000000000-"+trace.SpanID().String()+"-00", trace.String())
	assert.Equal(t, trace.SpanID().String(), trace.GetSpanID().String())
	assert.Len(t, trace.String(), 55)
}

func TestTraceRandom(t *testing.T) {
	var trace trace.Trace
	trace.Version().SetRandom()
	trace.SpanID().SetRandom()
	trace.TraceID().SetRandom()
	trace.Flags().SetRandom()
	assert.NotEqual(t, "00", trace.Version().String())
	assert.NotEqual(t, "00000000000000000000000000000000", trace.TraceID().String())
	assert.NotEqual(t, "0000000000000000", trace.SpanID().String())
	assert.NotEqual(t, "00", trace.Flags().String())
	assert.Equal(t,
		trace.Version().String()+"-"+
			trace.TraceID().String()+"-"+
			trace.SpanID().String()+"-"+
			trace.Flags().String(), trace.String())
	assert.Equal(t, trace.Version().String(), trace.GetVersion().String())
	assert.Equal(t, trace.SpanID().String(), trace.GetSpanID().String())
	assert.Equal(t, trace.TraceID().String(), trace.GetTraceID().String())
	assert.Equal(t, trace.Flags().String(), trace.GetFlags().String())
	assert.Len(t, trace.String(), 55)
}

func TestTraceCopy(t *testing.T) {
	var t1 trace.Trace
	t1.Version().SetRandom()
	t1.SpanID().SetRandom()
	t1.TraceID().SetRandom()
	t1.Flags().SetRandom()
	t2 := t1
	assert.Equal(t, t1, t2)
	t2.TraceID().SetRandom()
	assert.NotEqual(t, t1, t2)
}
