// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"context"
	"encoding/json"
	"reflect"
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

func New(w xopbytes.BytesWriter) *Logger {
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

func (logger *Logger) Request(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string) xopbase.Request {
	request := &request{
		span: span{
			logger:    logger,
			bundle:    bundle,
			endTime:   ts.UnixNano(),
			isRequest: true,
			protoSpan: xopproto.Span{
				Name:      name,
				StartTime: ts.UnixNano(),
			},
		},
	}
	if logger.tagOption&TraceNumberTagOption != 0 {
		request.idNum = atomic.AddInt64(&logger.requestCount, 1)
	}
	request.request = request
	request.writer = logger.writer.Request(request)

	return request
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
		logger:  s.logger,
		writer:  s.writer,
		bundle:  bundle,
		request: s.request,
		endTime: ts.UnixNano(),
		protoSpan: xopproto.Span{
			Name:         name,
			StartTime:    ts.UnixNano(),
			SequenceCode: spanSequenceCode,
		},
	}
	return n
}

func (s *span) Done(t time.Time, _ bool) {
	atomic.StoreInt64(&s.endTime, t.UnixNano())
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
}

func (s *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: s.builder(),
	}
}

func (p *prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return prefilled
}

func (p *prefilled) Line(level xopnum.Level, t time.Time, pc []uintptr) xopbase.Line {
	atomic.StoreInt64(&p.span.endTime, t.UnixNano())
	if level >= xopnum.ErrorLevel {
		if level >= xopnum.AlertLevel {
			_ = atomic.AddInt32(&p.span.request.alertCount, 1) // XXX move to logger and include in flush?
		} else {
			_ = atomic.AddInt32(&p.span.request.errorCount, 1)
		}
	}
	l = &line{
		builder: p.span.builder(),
		Line: xopproto.Line{
			LogLevel:  int32(level),
			TimeStamp: t.UnixNano(),
		},
		prefillMsgPreEncoded: p.preEncodedMsg,
	}
	return l
}

func (l *line) Template(m string) {
	l.LineKind = xopproto.KindLine
	l.MessageTemplate = m
	l.done()
}

func (l *line) Msg(m string) {
	l.LineKind = xopproto.KindLine
	l.Message = m
	l.done()
}

func (l *line) done() {
	// XXX
}

func (l *line) Model(k string, v xopbase.ModelArg) {
	l.Model = &xopbase.Model{}
	enc, err := json.Marshal(v.Model)
	if err != nil {
		l.Model.Error = err.Error()
	} else {
		l.Model.Json = enc
	}
	if v.TypeName == "" {
		l.Model.Type = reflect.TypeOf(v.Model).Name()
	} else {
		l.Model.Type = v.TypeName
	}
	l.LineKind = xopproto.KindModel
	l.Message = k
	l.done()
}

func (l *line) Link(k string, v xoptrace.Trace) {
	l.LineKind = xopproto.KindLink
	l.Message = v.Trace.String() // XXX custom type?
	l.done()
}

/* XXX
func (b *builder) Link(k string, v xoptrace.Trace) {
	b.startAttributes()
	b.AddKey(k)
	b.AddLink(v)
}
*/

func (b *builder) ReclaimMemory() {
	// XXX
}

func (b *builder) AsBytes() []byte            // XXX
func (l *line) GetSpanID() xoptrace.HexBytes8 { return l.span.bundle.Trace.GetSpanID() }
func (l *line) GetLevel() xopnum.Level        { return xopnum.Level(l.LogLevel) }
func (l *line) GetTime() time.Time            { return time.UnixNano(0, l.Timestamp) }

func (l *line) ReclaimMemory() {
}

func (b *builder) Any(k string, v xopbase.ModelArg) {
	var intValue int64
	enc, err := json.Marshal(v.Model)
	if err != nil {
		enc = []byte(err.Error())
		intValue = -1
	}
	typeName := reflect.TypeOf(v).Name()
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Any,
		Value: &xopproto.AttributeValue{
			StringValue: typeName,
			BytesValue:  enc,
			IntValue:    intValue,
		},
	})
}

func (b *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k.Key(),
		Type: xopproto.AttributeType_Enum,
		Value: &xopproto.AttributeValue{
			StringValue: v.String(),
			IntValue:    v.Int64(),
		},
	})
}

func (b *builder) Time(k string, t time.Time) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Time,
		AttributeValue: &xopproto.AttributeValue{
			IntValue: t.UnixNano(),
		},
	})
}

func (b *builder) Bool(k string, v bool) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Bool,
		Value: &xopproto.AttributeValue{
			IntValue: boolToInt64(v),
		},
	})
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (b *builder) Int64(k string, v int64, _ xopbase.DataType) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Int64, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			IntValue: v,
		},
	})
}

func (b *builder) Uint64(k string, v uint64, _ xopbase.DataType) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Uint64, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			UintValue: v,
		},
	})
}

func (b *builder) String(k string, v string, _ xopbase.DataType) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_String, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			StringValue: v,
		},
	})
}

func (b *builder) Float64(k string, v float64, _ xopbase.DataType) {
	b.attributes = append(b.attributes, xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Float64, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			FloatValue: v,
		},
	})
}

func (b *builder) Duration(k string, v time.Duration) {
	b.Int64(k, int64(v), xopbase.DataTypeDuration)
}

func (s *span) MetadataAny(k *xopat.AnyAttribute, v interface{}) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	typeName := reflect.TypeOf(v).Name()
	enc, err := json.Marshal(v)
	if err != nil {
		if !existingAttribute {
			attribute.Values = append(attribute.Values, &xopproto.AttributeValue{
				StringValue: typeName,
				BytesValue:  []byte(err.Error()),
				IntValue:    -1,
			})
		}
		return
	}
	var dkAny string
	if k.Distinct() {
		dkAny = string(enc)
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.StringValue = typeName
		value.BytesValue = enc
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenString == nil {
					distinct.seenString = make(map[string]struct{})
				}
				dk := dkAny
				if _, ok := distinct.seenString[dk]; ok {
					return
				}
				distinct.seenString[dk] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataBool(k *xopat.BoolAttribute, v bool) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		if v {
			value.IntValue = 1
		} else {
			value.IntValue = 0
		}
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenInt == nil {
					distinct.seenInt = make(map[int64]struct{})
				}
				var dk int64
				if v {
					dk = 1
				}
				if _, ok := distinct.seenInt[dk]; ok {
					return
				}
				distinct.seenInt[dk] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.IntValue = v.Int()
		value.StringValue = v.String()
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.FloatValue = v
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenFloat64 == nil {
					distinct.seenFloat64 = make(map[float64]struct{})
				}
				if _, ok := distinct.seenFloat[v]; ok {
					return
				}
				distinct.seenFloat[v] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.IntValue = v
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenInt == nil {
					distinct.seenInt = make(map[int64]struct{})
				}
				dk := v
				if _, ok := distinct.seenInt[dk]; ok {
					return
				}
				distinct.seenInt[dk] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.StringValue = v.String()
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenString == nil {
					distinct.seenString = make(map[string]struct{})
				}
				dk := v.String()
				if _, ok := distinct.seenString[dk]; ok {
					return
				}
				distinct.seenString[dk] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataString(k *xopat.StringAttribute, v string) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.StringValue = v
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenString == nil {
					distinct.seenString = make(map[string]struct{})
				}
				dk := v
				if _, ok := distinct.seenString[dk]; ok {
					return
				}
				distinct.seenString[dk] = struct{}{}
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}

func (s *span) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	var distinct *distinction
	s.mu.Lock()
	defer s.mu.Unlock()
	attribute, existingAttribute := s.attributeMap[k.Key()]
	if !existingAttribute {
		var c int
		if !k.Multiple() {
			c = 1
		}
		attribute = &xopproto.SpanAttribute{
			// XXX AttributeDefinitionSequenceNumber:
			Values: make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.IntValue = v.UnixNano()
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
			}()
		}
		var value xopproto.AttributeValue
		setValue(&value)
		attribute.Values = append(attribute.Values, &value)
	} else {
		if k.Locked() && existingAttribute {
			return
		}
		setValue(attribute.Values[0])
	}
}
