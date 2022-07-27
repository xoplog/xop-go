// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"sync"
	"time"

	"github.com/muir/xop/trace"
	"github.com/muir/xop/xopbase"
	"github.com/muir/xop/xopconst"
)

type (
	baseLoggers  []xopbase.Logger
	baseRequests struct {
		baseSpans
		baseRequests []xopbase.Request
	}
)
type (
	baseSpans []xopbase.Span
	lines     []xopbase.Line
)

var (
	_ xopbase.Request = baseRequests{}
	_ xopbase.Span    = baseSpans{}
	_ xopbase.Line    = lines{}
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

func (l baseLoggers) StackFramesWanted() map[xopconst.Level]int {
	combined := make(map[xopconst.Level]int)
	for _, logger := range l {
		for level, frames := range logger.StackFramesWanted() {
			if frames > combined[level] {
				combined[level] = frames
			}
		}
	}
	return combined
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

func (s baseSpans) Span(span trace.Bundle, descriptionOrName string) xopbase.Span {
	baseSpans := make(baseSpans, len(s))
	for i, ele := range s {
		baseSpans[i] = ele.Span(span, descriptionOrName)
	}
	return baseSpans
}

func (s baseSpans) ID() string {
	panic("this is not expected to be called")
}

func (s baseSpans) Boring(b bool) {
	for _, span := range s {
		span.Boring(b)
	}
}

func (s baseSpans) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	lines := make(lines, len(s))
	for i, span := range s {
		lines[i] = span.Line(level, t, pc)
	}
	return lines
}

func (l lines) Recycle(level xopconst.Level, t time.Time, pc []uintptr) {
	for _, line := range l {
		line.Recycle(level, t, pc)
	}
}

func (l lines) SetAsPrefill(m string) {
	for _, line := range l {
		line.SetAsPrefill(m)
	}
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

// Any adds a interface{} key/value pair to a line that is in progress
func (l lines) Any(k string, v interface{}) {
	for _, line := range l {
		line.Any(k, v)
	}
}

// Bool adds a bool key/value pair to a line that is in progress
func (l lines) Bool(k string, v bool) {
	for _, line := range l {
		line.Bool(k, v)
	}
}

// Duration adds a time.Duration key/value pair to a line that is in progress
func (l lines) Duration(k string, v time.Duration) {
	for _, line := range l {
		line.Duration(k, v)
	}
}

// Error adds a error key/value pair to a line that is in progress
func (l lines) Error(k string, v error) {
	for _, line := range l {
		line.Error(k, v)
	}
}

// Int adds a int64 key/value pair to a line that is in progress
func (l lines) Int(k string, v int64) {
	for _, line := range l {
		line.Int(k, v)
	}
}

// Link adds a trace.Trace key/value pair to a line that is in progress
func (l lines) Link(k string, v trace.Trace) {
	for _, line := range l {
		line.Link(k, v)
	}
}

// Str adds a string key/value pair to a line that is in progress
func (l lines) Str(k string, v string) {
	for _, line := range l {
		line.Str(k, v)
	}
}

// Time adds a time.Time key/value pair to a line that is in progress
func (l lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}

// Uint adds a uint64 key/value pair to a line that is in progress
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

func (s baseSpans) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	for _, span := range s {
		span.MetadataNumber(k, v)
	}
}

func (s baseSpans) MetadataStr(k *xopconst.StrAttribute, v string) {
	for _, span := range s {
		span.MetadataStr(k, v)
	}
}

func (s baseSpans) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	for _, span := range s {
		span.MetadataTime(k, v)
	}
}
