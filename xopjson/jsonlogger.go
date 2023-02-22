// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/google/uuid"
)

const (
	maxBufferToKeep = 1024 * 10
	minBuffer       = 1024
	lineChanDepth   = 32
)

func New(w xopbytes.BytesWriter, opts ...Option) *Logger {
	log := &Logger{
		writer:        w,
		id:            uuid.New(),
		timeFormatter: defaultTimeFormatter,
	}
	prealloc := xoputil.NewPrealloc(log.preallocatedKeys[:])
	for _, f := range opts {
		f(log, prealloc)
	}
	if log.tagOption == 0 {
		if w.Buffered() {
			log.tagOption = SpanIDTagOption
		} else {
			log.tagOption = SpanIDTagOption | TraceIDTagOption
		}
	}
	log.activeRequests.Add(1)
	return log
}

func (logger *Logger) ID() string           { return logger.id.String() }
func (logger *Logger) Buffered() bool       { return logger.writer.Buffered() }
func (logger *Logger) ReferencesKept() bool { return false }

func (logger *Logger) Request(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string, s xopbase.SourceInfo) xopbase.Request {
	// XXX SourceInfo
	request := &request{
		span: span{
			logger:    logger,
			bundle:    bundle,
			name:      name,
			startTime: ts,
			endTime:   ts.UnixNano(),
			isRequest: true,
		},
		sourceInfo: s,
	}
	if logger.tagOption&TraceNumberTagOption != 0 {
		request.idNum = atomic.AddInt64(&logger.requestCount, 1)
	}
	request.attributes.Init(&request.span)
	request.request = request
	request.span.setSpanIDPrefill()
	request.writer = logger.writer.Request(request)

	if logger.spanStarts {
		rq := request.span.builder()
		rq.AppendBytes([]byte(`{"type":"request","span.ver":0`))
		request.span.addRequestStartData(rq)
		rq.AppendBytes([]byte{'}', '\n'})
		request.serializationCount++
		err := request.writer.Span(request, rq)
		if err != nil {
			request.errorFunc(err)
		}
	}

	return request
}

func (s *span) addRequestStartData(rq *builder) {
	rq.AddSafeKey("trace.id")
	rq.AddSafeString(s.bundle.Trace.TraceID().String())
	rq.AddSafeKey("span.id")
	rq.AddSafeString(s.bundle.Trace.SpanID().String())
	if !s.bundle.Parent.TraceID().IsZero() {
		rq.AddSafeKey("trace.parent")
		rq.AddSafeString(s.bundle.Parent.String())
	}
	if !s.bundle.State.IsZero() {
		rq.AddSafeKey("trace.state")
		rq.AddSafeString(s.bundle.State.String())
	}
	if !s.bundle.Baggage.IsZero() {
		rq.AddSafeKey("trace.baggage")
		rq.AddSafeString(s.bundle.Baggage.String())
	}
	rq.AddSafeKey("span.name")
	rq.AddString(s.name)
	rq.Time("ts", s.startTime)
	rq.AddSafeKey("source")
	rq.AppendByte('"')
	rq.AddStringBody(s.request.sourceInfo.Source)
	rq.AppendByte(' ')
	rq.AddStringBody(s.request.sourceInfo.SourceVersion.String())
	rq.AppendByte('"')
	rq.AddSafeKey("ns")
	rq.AppendByte('"')
	rq.AddStringBody(s.request.sourceInfo.Namespace)
	rq.AppendByte(' ')
	rq.AddStringBody(s.request.sourceInfo.NamespaceVersion.String())
	rq.AppendByte('"')
}

func (r *request) Flush() {
	r.writer.Flush()
}

func (r *request) Final() {
	r.writer.ReclaimMemory()
}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }
func (r *request) GetErrorCount() int32                  { return atomic.LoadInt32(&r.errorCount) }
func (r *request) GetAlertCount() int32                  { return atomic.LoadInt32(&r.alertCount) }

func (s *span) Span(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	n := &span{
		logger:       s.logger,
		writer:       s.writer,
		bundle:       bundle,
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
		rq.AddSafeString(bundle.Trace.SpanID().String())
		n.spanStartData(rq)
		rq.AppendBytes([]byte{'}', '\n'})
		n.serializationCount++
		err := s.request.writer.Span(s, rq)
		if err != nil {
			s.request.errorFunc(err)
		}
	}

	return n
}

func (s *span) spanStartData(rq *builder) {
	rq.stringKV("span.name", s.name)
	rq.Time("ts", s.startTime)
	rq.stringKV("span.parent_span", s.bundle.Parent.SpanID().String())
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
		b.AddSafeString(s.bundle.Trace.SpanID().String())
	}
	if s.logger.tagOption&TraceIDTagOption != 0 {
		b.AddSafeKey("trace.id")
		b.AddSafeString(s.bundle.Trace.TraceID().String())
	}
	if s.logger.tagOption&TraceNumberTagOption != 0 {
		b.AddSafeKey("trace.num")
		b.AddInt64(s.request.idNum)
	}
	if s.logger.tagOption&SpanSequenceTagOption != 0 {
		b.AddSafeKey("span.seq")
		b.AddSafeString(s.sequenceCode)
	}
}

func (s *span) flushAttributes() {
	rq := s.builder()
	if s == &s.request.span {
		rq.AppendBytes([]byte(`{"type":"request"`)) // }
		if !s.logger.spanStarts || !s.logger.spanChangesOnly {
			s.addRequestStartData(rq)
		} else {
			s.identifySpan(&rq.JBuilder)
		}
		rq.AddSafeKey("span.ver")
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
	err := s.request.writer.Span(s, rq)
	if err != nil {
		s.request.errorFunc(err)
	}
}

func (s *span) Done(t time.Time, _ bool) {
	atomic.StoreInt64(&s.endTime, t.UnixNano())
	s.flushAttributes()
}

func (s *span) Boring(bool)                {}
func (s *span) ID() string                 { return s.logger.id.String() }
func (s *span) GetBundle() xoptrace.Bundle { return s.bundle }
func (s *span) GetStartTime() time.Time    { return s.startTime }
func (s *span) GetEndTimeNano() int64      { return s.endTime }
func (s *span) IsRequest() bool            { return s.isRequest }

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
	if level >= xopnum.ErrorLevel {
		if level >= xopnum.AlertLevel {
			_ = atomic.AddInt32(&p.span.request.alertCount, 1)
		} else {
			_ = atomic.AddInt32(&p.span.request.errorCount, 1)
		}
	}
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
	l.done()
}

func (l *line) done() {
	err := l.span.writer.Line(l)
	if err != nil {
		l.span.request.errorFunc(err)
	}
}

func (l *line) Model(k string, v xopbase.ModelArg) {
	if l.attributesStarted {
		l.AppendByte( /*{*/ '}')
	}
	v.Encode()
	l.AppendBytes([]byte(`,"type":"model","model":`))
	if v.Encoding == xopproto.Encoding_JSON {
		l.AppendBytes(v.Encoded)
	} else {
		l.AddString(string(v.Encoded))
		l.AppendBytes([]byte(`,"encoding":`))
		l.AddSafeString(v.Encoding.String())
	}
	l.AppendBytes([]byte(`,"modelType":`))
	l.AddString(v.TypeName)
	l.AppendBytes([]byte(`,"msg":"`))
	if len(l.prefillMsgPreEncoded) != 0 {
		l.AppendBytes(l.prefillMsgPreEncoded)
	}
	l.AddStringBody(k)
	l.AppendBytes([]byte{
		'"', // {
		'}',
		'\n',
	})
	l.done()
}

func (l *line) Link(k string, v xoptrace.Trace) {
	if l.attributesStarted {
		l.AppendByte( /*{*/ '}')
	}
	l.AppendBytes([]byte(`,"type":"link","link":"`))
	l.AppendString(v.String())
	l.AppendBytes([]byte(`","msg":"`))
	if len(l.prefillMsgPreEncoded) != 0 {
		l.AppendBytes(l.prefillMsgPreEncoded)
	}
	l.AddStringBody(k)
	l.AppendBytes([]byte{
		'"', // {
		'}',
		'\n',
	})
	l.done()
}

/* XXX
func (b *builder) Link(k string, v xoptrace.Trace) {
	b.startAttributes()
	b.AddKey(k)
	b.AddLink(v)
}
*/

func (b *builder) AddLink(v xoptrace.Trace) {
	b.AppendBytes([]byte(`{"xop.link":"`))
	b.AppendString(v.String())
	b.AppendBytes([]byte(`"}`))
}

func (b *builder) ReclaimMemory() {
	if len(b.B) > maxBufferToKeep {
		return
	}
	b.span.request.logger.builderPool.Put(b)
}

func (b *builder) AsBytes() []byte            { return b.B }
func (l *line) GetSpanID() xoptrace.HexBytes8 { return l.span.bundle.Trace.GetSpanID() }
func (l *line) GetLevel() xopnum.Level        { return l.level }
func (l *line) GetTime() time.Time            { return l.timestamp }

func (l *line) ReclaimMemory() {
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
	err := l.span.writer.Line(l)
	if err != nil {
		l.span.request.errorFunc(err)
	}
}

func (b *builder) startAttributes() {
	if b.attributesWanted && !b.attributesStarted {
		b.attributesStarted = true
		b.AppendBytes([]byte(`,"attributes":{`)) // }
	}
}

func (b *builder) Any(k string, v xopbase.ModelArg) {
	b.startAttributes()
	b.AddKey(k)
	v.Encode()
	b.AppendBytes([]byte(`{"type":"model","model":`)) // }
	if v.Encoding == xopproto.Encoding_JSON {
		b.AppendBytes(v.Encoded)
	} else {
		b.AddString(string(v.Encoded))
		b.AppendBytes([]byte(`,"encoding":`))
		b.AddSafeString(v.Encoding.String())
	}
	b.AppendBytes([]byte(`,"modelType":`))
	b.AddString(v.TypeName)
	b.AppendByte( /* { */ '}')
}

func (b *builder) AddAny(v interface{}) {
	before := len(b.B)
	err := b.encoder.Encode(v)
	if err != nil {
		b.B = b.B[:before]
		b.span.request.errorFunc(err)
		b.stringKV("encode", err.Error())
	} else {
		// remove \n added by json.Encoder.Encode.  So helpful!
		if b.B[len(b.B)-1] == '\n' {
			b.B = b.B[:len(b.B)-1]
		}
	}
}

func (b *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	b.startAttributes()
	b.Comma()
	b.AppendString(k.JSONKey().String())
	b.AddEnum(v)
}

func (b *builder) AddEnum(v xopat.Enum) {
	b.AddString(v.String())
}

func (b *builder) Time(k string, t time.Time) {
	b.startAttributes()
	b.AddKey(k)
	b.AddTime(t)
}

func (b *builder) AddTime(t time.Time) {
	b.B = b.span.logger.timeFormatter(b.B, t)
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

func (b *builder) stringKV(k string, v string) {
	b.startAttributes()
	b.AddKey(k)
	b.AddString(v)
}

func (b *builder) String(k string, v string, _ xopbase.DataType) {
	b.stringKV(k, v)
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

func (s *span) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	s.attributes.MetadataAny(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	s.attributes.MetadataBool(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	s.attributes.MetadataEnum(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
	s.logger.writer.DefineEnum(k, v)
}

func (s *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	s.attributes.MetadataFloat64(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	s.attributes.MetadataInt64(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	s.attributes.MetadataLink(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataString(k *xopat.StringAttribute, v string) {
	s.attributes.MetadataString(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

func (s *span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	s.attributes.MetadataTime(k, v)
	s.logger.writer.DefineAttribute(&k.Attribute)
}

// end
