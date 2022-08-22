package trace_test

import (
	"testing"

	"github.com/muir/xop-go/trace"

	"github.com/stretchr/testify/assert"
)

func TestSetRandom(t *testing.T) {
	trace := trace.NewTrace()
	trace.RebuildSetNonZero()
	trace.RandomizeSpanID()
	a := trace.SpanID().String()
	trace.RandomizeSpanID()
	b := trace.SpanID().String()
	assert.NotEqual(t, a, b)
}
