// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptest

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/trace"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
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
	requestCount int
	traceNum     int
	spans        map[string]*Span
}

type Span struct {
	lock         sync.Mutex
	testLogger   *TestLogger
	RequestNum   int // sequence of requests with the same traceID
	Bundle       trace.Bundle
	IsRequest    bool
	Parent       *Span
	Spans        []*Span
	RequestLines []*Line
	Lines        []*Line
	Short        string // Tx.y where x is a sequence of requests and y is a sequence of spans within the request
	Metadata     map[string]interface{}
	MetadataType map[string]xopbase.DataType
	metadataSeen map[string]interface{}
	StartTime    time.Time
	EndTime      int64
	Name         string
	SequenceCode string
	Ctx          context.Context
}

type Prefilling struct {
	Builder
}

type Builder struct {
	Data     map[string]interface{}
	DataType map[string]xopbase.DataType
	Span     *Span
	kvText   []string
}

type Prefilled struct {
	Data     map[string]interface{}
	DataType map[string]xopbase.DataType
	Span     *Span
	Msg      string
	kvText   []string
}

type Line struct {
	Builder
	Level     xopnum.Level
	Timestamp time.Time
	Message   string // Prefill text + line text (template evaluated)
	Text      string // Complete text of line including key=value pairs
	Tmpl      string // un-evaluated template
	Stack     []runtime.Frame
}

type Event struct {
	Type EventType
	Line *Line
	Span *Span
	Msg  string
	Done bool
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

// ID is a required method for xopbase.Logger
func (log *TestLogger) ID() string { return log.id }

// Buffered is a required method for xopbase.Logger
func (log *TestLogger) Buffered() bool { return false }

// ReferencesKept is a required method for xopbase.Logger
func (log *TestLogger) ReferencesKept() bool { return true }

// SetErrorReporter is a required method for xopbase.Logger
func (log *TestLogger) SetErrorReporter(func(error)) {}

// Request is a required method for xopbase.Logger
func (log *TestLogger) Request(ctx context.Context, ts time.Time, bundle trace.Bundle, name string) xopbase.Request {
	log.lock.Lock()
	defer log.lock.Unlock()
	s := &Span{
		testLogger:   log,
		IsRequest:    true,
		Bundle:       bundle,
		StartTime:    ts,
		Name:         name,
		Metadata:     make(map[string]interface{}),
		MetadataType: make(map[string]xopbase.DataType),
		metadataSeen: make(map[string]interface{}),
		Ctx:          ctx,
	}
	s.setShortRequest()
	log.Requests = append(log.Requests, s)
	log.Events = append(log.Events, &Event{
		Type: RequestStart,
		Span: s,
	})
	return s
}

// must hold a lock to call setShortRequest
func (span *Span) setShortRequest() {
	ts := span.Bundle.Trace.GetTraceID().String()
	if ti, ok := span.testLogger.traceMap[ts]; ok {
		ti.requestCount++
		span.RequestNum = ti.requestCount
		ti.spans[span.Bundle.Trace.SpanID().String()] = span
		span.Short = fmt.Sprintf("T%d.%d", ti.traceNum, ti.requestCount)
		span.testLogger.t.Log("Start request " + span.Short + "=" + span.Bundle.Trace.String() + " " + span.Name)
		return
	}
	span.testLogger.traceCount++
	span.RequestNum = 1
	span.testLogger.traceMap[ts] = &traceInfo{
		requestCount: 1,
		traceNum:     span.testLogger.traceCount,
		spans: map[string]*Span{
			span.Bundle.Trace.SpanID().String(): span,
		},
	}
	span.Short = fmt.Sprintf("T%d.%d", span.testLogger.traceCount, 1)
	span.testLogger.t.Log("Start request " + span.Short + "=" + span.Bundle.Trace.String() + " " + span.Name)
}

// must hold a lock to call setShortSpan
func (span *Span) setShortSpan() {
	ts := span.Bundle.Trace.GetTraceID().String()
	ti := span.testLogger.traceMap[ts]
	span.RequestNum = span.Parent.RequestNum
	ti.spans[span.Bundle.Trace.SpanID().String()] = span
	span.Short = fmt.Sprintf("T%d.%d%s", ti.traceNum, span.RequestNum, span.SequenceCode)
	span.testLogger.t.Log("Start span " + span.Short + "=" + span.Bundle.Trace.String() + " " + span.Name)
}

// Done is a required method for xopbase.Span
func (span *Span) Done(t time.Time, final bool) {
	atomic.StoreInt64(&span.EndTime, t.UnixNano())
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	if span.IsRequest {
		span.testLogger.Events = append(span.testLogger.Events, &Event{
			Type: RequestDone,
			Span: span,
			Done: final,
		})
	} else {
		span.testLogger.Events = append(span.testLogger.Events, &Event{
			Type: SpanDone,
			Span: span,
			Done: final,
		})
	}
}

// Done is a required method for xopbase.Request
func (span *Span) Flush() {
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	span.testLogger.Events = append(span.testLogger.Events, &Event{
		Type: FlushEvent,
		Span: span,
	})
}

// Final is a required method for xopbase.Request
func (span *Span) Final() {}

// Boring is a required method for xopbase.Span
func (span *Span) Boring(bool) {}

// ID is a required method for xopbase.Span
func (span *Span) ID() string { return span.testLogger.id }

// ID is a required method for xopbase.Request
func (span *Span) SetErrorReporter(func(error)) {}

// Span is a required method for xopbase.Span
func (span *Span) Span(ctx context.Context, ts time.Time, bundle trace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	span.testLogger.lock.Lock()
	defer span.testLogger.lock.Unlock()
	span.lock.Lock()
	defer span.lock.Unlock()
	n := &Span{
		testLogger:   span.testLogger,
		Bundle:       bundle,
		StartTime:    ts,
		Name:         name,
		Metadata:     make(map[string]interface{}),
		MetadataType: make(map[string]xopbase.DataType),
		metadataSeen: make(map[string]interface{}),
		SequenceCode: spanSequenceCode,
		Ctx:          ctx,
		Parent:       span,
	}
	n.setShortSpan()
	span.Spans = append(span.Spans, n)
	span.testLogger.Spans = append(span.testLogger.Spans, n)
	span.testLogger.Events = append(span.testLogger.Events, &Event{
		Type: SpanStart,
		Span: n,
	})
	return n
}

// ParentRequest returns the span that is the request-level parent
// of the current span. If the current span is a request, it returns
// the current span.
func (span *Span) ParentRequest() *Span {
	for {
		if span.IsRequest {
			return span
		}
		span = span.Parent
	}
}

// NoPrefill is a required method for xopbase.Span
func (span *Span) NoPrefill() xopbase.Prefilled {
	return &Prefilled{
		Span: span,
	}
}

// StartPrefill is a required method for xopbase.Span
func (span *Span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		Builder: Builder{
			Data:     make(map[string]interface{}),
			DataType: make(map[string]xopbase.DataType),
			Span:     span,
		},
	}
}

// PrefillComplete is a required method for xopbase.Prefilling
func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &Prefilled{
		Data:     p.Data,
		DataType: p.DataType,
		Span:     p.Span,
		kvText:   p.kvText,
		Msg:      m,
	}
}

// Line is a required method for xopbase.Prefilled
func (p *Prefilled) Line(level xopnum.Level, t time.Time, pc []uintptr) xopbase.Line {
	atomic.StoreInt64(&p.Span.EndTime, t.UnixNano())
	line := &Line{
		Builder: Builder{
			Data:     make(map[string]interface{}),
			DataType: make(map[string]xopbase.DataType),
			Span:     p.Span,
		},
		Level:     level,
		Timestamp: t,
	}
	if len(pc) > 0 {
		frames := runtime.CallersFrames(pc)
		stack := make([]runtime.Frame, 0, len(pc))
		for {
			frame, more := frames.Next()
			if !strings.Contains(frame.File, "runtime/") {
				break
			}
			stack = append(stack, frame)
			if !more {
				break
			}
		}
		line.Stack = stack
	}
	for k, v := range p.Data {
		line.Data[k] = v
		line.DataType[k] = p.DataType[k]
	}
	if len(p.kvText) != 0 {
		line.kvText = make([]string, len(p.kvText), len(p.kvText)+5)
		copy(line.kvText, p.kvText)
	}
	line.Message = p.Msg
	return line
}

// Static is a required method for xopbase.Line
func (line *Line) Static(m string) {
	line.Msg(m)
}

// Msg is a required method for xopbase.Line
func (line *Line) Msg(m string) {
	line.Message += m
	text := line.Span.Short + ": " + line.Message
	if len(line.kvText) > 0 {
		text += " " + strings.Join(line.kvText, " ")
		line.kvText = nil
	}
	line.Text = text
	line.send(text)
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

// Template is a required method for xopbase.Line
func (line *Line) Template(m string) {
	line.Tmpl = line.Message + m
	used := make(map[string]struct{})
	msg := templateRE.ReplaceAllStringFunc(line.Tmpl, func(k string) string {
		k = k[1 : len(k)-1]
		if v, ok := line.Data[k]; ok {
			used[k] = struct{}{}
			return fmt.Sprint(v)
		}
		return "''"
	})
	line.Message = msg
	text := line.Span.Short + ": " + msg
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

// TemplateOrMessage returns the line template (if set) or the template
// message (Msg) if there is no template
func (line *Line) TemplateOrMessage() string {
	if line.Tmpl != "" {
		return line.Tmpl
	}
	return line.Message
}

func (b *Builder) any(k string, v interface{}, dt xopbase.DataType) {
	b.Data[k] = v
	b.DataType[k] = dt
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%+v", k, v))
}

// Enum is a required method for xopbase.ObjectParts
func (b *Builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	ks := k.Key()
	b.Data[ks] = v.String()
	b.DataType[ks] = xopbase.EnumDataType
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%s(%d)", ks, v.String(), v.Int64()))
}

// Link is a required method for xopbase.ObjectParts
func (b *Builder) Link(k string, v trace.Trace) {
	b.Data[k] = v
	b.DataType[k] = xopbase.LinkDataType
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%+v", k, v.String()))
}

// Any is a required method for xopbase.ObjectParts
func (b *Builder) Any(k string, v interface{}) { b.any(k, v, xopbase.AnyDataType) }

// Bool is a required method for xopbase.ObjectParts
func (b *Builder) Bool(k string, v bool) { b.any(k, v, xopbase.BoolDataType) }

// Duration is a required method for xopbase.ObjectParts
func (b *Builder) Duration(k string, v time.Duration) { b.any(k, v, xopbase.DurationDataType) }

// Time is a required method for xopbase.ObjectParts
func (b *Builder) Time(k string, v time.Time) { b.any(k, v, xopbase.TimeDataType) }

// Float64 is a required method for xopbase.ObjectParts
func (b *Builder) Float64(k string, v float64, dt xopbase.DataType) { b.any(k, v, dt) }

// Int64 is a required method for xopbase.ObjectParts
func (b *Builder) Int64(k string, v int64, dt xopbase.DataType) { b.any(k, v, dt) }

// String is a required method for xopbase.ObjectParts
func (b *Builder) String(k string, v string, dt xopbase.DataType) { b.any(k, v, dt) }

// Uint64 is a required method for xopbase.ObjectParts
func (b *Builder) Uint64(k string, v uint64, dt xopbase.DataType) { b.any(k, v, dt) }

// MetadataAny is a required method for xopbase.Span
func (s *Span) MetadataAny(k *xopat.AnyAttribute, v interface{}) {
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
		value := v
		if k.Distinct() {
			var key string
			enc, err := json.Marshal(v)
			if err != nil {
				key = fmt.Sprintf("%+v", v)
			} else {
				key = string(enc)
			}
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[string]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.AnyArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.AnyDataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataBool is a required method for xopbase.Span
func (s *Span) MetadataBool(k *xopat.BoolAttribute, v bool) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[bool]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[bool]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.BoolArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.BoolDataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataEnum is a required method for xopbase.Span
func (s *Span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
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
		value := v.String()
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[string]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.EnumArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.EnumDataType
		}
		s.Metadata[k.Key()] = v.String()
	}
}

// MetadataFloat64 is a required method for xopbase.Span
func (s *Span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[float64]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[float64]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.Float64ArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.Float64DataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataInt64 is a required method for xopbase.Span
func (s *Span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[int64]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[int64]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.Int64ArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.Int64DataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataLink is a required method for xopbase.Span
func (s *Span) MetadataLink(k *xopat.LinkAttribute, v trace.Trace) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[trace.Trace]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[trace.Trace]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.LinkArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.LinkDataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataString is a required method for xopbase.Span
func (s *Span) MetadataString(k *xopat.StringAttribute, v string) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[string]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.StringArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.StringDataType
		}
		s.Metadata[k.Key()] = v
	}
}

// MetadataTime is a required method for xopbase.Span
func (s *Span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
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
		value := v
		if k.Distinct() {
			key := value
			seenRaw, ok := s.metadataSeen[k.Key()]
			if !ok {
				seen := make(map[time.Time]struct{})
				s.metadataSeen[k.Key()] = seen
				seen[key] = struct{}{}
			} else {
				seen := seenRaw.(map[time.Time]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if p, ok := s.Metadata[k.Key()]; ok {
			s.Metadata[k.Key()] = append(p.([]interface{}), value)
		} else {
			s.Metadata[k.Key()] = []interface{}{value}
			s.MetadataType[k.Key()] = xopbase.TimeArrayDataType
		}
	} else {
		if _, ok := s.Metadata[k.Key()]; ok {
			if k.Locked() {
				return
			}
		} else {
			s.MetadataType[k.Key()] = xopbase.TimeDataType
		}
		s.Metadata[k.Key()] = v
	}
}
