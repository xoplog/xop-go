// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"sync"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
)

type (
	baseLoggers  []xopbase.Logger
	baseRequests struct {
		baseSpans
		baseRequests []xopbase.Request
	}
)
type (
	baseSpans   []xopbase.Span
	lines       []xopbase.Line
	prefilleds  []xopbase.Prefilled
	prefillings []xopbase.Prefilling
)

var (
	_ xopbase.Request    = baseRequests{}
	_ xopbase.Span       = baseSpans{}
	_ xopbase.Line       = lines{}
	_ xopbase.Prefilled  = prefilleds{}
	_ xopbase.Prefilling = prefillings{}
)

func (l baseLoggers) StartRequests(span trace.Bundle, descriptionOrName string) (xopbase.Request, map[string]xopbase.Request) {
	if len(l) == 1 {
		req := l[0].Request(span, descriptionOrName)
		return req, map[string]xopbase.Request{l[0].ID(): req}
	}
	m := make(map[string]xopbase.Request)
	r := baseRequests{
		baseSpans:    make(baseSpans, 0, len(l)),
		baseRequests: make([]xopbase.Request, 0, len(l)),
	}
	for _, logger := range l {
		id := logger.ID()
		if _, ok := m[id]; ok {
			// duplicate!
			continue
		}
		req := logger.Request(span, descriptionOrName)
		r.baseRequests = append(r.baseRequests, req)
		r.baseSpans = append(r.baseSpans, req.(xopbase.Span))
		m[id] = req
	}
	return r, m
}

func (l baseLoggers) ReferencesKept() bool {
	for _, logger := range l {
		if logger.ReferencesKept() {
			return true
		}
	}
	return false
}

func (l baseLoggers) Buffered() bool {
	for _, logger := range l {
		if logger.Buffered() {
			return true
		}
	}
	return false
}

func (s baseRequests) SetErrorReporter(f func(error)) {
	for _, request := range s.baseRequests {
		request.SetErrorReporter(f)
	}
}

func (s baseRequests) Flush() {
	var wg sync.WaitGroup
	wg.Add(len(s.baseRequests))
	for _, request := range s.baseRequests {
		go func() {
			defer wg.Done()
			request.Flush()
		}()
	}
	wg.Wait()
}

func (s baseSpans) Span(t time.Time, span trace.Bundle, descriptionOrName string) xopbase.Span {
	baseSpans := make(baseSpans, len(s))
	for i, ele := range s {
		baseSpans[i] = ele.Span(t, span, descriptionOrName)
	}
	return baseSpans
}

func (s baseSpans) Done(t time.Time) {
	for _, span := range s {
		span.Done(t)
	}
}

func (s baseSpans) ID() string {
	panic("this is not expected to be called")
}

func (s baseSpans) Boring(b bool) {
	for _, span := range s {
		span.Boring(b)
	}
}

func (s baseSpans) NoPrefill() xopbase.Prefilled {
	prefilled := make(prefilleds, len(s))
	for i, span := range s {
		prefilled[i] = span.NoPrefill()
	}
	return prefilled
}

func (s baseSpans) StartPrefill() xopbase.Prefilling {
	prefilling := make(prefillings, len(s))
	for i, span := range s {
		prefilling[i] = span.StartPrefill()
	}
	return prefilling
}

func (p prefillings) PrefillComplete(m string) xopbase.Prefilled {
	prefilled := make(prefilleds, len(p))
	for i, prefilling := range p {
		prefilled[i] = prefilling.PrefillComplete(m)
	}
	return prefilled
}

func (p prefilleds) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	lines := make(lines, len(p))
	for i, prefilled := range p {
		lines[i] = prefilled.Line(level, t, pc)
	}
	return lines
}

func (l lines) Template(m string) {
	for _, line := range l {
		line.Template(m)
	}
}

func (l lines) Msg(m string) {
	for _, line := range l {
		line.Msg(m)
	}
}

func (l lines) Static(m string) {
	for _, line := range l {
		line.Static(m)
	}
}

func (l lines) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	for _, line := range l {
		line.Enum(k, v)
	}
}

func (p prefillings) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	for _, prefilling := range p {
		prefilling.Enum(k, v)
	}
}

func (p prefillings) Any(k string, v interface{}) {
	for _, prefilling := range p {
		prefilling.Any(k, v)
	}
}

func (p prefillings) Bool(k string, v bool) {
	for _, prefilling := range p {
		prefilling.Bool(k, v)
	}
}

func (p prefillings) Duration(k string, v time.Duration) {
	for _, prefilling := range p {
		prefilling.Duration(k, v)
	}
}

func (p prefillings) Error(k string, v error) {
	for _, prefilling := range p {
		prefilling.Error(k, v)
	}
}

func (p prefillings) Float64(k string, v float64) {
	for _, prefilling := range p {
		prefilling.Float64(k, v)
	}
}

func (p prefillings) Int(k string, v int64) {
	for _, prefilling := range p {
		prefilling.Int(k, v)
	}
}

func (p prefillings) Link(k string, v trace.Trace) {
	for _, prefilling := range p {
		prefilling.Link(k, v)
	}
}

func (p prefillings) String(k string, v string) {
	for _, prefilling := range p {
		prefilling.String(k, v)
	}
}

func (p prefillings) Time(k string, v time.Time) {
	for _, prefilling := range p {
		prefilling.Time(k, v)
	}
}

func (p prefillings) Uint(k string, v uint64) {
	for _, prefilling := range p {
		prefilling.Uint(k, v)
	}
}

func (l lines) Any(k string, v interface{}) {
	for _, line := range l {
		line.Any(k, v)
	}
}

func (l lines) Bool(k string, v bool) {
	for _, line := range l {
		line.Bool(k, v)
	}
}

func (l lines) Duration(k string, v time.Duration) {
	for _, line := range l {
		line.Duration(k, v)
	}
}

func (l lines) Error(k string, v error) {
	for _, line := range l {
		line.Error(k, v)
	}
}

func (l lines) Float64(k string, v float64) {
	for _, line := range l {
		line.Float64(k, v)
	}
}

func (l lines) Int(k string, v int64) {
	for _, line := range l {
		line.Int(k, v)
	}
}

func (l lines) Link(k string, v trace.Trace) {
	for _, line := range l {
		line.Link(k, v)
	}
}

func (l lines) String(k string, v string) {
	for _, line := range l {
		line.String(k, v)
	}
}

func (l lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}

func (l lines) Uint(k string, v uint64) {
	for _, line := range l {
		line.Uint(k, v)
	}
}

func (s baseSpans) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	for _, span := range s {
		span.MetadataAny(k, v)
	}
}

func (s baseSpans) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	for _, span := range s {
		span.MetadataBool(k, v)
	}
}

func (s baseSpans) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	for _, span := range s {
		span.MetadataEnum(k, v)
	}
}

func (s baseSpans) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	for _, span := range s {
		span.MetadataFloat64(k, v)
	}
}

func (s baseSpans) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
	for _, span := range s {
		span.MetadataInt64(k, v)
	}
}

func (s baseSpans) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	for _, span := range s {
		span.MetadataLink(k, v)
	}
}

func (s baseSpans) MetadataStr(k *xopconst.StringAttribute, v string) {
	for _, span := range s {
		span.MetadataStr(k, v)
	}
}

func (s baseSpans) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	for _, span := range s {
		span.MetadataTime(k, v)
	}
}
