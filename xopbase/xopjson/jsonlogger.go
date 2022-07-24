// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file
package xopjson

import (
	"sync/atomic"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopbase"
	"github.com/muir/xoplog/xopconst"
	"github.com/muir/xoplog/xoputil"
)

const maxBufferToKeep = 1024 * 10

var (
	_ xopbase.Logger  = &Logger{}
	_ xopbase.Request = &Span{}
	_ xopbase.Span    = &Span{}
	_ xopbase.Line    = &Line{}
)

type Option func(*Logger)

type timeOption int

const (
	noTime timeOption = iota
	strftimeTime
	timeTime
	epochTime
)

type DurationOption int

const (
	AsNanos   DurationOption = iota // int64(duration)
	AsMillis                        // int64(duration / time.Milliscond)
	AsSeconds                       // int64(duration / time.Second)
	AsString                        // duration.String()
)

type AsynchronousWriter interface {
	Write([]byte) (int, error)
	Flush() error
	Close() error
	Buffered() bool
}

type Logger struct {
	writer         AsynchronousWriter
	timeOption     timeOption
	timeFormat     string
	framesAtLevel  map[xopconst.Level]int
	withGoroutine  bool
	fastKeys       bool
	durationFormat DurationOption
	errorFunc      func(error)
}

type Request struct {
	*Span
}

type prefill struct {
	data []byte
	msg  string
}

type Span struct {
	attributes xoputil.AttributeBuilder
	trace      trace.Bundle
	logger     *Logger
	prefill    atomic.Value
}

type Line struct {
	dataBuffer xoputil.JBuilder
	level      xopconst.Level
	timestamp  time.Time
	span       *Span
	prefillLen int
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

func WithCallersAtLevel(logLevel xopconst.Level, framesWanted int) Option {
	return func(l *Logger) {
		l.framesAtLevel[logLevel] = framesWanted
	}
}

func WithGoroutineID(b bool) Option {
	return func(l *Logger) {
		l.withGoroutine = b
	}
}

func New(w AsynchronousWriter, opts ...Option) *Logger {
	logger := &Logger{
		writer:        w,
		framesAtLevel: make(map[xopconst.Level]int),
	}
	for _, f := range opts {
		f(logger)
	}
	return logger
}

func (l *Logger) Buffered() bool                            { return l.writer.Buffered() }
func (l *Logger) ReferencesKept() bool                      { return false }
func (l *Logger) StackFramesWanted() map[xopconst.Level]int { return l.framesAtLevel }
func (l *Logger) SetErrorReporter(reporter func(error))     { l.errorFunc = reporter }

func (l *Logger) Close() {
	err := l.writer.Close()
	if err != nil {
		l.errorFunc(err)
	}
}

func (l *Logger) Request(span trace.Bundle, name string) xopbase.Request {
	s := &Span{
		logger: l,
	}
	s.attributes.Reset()
	return s
}

func (s *Span) Flush() {
	s.logger.writer.Flush()
}

func (s *Span) Boring(bool) {} // TODO

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	return s.logger.Request(span, name)
}

func (s *Span) getPrefill() *prefill {
	p := s.prefill.Load()
	if p == nil {
		return nil
	}
	return p.(*prefill)
}

func (s *Span) Line(level xopconst.Level, t time.Time) xopbase.Line {
	l := &Line{
		level:     level,
		timestamp: t,
		span:      s,
	}
	l.dataBuffer.FastKeys = s.logger.fastKeys
	l.getPrefill()
	return l
}

func (l *Line) Recycle(level xopconst.Level, t time.Time) {
	l.level = level
	l.timestamp = t
	l.dataBuffer.Reset()
	l.getPrefill()
}

func (l *Line) getPrefill() {
	l.dataBuffer.Byte('{') // }
	prefill := l.span.getPrefill()
	l.prefillLen = len(prefill.data)
	if prefill != nil {
		l.dataBuffer.Append(prefill.data)
	}
}

func (l *Line) SetAsPrefill(m string) {
	skip := 1 + l.prefillLen
	prefill := prefill{
		msg:  m,
		data: make([]byte, len(l.dataBuffer.B)-skip),
	}
	copy(prefill.data, l.dataBuffer.B[skip:])
	l.span.prefill.Store(prefill)
	// this Line will not be recycled so destory its buffers
	l.reclaimMemory()
}

func (l *Line) Static(m string) {
	l.Msg(m) // TODO
}

func (l *Line) Msg(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.Append([]byte(`"msg":`))
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.Byte('}')
	_, err := l.span.logger.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.logger.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) reclaimMemory() {
	if len(l.dataBuffer.B) > maxBufferToKeep {
		l.dataBuffer = xoputil.JBuilder{FastKeys: l.span.logger.fastKeys}
	}
}

func (l *Line) Template(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendString(`"xop":"template","msg":`)
	l.dataBuffer.String(m)
	// {
	l.dataBuffer.Byte('}')
	_, err := l.span.logger.writer.Write(l.dataBuffer.B)
	if err != nil {
		l.span.logger.errorFunc(err)
	}
	l.reclaimMemory()
}

func (l *Line) Any(k string, v interface{}) {
	// XXX
}

func (l *Line) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	// XXX
}

func (l *Line) Time(k string, v time.Time) { // XXX
}

func (l *Line) Link(k string, v trace.Trace) { // XXX
}

func (l *Line) Bool(k string, v bool) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Bool(v)
}

func (l *Line) Int(k string, v int64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Int64(v)
}

func (l *Line) Uint(k string, v uint64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Uint64(v)
}

func (l *Line) Str(k string, v string) {
	l.dataBuffer.Key(k)
	l.dataBuffer.String(v)
}

func (l *Line) Number(k string, v float64) {
	l.dataBuffer.Key(k)
	l.dataBuffer.Float64(v)
}

func (l *Line) Duration(k string, v time.Duration) {
	l.dataBuffer.Key(k)
	switch l.span.logger.durationFormat {
	case AsNanos:
		l.dataBuffer.Int64(int64(v / time.Nanosecond))
	case AsMillis:
		l.dataBuffer.Int64(int64(v / time.Millisecond))
	case AsSeconds:
		l.dataBuffer.Int64(int64(v / time.Second))
	case AsString:
		fallthrough
	default:
		l.dataBuffer.UncheckedString(v.String())
	}
}

// TODO: allow custom formats
func (l *Line) Error(k string, v error) {
	l.dataBuffer.Key(k)
	l.dataBuffer.String(v.Error())
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
