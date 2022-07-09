package multibase

import (
	"sync"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type Loggers []xopbase.Logger
type Requests []xopbase.Request
type Spans []xopbase.Span
type Lines []xopbase.Line

var _ xopbase.Logger = Loggers{}
var _ xopbase.Request = Requests{}
var _ xopbase.Span = Spans{}
var _ xopbase.Line = Lines{}

func CombineLoggers(loggers []xopbase.Logger) xopbase.Logger {
	if len(loggers) == 1 {
		return loggers[0]
	}
	return Loggers(loggers)
}

func (l Loggers) Request(span trace.Bundle) xopbase.Request {
	r := make(Requests, len(l))
	for i, logger := range l {
		r[i] = logger.Request(span)
	}
	return r
}

func (l Loggers) ReferencesKept() bool {
	for _, logger := range l {
		if logger.ReferencesKept() {
			return true
		}
	}
	return false
}

func (l Loggers) Close() {
	for _, logger := range l {
		logger.Close()
	}
}

func (s Requests) Flush() {
	var wg sync.WaitGroup
	wg.Add(len(s))
	for _, request := range s {
		go func() {
			defer wg.Done()
			request.Flush()
		}()
	}
	wg.Wait()
}

func (s Requests) Span(span trace.Bundle) xopbase.Span {
	spans := make(Spans, len(s))
	for i, ele := range s {
		spans[i] = ele.Span(span)
	}
	return spans
}
func (s Spans) Span(span trace.Bundle) xopbase.Span {
	spans := make(Spans, len(s))
	for i, ele := range s {
		spans[i] = ele.Span(span)
	}
	return spans
}

func (s Requests) SpanInfo(st xopconst.SpanType, things []xop.Thing) {
	for _, span := range s {
		span.SpanInfo(st, things)
	}
}
func (s Spans) SpanInfo(st xopconst.SpanType, things []xop.Thing) {
	for _, span := range s {
		span.SpanInfo(st, things)
	}
}

func (s Requests) AddPrefill(things []xop.Thing) {
	for _, span := range s {
		span.AddPrefill(things)
	}
}
func (s Spans) AddPrefill(things []xop.Thing) {
	for _, span := range s {
		span.AddPrefill(things)
	}
}

func (s Requests) ResetLinePrefill() {
	for _, span := range s {
		span.ResetLinePrefill()
	}
}
func (s Spans) ResetLinePrefill() {
	for _, span := range s {
		span.ResetLinePrefill()
	}
}

func (s Requests) Line(level xopconst.Level, t time.Time) xopbase.Line {
	lines := make(Lines, len(s))
	for i, span := range s {
		lines[i] = span.Line(level, t)
	}
	return lines
}
func (s Spans) Line(level xopconst.Level, t time.Time) xopbase.Line {
	lines := make(Lines, len(s))
	for i, span := range s {
		lines[i] = span.Line(level, t)
	}
	return lines
}

func (l Lines) Int(k string, v int64) {
	for _, line := range l {
		line.Int(k, v)
	}
}
func (l Lines) Str(k string, v string) {
	for _, line := range l {
		line.Str(k, v)
	}
}
func (l Lines) Bool(k string, v bool) {
	for _, line := range l {
		line.Bool(k, v)
	}
}
func (l Lines) Uint(k string, v uint64) {
	for _, line := range l {
		line.Uint(k, v)
	}
}
func (l Lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}
func (l Lines) Any(k string, v interface{}) {
	for _, line := range l {
		line.Any(k, v)
	}
}
func (l Lines) Error(k string, v error) {
	for _, line := range l {
		line.Error(k, v)
	}
}
func (l Lines) Msg(m string) {
	for _, line := range l {
		line.Msg(m)
	}
}

func (l Lines) Things(things []xop.Thing) {
	for _, line := range l {
		xopbase.LineThings(line, things)
	}
}
