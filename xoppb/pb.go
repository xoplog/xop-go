// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	"github.com/muir/list"
)

func New(w Writer) *Logger {
	log := &Logger{
		writer: w,
		id:     uuid.New(),
	}
	return log
}

func (logger *Logger) ID() string           { return logger.id.String() }
func (logger *Logger) Buffered() bool       { return false }
func (logger *Logger) ReferencesKept() bool { return false }

func (logger *Logger) Request(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	request := &request{
		span: span{
			logger:  logger,
			bundle:  bundle,
			endTime: ts.UnixNano(),
			protoSpan: xopproto.Span{
				Name:      name,
				StartTime: ts.UnixNano(),
				IsRequest: true,
				ParentID:  bundle.Parent.GetSpanID().Bytes(),
				SpanID:    bundle.Trace.GetSpanID().Bytes(),
			},
		},
		sourceInfo:     sourceInfo,
		spans:          make([]*xopproto.Span, 1, 20), // reserving space for self
		lines:          make([]*xopproto.Line, 0, 200),
		attributeIndex: make(map[int32]uint32),
	}
	request.request = request
	return request
}

func (r *request) Flush() {
	var rproto xopproto.Request
	nSpans := make([]*xopproto.Span, 1, 20)
	nLines := make([]*xopproto.Line, 0, 200)
	func() {
		r.spanLock.Lock()
		defer r.spanLock.Unlock()
		rproto.Spans = r.spans
		r.spans = nSpans
		rproto.AttributeDefinitions = r.attributeDefinitions[:]
	}()
	rproto.Spans[0] = r.span.getProto()
	func() {
		r.lineLock.Lock()
		defer r.lineLock.Unlock()
		rproto.PriorLinesInRequest = int32(r.priorLines)
		r.priorLines += len(r.lines)
		rproto.Lines = r.lines
		r.lines = nLines
		rproto.AlertCount = atomic.LoadInt32(&r.alertCount)
		rproto.ErrorCount = atomic.LoadInt32(&r.errorCount)
	}()
	rproto.RequestID = r.bundle.Trace.GetSpanID().Bytes()
	rproto.ParentTraceID = r.bundle.Parent.GetTraceID().Bytes()
	rproto.SourceNamespace = r.sourceInfo.Namespace
	rproto.SourceNamespaceVersion = r.sourceInfo.NamespaceVersion.String()
	rproto.SourceID = r.sourceInfo.Source
	rproto.SourceVersion = r.sourceInfo.SourceVersion.String()
	r.logger.writer.Request(r.bundle.Trace.GetTraceID(), &rproto)
	r.logger.writer.Flush()
}

func (r *request) Final() {}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }
func (r *request) GetErrorCount() int32                  { return atomic.LoadInt32(&r.errorCount) }
func (r *request) GetAlertCount() int32                  { return atomic.LoadInt32(&r.alertCount) }

func (r *request) defineAttribute(k xopat.AttributeInterface) uint32 {
	r.spanLock.Lock()
	defer r.spanLock.Unlock()
	n := k.RegistrationNumber()
	if i, ok := r.attributeIndex[n]; ok {
		return i
	}
	i := uint32(len(r.attributeDefinitions))
	r.attributeIndex[n] = i
	r.attributeDefinitions = append(r.attributeDefinitions, &xopproto.AttributeDefinition{
		Key:             k.Key(),
		Description:     k.Description(),
		Namespace:       k.Namespace(),
		NamespaceSemver: k.SemverString(),
		Type:            k.ProtoType(),
		ShouldIndex:     k.Indexed(),
		Prominence:      int32(k.Prominence()),
		Locked:          k.Locked(),
		Distinct:        k.Distinct(),
		Multiple:        k.Multiple(),
	})
	return i
}

func (s *span) Span(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	n := &span{
		logger:  s.logger,
		bundle:  bundle,
		request: s.request,
		endTime: ts.UnixNano(),
		protoSpan: xopproto.Span{
			Name:         name,
			StartTime:    ts.UnixNano(),
			SequenceCode: spanSequenceCode,
			ParentID:     bundle.Parent.GetSpanID().Bytes(),
			SpanID:       bundle.Trace.GetSpanID().Bytes(),
		},
	}
	return n
}

func (s *span) getProto() *xopproto.Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.protoSpan.Version++
	c := s.protoSpan
	c.Attributes = list.Copy(c.Attributes)
	return &c
}

func (s *span) Done(t time.Time, _ bool) {
	atomic.StoreInt64(&s.endTime, t.UnixNano())
	if s.protoSpan.IsRequest {
		return
	}
	p := s.getProto()
	s.request.spanLock.Lock()
	defer s.request.spanLock.Unlock()
	fmt.Printf("XXX in span.Done, appending %s\n", s.bundle.Trace.GetSpanID())
	s.request.spans = append(s.request.spans, p)
}

func (s *span) Boring(bool)                {}
func (s *span) ID() string                 { return s.logger.id.String() }
func (s *span) GetBundle() xoptrace.Bundle { return s.bundle }
func (s *span) GetStartTime() time.Time    { return time.Unix(0, s.protoSpan.StartTime) }
func (s *span) GetEndTimeNano() int64      { return s.endTime }
func (s *span) IsRequest() bool            { return s.protoSpan.IsRequest }

func (s *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		span: s,
	}
}

func (s *span) builder() *builder {
	return &builder{
		span: s,
	}
}

func (s *span) StartPrefill() xopbase.Prefilling {
	return &prefilling{
		builder: s.builder(),
	}
}

func (p *prefilling) PrefillComplete(m string) xopbase.Prefilled {
	return &prefilled{
		data:       p.attributes,
		prefillMsg: m,
		span:       p.span,
	}
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
	l := &line{
		builder: p.span.builder(),
		protoLine: &xopproto.Line{
			LogLevel:   int32(level),
			Timestamp:  t.UnixNano(),
			Attributes: list.Copy(p.data),
		},
	}
	return l
}

func (l *line) Template(m string) {
	l.protoLine.LineKind = xopproto.LineKind_KindLine
	l.protoLine.MessageTemplate = m
	l.done()
}

func (l *line) Msg(m string) {
	l.protoLine.LineKind = xopproto.LineKind_KindLine
	l.protoLine.Message = m
	l.done()
}

func (l *line) done() {
	l.span.request.lineLock.Lock()
	defer l.span.request.lineLock.Unlock()
	l.span.request.lines = append(l.span.request.lines, l.protoLine)
}

func (l *line) Model(k string, v xopbase.ModelArg) {
	l.protoLine.Model = &xopproto.Model{}
	v.Encode()
	l.protoLine.Model.Encoded = v.Encoded
	l.protoLine.Model.Type = v.TypeName
	l.protoLine.Model.Encoding = v.Encoding
	l.protoLine.LineKind = xopproto.LineKind_KindModel
	l.protoLine.Message = k
	l.done()
}

func (l *line) Link(k string, v xoptrace.Trace) {
	l.protoLine.LineKind = xopproto.LineKind_KindLink
	l.protoLine.Message = k
	l.protoLine.Link = v.String()
	l.done()
}

func (b *builder) ReclaimMemory() {
	// XXX
}

// func (b *builder) AsBytes() []byte            // XXX
func (l *line) GetSpanID() xoptrace.HexBytes8 { return l.span.bundle.Trace.GetSpanID() }
func (l *line) GetLevel() xopnum.Level        { return xopnum.Level(l.protoLine.LogLevel) }
func (l *line) GetTime() time.Time            { return time.Unix(0, l.protoLine.Timestamp) }

func (l *line) ReclaimMemory() {
}

func (b *builder) Any(k string, v xopbase.ModelArg) {
	v.Encode()
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Any,
		Value: &xopproto.AttributeValue{
			StringValue: v.TypeName,
			BytesValue:  v.Encoded,
			IntValue:    int64(v.Encoding),
		},
	})
}

func (b *builder) Enum(k *xopat.EnumAttribute, v xopat.Enum) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k.Key(),
		Type: xopproto.AttributeType_Enum,
		Value: &xopproto.AttributeValue{
			StringValue: v.String(),
			IntValue:    v.Int64(),
		},
	})
}

func (b *builder) Time(k string, t time.Time) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Time,
		Value: &xopproto.AttributeValue{
			IntValue: t.UnixNano(),
		},
	})
}

func (b *builder) Bool(k string, v bool) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
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
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Int64, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			IntValue: v,
		},
	})
}

func (b *builder) Uint64(k string, v uint64, _ xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Uint64, // Convert to xopproto
		Value: &xopproto.AttributeValue{
			UintValue: v,
		},
	})
}

func (b *builder) String(k string, v string, typ xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(typ),
		Value: &xopproto.AttributeValue{
			StringValue: v,
		},
	})
}

func (b *builder) Float64(k string, v float64, typ xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(typ),
		Value: &xopproto.AttributeValue{
			FloatValue: v,
		},
	})
}

func (b *builder) Duration(k string, v time.Duration) {
	b.Int64(k, int64(v), xopbase.DurationDataType)
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
			s.distinctMaps[k.Key()] = distinct
		}
	}
	setValue := func(value *xopproto.AttributeValue) {
		value.IntValue = v.Int64()
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
				if distinct.seenFloat == nil {
					distinct.seenFloat = make(map[float64]struct{})
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
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
			AttributeDefinitionSequenceNumber: s.request.defineAttribute(k),
			Values:                            make([]*xopproto.AttributeValue, c, 1),
		}
		s.protoSpan.Attributes = append(s.protoSpan.Attributes, attribute)
		if k.Distinct() {
			distinct = &distinction{}
			if s.distinctMaps == nil {
				s.distinctMaps = make(map[string]*distinction)
			}
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
		if existingAttribute {
			if k.Locked() {
				return
			}
		} else {
			attribute.Values[0] = &xopproto.AttributeValue{}
		}
		setValue(attribute.Values[0])
	}
}
