package xopoteltest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type keyConstraint interface {
	comparable
	fmt.Stringer
}

func makeListCompare[T any, K keyConstraint](mapper func([]T) map[K]T, compare func(string, T, T) []Diff) func(string, []T, []T) []Diff {
	return func(name string, a []T, b []T) []Diff {
		aMap := mapper(a)
		bMap := mapper(b)
		var diffs []Diff
		for key, aThing := range aMap {
			if bThing, ok := bMap[key]; ok {
				diffs = append(diffs, compare(key.String(), aThing, bThing)...)
			} else {
				diffs = append(diffs, Diff{Path: []string{key.String()}, A: aThing})
			}
		}
		for key, bThing := range bMap {
			if _, ok := aMap[key]; !ok {
				diffs = append(diffs, Diff{Path: []string{key.String()}, B: bThing})
			}
		}
		return diffPrefix(name, diffs)
	}
}

type Diff struct {
	Path []string
	A    any
	B    any
}

func (d Diff) String() string {
	return fmt.Sprintf("%s: %s vs %s", strings.Join(d.Path, "."), toString(d.A), toString(d.B))
}

func (d Diff) MatchTail(tail ...string) bool {
	if len(tail) > len(d.Path) {
		return false
	}
	for i, tailPart := range tail {
		if d.Path[len(d.Path)-len(tail)+i] != tailPart {
			return false
		}
	}
	return true
}

var stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func toString(a any) string {
	if a == nil {
		return "missing"
	}
	if stringer, ok := a.(fmt.Stringer); ok {
		return stringer.String()
	}
	v := reflect.ValueOf(a)
	if v.IsValid() &&
		v.Type().Kind() == reflect.Func &&
		v.Type().NumIn() == 0 &&
		v.Type().NumOut() == 1 &&
		v.Type().Out(0).AssignableTo(stringerType) {
		out := v.Call([]reflect.Value{})
		ov := out[0].Interface()
		return ov.(fmt.Stringer).String()
	}
	return reflect.TypeOf(a).String() + "/" + fmt.Sprint(a)
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

var CompareSpanStubSlice = makeListCompare(makeSpanSubIndex, CompareSpanStub)

func makeSpanSubIndex(list []SpanStub) map[oteltrace.SpanID]SpanStub {
	m := make(map[oteltrace.SpanID]SpanStub)
	for _, e := range list {
		m[e.SpanContext.SpanID()] = e
	}
	return m
}

func CompareSpanStub(name string, a SpanStub, b SpanStub) []Diff {
	var diffs []Diff
	diffs = append(diffs, Compare("Name", a.Name, b.Name)...)
	diffs = append(diffs, CompareSpanContext("SpanContext", a.SpanContext, b.SpanContext)...)
	diffs = append(diffs, CompareSpanContext("Parent", a.Parent, b.Parent)...)
	diffs = append(diffs, CompareTime("StartTime", a.StartTime, b.StartTime)...)
	diffs = append(diffs, CompareTime("EndTime", a.EndTime, b.EndTime)...)
	diffs = append(diffs, CompareAttributes("Attributes", a.Attributes, b.Attributes)...)
	diffs = append(diffs, CompareEvents("Events", a.Events, b.Events)...)
	// XXX diffs = append(diffs, CompareLinks("Links", a.Links, b.Links)...)
	// XXX diffs = append(diffs, CompareStatus("Status", a.Status, b.Status)...)
	diffs = append(diffs, Compare("DroppedAttributes", a.DroppedAttributes, b.DroppedAttributes)...)
	diffs = append(diffs, Compare("DroppedEvents", a.DroppedEvents, b.DroppedEvents)...)
	diffs = append(diffs, Compare("DroppedLinks", a.DroppedLinks, b.DroppedLinks)...)
	diffs = append(diffs, Compare("ChildSpanCount", a.ChildSpanCount, b.ChildSpanCount)...)
	// XXX diffs = append(diffs, CompareResource("Resource", a.Resource, b.Resource)...)
	// XXX diffs = append(diffs, CompareInstrumentationLibrary("InstrumentationLibrary", a.InstrumentationLibrary, b.InstrumentationLibrary)...)
	return diffPrefix(name, diffs)
}

// XXX var CompareLinks = makeListCompare(makeLinkMap, CompareLink)

var CompareEvents = makeListCompare(makeEventMap, CompareEvent)

type eventKey struct {
	name string
	ts   int64
}

func (e eventKey) String() string {
	return e.name + "@" + time.Unix(0, e.ts).Format(time.RFC3339Nano)
}

func makeEventMap(events []sdktrace.Event) map[eventKey]sdktrace.Event {
	m := make(map[eventKey]sdktrace.Event)
	for _, event := range events {
		m[eventKey{
			name: event.Name,
			ts:   event.Time.UnixNano(),
		}] = event
	}
	return m
}

func CompareEvent(name string, a sdktrace.Event, b sdktrace.Event) []Diff {
	var diffs []Diff
	diffs = append(diffs, Compare("Name", a.Name, b.Name)...)
	diffs = append(diffs, CompareAttributes("Attributes", a.Attributes, b.Attributes)...)
	diffs = append(diffs, Compare("DroppedAttributeCount", a.DroppedAttributeCount, b.DroppedAttributeCount)...)
	diffs = append(diffs, CompareTime("Time", a.Time, b.Time)...)
	return diffPrefix(name, diffs)
}

var CompareAttributes = makeListCompare(makeAttributesIndex, CompareAttribute)

type stringKey string

func (s stringKey) String() string { return string(s) }

func makeAttributesIndex(list []attribute.KeyValue) map[stringKey]attribute.KeyValue {
	m := make(map[stringKey]attribute.KeyValue)
	for _, kv := range list {
		m[stringKey(kv.Key)] = kv
	}
	return m
}

func CompareAttribute(name string, a attribute.KeyValue, b attribute.KeyValue) []Diff {
	if a.Value.Type() != b.Value.Type() {
		return []Diff{{Path: []string{name, "Type"}, A: a.Value.Type().String(), B: b.Value.Type().String()}}
	}
	switch a.Value.Type() {
	case attribute.STRING:
		return diffPrefix(name, Compare("String", a.Value.AsString(), b.Value.AsString()))
	case attribute.BOOL:
		return diffPrefix(name, Compare("Bool", a.Value.AsBool(), b.Value.AsBool()))
	case attribute.INT64:
		return diffPrefix(name, Compare("Int64", a.Value.AsInt64(), b.Value.AsInt64()))
	case attribute.FLOAT64:
		return diffPrefix(name, Compare("Float64", a.Value.AsFloat64(), b.Value.AsFloat64()))
	case attribute.STRINGSLICE:
		return diffPrefix(name, CompareSlice("StringSlice", a.Value.AsStringSlice(), b.Value.AsStringSlice()))
	case attribute.BOOLSLICE:
		return diffPrefix(name, CompareSlice("BoolSlice", a.Value.AsBoolSlice(), b.Value.AsBoolSlice()))
	case attribute.INT64SLICE:
		return diffPrefix(name, CompareSlice("Int64Slice", a.Value.AsInt64Slice(), b.Value.AsInt64Slice()))
	case attribute.FLOAT64SLICE:
		return diffPrefix(name, CompareSlice("Float64Slice", a.Value.AsFloat64Slice(), b.Value.AsFloat64Slice()))
	default:
		return nil
	}
}

func CompareSlice[T comparable](name string, a []T, b []T) []Diff {
	var diffs []Diff
	if len(a) != len(b) {
		return []Diff{{Path: []string{name, "len"}, A: strconv.Itoa(len(a)), B: strconv.Itoa(len(b))}}
	}
	for i := 0; i < len(a); i++ {
		diffs = append(diffs, Compare("["+strconv.Itoa(i)+"]", a[i], b[i])...)
	}
	return diffPrefix(name, diffs)
}

func CompareTime(name string, a time.Time, b time.Time) []Diff {
	if a.Equal(b) {
		return nil
	}
	return []Diff{{Path: []string{name}, A: a.Format(time.RFC3339Nano), B: b.Format(time.RFC3339Nano)}}
}

func CompareSpanContext(name string, a SpanContext, b SpanContext) []Diff {
	if a.IsValid() != b.IsValid() {
		if a.IsValid() {
			return []Diff{{Path: []string{name}, A: a}}
		} else {
			return []Diff{{Path: []string{name}, B: b}}
		}
	}
	if !a.IsValid() {
		return nil
	}
	var diffs []Diff
	if a.SpanID() != b.SpanID() {
		diffs = append(diffs, Diff{Path: []string{"SpanID"}, A: a.SpanID(), B: b.SpanID()})
	}
	if a.TraceID() != b.TraceID() {
		diffs = append(diffs, Diff{Path: []string{"TraceID"}, A: a.TraceID(), B: b.TraceID()})
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
	return []Diff{{Path: []string{name}, A: a, B: b}}
}
