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
	a := trace.SpanIDString()
	trace.RandomizeSpanID()
	b := trace.SpanIDString()
	assert.NotEqual(t, a, b)
}
