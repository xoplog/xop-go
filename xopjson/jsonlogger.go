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

func New(w xopbytes.BytesWriter, opts ...Option) *Logger {
	log := &Logger{
		writer:      w,
		id:          uuid.New(),
		timeDivisor: time.Millisecond,
	}
	for _, f := range opts {
		f(log)
	}
	if log.tagOption == DefaultTagOption {
		if log.perRequestBufferLimit > 0 {
			log.tagOption = OmitTagOption
		} else {
			log.tagOption = FullIDTagOption
		}
	}
	return log
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
		trace:  span,
		name:   name,
	}
	if l.tagOption == TraceSequenceNumberTagOption {
		s.idNum = atomic.AddInt64(&l.requestCount, 1)
	}
	s.attributes.Reset()
	s.requestSpan = s
	return s
}

func (s *Span) Span(span trace.Bundle, name string) xopbase.Span {
	n := &Span{
		logger: s.logger,
		writer: s.writer,
		trace:  span,
		name:   name,
	}
	n.attributes.Reset()
	n.requestSpan = s
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
		data: make([]byte, len(p.Builder.dataBuffer.B)),
		span: p.Builder.span,
	}
	copy(prefilled.data, p.Builder.dataBuffer.B)
	if len(m) > 0 {
		msgBuffer := xoputil.JBuilder{
			B: make([]byte, len(m)), // alloc-per-prefill
		}
		msgBuffer.StringBody(m)
		prefilled.preEncodedMsg = msgBuffer.B
	}
	return prefilled
}

func (p *Prefilled) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
	l := &Line{
		Builder:              p.span.builder(),
		level:                level,
		timestamp:            t,
		prefillMsgPreEncoded: p.preEncodedMsg,
	}
	l.dataBuffer.AppendByte('{') // }
	if len(p.data) != 0 {
		l.dataBuffer.AppendBytes(p.data)
	}
	l.dataBuffer.Comma()
	l.dataBuffer.AppendByte('{')
	l.Int("level", int64(level))
	l.Time("time", t)
	if len(pc) > 0 {
		frames := runtime.CallersFrames(pc)
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
	switch l.span.logger.tagOption {
	case SpanIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"span_id":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.SpanIDString())
		l.dataBuffer.AppendByte('"')
	case FullIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"trace_header":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.HeaderString())
		l.dataBuffer.AppendByte('"')
	case TraceIDTagOption:
		l.dataBuffer.Comma()
		l.dataBuffer.AppendBytes([]byte(`"trace_id":"`))
		l.dataBuffer.AppendString(l.span.trace.Trace.TraceIDString())
		l.dataBuffer.AppendByte('"')
	case TraceSequenceNumberTagOption:
		l.dataBuffer.Key("trace_num")
		l.dataBuffer.Int64(l.span.requestSpan.idNum)
	case OmitTagOption:
		// yay!
	}
	l.dataBuffer.AppendByte('}')
	return l
}

func (l *Line) Static(m string) {
	l.Msg(m) // TODO
}

func (l *Line) Msg(m string) {
	l.dataBuffer.Comma()
	l.dataBuffer.AppendBytes([]byte(`"msg":"`))
	if len(l.prefillMsgPreEncoded) != 0 {
		l.dataBuffer.AppendBytes(l.prefillMsgPreEncoded)
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

func (b *Builder) Any(k string, v interface{}) {
	b.dataBuffer.Key(k)
	before := len(b.dataBuffer.B)
	err := b.encoder.Encode(v)
	if err != nil {
		b.dataBuffer.B = b.dataBuffer.B[:before]
		b.span.errorFunc(err)
		b.Error("encode:"+k, err)
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if b.dataBuffer.B[len(b.dataBuffer.B)-1] == '\n' {
			b.dataBuffer.B = b.dataBuffer.B[:len(b.dataBuffer.B)-1]
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
	case timeTimeFormat:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.B = t.AppendFormat(b.dataBuffer.B, b.span.logger.timeFormat)
		b.dataBuffer.AppendByte('"')
	case epochTime:
		b.dataBuffer.Key(k)
		b.dataBuffer.Int64(t.UnixNano() / int64(b.span.logger.timeDivisor))
	case epochQuoted:
		b.dataBuffer.Key(k)
		b.dataBuffer.AppendByte('"')
		b.dataBuffer.Int64(t.UnixNano() / int64(b.span.logger.timeDivisor))
		b.dataBuffer.AppendByte('"')
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

func (b *Builder) Float64(k string, v float64) {
	b.dataBuffer.Key(k)
	b.dataBuffer.Float64(v)
}

func (b *Builder) Duration(k string, v time.Duration) {
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
func (s *Span) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	s.attributes.MetadataFloat64(k, v)
}
func (s *Span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.attributes.MetadataInt64(k, v) }
func (s *Span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.attributes.MetadataLink(k, v)
}
func (s *Span) MetadataStr(k *xopconst.StrAttribute, v string)      { s.attributes.MetadataStr(k, v) }
func (s *Span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.attributes.MetadataTime(k, v) }

// end
