package testlogger

import (
	"fmt"
	"sync"
	"time"

	"github.com/muir/xoplog"
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
)

type testingT interface {
	Log(...interface{})
	Name() string
}

var _ xopbase.Logger = &TestLogger{}
var _ xopbase.Request = &Span{}
var _ xopbase.Span = &Span{}
var _ xopbase.Line = &Line{}

func New(t testingT) *TestLogger {
	return &TestLogger{
		t:        t,
		traceMap: make(map[string]*traceInfo),
	}
}

type TestLogger struct {
	lock       sync.Mutex
	t          testingT
	Requests   []*Span
	Spans      []*Span
	Lines      []*Line
	traceCount int
	traceMap   map[string]*traceInfo
}

type traceInfo struct {
	spanCount int
	traceNum  int
	spans     map[string]int
}

func (l *TestLogger) setShort(span trace.Bundle, name string) string {
	ts := span.Trace.GetTraceId().String()
	if ti, ok := l.traceMap[ts]; ok {
		ti.spanCount++
		ti.spans[span.Trace.GetSpanId().String()] = ti.spanCount
		short := fmt.Sprintf("T%d.%d", ti.traceNum, ti.spanCount)
		l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
		return short
	}
	l.traceCount++
	l.traceMap[ts] = &traceInfo{
		spanCount: 1,
		traceNum:  l.traceCount,
		spans: map[string]int{
			span.Trace.GetSpanId().String(): 1,
		},
	}
	short := fmt.Sprintf("T%d.%d", l.traceCount, 1)
	l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
	return short
}

type Span struct {
	lock         sync.Mutex
	testLogger   *TestLogger
	Trace        trace.Bundle
	IsRequest    bool
	Parent       *Span
	Spans        []*Span
	RequestLines []*Line
	Lines        []*Line
	short        string
}

type Line struct {
	Level     xopconst.Level
	Timestamp time.Time
	Span      *Span
	Message   string
	Data      map[string]interface{}
	Text      string
	kvText    []string
}

func (l *TestLogger) WithMe() xoplog.SeedModifier {
	return xoplog.WithBaseLogger("testing", l)
}

func (l *TestLogger) Close()               {}
func (l *TestLogger) ReferencesKept() bool { return true }
func (l *TestLogger) Request(span trace.Bundle, name string) xopbase.Request {
	l.lock.Lock()
	defer l.lock.Unlock()
	return &Span{
		testLogger: l,
		IsRequest:  true,
		Trace:      span,
		short:      l.setShort(span, name),
	}
}

func (l *Span) Flush() {}

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	s.testLogger.lock.Lock()
	defer s.testLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	n := &Span{
		testLogger: s.testLogger,
		Trace:      span,
		short:      s.testLogger.setShort(span, name),
	}
	s.Spans = append(s.Spans, n)
	s.testLogger.Spans = append(s.testLogger.Spans, n)
	return n
}

func (s *Span) SpanInfo(spanType xopconst.SpanType, data []xop.Thing) {
	s.SpanType = spanType
	s.Data = data
}

func (s *Span) Line(level xopconst.Level, t time.Time) xopbase.Line {
	line := &Line{
		Level:     level,
		Timestamp: t,
		Span:      s,
		Data:      make(map[string]interface{}),
	}
	return line
}

func (l *Line) Recycle(level xopconst.Level, t time.Time) {
	l.Level = level
	l.Timestamp = t
	l.kvText = nil
	l.Message = ""
	l.Data = make(map[string]interface{})
	l.Text = ""
}

func (l Line) Int(k string, v int64)      { l.Any(k, v) }
func (l Line) Uint(k string, v uint64)    { l.Any(k, v) }
func (l Line) Str(k string, v string)     { l.Any(k, v) }
func (l Line) Bool(k string, v bool)      { l.Any(k, v) }
func (l Line) Error(k string, v error)    { l.Any(k, v) }
func (l Line) Time(k string, v time.Time) { l.Any(k, v) }
func (l Line) Any(k string, v interface{}) {
	l.Data[k] = v
	l.kvText = append(l.kvText, fmt.Sprintf("%s=%+v", k, v))
}

func (l Line) Msg(m string) {
	l.Message = m
	text := l.Span.short + ": " + m
	if len(l.kvText) > 0 {
		text += " " + string.Join(l.kvText, " ")
		l.kvText = nil
	}
	l.Text = text

	l.Span.testLogger.t.Log(text)
	l.Span.testLogger.lock.Lock()
	defer l.Span.testLogger.lock.Unlock()
	l.Span.lock.Lock()
	defer l.Span.lock.Unlock()
	l.Span.testLogger.Lines = append(l.Span.testLogger.Lines, &l)
	l.Span.Lines = append(l.Span.Lines, &l)
}
