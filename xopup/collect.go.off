package xopup

import (
	"sync"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopproto"
)

type definitionComplete struct {
	once sync.Once
}

func (r *Request) AttributeReferenced(*xopat.Attribute) error { return nil } // TODO

func (u *Uploader) DefineAttribute(a *xopat.Attribute) {
	attributeKey := attributeKey{
		key:       a.Key(),
		namespace: a.Namespace(),
	}
	// We keep a pool of unused definition complete structs
	// so that when we get a hit, we're not creating another
	// throw-away object in the heap
	n := u.definitionsComplete.Get()
	v, ok := u.attributesDefined.LoadOrStore(attributeKey, n)
	if ok {
		u.definitionsComplete.Put(n)
		return
	}

	v.(*definitionComplete).once.Do(func() {
		definition := xopproto.AttributeDefinition{
			Key:             a.Key(),
			Description:     a.Description(),
			Namespace:       a.Namespace(),
			NamespaceSemver: a.SemverString(),
			Type:            xopproto.AttributeType(a.SubType()),
			ShouldIndex:     a.Indexed(),
			Prominence:      int32(a.Prominence()),
			Locked:          a.Locked(),
			Distinct:        a.Distinct(),
			Multiple:        a.Multiple(),
		}
		u.lock.Lock()
		defer u.lock.Unlock()
		fragment := u.getFragment()
		fragment.AttributeDefinitions = append(fragment.AttributeDefinitions, &definition)
	})
}

func (u *Uploader) DefineEnum(a *xopat.EnumAttribute, e xopat.Enum) {
	enumKey := enumKey{
		attributeKey: attributeKey{
			key:       a.Key(),
			namespace: a.Namespace(),
		},
		value: e.Int64(),
	}
	// We keep a pool of unused definition complete structs
	// so that when we get a hit, we're not creating another
	// throw-away object in the heap
	n := u.definitionsComplete.Get()
	v, ok := u.enumsDefined.LoadOrStore(enumKey, n)
	if ok {
		u.definitionsComplete.Put(n)
		return
	}

	v.(*definitionComplete).once.Do(func() {
		enum := xopproto.EnumDefinition{
			AttributeKey:    a.Key(),
			Namespace:       a.Namespace(),
			NamespaceSemver: a.SemverString(),
			String_:         e.String(),
			IntValue:        e.Int64(),
		}
		u.lock.Lock()
		defer u.lock.Unlock()
		fragment := u.getFragment()
		fragment.EnumDefinitions = append(fragment.EnumDefinitions, &enum)
	})
}

func (r *Request) Span(span xopbytes.Span, buffer xopbytes.Buffer) error {
	bundle := span.GetBundle()
	pbSpan := xopproto.Span{
		SpanID:    bundle.Trace.GetSpanID().Bytes(),
		ParentID:  bundle.Parent.GetSpanID().Bytes(),
		JsonData:  buffer.AsBytes(),
		StartTime: span.GetStartTime().UnixNano(),
		EndTime:   pointerToInt64OrNil(span.GetEndTimeNano()),
	}
	if span.IsRequest() {
		pbSpan.IsRequest = true
		pbSpan.Baggage = bundle.Baggage.Bytes()
		pbSpan.TraceState = bundle.State.Bytes()
	}
	r.uploader.lock.Lock()
	defer r.uploader.lock.Unlock()
	request, byteCount := r.uploader.getRequest(r, true)
	request.Spans = append(request.Spans, &pbSpan)
	return r.uploader.noteBytes(byteCount + sizeOfSpan + len(pbSpan.JsonData) + len(pbSpan.Baggage) + len(pbSpan.TraceState))
}

func (r *Request) Line(line xopbytes.Line) error {
	pbLine := xopproto.Line{
		SpanID:    line.GetSpanID().Bytes(),
		LogLevel:  int32(line.GetLevel()),
		Timestamp: line.GetTime().UnixNano(),
		JsonData:  line.AsBytes(),
	}
	r.uploader.lock.Lock()
	defer r.uploader.lock.Unlock()
	request, byteCount := r.uploader.getRequest(r, true)
	r.lineCount++
	request.Lines = append(request.Lines, &pbLine)
	return r.uploader.noteBytes(byteCount + sizeOfLine + len(pbLine.JsonData))
}
