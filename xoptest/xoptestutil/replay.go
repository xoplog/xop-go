package xoptestutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptrace"
)

func VerifyReplay(t *testing.T, want *xoptest.TestLogger, got *xoptest.TestLogger) {
	verifyReplaySpans(t, "request", want.Requests, got.Requests)
	verifyReplaySpans(t, "spans", want.Spans, got.Spans)
	// XXX improve lines
	require.Equal(t, len(want.Lines), len(got.Lines), "count of lines")
}

func verifyReplaySpans(t *testing.T, kind string, want []*xoptest.Span, got []*xoptest.Span) {
	require.Equalf(t, len(want), len(got), "count of %s", kind)
	for i := range want {
		verifyReplaySpan(t, want[i], got[i])
	}
}

func verifyReplaySpan(t *testing.T, want *xoptest.Span, got *xoptest.Span) {
	t.Logf("validating replay of span %s", want.Bundle.Trace)
	assert.Equal(t, want.IsRequest, got.IsRequest, "is request?")
	assert.Equal(t, want.RequestNum, got.RequestNum, "sequence number")
	if want.Parent != nil {
		if assert.NotNil(t, got.Parent, "parent not nil") {
			assert.Equal(t, want.Parent.Bundle.Trace.String(), got.Parent.Bundle.Trace.String(), "parent id")
		}
	} else {
		assert.Nil(t, got.Parent, "parent nil")
	}
	assert.Equal(t, want.Bundle.Parent.String(), got.Bundle.Parent.String(), "bundle parent")
	assert.Equal(t, want.Bundle.Baggage.String(), got.Bundle.Baggage.String(), "bundle baggage")
	assert.Equal(t, want.Bundle.State.String(), got.Bundle.State.String(), "bundle state")
	assert.Equal(t, want.Short, got.Short, "short span id for test output")
	assert.Truef(t, want.StartTime.Equal(got.StartTime), "start time %s vs %s", want.StartTime.Format(time.RFC3339), got.StartTime.Format(time.RFC3339))
	assert.Equal(t, want.EndTime, got.EndTime, "end time")
	assert.Equal(t, want.SequenceCode, got.SequenceCode, "sequence code")
	assert.Equal(t, want.SourceInfo, got.SourceInfo, "source info")
	for k, typ := range want.MetadataType {
		t.Logf(" validating metadata %s", k)
		if assert.Equal(t, want.MetadataType[k], got.MetadataType[k], "metadata type") {
			if ws, ok := want.Metadata[k].(fmt.Stringer); ok {
				gs := got.Metadata[k].(fmt.Stringer)
				assert.Equal(t, ws, gs, "metadata (as string)")
			}
			switch typ {
			case xopbase.LinkArrayDataType:
				wa := want.Metadata[k].([]interface{})
				ga := want.Metadata[k].([]interface{})
				if assert.Equal(t, len(wa), len(ga), "equal number of traces in link array") {
					for i := range wa {
						w := wa[i].(xoptrace.Trace)
						g := ga[i].(xoptrace.Trace)
						assert.Equalf(t, w.String(), g.String(), "link[%d]", i)
					}
				}
			case xopbase.LinkDataType:
				w := want.Metadata[k].(xoptrace.Trace)
				g := want.Metadata[k].(xoptrace.Trace)
				assert.Equal(t, w.String(), g.String(), "metadata")
			default:
				assert.Equalf(t, want.Metadata[k], got.Metadata[k], "metadata %s", typ)
			}
		}
	}
	for k := range got.MetadataType {
		_, ok := want.MetadataType[k]
		assert.Truef(t, ok, "extraneous metadata key '%s'", k)
	}
}
