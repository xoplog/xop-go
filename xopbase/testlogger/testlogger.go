package testlogger

import (
	"sync"
	"time"

	"github.com/muir/xoplog"
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xop"
)

type testingT interface {
	Log(...interface{})
}

var _ xoplog.BaseLogger = &TestLogger{}

func NewTestLogger(t testingT) *TestLogger {
	return &TestLogger{
		t: t,
	}
}

type TestLogger struct {
	lock  sync.Mutex
	t     testingT
	Lines []*Line
	Spans []*Span
}

type Span struct {
	lock        sync.Mutex
	Span        trace.Trace
	Parent      *Span
	Spans       []*Span
	Lines       []*Line
	TestLogger  *TestLogger
	Data        []xop.Thing
	SpanPrefill []xop.Thing
	LinePrefill []xop.Thing
}

type Line struct {
	xop.Things
	Level   xopconst.Level
	Message string
	Span    *Span
	Time    time.Time
}

func (l *TestLogger) Close()                      {}
func (l *TestLogger) Request() xoplog.BaseRequest { return l }
func (l *TestLogger) Flush()                      {}
func (l *TestLogger) Span(span trace.Trace) xoplog.BaseSpan {
	l.lock.Lock()
	defer l.lock.Unlock()
	s := &Span{
		Logger: l,
	}
	l.Spans = append(l.Spans, s)
	return s
}

func (s *Span) Data(data []xop.Thing) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Data = append(s.Data, data...)
}
func (s *Span) LinePrefil(data []xop.Thing) {
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
}
func (s *Span) Span(span trace.Trace) xoplog.BaseSpan {
	n := s.TestLogger.Span(span)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Spans = append(s.Spans, n)
	return n
}

func (l *Line) Msg(m string) {
	l.Message = m
	s.TestLogger.lock.Lock()
	defer s.TestLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.TestLogger.Lines = append(s.TestLogger.Lines, l)
	s.Lines = append(s.Lines, l)
}
