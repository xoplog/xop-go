// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptest

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muir/xoplog"
	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
	"github.com/muir/xoplog/xoputil"
)

type testingT interface {
	Log(...interface{})
	Name() string
}

var (
	_ xopbase.Logger  = &TestLogger{}
	_ xopbase.Request = &Span{}
	_ xopbase.Span    = &Span{}
	_ xopbase.Line    = &Line{}
)

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

type Span struct {
	Attributes   xoputil.AttributeBuilder
	lock         sync.Mutex
	testLogger   *TestLogger
	Trace        trace.Bundle
	IsRequest    bool
	Parent       *Span
	Spans        []*Span
	RequestLines []*Line
	Lines        []*Line
	short        string
	Metadata     map[string]interface{}
	prefill      atomic.Value
}

type Line struct {
	Level     xopconst.Level
	Timestamp time.Time
	Span      *Span
	Message   string
	Data      map[string]interface{}
	Text      string
	Tmpl      string
	kvText    []string
}

func (l *TestLogger) WithMe() xoplog.SeedModifier {
	return xoplog.WithBaseLogger("testing", l)
}

func (l *TestLogger) Close()                                    {}
func (l *TestLogger) Buffered() bool                            { return false }
func (l *TestLogger) ReferencesKept() bool                      { return true }
func (l *TestLogger) SetErrorReporter(func(error))              {}
func (l *TestLogger) StackFramesWanted() map[xopconst.Level]int { return nil }
func (l *TestLogger) Request(span trace.Bundle, name string) xopbase.Request {
	l.lock.Lock()
	defer l.lock.Unlock()
	s := &Span{
		testLogger: l,
		IsRequest:  true,
		Trace:      span,
		short:      l.setShort(span, name),
	}
	s.Attributes.Reset()
	return s
}

func (l *TestLogger) setShort(span trace.Bundle, name string) string {
	ts := span.Trace.GetTraceID().String()
	if ti, ok := l.traceMap[ts]; ok {
		ti.spanCount++
		ti.spans[span.Trace.GetSpanID().String()] = ti.spanCount
		short := fmt.Sprintf("T%d.%d", ti.traceNum, ti.spanCount)
		l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
		return short
	}
	l.traceCount++
	l.traceMap[ts] = &traceInfo{
		spanCount: 1,
		traceNum:  l.traceCount,
		spans: map[string]int{
			span.Trace.GetSpanID().String(): 1,
		},
	}
	short := fmt.Sprintf("T%d.%d", l.traceCount, 1)
	l.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
	return short
}

func (s *Span) Flush()      {}
func (s *Span) Boring(bool) {}

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
	n.Attributes.Reset()
	s.Spans = append(s.Spans, n)
	s.testLogger.Spans = append(s.testLogger.Spans, n)
	return n
}

func (s *Span) GetPrefill() *Line {
	p := s.prefill.Load()
	if p == nil {
		return nil
	}
	return p.(*Line)
}

func (s *Span) Line(level xopconst.Level, t time.Time) xopbase.Line {
	line := &Line{
		Level:     level,
		Timestamp: t,
		Span:      s,
		Data:      make(map[string]interface{}),
	}
	p := s.GetPrefill()
	if p != nil {
		if len(p.Data) != 0 {
			for k, v := range p.Data {
				line.Data[k] = v
			}
		}
		if len(p.kvText) != 0 {
			line.kvText = make([]string, len(p.kvText), len(p.kvText)+5)
			copy(line.kvText, p.kvText)
		}
		line.Tmpl = p.Tmpl
		line.Message = p.Message
	}
	return line
}

func (l *Line) Recycle(level xopconst.Level, t time.Time) {
	l.Level = level
	l.Timestamp = t
	l.kvText = nil
	l.Message = ""
	l.Tmpl = ""
	l.Data = make(map[string]interface{})
	l.Text = ""
}

// TODO: test SetAsPrefill
func (l *Line) SetAsPrefill(m string) {
	l.Span.prefill.Store(l)
}

func (l *Line) Static(m string) {
	l.Msg(m)
}

func (l *Line) Msg(m string) {
	l.Message += m
	text := l.Span.short + ": " + l.Message
	if len(l.kvText) > 0 {
		text += " " + strings.Join(l.kvText, " ")
		l.kvText = nil
	}
	l.Text = text
	l.send(text)
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (l *Line) Template(m string) {
	l.Tmpl += m
	used := make(map[string]struct{})
	text := l.Span.short + ": " +
		templateRE.ReplaceAllStringFunc(l.Tmpl, func(k string) string {
			k = k[1 : len(k)-1]
			if v, ok := l.Data[k]; ok {
				used[k] = struct{}{}
				return fmt.Sprint(v)
			}
			return "''"
		})
	for k, v := range l.Data {
		if _, ok := used[k]; !ok {
			text += " " + k + "=" + fmt.Sprint(v)
		}
	}
	l.Text = text
	l.send(text)
}

func (l Line) send(text string) {
	l.Span.testLogger.t.Log(text)
	l.Span.testLogger.lock.Lock()
	defer l.Span.testLogger.lock.Unlock()
	l.Span.lock.Lock()
	defer l.Span.lock.Unlock()
	l.Span.testLogger.Lines = append(l.Span.testLogger.Lines, &l)
	l.Span.Lines = append(l.Span.Lines, &l)
}

func (l *Line) Any(k string, v interface{}) {
	l.Data[k] = v
	l.kvText = append(l.kvText, fmt.Sprintf("%s=%+v", k, v))
}

func (l *Line) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	l.Data[k.Key()] = v.String()
	l.kvText = append(l.kvText, fmt.Sprintf("%s=%s(%d)", k.Key(), v.String(), v.Int64()))
}

func (l *Line) Bool(k string, v bool)              { l.Any(k, v) }
func (l *Line) Duration(k string, v time.Duration) { l.Any(k, v) }
func (l *Line) Error(k string, v error)            { l.Any(k, v) }
func (l *Line) Int(k string, v int64)              { l.Any(k, v) }
func (l *Line) Link(k string, v trace.Trace)       { l.Any(k, v) }
func (l *Line) Str(k string, v string)             { l.Any(k, v) }
func (l *Line) Time(k string, v time.Time)         { l.Any(k, v) }
func (l *Line) Uint(k string, v uint64)            { l.Any(k, v) }

func (s *Span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.Attributes.MetadataAny(k, v) }
func (s *Span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.Attributes.MetadataBool(k, v) }
func (s *Span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.Attributes.MetadataEnum(k, v)
}
func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.Attributes.MetadataInt64(k, v) }
func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.Attributes.MetadataLink(k, v)
}
func (s *Span) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	s.Attributes.MetadataNumber(k, v)
}
func (s *Span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.Attributes.MetadataStr(k, v) }
func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.Attributes.MetadataTime(k, v) }

// end
