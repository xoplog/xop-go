package testlogger

import (
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
}

var _ xopbase.Logger = &TestLogger{}
var _ xopbase.Request = &Span{}
var _ xopbase.Span = &Span{}
var _ xopbase.Line = &Line{}

func NewTestLogger(t testingT) *TestLogger {
	return &TestLogger{
		t: t,
	}
}

type TestLogger struct {
	lock     sync.Mutex
	t        testingT
	Requests []*Span
	Spans    []*Span
	Lines    []*Line
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
	Data         []xop.Thing
	SpanPrefill  []xop.Thing
	LinePrefill  []xop.Thing
}

type Line struct {
	xop.Things
	Level     xopconst.Level
	Time      time.Time
	Span      *Span
	Message   string
	Completed bool
}

func (l *TestLogger) Close()               {}
func (l *TestLogger) ReferencesKept() bool { return true }
func (l *TestLogger) Request(bundle trace.Bundle) xopbase.Request {
	return &Span{
		testLogger: l,
		IsRequest:  true,
		Trace:      bundle,
	}
}

func (l *Span) Flush() {}

func (s *Span) Span(span trace.Bundle) xoplog.Span {
	s.testLogger.lock.Lock()
	defer s.testLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	n := &Span{
		testLogger: s.testLogger,
	}
	s.Spans = append(s.Spans, n)
	s.testLogger.Spans = append(s.testLogger.Spans, n)
	return n
}

func (s *Span) SpanInfo(spanType xopconst.SpanType, data []xop.Thing) {
	s.SpanType = spanType
	s.Data = data
}

func (s *Span) AddPrefill(data []xop.Thing) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LinePrefill = append(s.LinePrefill, data...)
}

func (s *Span) ResetPrefil(data []xop.Thing) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LinePrefill = nil
}

func (s *Span) Line(level xopconst.Level, t time.Time) xoplog.BaseLine {
	line := &Line{
		Level: level,
		Time:  t,
		Span:  s,
	}
	s.testLogger.lock.Lock()
	defer s.testLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.testLogger.Lines = append(s.testLogger.Lines, line)
	s.Lines = append(s.Lines, line)
	return line
}

func (l *Line) Any(k string, v interface{}) { l.Things.Any(k, v) }
func (l *Line) Int(k string, v int64)       { l.Things.Int(k, v) }
func (l *Line) Str(k string, v string)      { l.Things.Str(k, v) }
func (l *Line) Bool(k string, v bool)       { l.Things.Bool(k, v) }
func (l *Line) Time(k string, v time.Time)  { l.Things.Time(k, v) }

func (l *Line) Msg(m string) {
	l.Message = m
	l.Completed = true
}
