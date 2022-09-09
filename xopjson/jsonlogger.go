// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopat"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopnum"
	"github.com/muir/xop-go/xoputil"

	"github.com/google/uuid"
	"github.com/phuslu/fasttime"
)

const (
	maxBufferToKeep = 1024 * 10
	minBuffer       = 1024
	lineChanDepth   = 32
)

func New(w xopbytes.BytesWriter, opts ...Option) *Logger {
	log := &Logger{
		writer:       w,
		id:           uuid.New(),
		timeDivisor:  time.Microsecond,
		closeRequest: make(chan struct{}),
	}
	prealloc := xoputil.NewPrealloc(log.preallocatedKeys[:])
	for _, f := range opts {
		f(log, prealloc)
	}
	if log.tagOption == 0 {
		if log.perRequestBufferLimit > 0 {
			log.tagOption = SpanIDTagOption
		} else {
			log.tagOption = SpanIDTagOption | TraceIDTagOption
		}
	}
	return log
}

func (logger *Logger) ID() string           { return logger.id.String() }
func (logger *Logger) Buffered() bool       { return logger.writer.Buffered() }
func (logger *Logger) ReferencesKept() bool { return false }

func (logger *Logger) Close() {
	logger.writer.Close()
}

func (logger *Logger) Request(_ context.Context, ts time.Time, trace trace.Bundle, name string) xopbase.Request {
	request := &request{
		span: span{
			logger:    logger,
			writer:    logger.writer.Request(trace),
			trace:     trace,
			name:      name,
			startTime: ts,
			endTime:   ts.UnixNano(),
		},
	}
	if logger.tagOption&TraceNumberTagOption != 0 {
		request.idNum = atomic.AddInt64(&logger.requestCount, 1)
	}
	request.attributes.Init(&request.span)
	request.request = request
	if logger.perRequestBufferLimit != 0 {
		request.maintainBuffer()
	}
	request.span.setSpanIDPrefill()

	if logger.spanStarts {
		rq := request.span.builder()
		rq.AppendBytes([]byte(`{"type":"request","request.ver":0`))
		request.span.addRequestStartData(rq)
		rq.AppendBytes([]byte{'}', '\n'})
		request.serializationCount++
		if logger.perRequestBufferLimit != 0 {
			request.completedBuilders <- rq
		} else {
			_, err := request.writer.Write(rq.B)
			if err != nil {
				request.errorFunc(err)
			}
			rq.reclaimMemory()
		}
	}

	return request
}

func (s *span) addRequestStartData(rq *builder) {
	rq.AddSafeKey("trace.id")
	rq.AddSafeString(s.trace.Trace.TraceIDString())
	rq.AddSafeKey("span.id")
	rq.AddSafeString(s.trace.Trace.SpanIDString())
	if !s.trace.TraceParent.IsZero() {
		rq.AddSafeKey("trace.parent")
		rq.AppendString(s.trace.TraceParent.String())
	}
	if !s.trace.State.IsZero() {
		rq.AddSafeKey("trace.state")
		rq.AppendString(s.trace.State.String())
	}
	if !s.trace.Baggage.IsZero() {
		rq.AddSafeKey("trace.baggage")
		rq.AppendString(s.trace.Baggage.String())
	}
	rq.AddSafeKey("span.name")
	rq.AddString(s.name)
	rq.Time("ts", s.startTime)
}

func (r *request) maintainBuffer() {
	r.flushRequest = make(chan struct{})
	r.flushComplete = make(chan struct{})
	r.completedLines = make(chan *line, lineChanDepth)
	r.completedBuilders = make(chan *builder, lineChanDepth)
	r.writeBuffer = make([]byte, 0, r.logger.perRequestBufferLimit/16)
	handleBuilder := func(builder *builder) {
		if len(builder.B)+len(r.writeBuffer) > r.logger.perRequestBufferLimit {
			r.flushBuffer()
		}
		if len(builder.B) > r.logger.perRequestBufferLimit {
			// TODO: split into multiple writes
		}
		r.writeBuffer = append(r.writeBuffer, builder.B...)
		builder.reclaimMemory()
	}
	handleLine := func(line *line) {
		if len(line.B)+len(r.writeBuffer) > r.logger.perRequestBufferLimit {
			r.flushBuffer()
		}
		if len(line.B) > r.logger.perRequestBufferLimit {
			// TODO: split into multiple writes
		}
		r.writeBuffer = append(r.writeBuffer, line.B...)
		line.reclaimMemory()
	}
	handleFlush := func() {
		// drains lines & builders before flush!
	Flush:
		for {
			select {
			case builder := <-r.completedBuilders:
				handleBuilder(builder)
			case line := <-r.completedLines:
				handleLine(line)
			default:
				break Flush
			}
		}
		r.flushBuffer()
		r.flushComplete <- struct{}{}
	}
	go func() {
		for {
			select {
			case builder := <-r.completedBuilders:
				handleBuilder(builder)
			case line := <-r.completedLines:
				handleLine(line)
			case <-r.flushRequest:
				handleFlush()
			case <-r.logger.closeRequest:
				// drain everything else first
			Request:
				for {
					select {
					case builder := <-r.completedBuilders:
						handleBuilder(builder)
					case line := <-r.completedLines:
						handleLine(line)
					case <-r.flushRequest:
						handleFlush()
					default:
						break Request
					}
				}
				r.flushBuffer()
				// TODO: have logger wait for requests to complete
				// WaitGroup?
				return
			}
		}
	}()
}

func (r *request) flushBuffer() {
	// TODO: trigger spans to write their stuff
	if len(r.writeBuffer) == 0 {
		return
	}
	_, err := r.writer.Write(r.writeBuffer)
	if err != nil {
		r.errorFunc(err)
	}
	r.writer.Flush()
	r.writeBuffer = r.writeBuffer[:0]
}

func (r *request) Flush() {
	if r.logger.perRequestBufferLimit != 0 {
		// TODO: improve this a bit by using a WaitGroup or something
		r.flushRequest <- struct{}{}
		<-r.flushComplete
	} else {
		r.writer.Flush()
	}
}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }

func (s *span) Span(_ context.Context, ts time.Time, trace trace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	n := &span{
		logger:       s.logger,
		writer:       s.writer,
		trace:        trace,
		name:         name,
		request:      s.request,
		startTime:    ts,
		endTime:      ts.UnixNano(),
		sequenceCode: spanSequenceCode,
	}
	n.attributes.Init(n)
	n.setSpanIDPrefill()

	if s.logger.spanStarts {
		rq := s.builder()
		rq.AppendBytes([]byte(`{"type":"span","span.ver":0,"span.id":`))
		rq.AddSafeString(trace.Trace.SpanIDString())
		n.spanStartData(rq)
		rq.AppendBytes([]byte{'}', '\n'})
		n.serializationCount++
		if s.request.logger.perRequestBufferLimit != 0 {
			s.request.completedBuilders <- rq
		} else {
			_, err := s.request.writer.Write(rq.B)
			if err != nil {
				s.request.errorFunc(err)
			}
			rq.reclaimMemory()
		}
	}

	return n
}

func (s *span) spanStartData(rq *builder) {
	rq.String("span.name", s.name)
	rq.Time("ts", s.startTime)
}

func (s *span) setSpanIDPrefill() {
	b := xoputil.JBuilder{
		B: s.spanIDBuffer[:0],
	}
	s.identifySpan(&b)
	s.spanIDPrebuilt = b
}

func (s *span) identifySpan(b *xoputil.JBuilder) {
	if s.logger.tagOption&SpanIDTagOption != 0 {
		b.AddSafeKey("span.id")
		b.AddSafeString(s.trace.Trace.SpanIDString())
	}
	if s.logger.tagOption&TraceIDTagOption != 0 {
		b.AddSafeKey("trace.id")
		b.AddSafeString(s.trace.Trace.TraceIDString())
	}
	if s.logger.tagOption&TraceNumberTagOption != 0 {
		b.AddSafeKey("trace.num")
		b.AddInt64(s.request.idNum)
	}
	if s.logger.tagOption&SpanSequenceTagOption != 0 {
		b.AddSafeKey("span.ctx")
		b.AddSafeString(s.request.sequenceCode)
	}
}

func (s *span) FlushAttributes() {
	rq := s.builder()
	if s == &s.request.span {
		rq.AppendBytes([]byte(`{"type":"request"`)) // }
		if !s.logger.spanStarts || !s.logger.spanChangesOnly {
			s.addRequestStartData(rq)
		} else {
			s.identifySpan(&rq.JBuilder)
		}
		rq.AddSafeKey("request.ver")
	} else {
		rq.AppendBytes([]byte(`{"type":"span"`)) // }
		if !s.logger.spanStarts || !s.logger.spanChangesOnly {
			s.spanStartData(rq)
		}
		s.identifySpan(&rq.JBuilder)
		rq.AddSafeKey("span.ver")
	}
	rq.AddInt64(int64(s.serializationCount))
	s.serializationCount++
	if s.request.logger.durationKey != nil {
		rq.AppendBytes(s.request.logger.durationKey)
		rq.AddDuration(time.Duration(s.endTime - s.startTime.UnixNano()))
	}
	s.attributes.Append(&rq.JBuilder, s.logger.spanChangesOnly)
	// {
	rq.AppendBytes([]byte{'}', '\n'})
	if s.request.logger.perRequestBufferLimit != 0 {
		s.request.completedBuilders <- rq
	} else {
		_, err := s.request.writer.Write(rq.B)
		if err != nil {
			s.request.errorFunc(err)
		}
		rq.reclaimMemory()
	}
}

func (s *span) Done(t time.Time, _ bool) {
	atomic.StoreInt64(&s.endTime, t.UnixNano())
	s.FlushAttributes()
}

func (s *span) Boring(bool) {} // TODO
func (s *span) ID() string  { return s.logger.id.String() }

func (s *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		span: s,
	}
}

func (b *builder) reset(s *span) {
	b.span = s
	b.B = b.B[:0]
	b.attributesWanted = false
	b.attributesStarted = false
}

func (s *span) builder() *builder {
	bRaw := s.request.logger.builderPool.Get()
	var b *builder
	if bRaw != nil {
		b = bRaw.(*builder)
		b.reset(s)
	} else {
		b = &builder{
			span: s,
			JBuilder: xoputil.JBuilder{
				B:        make([]byte, 0, minBuffer),
				FastKeys: s.logger.fastKeys,
			},
		}
		b.encoder = json.NewEncoder(&b.JBuilder)
		b.encoder.SetEscapeHTML(false)
	}
	return b
}

func (s *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: s.builder(),
	}
}

func (p *prefilling) PrefillComplete(m string) xopbase.Prefilled {
	prefilled := &prefilled{
		data: make([]byte, len(p.builder.B)),
		span: p.builder.span,
	}
	copy(prefilled.data, p.builder.B)
	if len(m) > 0 {
		msgBuffer := xoputil.JBuilder{
			B: make([]byte, 0, len(m)), // alloc-per-prefill
		}
		msgBuffer.AddStringBody(m)
		prefilled.preEncodedMsg = msgBuffer.B
	}
	return prefilled
}

func (p *prefilled) Line(level xopnum.Level, t time.Time, pc []uintptr) xopbase.Line {
	atomic.StoreInt64(&p.span.endTime, t.UnixNano())
	var l *line
	lRaw := p.span.request.logger.linePool.Get()
	if lRaw != nil {
		l = lRaw.(*line)
		l.builder.reset(p.span)
		l.level = level
		l.timestamp = t
		l.prefillMsgPreEncoded = p.preEncodedMsg
	} else {
		l = &line{
			builder:              p.span.builder(),
			level:                level,
			timestamp:            t,
			prefillMsgPreEncoded: p.preEncodedMsg,
		}
	}
	l.AppendByte('{') // }
	l.Int64("lvl", int64(level), 0)
	l.Time("ts", t)
	if len(pc) > 0 {
		// TODO: debug this!
		frames := runtime.CallersFrames(pc)
		l.AppendBytes([]byte(`,"stack":[`))
		for {
			frame, more := frames.Next()
			if strings.Contains(frame.File, "runtime/") {
				break
			}
			l.Comma()
			l.AppendByte('"')
			filename := frame.File
			if l.span.request.logger.stackLineRewrite != nil {
				filename = l.span.request.logger.stackLineRewrite(filename)
			}
			l.AddStringBody(filename)
			l.AppendByte(':')
			l.AddInt64(int64(frame.Line))
			l.AppendByte('"')
			if !more {
				break
			}
		}
		l.AppendByte(']')
	}
	if len(p.span.spanIDPrebuilt.B) != 0 {
		l.Comma()
		l.AppendBytes(p.span.spanIDPrebuilt.B)
	}
	l.attributesWanted = p.span.logger.attributesObject
	if len(p.data) != 0 {
		if l.attributesWanted {
			l.attributesStarted = true
			l.AppendBytes([]byte(`,"attributes":{`)) // }
		} else {
			l.AppendByte(',')
		}
		l.AppendBytes(p.data)
	}
	return l
}

func (l *line) Msg(m string) {
	if l.attributesStarted {
		l.AppendByte( /*{*/ '}')
	}
	l.AppendBytes([]byte(`,"msg":"`))
	if len(l.prefillMsgPreEncoded) != 0 {
		l.AppendBytes(l.prefillMsgPreEncoded)
	}
	l.AddStringBody(m)
	l.AppendBytes([]byte{
		'"', // {
		'}',
		'\n',
	})
	if l.span.logger.perRequestBufferLimit != 0 {
		l.span.request.completedLines <- l
	} else {
		_, err := l.span.writer.Write(l.B)
		if err != nil {
			l.span.request.errorFunc(err)
		}
		l.reclaimMemory()
	}
}

func (l *line) Static(m string) {
	l.Msg(m) // TODO
}

func (b *builder) reclaimMemory() {
	if len(b.B) > maxBufferToKeep {
		return
	}
	b.span.request.logger.builderPool.Put(b)
}

func (l *line) reclaimMemory() {
	if len(l.B) > maxBufferToKeep {
		return
	}
	l.span.request.logger.linePool.Put(l)
}

func (l *line) Template(m string) {
	if l.attributesStarted {
		// {
		l.AppendByte('}')
	}
	l.AppendString(`,"fmt":"tmpl","msg":`)
	l.AddString(m)
	// {
	l.AppendBytes([]byte{'}', '\n'})
	_, err := l.span.writer.Write(l.B)
	if err != nil {
		l.span.request.errorFunc(err)
	}
	l.reclaimMemory()
}

func (b *builder) startAttributes() {
	if b.attributesWanted && !b.attributesStarted {
		b.attributesStarted = true
		b.AppendBytes([]byte(`,"attributes":{`)) // }
	}
}

func (b *builder) Any(k string, v interface{}) {
	b.startAttributes()
	b.AddKey(k)
	b.AddAny(v)
}

func (b *builder) AddAny(v interface{}) {
	before := len(b.B)
	err := b.encoder.Encode(v)
	if err != nil {
		b.B = b.B[:before]
		b.span.request.errorFunc(err)
		b.Error("encode", err)
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if b.B[len(b.B)-1] == '\n' {
			b.B = b.B[:len(b.B)-1]
		}
	}
}

func (b *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	b.startAttributes()
	b.AddSafeKey(k.Key()) // TODO: check attribute keys at registration time
	b.AddEnum(v)
}

func (b *builder) AddEnum(v xopat.Enum) {
	// TODO: send dictionary and numbers
	b.AddString(v.String())
}

func (b *builder) Time(k string, t time.Time) {
	b.startAttributes()
	b.AddKey(k)
	b.AddTime(t)
}

func (b *builder) AddTime(t time.Time) {
	switch b.span.logger.timeOption {
	case strftimeTime:
		b.AppendByte('"')
		b.B = fasttime.AppendStrftime(b.B, b.span.logger.timeFormat, t)
		b.AppendByte('"')
	case timeTimeFormat:
		b.AppendByte('"')
		b.B = t.AppendFormat(b.B, b.span.logger.timeFormat)
		b.AppendByte('"')
	case epochTime:
		b.AddInt64(t.UnixNano() / int64(b.span.logger.timeDivisor))
	case epochQuoted:
		b.AppendByte('"')
		b.AddInt64(t.UnixNano() / int64(b.span.logger.timeDivisor))
		b.AppendByte('"')
	}
}

func (b *builder) Link(k string, v trace.Trace) {
	b.startAttributes()
	// TODO: is this the right format for links?
	b.AddKey(k)
	b.AddLink(v)
}

func (b *builder) AddLink(v trace.Trace) {
	b.AppendBytes([]byte(`{"xop.link":"`))
	b.AppendString(v.String())
	b.AppendBytes([]byte(`"}`))
}

func (b *builder) Bool(k string, v bool) {
	b.startAttributes()
	b.AddKey(k)
	b.AddBool(v)
}

func (b *builder) Int64(k string, v int64, _ xopbase.DataType) {
	b.startAttributes()
	b.AddKey(k)
	b.AddInt64(v)
}

func (b *builder) Uint64(k string, v uint64, _ xopbase.DataType) {
	b.startAttributes()
	b.AddKey(k)
	b.AddUint64(v)
}

func (b *builder) String(k string, v string) {
	b.startAttributes()
	b.AddKey(k)
	b.AddString(v)
}

func (b *builder) Float64(k string, v float64, _ xopbase.DataType) {
	b.startAttributes()
	b.AddKey(k)
	b.AddFloat64(v)
}

func (b *builder) Duration(k string, v time.Duration) {
	b.startAttributes()
	b.AddKey(k)
	b.AddDuration(v)
}

func (b *builder) AddDuration(v time.Duration) {
	switch b.span.logger.durationFormat {
	case AsNanos:
		b.AddInt64(int64(v / time.Nanosecond))
	case AsMicros:
		b.AddInt64(int64(v / time.Microsecond))
	case AsMillis:
		b.AddInt64(int64(v / time.Millisecond))
	case AsSeconds:
		b.AddInt64(int64(v / time.Second))
	case AsString:
		fallthrough
	default:
		b.AddSafeString(v.String())
	}
}

// TODO: allow custom formats
func (b *builder) Error(k string, v error) {
	b.startAttributes()
	b.AddKey(k)
	b.AddError(v)
}

func (b *builder) AddError(v error) {
	b.AddString(v.Error())
}

func (s *span) MetadataAny(k *xopat.AnyAttribute, v interface{})  { s.attributes.MetadataAny(k, v) }
func (s *span) MetadataBool(k *xopat.BoolAttribute, v bool)       { s.attributes.MetadataBool(k, v) }
func (s *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) { s.attributes.MetadataEnum(k, v) }
func (s *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	s.attributes.MetadataFloat64(k, v)
}
func (s *span) MetadataInt64(k *xopat.Int64Attribute, v int64)     { s.attributes.MetadataInt64(k, v) }
func (s *span) MetadataLink(k *xopat.LinkAttribute, v trace.Trace) { s.attributes.MetadataLink(k, v) }
func (s *span) MetadataString(k *xopat.StringAttribute, v string)  { s.attributes.MetadataString(k, v) }
func (s *span) MetadataTime(k *xopat.TimeAttribute, v time.Time)   { s.attributes.MetadataTime(k, v) }

// end
