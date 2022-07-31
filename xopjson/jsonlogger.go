// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"

	"github.com/google/uuid"
	"github.com/phuslu/fasttime"
)

const (
	maxBufferToKeep = 1024 * 10
	minBuffer       = 1024
)

var (
	_ xopbase.Logger     = &Logger{}
	_ xopbase.Request    = &Span{}
	_ xopbase.Span       = &Span{}
	_ xopbase.Line       = &Line{}
	_ xopbase.Prefilling = &Prefilling{}
	_ xopbase.Prefilled  = &Prefilled{}
)

type Option func(*Logger)

type timeOption int

const (
	epochTime timeOption = iota
	strftimeTime
	timeTime
	unixNano
)

type DurationOption int

const (
	AsNanos   DurationOption = iota // int64(duration)
	AsMillis                        // int64(duration / time.Milliscond)
	AsSeconds                       // int64(duration / time.Second)
	AsString                        // duration.String()
)

type Logger struct {
	writer         xopbytes.BytesWriter
	timeOption     timeOption
	timeFormat     string
	withGoroutine  bool
	fastKeys       bool
	durationFormat DurationOption
	id             uuid.UUID
}

type Request struct {
	*Span
}

type Span struct {
	attributes xoputil.AttributeBuilder
	writer     xopbytes.BytesRequest
	trace      trace.Bundle
	logger     *Logger
	prefill    atomic.Value
	errorFunc  func(error)
}

type Prefilling struct {
	Builder
}

type Prefilled struct {
	data          []byte
	preencodedMsg []byte
	span          *Span
}

type Line struct {
	Builder
	level     xopconst.Level
	timestamp time.Time
}

type Builder struct {
	dataBuffer xoputil.JBuilder
	encoder    *json.Encoder
	span       *Span
}

func WithUncheckedKeys(b bool) Option {
	return func(l *Logger) {
		l.fastKeys = b
	}
}

// TODO: allow custom error formats

// WithStrftime adds a timestamp to each log line.  See
// https://github.com/phuslu/fasttime for the supported
// formats.
func WithStrftime(format string) Option {
	return func(l *Logger) {
		l.timeOption = strftimeTime
		l.timeFormat = format
	}
}

func WithDuration(durationFormat DurationOption) Option {
	return func(l *Logger) {
		l.durationFormat = durationFormat
	}
}

func WithGoroutineID(b bool) Option {
	return func(l *Logger) {
		l.withGoroutine = b
	}
}

func New(w xopbytes.BytesWriter, opts ...Option) *Logger {
	logger := &Logger{
		writer:           w,
		id:               uuid.New(),
		framesAtLevelMap: make(map[xopconst.Level]int),
	}
	for _, f := range opts {
		f(logger)
	}
	return logger
}

func (l *Logger) ID() string           { return l.id.String() }
func (l *Logger) Buffered() bool       { return l.writer.Buffered() }
func (l *Logger) ReferencesKept() bool { return false }

func (l *Logger) Close() {
	l.writer.Close()
}

func (l *Logger) Request(span trace.Bundle, name string) xopbase.Request {
	s := &Span{
		logger: l,
		writer: l.writer.Request(span),
	}
	s.attributes.Reset()
	return s
}

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	n := &Span{
		logger: s.logger,
		writer: s.writer,
	}
	n.attributes.Reset()
	return n
}

func (s *Span) Flush() {
	s.writer.Flush()
}

func (s *Span) Boring(bool)                           {} // TODO
func (s *Span) ID() string                            { return s.logger.id.String() }
func (s *Span) SetErrorReporter(reporter func(error)) { s.errorFunc = reporter }

func (s *Span) NoPrefill() xopbase.Prefilled {
	return &Prefilled{
		span: s,
	}
}

func (s *Span) builder() Builder {
	b := Builder{
		span: s,
		dataBuffer: xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: s.logger.fastKeys,
		},
	}
	b.encoder = json.NewEncoder(&b.dataBuffer)
	return b
}

func (s *Span) StartPrefill() xopbase.Prefilling {
	return &Prefilling{
		Builder: s.builder(),
	}
}

func (p *Prefilling) PrefillComplete(m string) xopbase.Prefilled {
	prefilled := &Prefilled{
		data: make([]byte, len(p.Builder.databuffer.B)),
		msg:  m,
		span: p.Builder.span,
	}
	copy(prefilled.data, p.Builder.databuffer.B)
	return prefilled
}

func (p *Prefilled) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	l := &Line{
		Builder:      p.span.builder(),
		level:        level,
		timestamp:    t,
		prefilledMsg: p.msg,
	}
	l.dataBuffer.AppendByte('{') // }
	if len(p.data) != 0 {
		l.dataBuffer.AppendBytes(prefill.data)
	}
	l.dataBuffer.Comma()
	l.dataBuffer.AppendByte('{')
	l.Int("level", int64(level))
	l.Time("time", t)
	if len(pc) > 0 {
		n := l.span.logger.framesAtLevel[level]
		if n > len(pc) {
			n = len(pc)
		}
		frames := runtime.CallersFrames(pc[:n])
		l.dataBuffer.AppendBytes([]byte(`"stack":[`))
		for {
			frame, more := frames.Next()
			if !strings.Contains(frame.File, "runtime/") {
				break
			}
			l.dataBuffer.Comma()
			l.dataBuffer.AppendByte('"')
			l.dataBuffer.StringBody(frame.File)
			l.dataBuffer.AppendByte(':')
			l.dataBuffer.Int64(int64(frame.Line))
			l.dataBuffer.AppendByte('"')
			if !more {
				break
			}
		}
		l.dataBuffer.AppendByte(']')
	}
	l.dataBuffer.AppendByte('}')
}

func (l *Line) Static(m string) {
	l.Msg(m) // TODO
}

func (l *Line) Msg(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendBytes([]byte(`"msg":"`))
	if len(l.prefillMsgPreencoded) != 0 {
		l.dataBuffer.AppendBytes(l.prefillMsgPreencoded)
	}
	l.dataBuffer.StringBody(m)
	// {
	l.dataBuffer.AppendBytes([]byte{'"', '}'})
	_, err := l.span.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) reclaimMemory() { // XXX re-connect and have pool of Lines & Buffers
	if len(l.dataBuffer.B) > maxBufferToKeep {
		l.dataBuffer = xoputil.JBuilder{
			B:        make([]byte, 0, minBuffer),
			FastKeys: l.span.logger.fastKeys,
		}
		l.encoder = json.NewEncoder(&l.dataBuffer)
	}
}

func (l *Line) Template(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendString(`"xop":"template","msg":`)
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.AppendByte('}')
	_, err := l.span.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) Any(k string, v interface{}) {
	l.dataBuffer.Key(k)
	before := len(l.dataBuffer.B)
	err := l.encoder.Encode(v)
	if err != nil {
		l.dataBuffer.B = l.dataBuffer.B[:before]
		l.span.errorFunc(err)
		l.Error("encode:"+k, err)
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if l.dataBuffer.B[len(l.dataBuffer.B)-1] == '\n' {
			l.dataBuffer.B = l.dataBuffer.B[:len(l.dataBuffer.B)-1]
		}
	}
}

func (b *Builder) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	// TODO: send dictionary and numbers
	b.Int(k.Key(), v.Int64())
}

func (b *Builder) Time(k string, t time.Time) {
	switch b.span.logger.timeOption {
	case strftimeTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.B = fasttime.AppendStrftime(b.dataBuffer.B, b.span.logger.timeFormat, t)
		b.dataBuffer.AppendByte('"')
	case timeTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.B = t.AppendFormat(b.dataBuffer.B, b.span.logger.timeFormat)
		b.dataBuffer.AppendByte('"')
	case epochTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.Float64(float64(t.UnixNano()) / 1000000000.0) // TODO good enough?
	case unixNano:
		b.dataBuffer.Key(k)
		b.dataBuffer.Int64(t.UnixNano())
	}
}

func (b *Builder) Link(k string, v trace.Trace) {
	// TODO: is this the right format for links?
	b.dataBuffer.Key(k)
	b.dataBuffer.AppendBytes([]byte(`{"xop.link":"`))
	b.dataBuffer.AppendString(v.HeaderString())
	b.dataBuffer.AppendBytes([]byte(`"}`))
}

func (b *Builder) Bool(k string, v bool) {
	b.dataBuffer.Key(k)
	b.dataBuffer.Bool(v)
}

func (b *Builder) Int(k string, v int64) {
	b.dataBuffer.Key(k)
	b.dataBuffer.Int64(v)
}

func (b *Builder) Uint(k string, v uint64) {
	b.dataBuffer.Key(k)
	b.dataBuffer.Uint64(v)
}

func (b *Builder) Str(k string, v string) {
	b.dataBuffer.Key(k)
	b.dataBuffer.String(v)
}

func (b *Builder) Number(k string, v float64) {
	b.dataBuffer.Key(k)
	b.dataBuffer.Float64(v)
}

func (l *Builder) Duration(k string, v time.Duration) {
	b.dataBuffer.Key(k)
	switch b.span.logger.durationFormat {
	case AsNanos:
		b.dataBuffer.Int64(int64(v / time.Nanosecond))
	case AsMillis:
		b.dataBuffer.Int64(int64(v / time.Millisecond))
	case AsSeconds:
		b.dataBuffer.Int64(int64(v / time.Second))
	case AsString:
		fallthrough
	default:
		b.dataBuffer.UncheckedString(v.String())
	}
}

// TODO: allow custom formats
func (b *Builder) Error(k string, v error) {
	b.dataBuffer.Key(k)
	b.dataBuffer.String(v.Error())
}

func (s *Span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.attributes.MetadataAny(k, v) }
func (s *Span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.attributes.MetadataBool(k, v) }
func (s *Span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.attributes.MetadataEnum(k, v)
}
func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.attributes.MetadataInt64(k, v) }
func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.attributes.MetadataLink(k, v)
}
func (s *Span) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	s.attributes.MetadataNumber(k, v)
}
func (s *Span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.attributes.MetadataStr(k, v) }
func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.attributes.MetadataTime(k, v) }

// end
