package xoplog

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopconst"

	"github.com/mohae/deepcopy"
)

type spanData struct {
	intData      []intData
	strData      []strData
	anyData      []anyData
	timeData     []timeData
	linkData     []linkData
	durationData []durationData
}

type intData struct {
	key   *xopconst.IntAttribute
	value int64
}

type strData struct {
	key   *xopconst.StrAttribute
	value string
}

type anyData struct {
	key   *xopconst.StrAttribute
	value interface{}
}

type linkData struct {
	key   *xopconst.StrAttribute
	value trace.Trace
}

type timeData struct {
	key   *xopconst.TimeAttribute
	value time.Time
}

type timeData struct {
	key   *xopconst.TimeAttribute
	value time.Time
}

func (l *Log) Request() *Span {
	return l.request
}

func (l *Log) Span() *Span {
	return &l.span
}

func (s *Span) Data() *SpanData {
	// TODO: PEROFMANCE: use a pool
	return SpanDataSetter{
		span: s,
	}
}

type SpanData struct {
	span *Span
	data spanData
}

func (d *SpanData) Int(k *xopconst.IntAttribute, v int) *SpanData {
	d.data.intData = append(d.data.intData, intData{key: k, value: int64(v)})
	return d
}

func (d *SpanData) Int8(k *xopconst.IntAttribute, v int8) *SpanData {
	d.data.intData = append(d.data.intData, intData{key: k, value: int64(v)})
	return d
}

func (d *SpanData) Int16(k *xopconst.IntAttribute, v int16) *SpanData {
	d.data.intData = append(d.data.intData, intData{key: k, value: int64(v)})
	return d
}

func (d *SpanData) Int32(k *xopconst.IntAttribute, v int32) *SpanData {
	d.data.intData = append(d.data.intData, intData{key: k, value: int64(v)})
	return d
}

func (d *SpanData) Int64(k *xopconst.IntAttribute, v int64) *SpanData {
	d.data.intData = append(d.data.intData, intData{key: k, value: v})
	return d
}

func (d *SpanData) Str(k *xopconst.IntAttribute, v string) *SpanData {
	d.data.strData = append(d.data.strData, strData{key: k, value: v})
	return d
}

func (d *SpanData) Any(k *xopconst.IntAttribute, v interface{}) *SpanData {
	v = deepcopy.Copy(v)
	d.data.anyData = append(d.data.anyData, strData{key: k, value: v})
	return d
}

func (d *SpanData) AnyImmutable(k *xopconst.IntAttribute, v interface{}) *SpanData {
	d.data.anyData = append(d.data.anyData, strData{key: k, value: v})
	return d
}

func (d *SpanData) Link(k *xopconst.LinkAttribute, v trace.Trace) *SpanData {
	d.data.linkData = append(d.data.linkData, linkData{key: k, value: v})
	return d
}

func (d *SpanData) Duration(k *xopconst.DurationAttribute, v time.Duration) *SpanData {
	d.data.linkData = append(d.data.durationData, durationData{key: k, value: v})
	return d
}

func (d *SpanData) Add() {
	func() {
		s.dataLock.Lock()
		defer s.dataLock.Unlock()
		// !GENERATE!
		span.data.intData = append(span.data.intData, d.data.intData...)
		span.data.strData = append(span.data.intData, d.data.strData...)
		span.data.anyData = append(span.data.intData, d.data.anyData...)
		span.data.timeData = append(span.data.intData, d.data.timeData...)
		span.data.linkData = append(span.data.intData, d.data.linkData...)
		span.data.durationData = append(span.data.intData, d.data.durationData...)
	}()
	s.log.setDirty()
}
