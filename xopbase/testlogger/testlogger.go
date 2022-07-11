package testlogger

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func (l *TestLogger) setShort(span trace.Bundle) string {
	ts := span.Trace.GetTraceId().String()
	if ti, ok := l.traceMap[ts]; ok {
		ti.spanCount++
		ti.spans[span.Trace.GetSpanId().String()] = ti.spanCount
		short := fmt.Sprintf("T%d.%d", ti.traceNum, ti.spanCount)
		l.t.Log("Start span " + short + "=" + span.Trace.HeaderString())
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
	l.t.Log("Start span " + short + "=" + span.Trace.HeaderString())
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
	Data         []xop.Thing
	SpanType     xopconst.SpanType
	short        string
}

type Line struct {
	Things    xop.Things
	Level     xopconst.Level
	Timestamp time.Time
	Span      *Span
	Message   string
}

func (l *TestLogger) WithMe() xoplog.SeedModifier {
	return xoplog.WithBaseLogger("testing", l)
}

func (l *TestLogger) Close()               {}
func (l *TestLogger) ReferencesKept() bool { return true }
func (l *TestLogger) Request(span trace.Bundle) xopbase.Request {
	l.lock.Lock()
	defer l.lock.Unlock()
	return &Span{
		testLogger: l,
		IsRequest:  true,
		Trace:      span,
		short:      l.setShort(span),
	}
}

func (l *Span) Flush() {}

func (s *Span) Span(span trace.Bundle) xopbase.Span {
	s.testLogger.lock.Lock()
	defer s.testLogger.lock.Unlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	n := &Span{
		testLogger: s.testLogger,
		Trace:      span,
		short:      s.testLogger.setShort(span),
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
	}
	return line
}

func (l Line) Any(k string, v interface{}) { l.Things.AnyImmutable(k, v) }
func (l Line) Int(k string, v int64)       { l.Things.Int(k, v) }
func (l Line) Uint(k string, v uint64)     { l.Things.Uint(k, v) }
func (l Line) Str(k string, v string)      { l.Things.Str(k, v) }
func (l Line) Bool(k string, v bool)       { l.Things.Bool(k, v) }
func (l Line) Error(k string, v error)     { l.Things.Error(k, v) }
func (l Line) Time(k string, v time.Time)  { l.Things.Time(k, v) }
func (l *Line) Recycle(level xopconst.Level, t time.Time) {
	l.Level = level
	l.Timestamp = t
	l.Things = xop.Things{}
	l.Message = ""
}

func (l Line) Msg(m string) {
	l.Message = m
	text := l.Span.short + ": " + m
	// TODO: replace with higher performance version
	// TODO: move encoding somewhere else
	for _, thing := range l.Things.Things {
		text += " " + thing.Key + "="
		switch thing.Type {
		case xop.IntType:
			text += strconv.FormatInt(thing.Int, 64)
		case xop.UintType:
			text += strconv.FormatUint(thing.Any.(uint64), 64)
		case xop.BoolType:
			text += strconv.FormatBool(thing.Any.(bool))
		case xop.StringType:
			text += strconv.Quote(thing.String)
		case xop.TimeType:
			text += thing.Any.(time.Time).Format(time.RFC3339)
		case xop.AnyType:
			enc, err := json.Marshal(thing.Any)
			if err != nil {
				text += "???(marshal error:" + err.Error() + ")"
			} else {
				text += string(enc)
			}
		case xop.ErrorType:
			text += thing.Any.(error).Error()
		case xop.UnsetType:
			fallthrough
		// TODO: more types?
		default:
			text += "???(unknown thing" + strconv.Itoa(int(thing.Type)) + ")"
		}
	}

	l.Span.testLogger.t.Log(text)
	l.Span.testLogger.lock.Lock()
	defer l.Span.testLogger.lock.Unlock()
	l.Span.lock.Lock()
	defer l.Span.lock.Unlock()
	l.Span.testLogger.Lines = append(l.Span.testLogger.Lines, &l)
	l.Span.Lines = append(l.Span.Lines, &l)
}
