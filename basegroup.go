// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
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

// CombineBaseLoggers is probably only useful for testing xop because Seed already
// provides a better way to manage multiple base loggers.
func CombineBaseLoggers(first xopbase.Logger, more ...xopbase.Logger) xopbase.Logger {
	if len(more) == 0 {
		return first
	}
	return baseLoggers(append([]xopbase.Logger{first}, more...))
}

func (l baseLoggers) Request(ctx context.Context, ts time.Time, span xoptrace.Bundle, descriptionOrName string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	r, _ := l.startRequests(ctx, ts, span, descriptionOrName, sourceInfo)
	return r
}

func (l baseLoggers) ID() string {
	ids := make([]string, len(l))
	for i, logger := range l {
		ids[i] = logger.ID()
	}
	return strings.Join(ids, "/")
}

func (l baseLoggers) startRequests(ctx context.Context, ts time.Time, span xoptrace.Bundle, descriptionOrName string, sourceInfo xopbase.SourceInfo) (xopbase.Request, map[string]xopbase.Request) {
	if len(l) == 1 {
		req := l[0].Request(ctx, ts, span, descriptionOrName, sourceInfo)
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
		req := logger.Request(ctx, ts, span, descriptionOrName, sourceInfo)
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

func (s baseRequests) Boring(isBoring bool) {
	for _, request := range s.baseRequests {
		request.Boring(isBoring)
	}
}

func (s baseRequests) Final() {
	for _, request := range s.baseRequests {
		go func() {
			request.Final()
		}()
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

func (s baseSpans) Span(ctx context.Context, t time.Time, span xoptrace.Bundle, descriptionOrName string, spanSequenceCode string) xopbase.Span {
	baseSpans := make(baseSpans, len(s))
	for i, ele := range s {
		baseSpans[i] = ele.Span(ctx, t, span, descriptionOrName, spanSequenceCode)
	}
	return baseSpans
}

func (s baseSpans) Done(t time.Time, final bool) {
	for _, span := range s {
		span.Done(t, final)
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

func (p prefilleds) Line(level xopnum.Level, t time.Time, pc []uintptr) xopbase.Line {
	lines := make(lines, len(p))
	for i, prefilled := range p {
		lines[i] = prefilled.Line(level, t, pc)
	}
	return lines
}

func (l lines) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	for _, line := range l {
		line.Enum(k, v)
	}
}

func (p prefillings) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	for _, prefilling := range p {
		prefilling.Enum(k, v)
	}
}

func (l lines) Msg(m string) {
	for _, line := range l {
		line.Msg(m)
	}
}

func (l lines) Template(m string) {
	for _, line := range l {
		line.Template(m)
	}
}

func (l lines) Link(k string, v xoptrace.Trace) {
	for _, line := range l {
		line.Link(k, v)
	}
}

func (l lines) Model(k string, v xopbase.ModelArg) {
	for _, line := range l {
		line.Model(k, v)
	}
}

func (p prefillings) Any(k string, v xopbase.ModelArg) {
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

func (p prefillings) Time(k string, v time.Time) {
	for _, prefilling := range p {
		prefilling.Time(k, v)
	}
}

func (p prefillings) Float64(k string, v float64, dt xopbase.DataType) {
	for _, prefilling := range p {
		prefilling.Float64(k, v, dt)
	}
}

func (p prefillings) Int64(k string, v int64, dt xopbase.DataType) {
	for _, prefilling := range p {
		prefilling.Int64(k, v, dt)
	}
}

func (p prefillings) String(k string, v string, dt xopbase.DataType) {
	for _, prefilling := range p {
		prefilling.String(k, v, dt)
	}
}

func (p prefillings) Uint64(k string, v uint64, dt xopbase.DataType) {
	for _, prefilling := range p {
		prefilling.Uint64(k, v, dt)
	}
}

func (l lines) Any(k string, v xopbase.ModelArg) {
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

func (l lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}

func (l lines) Float64(k string, v float64, dt xopbase.DataType) {
	for _, line := range l {
		line.Float64(k, v, dt)
	}
}

func (l lines) Int64(k string, v int64, dt xopbase.DataType) {
	for _, line := range l {
		line.Int64(k, v, dt)
	}
}

func (l lines) String(k string, v string, dt xopbase.DataType) {
	for _, line := range l {
		line.String(k, v, dt)
	}
}

func (l lines) Uint64(k string, v uint64, dt xopbase.DataType) {
	for _, line := range l {
		line.Uint64(k, v, dt)
	}
}

func (s baseSpans) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	for _, span := range s {
		span.MetadataAny(k, v)
	}
}

func (s baseSpans) MetadataBool(k *xopat.BoolAttribute, v bool) {
	for _, span := range s {
		span.MetadataBool(k, v)
	}
}

func (s baseSpans) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	for _, span := range s {
		span.MetadataEnum(k, v)
	}
}

func (s baseSpans) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	for _, span := range s {
		span.MetadataFloat64(k, v)
	}
}

func (s baseSpans) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	for _, span := range s {
		span.MetadataInt64(k, v)
	}
}

func (s baseSpans) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	for _, span := range s {
		span.MetadataLink(k, v)
	}
}

func (s baseSpans) MetadataString(k *xopat.StringAttribute, v string) {
	for _, span := range s {
		span.MetadataString(k, v)
	}
}

func (s baseSpans) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	for _, span := range s {
		span.MetadataTime(k, v)
	}
}
