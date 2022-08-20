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
	if log.tagOption == DefaultTagOption {
		if log.perRequestBufferLimit > 0 {
			log.tagOption = OmitTagOption
		} else {
			log.tagOption = FullIDTagOption
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

func (logger *Logger) Request(ts time.Time, trace trace.Bundle, name string) xopbase.Request {
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
	if logger.tagOption == TraceSequenceNumberTagOption {
		request.idNum = atomic.AddInt64(&logger.requestCount, 1)
	}
	request.attributes.Init(&request.span)
	request.request = request
	if logger.perRequestBufferLimit != 0 {
		request.maintainBuffer()
	}
	request.allSpans = make([]*span, 1, 64)
	request.allSpans[0] = &request.span

	rq := request.span.builder()
	rq.AppendBytes([]byte(`{"type":"request","trace.id":"`))
	rq.AppendString(trace.Trace.TraceIDString())
	rq.AppendBytes([]byte(`","span.id":`))
	rq.AddSafeString(trace.Trace.SpanIDString())
	if !trace.TraceParent.IsZero() {
		rq.AppendBytes([]byte(`,"trace.parent":`))
		rq.AppendString(trace.TraceParent.HeaderString())
	}
	if !trace.State.IsZero() {
		rq.AppendBytes([]byte(`,"trace.state":`))
		rq.AppendString(trace.State.String())
	}
	if !trace.Baggage.IsZero() {
		rq.AppendBytes([]byte(`,"trace.baggage":`))
		rq.AppendString(trace.Baggage.String())
	}
	rq.String("span.name", name)
	rq.Time("ts", ts)
	rq.AppendBytes([]byte{'}', '\n'})
	if logger.perRequestBufferLimit != 0 {
		request.completedBuilders <- rq
	} else {
		_, err := request.writer.Write(rq.B)
		if err != nil {
			request.errorFunc(err)
		}
		rq.reclaimMemory()
	}

	return request
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
	func() {
		r.allSpansLock.Lock()
		defer r.allSpansLock.Unlock()
		for _, span := range r.allSpans {
			span.FlushAttributes()
		}
	}()
	if r.logger.perRequestBufferLimit != 0 {
		// TODO: improve this a bit by using a WaitGroup or something
		r.flushRequest <- struct{}{}
		<-r.flushComplete
	} else {
		r.writer.Flush()
	}
}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }

func (s *span) Span(ts time.Time, trace trace.Bundle, name string) xopbase.Span {
	n := &span{
		logger:    s.logger,
		writer:    s.writer,
		trace:     trace,
		name:      name,
		request:   s.request,
		startTime: ts,
		endTime:   ts.UnixNano(),
	}
	n.attributes.Init(n)
	func() {
		s.request.allSpansLock.Lock()
		defer s.request.allSpansLock.Unlock()
		s.request.allSpans = append(s.request.allSpans, n)
	}()

	rq := s.builder()
	rq.AppendBytes([]byte(`{"type":"span","span.ver":0,"span.id":`))
	rq.AddSafeString(trace.Trace.SpanIDString())
	rq.String("span.name", name)
	rq.Time("ts", ts)
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

	return n
}

func (s *span) FlushAttributes() {
	rq := s.builder()
	s.serializationCount++
	if s == &s.request.span {
		rq.AppendBytes([]byte(`{"type":"request","request.ver":`)) // }
	} else {
		rq.AppendBytes([]byte(`{"type":"span","span.ver":`)) // }
	}
	rq.AddInt64(int64(s.serializationCount))
	rq.AppendBytes([]byte(`,"span.id":`))
	rq.AddSafeString(s.trace.Trace.SpanIDString())
	if s.request.logger.durationKey != nil {
		rq.AppendBytes(s.request.logger.durationKey)
		rq.AddDuration(time.Duration(s.endTime - s.startTime.UnixNano()))
	}
	s.attributes.Append(&rq.JBuilder)
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

func (s *span) Done(t time.Time) { atomic.StoreInt64(&s.endTime, t.UnixNano()) }
func (s *span) Boring(bool)      {} // TODO
func (s *span) ID() string       { return s.logger.id.String() }

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
			B: make([]byte, len(m)), // alloc-per-prefill
		}
		msgBuffer.AddStringBody(m)
		prefilled.preEncodedMsg = msgBuffer.B
	}
	return prefilled
}

func (p *prefilled) Line(level xopconst.Level, t time.Time, pc []uintptr) xopbase.Line {
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
	l.Int("lvl", int64(level))
	l.Time("ts", t)
	if len(pc) > 0 {
		frames := runtime.CallersFrames(pc)
		l.AppendBytes([]byte(`,"stack":[`))
		for {
			frame, more := frames.Next()
			if !strings.Contains(frame.File, "runtime/") {
				break
			}
			l.Comma()
			l.AppendByte('"')
			l.AddStringBody(frame.File)
			l.AppendByte(':')
			l.AddInt64(int64(frame.Line))
			l.AppendByte('"')
			if !more {
				break
			}
		}
		l.AppendByte(']')
	}
	switch l.span.logger.tagOption {
	case SpanIDTagOption:
		l.Comma()
		l.AppendBytes([]byte(`"span.id":`))
		l.AddSafeString(l.span.trace.Trace.SpanIDString())
	case FullIDTagOption:
		l.Comma()
		l.AppendBytes([]byte(`"trace.header":`))
		l.AddSafeString(l.span.trace.Trace.HeaderString())
	case TraceIDTagOption:
		l.Comma()
		l.AppendBytes([]byte(`"trace.id":`))
		l.AddSafeString(l.span.trace.Trace.TraceIDString())
	case TraceSequenceNumberTagOption:
		l.AddKey("trace_num")
		l.AddInt64(l.span.request.idNum)
	case OmitTagOption:
		// yay!
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

func (b *builder) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	b.startAttributes()
	b.AddUncheckedKey(k.Key()) // TODO: check attribute keys at registration time
	b.AddEnum(v)
}

func (b *builder) AddEnum(v xopconst.Enum) {
	// TODO: send dictionary and numbers
	b.AddInt64(v.Int64())
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
	b.AppendString(v.HeaderString())
	b.AppendBytes([]byte(`"}`))
}

func (b *builder) Bool(k string, v bool) {
	b.startAttributes()
	b.AddKey(k)
	b.AddBool(v)
}

func (b *builder) Int(k string, v int64) {
	b.startAttributes()
	b.AddKey(k)
	b.AddInt64(v)
}

func (b *builder) Uint(k string, v uint64) {
	b.startAttributes()
	b.AddKey(k)
	b.AddUint64(v)
}

func (b *builder) String(k string, v string) {
	b.startAttributes()
	b.AddKey(k)
	b.AddString(v)
}

func (b *builder) Float64(k string, v float64) {
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

func (s *span) MetadataAny(k *xopconst.AnyAttribute, v interface{}) { s.attributes.MetadataAny(k, v) }
func (s *span) MetadataBool(k *xopconst.BoolAttribute, v bool)      { s.attributes.MetadataBool(k, v) }
func (s *span) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	s.attributes.MetadataEnum(k, v)
}
func (s *span) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	s.attributes.MetadataFloat64(k, v)
}
func (s *span) MetadataInt64(k *xopconst.Int64Attribute, v int64) { s.attributes.MetadataInt64(k, v) }
func (s *span) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	s.attributes.MetadataLink(k, v)
}
func (s *span) MetadataString(k *xopconst.StringAttribute, v string) {
	s.attributes.MetadataString(k, v)
}
func (s *span) MetadataTime(k *xopconst.TimeAttribute, v time.Time) { s.attributes.MetadataTime(k, v) }

// end
