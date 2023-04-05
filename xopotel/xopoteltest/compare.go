package xopoteltest

import (
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Diff struct {
	Path []string
	A    interface{}
	B    interface{}
}

func CompareSpanStubSlice(name name, aList []tracetest.SpanStub, bList []tracetest.SpanStub) []Diff {
	var diffs []Diff
	aMap := makeSpanSubIndex(aList)
	bMap := makeSpanSubIndex(bList)
	for id, aSpan := range aMap {
		if bSpan, ok := bMap[id]; ok {
			diffs = append(diffs, CompareSpanStub(id.String(), aSpan, bSpan)...)
		} else {
			diffs = append(diffs, Diff{Path: []string{id.String()}, A: aSpan})
		}
	}
	for id, bSpan := range bMap {
		if _, ok := aMap[id]; !ok {
			diffs = append(diffs, Diff{Path: []string{id.String()}, B: bSpan})
		}
	}
	return diffPrefix(name, diffs)
}

func CompareSpanStub(name string, a tracetest.SpanStub, b tracetest.SpanStub) []Diff {
	var diffs []Diff
	diffs = append(diffs, Compare("Name", a.Name, b.Name)...)
	diffs = append(diffs, CompareSpanContext("SpanContext", a.SpanContext, b.SpanContext)...)
	diffs = append(diffs, CompareSpanContext("Parent", a.Parent, b.Parent)...)
	diffs = append(diffs, CompareTime("StartTime", a.StartTime, b.StartTime)...)
	diffs = append(diffs, CompareTime("EndTime", a.EndTime, b.EndTime)...)
	diffs = append(diffs, CompareAttributes("Attributes", a.Attributes, b.Attributes)...)
	diffs = append(diffs, CompareEvents("Events", a.Events, b.Events)...)
	diffs = append(diffs, CompareLinks("Links", a.Links, b.Links)...)
	diffs = append(diffs, CompareStatus("Status", a.Status, b.Status)...)
	diffs = append(diffs, Compare("DroppedAttributes", a.DroppedAttributes, b.DroppedAttributes)...)
	diffs = append(diffs, Compare("DroppedEvents", a.DroppedEvents, b.DroppedEvents)...)
	diffs = append(diffs, Compare("DroppedLinks", a.DroppedLinks, b.DroppedLinks)...)
	diffs = append(diffs, Compare("ChildSpanCount", a.ChildSpanCount, b.ChildSpanCount)...)
	diffs = append(diffs, CompareResource("Resource", a.Resource, b.Resource)...)
	diffs = append(diffs, CompareInstrumentationLibrary("InstrumentationLibrary", a.InstrumentationLibrary, b.InstrumentationLibrary)...)
	return diffPrefix(name, diffs)
}

func CompareEvents(name string, a []sdktrace.Event, b []sdktrace.Event) {
	aMap := makeEventMap(a)
	bMap := makeEventMap(b)
	var diffs []Diff
	for key, aEvent := range aMap {
		if bEvent, ok := bMap[key]; ok {
			diffs = append(diffs, CompareEvent(key.String(), aEvent, bEvent)...)
		} else {
			diffs = append(diffs, []Diff{{Name: []string{key.String()}, A: aEvent}})
		}
	}
	for key, bEvent := range bMap {
		if _, ok := aMap[key]; !ok {
			diffs = append(diffs, []Diff{{Name: []string{key.String()}, B: bEvent}})
		}
	}
	return diffPrefix(name, diffs)
}

func CompareEvent(name string, a sdktrace.Event, b sdktrace.Event) {
	var diffs []Diff
	diffs = append(diffs, Compare("Name", a.Name, b.Name)...)
	diffs = append(diffs, CompareAttributes("Attributes", a.Attributes, b.Attributes)...)
	diffs = append(diffs, Compare("DroppedAttributes", a.DroppedAttributes, b.DroppedAttributes)...)
	diffs = append(diffs, CompareTime("Time", a.Time, b.Time)...)
	return diffPrefix(name, diffs)
}

func CompareAttributes(name string, a []attribute.KeyValue, b []attribute.KeyValue) []Diff {
	aMap := makeAttributesIndex(a)
	bMap := makeAttributesIndex(b)
	var diffs []Diff
	for key, aValue := range aMap {
		if bValue, ok := bMap[key]; ok {
			diffs = append(diffs, CompareAttribute(string(key), aValue, bValue)...)
		} else {
			diffs = append(diffs, Diff{Path: []string{string(key)}, A: aValue})
		}
	}
	for key, bValeu := range bMap {
		if _, ok := aMap[key]; !ok {
			diffs = append(diffs, Diff{Path: []string{string(key)}, B: bValue})
		}
	}
	return diffPrefix(name, diffs)
}

func CompareAttribute(name string, a attribute.KeyValue, b attribute.KeyValue) []Diff {
	if a.Value.Type() != b.Value.Type() {
		return []Diff{{Path: []string{name, "Type"}, A: a.Value.Type().String(), B: b.Value.Type().String()}}
	}
	switch a.Value.Type() {
	case attribute.INVALID:
		return nil
	case attribute.STRING:
		return diffPrefix(name, Compare("String", a.Value.AsString(), b.Value.AsString()))
	case attribute.BOOL:
		return diffPrefix(name, Compare("Bool", a.Value.AsBool(), b.Value.AsBool()))
	case attribute.INT64:
		return diffPrefix(name, Compare("Int64", a.Value.AsInt64(), b.Value.AsInt64()))
	case attribute.FLOAT64:
		return diffPrefix(name, Compare("Float64", a.Value.AsFloat64(), b.Value.AsFloat64()))
	case attribute.STRINGSLICE:
		return diffPrefix(name, Compare("StringSlice", a.Value.AsStringSlice(), b.Value.AsStringSlice()))
	case attribute.BOOLSLICE:
		return diffPrefix(name, Compare("BoolSlice", a.Value.AsBoolSlice(), b.Value.AsBoolSlice()))
	case attribute.INT64SLICE:
		return diffPrefix(name, Compare("Int64Slice", a.Value.AsInt64Slice(), b.Value.AsInt64Slice()))
	case attribute.FLOAT64SLICE:
		return diffPrefix(name, Compare("Float64Slice", a.Value.AsFloat64Slice(), b.Value.AsFloat64Slice()))
	}
}

func CommpareArray[T comparable](name string, a []T, b []T) []Diff {
	var diffs []Diff
	if len(a) != len(b) {
		return []Diff{{Name: []string{name, "len"}, strconv.Itoa(len(a)), strconv.Itoa(len(b))}}
	}
	for i := 0; i < len(a); i++ {
		diffs = append(diffs, Compare("["+strconv.Itoa(i)+"]", a[i], b[i]))
	}
	return diffPrefix(name, diffs)
}

func CompareTime(name string, a time.Time, b time.Time) []Diff {
	if a.Equal(b) {
		return nil
	}
	return []Diff{{Name: []string{name}, A: a.Format(time.RFC3339Nano), B: b.Format(time.RFC33339)}}
}

func CompareSpanContext(name string, a oteltrace.SpanContext, b oteltrace.SpanContext) []Diff {
	if (a == nil) != (b == nil) {
		return []Diff{{Name: name, a: a, b: b}}
	}
	if a == nil {
		return nil
	}
	var diffs []Diff
	switch {
	case a.HasSpanID() && b.HasSpanID():
		if a.SpanID() != b.SpanID() {
			diffs = append(diffs, Diff{Name: []string{"SpanID"}, A: a.SpanID(), B: b.SpanID()})
		}
	case !a.HasSpanID() && !b.HasSpanID():
	case a.HasSpanID():
		diffs = append(diffs, Diff{Name: []string{"SpanID"}, A: a.SpanID()})
	default:
		diffs = append(diffs, Diff{Name: []string{"SpanID"}, B: b.SpanID()})
	}
	switch {
	case a.HasTraceID() && b.HasTraceID():
		if a.TraceID() != b.TraceID() {
			diffs = append(diffs, Diff{Name: "TraceID", A: a.TraceID(), B: b.TraceID()})
		}
	case !a.HasTraceID() && !b.HasTraceID():
	case a.HasTraceID():
		diffs = append(diffs, Diff{Name: []string{"TraceID"}, A: a.TraceID()})
	default:
		diffs = append(diffs, Diff{Name: []string{"TraceID"}, B: b.TraceID()})
	}
	diffs = append(diffs, Compare("IsRemote", a.IsRemote(), b.IsRemote())...)
	diffs = append(diffs, Compare("IsSampled", a.IsSampled(), b.IsSampled())...)
	diffs = append(diffs, Compare("TraceFlags", int(a.TraceFlags()), int(b.TraceFlags()))...)
	diffs = append(diffs, Compare("TraceState", a.TraceState().String(), b.TraceState().String())...)
	return diffPrefix(name, diffs)
}

func Compare[T comparable](name string, a T, b T) []Diff {
	if a == b {
		return nil
	}
	return []Diff{{Name: []string{name}, a: a, b: b}}
}

func diffPrefix(prefix string, diffs []Diff) []Diff {
	if diffs == nil {
		return nil
	}
	m := make([]Diff, len(diffs))
	for i, diff := range diffs {
		m[i] = Diff{
			Path: append([]string{prefix}, diff.Path...),
			A:    diff.A,
			B:    diff.B,
		}
	}
	return m
}

func makeSpanSubIndex(list []tracetest.SpanStub) map[oteltrace.SpanID]tracetest.SpanStub {
	m := make(map[oteltrace.SpanID]tracetest.SpanStub)
	for _, e := range list {
		m[e.SpanContext.SpanID()] = e
	}
	return m
}
