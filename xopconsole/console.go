/*
Package xopconsole provides a xopbase.Logger that is partially meant for human consumption.
It fully suppors replay without data loss and that requires all details to be
output. Since console loggers exist in situations with other writers, all lines
are prefixed so that lines that did not come from xopconsole can be ignored.
*/
package xopconsole

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/google/uuid"
	"github.com/muir/list"
)

var _ xopbase.Logger = &Logger{}
var _ xopbase.Request = &Span{}
var _ xopbase.Span = &Span{}
var _ xopbase.Prefilling = &Prefilling{}
var _ xopbase.Prefilled = &Prefilled{}
var _ xopbase.Line = &Line{}

const timeFormat = "2006-01-02 15:04:05.00000000"

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
	AttributeBuilder
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
	VersionNumber      int32
	Request            *Span
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
	PrefillMsg string
}

type Line struct {
	builder
	Stack      []runtime.Frame
	PrefillMsg string
	Timestamp  time.Time
	Level      xopnum.Level
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
	s.AttributeBuilder.Init(s)
	var buf [200]byte
	b := xoputil.JBuilder{
		B: buf[:0],
	}
	b.AppendBytes([]byte("xop Request "))
	b.B = DefaultTimeFormatter(b.B, ts)
	b.AppendBytes([]byte(" Start1 "))
	b.AppendString(bundle.Trace.String())
	b.AppendByte(' ')
	b.AddConsoleString(name)
	b.AppendByte(' ')
	b.AddConsoleString(sourceInfo.Source + "-" + sourceInfo.SourceVersion.String())
	b.AppendByte(' ')
	b.AddConsoleString(sourceInfo.Namespace + "-" + sourceInfo.NamespaceVersion.String())
	if !bundle.Parent.IsZero() {
		b.AppendBytes([]byte(" parent:"))
		if bundle.Parent.GetTraceID() != bundle.Trace.GetTraceID() {
			b.AppendString(bundle.Parent.String())
		} else {
			b.AppendBytes(bundle.Parent.GetSpanID().HexBytes())
		}
	}

	if !bundle.State.IsZero() {
		b.AppendBytes([]byte(" state:"))
		b.AppendBytes(bundle.State.Bytes())
	}
	if !bundle.Baggage.IsZero() {
		b.AppendBytes([]byte(" baggage:"))
		b.AppendBytes(bundle.Baggage.Bytes())
	}
	b.AppendByte('\n')
	_, err := log.out.Write(b.B)
	if err != nil {
		log.errorReporter(err)
	}
	return s
}

// Done is a required method for xopbase.Span
func (span *Span) Done(ts time.Time, final bool) {
	xoputil.AtomicMaxInt64(&span.EndTime, xoputil.AtomicMaxInt64(&span.provisionalEndTime, ts.UnixNano()))
	var buf [200]byte
	b := &Builder{
		JBuilder: xoputil.JBuilder{
			B: buf[:0],
		},
	} // TODO: pull from pool
	b.Init()
	if span.IsRequest {
		b.AppendBytes([]byte("xop Request "))
	} else {
		b.AppendBytes([]byte("xop Span "))
	}
	b.B = DefaultTimeFormatter(b.B, ts)
	b.AppendBytes([]byte(" v"))
	span.VersionNumber++
	b.B = strconv.AppendInt(b.B, int64(span.VersionNumber), 10)
	b.AppendByte(' ')
	b.AppendBytes(span.Bundle.Trace.SpanID().HexBytes())
	span.AttributeBuilder.Append(b, false, false)
	fmt.Println("XXX done with append:", string(b.B))
	b.AppendByte('\n')
	_, err := span.logger.out.Write(b.B)
	if err != nil {
		span.logger.errorReporter(err)
	}
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
	n.AttributeBuilder.Init(span.request)
	var buf [200]byte
	b := xoputil.JBuilder{
		B: buf[:0],
	}
	b.AppendBytes([]byte("xop Span "))
	b.B = DefaultTimeFormatter(b.B, ts)
	b.AppendBytes([]byte(" Start "))
	b.AppendBytes(bundle.Trace.SpanID().HexBytes())
	b.AppendByte(' ')
	b.AppendBytes(span.Bundle.Trace.SpanID().HexBytes())
	b.AppendByte(' ')
	b.AddConsoleString(name)
	b.AppendByte(' ')
	b.AppendString(n.Short)
	b.AppendByte('\n')
	_, err := span.logger.out.Write(b.B)
	if err != nil {
		span.logger.errorReporter(err)
	}
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
		builder:    p.builder,
		PrefillMsg: m,
	}
}

// Line is a required method for xopbase.Prefilled
func (p *Prefilled) Line(level xopnum.Level, t time.Time, frames []runtime.Frame) xopbase.Line {
	xoputil.AtomicMaxInt64(&p.Span.provisionalEndTime, t.UnixNano())
	line := &Line{
		PrefillMsg: p.PrefillMsg,
		builder:    p.builder,
		Level:      level,
		Timestamp:  t,
		Stack:      list.Copy(frames),
	}
	return line
}

// Link is a required method for xopbase.Line
func (line *Line) Link(m string, v xoptrace.Trace) {
	line.send([]byte("LINK:"), m, []byte(v.String()))
}

// Model is a required method for xopbase.Line
func (line *Line) Model(m string, v xopbase.ModelArg) {
	var b Builder
	b.Init()
	b.AnyCommon(v)
	line.send([]byte("MODEL:"), m, b.B)
}

// Msg is a required method for xopbase.Line
func (line *Line) Msg(m string) {
	line.send(nil, line.PrefillMsg+m, nil)
}

// Template is a required method for xopbase.Line
func (line *Line) Template(m string) {
	line.send([]byte("TEMPLATE:"), line.PrefillMsg+m, nil)
}

func (line Line) send(prefix []byte, text string, postfix []byte) {
	b := xoputil.JBuilder{
		B: make([]byte, 0, len(line.B)+len(text)+len(prefix)+len(postfix)+50),
	}
	b.AppendBytes([]byte("xop "))
	b.AppendString(line.Level.String())
	b.AppendByte(' ')
	b.B = DefaultTimeFormatter(b.B, line.Timestamp)
	b.AppendByte(' ')
	b.AppendBytes(line.Span.Bundle.Trace.SpanID().HexBytes())
	b.AppendByte(' ')
	if len(prefix) > 0 {
		b.AppendBytes(prefix)
	}
	b.AddConsoleString(text)
	if len(postfix) > 0 {
		b.AppendByte(' ')
		b.AppendBytes(postfix)
	}
	if len(line.B) > 0 {
		b.AppendBytes(line.B)
	}
	if len(line.Stack) > 0 {
		b.AppendBytes([]byte(" STACK:"))
		for _, frame := range line.Stack {
			b.AppendByte(' ')
			b.AddConsoleString(frame.File)
			b.AppendByte(':')
			b.AddInt64(int64(frame.Line))
		}
	}
	b.AppendByte('\n')
	_, err := line.Span.logger.out.Write(b.B)
	if err != nil {
		line.Span.logger.errorReporter(err)
	}
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
	b.B = strconv.AppendInt(b.B, v, 10)
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
	b.B = append(b.B, ' ')
	b.AppendBytes(k.ConsoleKey())
	b.AttributeEnum(v)
}

func (b *Builder) Any(k string, v xopbase.ModelArg) {
	b.addKey(k)
	b.AnyCommon(v)
}
