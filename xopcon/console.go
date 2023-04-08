// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

/*
Package xopcon provides a xopbase.Logger that is meant human consumption.
It does not support replay. Data is omitted to increase readability.
*/
package xopcon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbase/xopbaseutil"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/google/uuid"
	"github.com/muir/list"
)

var (
	_ xopbase.Logger     = &Logger{}
	_ xopbase.Request    = &Span{}
	_ xopbase.Span       = &Span{}
	_ xopbase.Prefilling = &Prefilling{}
	_ xopbase.Prefilled  = &Prefilled{}
	_ xopbase.Line       = &Line{}
)

type Opt func(*Logger)

func WithWriter(w io.Writer) Opt {
	return func(log *Logger) {
		log.out = w
	}
}

func WithRequestCounter(c *xoputil.RequestCounter) Opt {
	return func(log *Logger) {
		log.requestCounter = c
	}
}

func New(opts ...Opt) *Logger {
	log := &Logger{
		out:            os.Stdout,
		id:             "xopcon" + "-" + uuid.New().String(),
		requestCounter: xoputil.NewRequestCounter(),
	}
	for _, opt := range opts {
		opt(log)
	}
	return log
}

type Logger struct {
	lock           sync.Mutex
	out            io.Writer
	traceCount     int
	id             string
	linePrefix     string
	requestCounter *xoputil.RequestCounter
	errorReporter  func(error)
}

type Span struct {
	xopbaseutil.SpanMetadata
	EndTime            int64
	provisionalEndTime int64
	lock               sync.Mutex
	logger             *Logger
	RequestNum         int // sequence of requests with the same traceID
	TraceNum           int // sequence of traces
	Bundle             xoptrace.Bundle
	IsRequest          bool
	Parent             *Span
	Short              string // Tx.y where x is a sequence of requests and y is a sequence of spans within the request
	StartTime          time.Time
	Name               string
	SequenceCode       string
	Ctx                context.Context
	SourceInfo         *xopbase.SourceInfo
}

type Prefilling struct {
	Builder
}

type Builder struct {
	Enums    map[string]*xopat.EnumAttribute
	Data     map[string]interface{}
	DataType map[string]xopbase.DataType
	Span     *Span
	kvText   []string
}

type Prefilled struct {
	Enums    map[string]*xopat.EnumAttribute
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
	AsLink    *xoptrace.Trace
	AsModel   *xopbase.ModelArg
	Stack     []runtime.Frame
}

func (log *Logger) SetPrefix(p string) {
	log.linePrefix = p
}

func (log *Logger) output(s string) {
	s += "\n"
	_, err := log.out.Write([]byte(s))
	if err != nil {
		log.errorReporter(err)
	}
}

// WithLock is provided for thread-safe introspection of the logger
func (log *Logger) WithLock(f func(*Logger) error) error {
	log.lock.Lock()
	defer log.lock.Unlock()
	return f(log)
}

// ID is a required method for xopbase.Logger
func (log *Logger) ID() string { return log.id }

// Buffered is a required method for xopbase.Logger
func (log *Logger) Buffered() bool { return false }

// ReferencesKept is a required method for xopbase.Logger
func (log *Logger) ReferencesKept() bool { return true }

// SetErrorReporter is a required method for xopbase.Logger
func (log *Logger) SetErrorReporter(f func(error)) { log.errorReporter = f }

// Request is a required method for xopbase.Logger
func (log *Logger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, name string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	log.lock.Lock()
	defer log.lock.Unlock()
	traceNum, requestNum, isNew := log.requestCounter.GetNumber(bundle.Trace)
	s := &Span{
		logger:     log,
		IsRequest:  true,
		Bundle:     bundle,
		StartTime:  ts,
		Name:       name,
		Ctx:        ctx,
		SourceInfo: &sourceInfo,
		RequestNum: requestNum,
		TraceNum:   traceNum,
		Short:      fmt.Sprintf("T%d.%d", traceNum, requestNum),
	}
	if isNew {
		s.logger.output("Start request " + s.Short + "=" + bundle.Trace.String() + " " + name)
	}
	return s
}

// Done is a required method for xopbase.Span
func (span *Span) Done(t time.Time, final bool) {
	xoputil.AtomicMaxInt64(&span.EndTime, xoputil.AtomicMaxInt64(&span.provisionalEndTime, t.UnixNano()))
}

// Flush is a required method for xopbase.Request
func (span *Span) Flush() {}

// Final is a required method for xopbase.Request
func (span *Span) Final() {}

// Boring is a required method for xopbase.Span
func (span *Span) Boring(bool) {}

// ID is a required method for xopbase.Span
func (span *Span) ID() string { return span.logger.id }

// ID is a required method for xopbase.Request
func (span *Span) SetErrorReporter(func(error)) {}

// Span is a required method for xopbase.Span
func (span *Span) Span(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	span.logger.lock.Lock()
	defer span.logger.lock.Unlock()
	span.lock.Lock()
	defer span.lock.Unlock()
	n := &Span{
		logger:       span.logger,
		Bundle:       bundle,
		StartTime:    ts,
		Name:         name,
		SequenceCode: spanSequenceCode,
		Ctx:          ctx,
		Parent:       span,
		RequestNum:   span.Parent.RequestNum,
		TraceNum:     span.Parent.TraceNum,
	}
	span.Short = fmt.Sprintf("T%d.%d%s", span.TraceNum, span.RequestNum, span.SequenceCode)
	span.logger.output("Start span " + span.Short + "=" + span.Bundle.Trace.String() + " " + span.Name)
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
			Enums:    make(map[string]*xopat.EnumAttribute),
			Data:     make(map[string]interface{}),
			DataType: make(map[string]xopbase.DataType),
			Span:     span,
		},
	}
}

// PrefillComplete is a required method for xopbase.Prefilling
func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &Prefilled{
		Enums:    p.Enums,
		Data:     p.Data,
		DataType: p.DataType,
		Span:     p.Span,
		kvText:   p.kvText,
		Msg:      m,
	}
}

// Line is a required method for xopbase.Prefilled
func (p *Prefilled) Line(level xopnum.Level, t time.Time, frames []runtime.Frame) xopbase.Line {
	xoputil.AtomicMaxInt64(&p.Span.provisionalEndTime, t.UnixNano())
	line := &Line{
		Builder: Builder{
			Enums:    make(map[string]*xopat.EnumAttribute),
			Data:     make(map[string]interface{}),
			DataType: make(map[string]xopbase.DataType),
			Span:     p.Span,
		},
		Level:     level,
		Timestamp: t,
		Stack:     list.Copy(frames),
	}
	for k, v := range p.Data {
		line.Data[k] = v
		line.DataType[k] = p.DataType[k]
		if e, ok := p.Enums[k]; ok {
			line.Enums[k] = e
		}
	}
	if len(p.kvText) != 0 {
		line.kvText = make([]string, len(p.kvText), len(p.kvText)+5)
		copy(line.kvText, p.kvText)
	}
	line.Message = p.Msg
	return line
}

// Link is a required method for xopbase.Line
func (line *Line) Link(m string, v xoptrace.Trace) {
	line.AsLink = &v
	line.Message += m
	text := line.Span.Short + " LINK:" + line.Message + " " + v.String()
	if len(line.kvText) > 0 {
		text += " " + strings.Join(line.kvText, " ")
		line.kvText = nil
	}
	line.Text = text
	line.send(text)
}

// Model is a required method for xopbase.Line
func (line *Line) Model(m string, v xopbase.ModelArg) {
	line.AsModel = &v
	line.Message += m
	enc, _ := json.Marshal(v.Model)
	text := line.Span.Short + " MODEL:" + line.Message + " " + string(enc)
	if len(line.kvText) > 0 {
		text += " " + strings.Join(line.kvText, " ")
		line.kvText = nil
	}
	line.Text = text
	line.send(text)
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
	line.Span.logger.output(text)
}

func (b *Builder) any(k string, v interface{}, dt xopbase.DataType) {
	b.Data[k] = v
	b.DataType[k] = dt
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%+v", k, v))
}

// Enum is a required method for xopbase.ObjectParts
func (b *Builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	ks := k.Key()
	b.Enums[ks] = k
	b.Data[ks] = v
	b.DataType[ks] = xopbase.EnumDataType
	b.kvText = append(b.kvText, fmt.Sprintf("%s=%s(%d)", ks, v.String(), v.Int64()))
}

// Any is a required method for xopbase.ObjectParts
func (b *Builder) Any(k string, v xopbase.ModelArg) { b.any(k, v, xopbase.AnyDataType) }

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