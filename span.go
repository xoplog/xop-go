package xoplog

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopconst"
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

// Data adds one key/value pair to describe the span.  The
// Span is returned so data methods can be chained.
func (s *Span) Int64Data(key *xopconst.IntAttribute, value int64) *Span {
	func() {
		s.dataLock.Lock()
		defer s.dataLock.Unlock()
		s.data.intData = append(s.data.intData, intData{key: key, value: value})
	}()
	s.log.setDirty()
	return s
}
func (s *Span) IntData(key *xopconst.IntAttribute, value int) *Span {
	return s.Int64Data(key, int64(value))
}
func (s *Span) Int8Data(key *xopconst.IntAttribute, value int8) *Span {
	return s.Int64Data(key, int64(value))
}
func (s *Span) Int16Data(key *xopconst.IntAttribute, value int16) *Span {
	return s.Int64Data(key, int64(value))
}
func (s *Span) Int32ata(key *xopconst.IntAttribute, value int32) *Span {
	return s.Int64Data(key, int64(value))
}

func (s *Span) StrData(key *xopconst.StrAttribute, value string) *Span {
	func() {
		s.dataLock.Lock()
		defer s.dataLock.Unlock()
		s.data.strData = append(s.data.strData, intData{key: key, value: value})
	}()
	s.log.setDirty()
	return s
}
