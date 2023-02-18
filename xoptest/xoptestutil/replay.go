package xoptestutil

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/muir/list"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func VerifyReplay(t *testing.T, want *xoptest.TestLogger, got *xoptest.TestLogger) {
	verifyReplaySpans(t, "request", want.Requests, got.Requests)
	verifyReplaySpans(t, "spans", want.Spans, got.Spans)
	verifyReplayLines(t, want.Lines, got.Lines)
}

func verifyReplayLines(t *testing.T, want []*xoptest.Line, got []*xoptest.Line) {
	require.Equal(t, len(want), len(got), "count of lines")
	for i := range want {
		verifyReplayLine(t, want[i], got[i])
	}
}

func verifyReplayLine(t *testing.T, want *xoptest.Line, got *xoptest.Line) {
	t.Log("verify line", want.Text)
	assert.Equal(t, want.Level, got.Level, "level")
	assert.Truef(t, want.Timestamp.Equal(got.Timestamp), "timestamp %s vs %s", want.Timestamp.Format(time.RFC3339), got.Timestamp.Format(time.RFC3339))
	assert.Equal(t, want.Message, got.Message, "message")
	assert.Equal(t, want.Tmpl, got.Tmpl, "template")
	if want.AsLink != nil && assert.NotNil(t, got.AsLink, "link") {
		assert.Equal(t, want.AsLink.String(), got.AsLink.String(), "link")
	}
	if want.AsModel != nil && assert.NotNil(t, got.AsModel, "model") {
		want.AsModel.Encode()
		got.AsModel.Encode()
		assert.Equal(t, want.AsModel.Encoding, got.AsModel.Encoding, "encoding")
		assert.Equal(t, want.AsModel.TypeName, got.AsModel.TypeName, "encoding")
		assert.Equal(t, want.AsModel.Encoded, got.AsModel.Encoded, "encoded")
	}
	assert.Equal(t, want.Tmpl, got.Tmpl, "template")
	for key, wdata := range want.Data {
		gdata, ok := got.Data[key]
		if !assert.True(t, ok, "data element '%s' in want, but not got", key) {
			continue
		}
		dt := want.DataType[key]
		if !assert.Equalf(t, dt.String(), got.DataType[key].String(), "data type for key '%s'", key) {
			continue
		}
		switch dt {
		case xopbase.AnyDataType:
			wany := wdata.(xopbase.ModelArg)
			gany := wdata.(xopbase.ModelArg)
			wany.Encode()
			gany.Encode()
			assert.Equalf(t, wany.Encoding, gany.Encoding, "encoding %s", key)
			assert.Equalf(t, wany.TypeName, gany.TypeName, "encoding %s", key)
			assert.Equalf(t, wany.Encoded, gany.Encoded, "encoded %s", key)
		case xopbase.EnumDataType:
			wenum := wdata.(xopat.Enum)
			genum := gdata.(xopat.Enum)
			assert.Equalf(t, wenum.String(), genum.String(), "enum %s", key)
			assert.Equalf(t, wenum.Int64(), genum.Int64(), "enum %s", key)
		default:
			assert.Equal(t, wdata, gdata, "data")
		}
	}

	for key := range got.Data {
		_, ok := want.Data[key]
		assert.Truef(t, ok, "data element '%s' in got, but not want", key)
	}
}

func verifyReplaySpans(t *testing.T, kind string, want []*xoptest.Span, got []*xoptest.Span) {
	if !assert.Equalf(t, len(want), len(got), "count of %s", kind) {
		return
	}
	want = sortSpans(want)
	got = sortSpans(got)
	for i := range want {
		verifyReplaySpan(t, want[i], got[i])
	}
}

func sortSpans(spans []*xoptest.Span) []*xoptest.Span {
	spans = list.Copy(spans)
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].Bundle.Trace.GetSpanID().String() < spans[j].Bundle.Trace.GetSpanID().String()
	})
	return spans
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
		if _, ok := got.MetadataType[k]; !assert.Truef(t, ok, "missing metadata %s", k) {
			continue
		}
		if assert.Equal(t, want.MetadataType[k].String(), got.MetadataType[k].String(), "metadata type") {
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
