// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptest

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"
)

//go:generate enumer -type=EventType -linecomment -json -sql

type EventType int

const (
	LineEvent    EventType = iota // line
	RequestStart                  // requestStart
	RequestDone                   // requestDone
	SpanStart                     // spanStart
	SpanDone                      // spanStart
	FlushEvent                    // flush
	MetadataSet                   // metadata
	CustomEvent                   // custom
)

type testingT interface {
	Log(...interface{})
	Name() string
}

var (
	_ xopbase.Logger     = &TestLogger{}
	_ xopbase.Request    = &Span{}
	_ xopbase.Span       = &Span{}
	_ xopbase.Prefilling = &Prefilling{}
	_ xopbase.Prefilled  = &Prefilled{}
	_ xopbase.Line       = &Line{}
)

func New(t testingT) *TestLogger {
	return &TestLogger{
		t:        t,
		id:       t.Name() + "-" + uuid.New().String(),
		traceMap: make(map[string]*traceInfo),
	}
}

type TestLogger struct {
	lock       sync.Mutex
	t          testingT
	Requests   []*Span
	Spans      []*Span
	Lines      []*Line
	Events     []*Event
	traceCount int
	traceMap   map[string]*traceInfo
	id         string
}

type traceInfo struct {
	spanCount int
	traceNum  int
	spans     map[string]int
}

type Span struct {
	lock          sync.Mutex
	testLogger    *TestLogger
	Trace         trace.Bundle
	IsRequest     bool
	Parent        *Span
	Spans         []*Span
	RequestLines  []*Line
	Lines         []*Line
	short         string
	Metadata      map[string]interface{}
	MetadataTypes map[string]xoputil.BaseAttributeType
	StartTime     time.Time
	EndTime       int64
	Name          string
	SequenceCode  string
}

type Prefilling struct {
	Builder
}

type Builder struct {
	Data   map[string]interface{}
	Span   *Span
	kvText []string
}

type Prefilled struct {
	Data   map[string]interface{}
	Span   *Span
	Msg    string
	kvText []string
}

type Line struct {
	Builder
	Level     xopconst.Level
	Timestamp time.Time
	Message   string
	Text      string
	Tmpl      bool
}

type Event struct {
	Type EventType
	Line *Line
	Span *Span
	Msg  string
}

func (log *TestLogger) Log() *xop.Log {
	return xop.NewSeed(xop.WithBase(log)).Request(log.t.Name())
}

// WithLock is provided for thread-safe introspection of the logger
func (log *TestLogger) WithLock(f func(*TestLogger) error) error {
	log.lock.Lock()
	defer log.lock.Unlock()
	return f(log)
}

func (log *TestLogger) CustomEvent(msg string, args ...interface{}) {
	log.lock.Lock()
	defer log.lock.Unlock()
	log.Events = append(log.Events, &Event{
		Type: CustomEvent,
		Msg:  fmt.Sprintf(msg, args...),
	})
}

func (log *TestLogger) ID() string                   { return log.id }
func (log *TestLogger) Close()                       {}
func (log *TestLogger) Buffered() bool               { return false }
func (log *TestLogger) ReferencesKept() bool         { return true }
func (log *TestLogger) SetErrorReporter(func(error)) {}
func (log *TestLogger) Request(ts time.Time, span trace.Bundle, name string) xopbase.Request {
	log.lock.Lock()
	defer log.lock.Unlock()
	s := &Span{
		testLogger:    log,
		IsRequest:     true,
		Trace:         span,
		short:         log.setShort(span, name),
		StartTime:     ts,
		Name:          name,
		Metadata:      make(map[string]interface{}),
		MetadataTypes: make(map[string]xoputil.BaseAttributeType),
	}
	log.Requests = append(log.Requests, s)
	log.Events = append(log.Events, &Event{
		Type: RequestStart,
		Span: s,
	})
	return s
}

// must hold a lock to call setShort
func (log *TestLogger) setShort(span trace.Bundle, name string) string {
	ts := span.Trace.GetTraceID().String()
	if ti, ok := log.traceMap[ts]; ok {
		ti.spanCount++
		ti.spans[span.Trace.GetSpanID().String()] = ti.spanCount
		short := fmt.Sprintf("T%d.%d", ti.traceNum, ti.spanCount)
		log.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
		return short
	}
	log.traceCount++
	log.traceMap[ts] = &traceInfo{
		spanCount: 1,
		traceNum:  log.traceCount,
		spans: map[string]int{
			span.Trace.GetSpanID().String(): 1,
		},
	}
	short := fmt.Sprintf("T%d.%d", log.traceCount, 1)
	log.t.Log("Start span " + short + "=" + span.Trace.HeaderString() + " " + name)
	return short
}

func (span *Span) Done(t time.Time) {
	atomic.StoreInt64(&span.EndTime, t.UnixNano())
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	if span.IsRequest {
		span.testLogger.Events = append(span.testLogger.Events, &Event{
			Type: RequestDone,
			Span: span,
		})
	} else {
		span.testLogger.Events = append(span.testLogger.Events, &Event{
			Type: SpanDone,
			Span: span,
		})
	}
}

func (span *Span) Flush() {
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	span.testLogger.Events = append(span.testLogger.Events, &Event{
		Type: FlushEvent,
		Span: span,
	})
}

func (span *Span) Boring(bool)                  {}
func (span *Span) ID() string                   { return span.testLogger.id }
func (span *Span) SetErrorReporter(func(error)) {}

func (span *Span) Span(ts time.Time, traceSpan trace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	span.lock.Lock()
	defer span.lock.Unlock()
	n := &Span{
		testLogger:    span.testLogger,
		Trace:         traceSpan,
		short:         span.testLogger.setShort(traceSpan, name),
		StartTime:     ts,
		Name:          name,
		Metadata:      make(map[string]interface{}),
		MetadataTypes: make(map[string]xoputil.BaseAttributeType),
		SequenceCode:  spanSequenceCode,
	}
	span.Spans = append(span.Spans, n)
	span.testLogger.Spans = append(span.testLogger.Spans, n)
	span.testLogger.Events = append(span.testLogger.Events, &Event{
		Type: SpanStart,
		Span: n,
	})
	return n
}

func (span *Span) NoPrefill() xopbase.Prefilled {
	return &Prefilled{
		Span: span,
	}
}

func (span *Span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		Builder: Builder{
			Data: make(map[string]interface{}),
			Span: span,
		},
	}
}

func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &Prefilled{
		Data:   p.Data,
		Span:   p.Span,
		kvText: p.kvText,
		Msg:    m,
	}
}

func (p *Prefilled) Line(level xopconst.Level, t time.Time, _ []uintptr) xopbase.Line {
	atomic.StoreInt64(&p.Span.EndTime, t.UnixNano())
	// TODO: stack traces
	line := &Line{
		Builder: Builder{
			Data: make(map[string]interface{}),
			Span: p.Span,
		},
		Level:     level,
		Timestamp: t,
	}
	for k, v := range p.Data {
		line.Data[k] = v
	}
	if len(p.kvText) != 0 {
		line.kvText = make([]string, len(p.kvText), len(p.kvText)+5)
		copy(line.kvText, p.kvText)
	}
	line.Message = p.Msg
	return line
}

func (line *Line) Static(m string) {
	line.Msg(m)
}

func (line *Line) Msg(m string) {
	line.Message += m
	text := line.Span.short + ": " + line.Message
	if len(line.kvText) > 0 {
		text += " " + strings.Join(line.kvText, " ")
		line.kvText = nil
	}
	line.Text = text
	line.send(text)
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

func (line *Line) Template(m string) {
	line.Tmpl = true
	line.Message += m
	used := make(map[string]struct{})
	text := line.Span.short + ": " +
		templateRE.ReplaceAllStringFunc(line.Message, func(k string) string {
			k = k[1 : len(k)-1]
			if v, ok := line.Data[k]; ok {
				used[k] = struct{}{}
				return fmt.Sprint(v)
			}
			return "''"
		})
	for k, v := range line.Data {
		if _, ok := used[k]; !ok {
			text += " " + k + "=" + fmt.Sprint(v)
		}
	}
	line.Text = text
	line.send(text)
}

func (line Line) send(text string) {
	line.Span.testLogger.t.Log(text)
	line.Span.testLogger.lock.Lock()
	defer line.Span.testLogger.lock.Unlock()
	line.Span.lock.Lock()
	defer line.Span.lock.Unlock()
	line.Span.testLogger.Lines = append(line.Span.testLogger.Lines, &line)
	line.Span.testLogger.Events = append(line.Span.testLogger.Events, &Event{
		Type: LineEvent,
		Line: &line,
	})
	line.Span.Lines = append(line.Span.Lines, &line)
}

func (b *Builder) Any(k string, v interface{}) {
	b.Data[k] = v
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%+v", k, v))
}

func (b *Builder) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	b.Data[k.Key()] = v.String()
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%s(%d)", k.Key(), v.String(), v.Int64()))
}

func (b *Builder) Bool(k string, v bool)              { b.Any(k, v) }
func (b *Builder) Duration(k string, v time.Duration) { b.Any(k, v) }
func (b *Builder) Error(k string, v error)            { b.Any(k, v) }
func (b *Builder) Float64(k string, v float64)        { b.Any(k, v) }
func (b *Builder) Int(k string, v int64)              { b.Any(k, v) }
func (b *Builder) Link(k string, v trace.Trace)       { b.Any(k, v) }
func (b *Builder) String(k string, v string)          { b.Any(k, v) }
func (b *Builder) Time(k string, v time.Time)         { b.Any(k, v) }
func (b *Builder) Uint(k string, v uint64)            { b.Any(k, v) }

func (s *Span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeAnyArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeAny
	}
}

func (s *Span) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeBoolArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeBool
	}
}

func (s *Span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeEnumArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeEnum
	}
}

func (s *Span) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeFloat64Array
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeFloat64
	}
}

func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeInt64Array
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeInt64
	}
}

func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeLinkArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeLink
	}
}

func (s *Span) MetadataString(k *xopconst.StringAttribute, v string) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeStringArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeString
	}
}

func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	func() {
		s.testLogger.lock.Lock()
		defer s.testLogger.lock.Unlock()
		s.testLogger.Events = append(s.testLogger.Events, &Event{
			Type: MetadataSet,
			Msg:  k.Key(),
			Span: s,
		})
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	if k.Multiple() {
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), v)
		} else {
			s.Metadata[k.Key()] = []interface{}{v}
			s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeTimeArray
		}
	} else {
		s.Metadata[k.Key()] = v
		s.MetadataTypes[k.Key()] = xoputil.BaseAttributeTypeTime
	}
}
