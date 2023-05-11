// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

/*
Package xopconsole provides a xopbase.Logger that is partially meant for human consumption.
It fully suppors replay without data loss and that requires all details to be
output. Since console loggers exist in situations with other writers, all lines
are prefixed so that lines that did not come from xopconsole can be ignored.
*/
package xopconsole

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
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

const (
	universalPrefix = "xop"
	timeFormat      = "2006-01-02 15:04:05.00000000"
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
	builder
}

type builder struct {
	Builder
	Span *Span
}

type Prefilled struct {
	builder
	Msg string
}

type Line struct {
	builder
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
	_, err := log.out.Write([]byte(log.linePrefix + s))
	if err != nil {
		log.errorReporter(err)
	}
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
	traceNum, requestNum, _ := log.requestCounter.GetNumber(bundle.Trace)
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
	s.Parent = s
	s.logger.output("Start request " + s.Short + "=" + bundle.Trace.String() + " " + name)
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
	n.Short = fmt.Sprintf("T%d.%d%s", n.TraceNum, n.RequestNum, n.SequenceCode)
	span.logger.output("Start span " + n.Short + "=" + bundle.Trace.String() + " " + n.Name)
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
		builder: builder{
			Span: span,
		},
	}
}

// StartPrefill is a required method for xopbase.Span
func (span *Span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		builder: builder{
			Span: span,
		},
	}
}

// PrefillComplete is a required method for xopbase.Prefilling
func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &Prefilled{
		builder: p.builder,
		Msg:     m,
	}
}

// Line is a required method for xopbase.Prefilled
func (p *Prefilled) Line(level xopnum.Level, t time.Time, frames []runtime.Frame) xopbase.Line {
	xoputil.AtomicMaxInt64(&p.Span.provisionalEndTime, t.UnixNano())
	line := &Line{
		builder:   p.builder,
		Level:     level,
		Timestamp: t,
		Stack:     list.Copy(frames),
	}
	line.Message = p.Msg
	return line
}

// Link is a required method for xopbase.Line
func (line *Line) Link(m string, v xoptrace.Trace) {
	line.AsLink = &v
	line.Message += m
	text := line.Span.Short + " LINK:" + line.Message + " " + v.String()
	line.Text = text
	line.send(text)
}

// Model is a required method for xopbase.Line
func (line *Line) Model(m string, v xopbase.ModelArg) {
	line.AsModel = &v
	line.Message += m
	enc, _ := json.Marshal(v.Model)
	text := line.Span.Short + " MODEL:" + line.Message + " " + string(enc)
	line.Text = text
	line.send(text)
}

// Msg is a required method for xopbase.Line
func (line *Line) Msg(m string) {
	line.Message += m
	text := line.Span.Short + ": " + line.Message
	line.Text = text
	line.send(text)
}

var templateRE = regexp.MustCompile(`\{.+?\}`)

// Template is a required method for xopbase.Line
func (line *Line) Template(m string) {
	line.Tmpl = line.Message + m
	line.Message = m
	text := universalPrefix + line.Timestamp.UTC().Format(timeFormat) + "-" + line.Level.String() + "-" + line.Span.Short + ": " + m
	line.Text = text
	line.send(text)
}

func (line Line) send(text string) {
	line.Span.logger.output(text)
}

func (b *Builder) addKey(k string) {
	b.B = append(b.B, ' ')
	b.AddConsoleString(k)
	b.B = append(b.B, '=')
}

func (b *Builder) addType(t string) {
	b.B = append(b.B, '(')
	b.B = append(b.B, []byte(t)...)
	b.B = append(b.B, ')')
}

func (b *Builder) Duration(k string, v time.Duration) {
	b.addKey(k)
	b.B = append(b.B, []byte(v.String())...)
	b.addType(xopbase.DurationDataTypeAbbr)
}

func (b *Builder) Float64(k string, v float64, t xopbase.DataType) {
	b.addKey(k)
	b.AddFloat64(v)
	b.addType(xopbase.DataTypeToString[t])
}

func (b *Builder) String(k string, v string, t xopbase.DataType) {
	b.addKey(k)
	b.AddConsoleString(v)
	if t != xopbase.StringDataType {
		b.addType(xopbase.DataTypeToString[t])
	}
}

func (b *Builder) Int64(k string, v int64, t xopbase.DataType) {
	b.addKey(k)
	b.B = strconv.AppendInt(b.B, v, 64)
	if t != xopbase.IntDataType {
		b.addType(xopbase.DataTypeToString[t])
	}
}

func (b *Builder) Uint64(k string, v uint64, t xopbase.DataType) {
	b.addKey(k)
	b.AddUint64(v)
	b.addType(xopbase.DataTypeToString[t])
}

func (b *Builder) Bool(k string, v bool) {
	b.addKey(k)
	if v {
		b.B = append(b.B, 't')
	} else {
		b.B = append(b.B, 'f')
	}
	b.addType(xopbase.BoolDataTypeAbbr)
}

func (b *Builder) Time(k string, t time.Time) {
	b.addKey(k)
	b.B = t.AppendFormat(b.B, time.RFC3339Nano)
	b.addType(xopbase.TimeDataTypeAbbr)
}

// Enum doesn't need a type indicator
func (b *Builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	b.AppendBytes(k.ConsoleKey())
	b.B = append(b.B, '=')
	b.AddInt64(v.Int64())
	b.B = append(b.B, '/')
	b.AddConsoleString(v.String())
}

func (b *Builder) Any(k string, v xopbase.ModelArg) {
	b.addKey(k)
	v.Encode()
	b.AddConsoleString(strconv.Quote(string(v.Encoded)))
	b.B = append(b.B, '/')
	b.AddConsoleString(v.Encoding.ToString())
}

// MetadataAny is a required method for xopbase.Span
func (s *Span) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	s.SpanMetadata.MetadataAny(k, v)
}

// MetadataBool is a required method for xopbase.Span
func (s *Span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	s.SpanMetadata.MetadataBool(k, v)
}

// MetadataEnum is a required method for xopbase.Span
func (s *Span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	s.SpanMetadata.MetadataEnum(k, v)
}

// MetadataFloat64 is a required method for xopbase.Span
func (s *Span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	s.SpanMetadata.MetadataFloat64(k, v)
}

// MetadataInt64 is a required method for xopbase.Span
func (s *Span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	s.SpanMetadata.MetadataInt64(k, v)
}

// MetadataLink is a required method for xopbase.Span
func (s *Span) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	s.SpanMetadata.MetadataLink(k, v)
}

// MetadataString is a required method for xopbase.Span
func (s *Span) MetadataString(k *xopat.StringAttribute, v string) {
	s.SpanMetadata.MetadataString(k, v)
}

// MetadataTime is a required method for xopbase.Span
func (s *Span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	s.SpanMetadata.MetadataTime(k, v)
}
