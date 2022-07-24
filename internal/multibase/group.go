// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package multibase

import (
	"sync"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type (
	Loggers  []xopbase.Logger
	Requests struct {
		Spans
		Requests []xopbase.Request
	}
)
type (
	Spans []xopbase.Span
	Lines []xopbase.Line
)

var (
	_ xopbase.Logger  = Loggers{}
	_ xopbase.Request = Requests{}
	_ xopbase.Span    = Spans{}
	_ xopbase.Line    = Lines{}
)

func CombineLoggers(loggers []xopbase.Logger) xopbase.Logger {
	if len(loggers) == 1 {
		return loggers[0]
	}
	return Loggers(loggers)
}

func (l Loggers) Request(span trace.Bundle, descriptionOrName string) xopbase.Request {
	r := Requests{
		Spans:    make(Spans, len(l)),
		Requests: make([]xopbase.Request, len(l)),
	}
	for i, logger := range l {
		r.Requests[i] = logger.Request(span, descriptionOrName)
		r.Spans[i] = r.Requests[i].(xopbase.Span)
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

func (l Loggers) Buffered() bool {
	for _, logger := range l {
		if logger.Buffered() {
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

func (l Loggers) StackFramesWanted() map[xopconst.Level]int {
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

func (l Loggers) SetErrorReporter(f func(error)) {
	for _, logger := range l {
		logger.SetErrorReporter(f)
	}
}

func (s Requests) Flush() {
	var wg sync.WaitGroup
	wg.Add(len(s.Requests))
	for _, request := range s.Requests {
		go func() {
			defer wg.Done()
			request.Flush()
		}()
	}
	wg.Wait()
}

func (s Spans) Span(span trace.Bundle, descriptionOrName string) xopbase.Span {
	spans := make(Spans, len(s))
	for i, ele := range s {
		spans[i] = ele.Span(span, descriptionOrName)
	}
	return spans
}

func (s Spans) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	for _, span := range s {
		span.MetadataAny(k, v)
	}
}

func (s Spans) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	for _, span := range s {
		span.MetadataBool(k, v)
	}
}

func (s Spans) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	for _, span := range s {
		span.MetadataEnum(k, v)
	}
}

func (s Spans) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
	for _, span := range s {
		span.MetadataInt64(k, v)
	}
}

func (s Spans) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	for _, span := range s {
		span.MetadataLink(k, v)
	}
}

func (s Spans) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	for _, span := range s {
		span.MetadataNumber(k, v)
	}
}

func (s Spans) MetadataStr(k *xopconst.StrAttribute, v string) {
	for _, span := range s {
		span.MetadataStr(k, v)
	}
}

func (s Spans) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	for _, span := range s {
		span.MetadataTime(k, v)
	}
}

func (s Spans) Boring(b bool) {
	for _, span := range s {
		span.Boring(b)
	}
}

func (s Spans) Line(level xopconst.Level, t time.Time) xopbase.Line {
	lines := make(Lines, len(s))
	for i, span := range s {
		lines[i] = span.Line(level, t)
	}
	return lines
}

func (l Lines) Recycle(level xopconst.Level, t time.Time) {
	for _, line := range l {
		line.Recycle(level, t)
	}
}

func (l Lines) SetAsPrefill(m string) {
	for _, line := range l {
		line.SetAsPrefill(m)
	}
}

func (l Lines) Template(m string) {
	for _, line := range l {
		line.Template(m)
	}
}

func (l Lines) Msg(m string) {
	for _, line := range l {
		line.Msg(m)
	}
}

func (l Lines) Static(m string) {
	for _, line := range l {
		line.Static(m)
	}
}

func (l Lines) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	for _, line := range l {
		line.Enum(k, v)
	}
}

// Any adds a interface{} key/value pair to a line that is in progress
func (l Lines) Any(k string, v interface{}) {
	for _, line := range l {
		line.Any(k, v)
	}
}

// Bool adds a bool key/value pair to a line that is in progress
func (l Lines) Bool(k string, v bool) {
	for _, line := range l {
		line.Bool(k, v)
	}
}

// Duration adds a time.Duration key/value pair to a line that is in progress
func (l Lines) Duration(k string, v time.Duration) {
	for _, line := range l {
		line.Duration(k, v)
	}
}

// Error adds a error key/value pair to a line that is in progress
func (l Lines) Error(k string, v error) {
	for _, line := range l {
		line.Error(k, v)
	}
}

// Int adds a int64 key/value pair to a line that is in progress
func (l Lines) Int(k string, v int64) {
	for _, line := range l {
		line.Int(k, v)
	}
}

// Link adds a trace.Trace key/value pair to a line that is in progress
func (l Lines) Link(k string, v trace.Trace) {
	for _, line := range l {
		line.Link(k, v)
	}
}

// Str adds a string key/value pair to a line that is in progress
func (l Lines) Str(k string, v string) {
	for _, line := range l {
		line.Str(k, v)
	}
}

// Time adds a time.Time key/value pair to a line that is in progress
func (l Lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}

// Uint adds a uint64 key/value pair to a line that is in progress
func (l Lines) Uint(k string, v uint64) {
	for _, line := range l {
		line.Uint(k, v)
	}
}
