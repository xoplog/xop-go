// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoppb

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

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
			logger: logger,
			bundle: bundle,
			protoSpan: xopproto.Span{
				Name:      name,
				StartTime: ts.UnixNano(),
				IsRequest: true,
				ParentID:  bundle.Parent.GetSpanID().Bytes(),
				SpanID:    bundle.Trace.GetSpanID().Bytes(),
			},
			needFlushing: make([]*span, 0, 10),
		},
		sourceInfo:     sourceInfo,
		lines:          make([]*xopproto.Line, 0, 200),
		attributeIndex: make(map[int32]uint32),
	}
	request.request = request
	request.parent = &request.span
	return request
}

func (r *request) Flush() {
	r.flushGeneration++
	rproto := xopproto.Request{
		Span:                   r.span.getProto(r.flushGeneration),
		ParentTraceID:          r.bundle.Parent.GetTraceID().Bytes(),
		SourceNamespace:        r.sourceInfo.Namespace,
		SourceNamespaceVersion: r.sourceInfo.NamespaceVersion.String(),
		SourceID:               r.sourceInfo.Source,
		SourceVersion:          r.sourceInfo.SourceVersion.String(),
		Baggage:                r.bundle.Baggage.String(),
		TraceState:             r.bundle.State.String(),
	}
	nLines := make([]*xopproto.Line, 0, 200)
	func() {
		r.requestLock.Lock()
		defer r.requestLock.Unlock()
		rproto.AttributeDefinitions = r.attributeDefinitions[:]
	}()
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
	r.logger.writer.Request(r.bundle.Trace.GetTraceID(), &rproto)
	r.logger.writer.Flush()
}

func (s *span) getProto(flushGeneration int) *xopproto.Span {
	if s.lastFlush == flushGeneration {
		return nil
	}
	s.lastFlush = flushGeneration
	var c xopproto.Span
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.protoSpan.Version++
		if s.endTime != 0 {
			s.protoSpan.EndTime = &s.endTime
		}
		c = s.protoSpan
		c.Attributes = list.Copy(c.Attributes)
	}()
	nSpans := make([]*span, 0, 5)
	var needFlushing []*span
	func() {
		s.spanLock.Lock()
		defer s.spanLock.Unlock()
		needFlushing = s.needFlushing
		s.needFlushing = nSpans
	}()
	c.Spans = make([]*xopproto.Span, 0, len(needFlushing))
	for _, span := range needFlushing {
		p := span.getProto(flushGeneration)
		if p != nil {
			c.Spans = append(c.Spans, p)
		}
	}
	return &c
}

func (r *request) Final() {}

func (r *request) SetErrorReporter(reporter func(error)) { r.errorFunc = reporter }
func (r *request) GetErrorCount() int32                  { return atomic.LoadInt32(&r.errorCount) }
func (r *request) GetAlertCount() int32                  { return atomic.LoadInt32(&r.alertCount) }

func (r *request) defineAttribute(k xopat.AttributeInterface) uint32 {
	r.requestLock.Lock()
	defer r.requestLock.Unlock()
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
		Ranged:          k.Ranged(),
		Multiple:        k.Multiple(),
	})
	return i
}

func (s *span) Span(_ context.Context, ts time.Time, bundle xoptrace.Bundle, name string, spanSequenceCode string) xopbase.Span {
	n := &span{
		logger:  s.logger,
		bundle:  bundle,
		request: s.request,
		parent:  s,
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

func (s *span) Done(t time.Time, _ bool) {
	xoputil.AtomicMaxInt64(&s.endTime, t.UnixNano())
	s.done()
}

func (s *span) done() {
	if s.protoSpan.IsRequest {
		return
	}
	func() {
		s.parent.spanLock.Lock()
		defer s.parent.spanLock.Unlock()
		s.parent.needFlushing = append(s.parent.needFlushing, s)
	}()
	s.parent.done()
}

func (s *span) Boring(bool)                {}
func (s *span) ID() string                 { return s.logger.id.String() }
func (s *span) GetBundle() xoptrace.Bundle { return s.bundle }
func (s *span) GetStartTime() time.Time    { return time.Unix(0, s.protoSpan.StartTime) }
func (s *span) GetEndTimeNano() int64      { return s.endTime }
func (s *span) IsRequest() bool            { return s.protoSpan.IsRequest }

func (s *span) NoPrefill() xopbase.Prefilled {
	return &prefilled{
		builder: &builder{
			span: s,
		},
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
		builder:    p.builder,
		prefillMsg: m,
	}
}

func (p *prefilled) Line(level xopnum.Level, t time.Time, frames []runtime.Frame) xopbase.Line {
	xoputil.AtomicMaxInt64(&p.span.endTime, t.UnixNano())
	if level >= xopnum.ErrorLevel {
		if level >= xopnum.AlertLevel {
			_ = atomic.AddInt32(&p.span.request.alertCount, 1) // TODO: move to logger and include in flush?
		} else {
			_ = atomic.AddInt32(&p.span.request.errorCount, 1)
		}
	}
	l := &line{
		builder: p.span.builder(),
		protoLine: &xopproto.Line{
			LogLevel:  int32(level),
			Timestamp: t.UnixNano(),
			SpanID:    p.span.bundle.Trace.SpanID().Bytes(),
		},
		prefillMsg: p.prefillMsg,
	}
	l.builder.attributes = list.Copy(p.attributes)
	if len(frames) > 0 {
		l.protoLine.StackFrames = make([]*xopproto.StackFrame, len(frames))
		for i, frame := range frames {
			l.protoLine.StackFrames[i] = &xopproto.StackFrame{
				File:       frame.File,
				LineNumber: int32(frame.Line),
			}
		}
	}
	return l
}

func (l *line) Template(m string) {
	l.protoLine.LineKind = xopproto.LineKind_KindLine
	l.protoLine.MessageTemplate = l.prefillMsg + m
	l.done()
}

func (l *line) Msg(m string) {
	l.protoLine.LineKind = xopproto.LineKind_KindLine
	l.protoLine.Message = l.prefillMsg + m
	l.done()
}

func (l *line) done() {
	l.protoLine.Attributes = l.builder.attributes
	l.span.request.lineLock.Lock()
	defer l.span.request.lineLock.Unlock()
	l.span.request.lines = append(l.span.request.lines, l.protoLine)
}

func (l *line) Model(k string, v xopbase.ModelArg) {
	l.protoLine.Model = &xopproto.Model{}
	v.Encode()
	l.protoLine.Model.Encoded = v.Encoded
	l.protoLine.Model.Type = v.ModelType
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

func (b *builder) ReclaimMemory() {}

func (l *line) GetSpanID() xoptrace.HexBytes8 { return l.span.bundle.Trace.GetSpanID() }
func (l *line) GetLevel() xopnum.Level        { return xopnum.Level(l.protoLine.LogLevel) }
func (l *line) GetTime() time.Time            { return time.Unix(0, l.protoLine.Timestamp) }

func (l *line) ReclaimMemory() {}

func (b *builder) Any(k string, v xopbase.ModelArg) {
	v.Encode()
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType_Any,
		Value: &xopproto.AttributeValue{
			StringValue: v.ModelType,
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

func (b *builder) Int64(k string, v int64, dataType xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(dataType),
		Value: &xopproto.AttributeValue{
			IntValue: v,
		},
	})
}

func (b *builder) Uint64(k string, v uint64, dataType xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(dataType),
		Value: &xopproto.AttributeValue{
			UintValue: v,
		},
	})
}

func (b *builder) String(k string, v string, dataType xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(dataType),
		Value: &xopproto.AttributeValue{
			StringValue: v,
		},
	})
}

func (b *builder) Float64(k string, v float64, dataType xopbase.DataType) {
	b.attributes = append(b.attributes, &xopproto.Attribute{
		Key:  k,
		Type: xopproto.AttributeType(dataType),
		Value: &xopproto.AttributeValue{
			FloatValue: v,
		},
	})
}

func (b *builder) Duration(k string, v time.Duration) {
	b.Int64(k, int64(v), xopbase.DurationDataType)
}

func (s *span) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
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
	v.Encode()
	setValue := func(value *xopproto.AttributeValue) {
		value.StringValue = v.ModelType
		value.BytesValue = v.Encoded
		value.IntValue = int64(v.Encoding)
	}
	if k.Multiple() {
		if k.Distinct() {
			func() {
				distinct.mu.Lock()
				defer distinct.mu.Unlock()
				if distinct.seenString == nil {
					distinct.seenString = make(map[string]struct{})
				}
				dk := string(v.Encoded)
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
