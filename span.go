// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file
package xoplog

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopconst"

	"github.com/mohae/deepcopy"
)

// Request provides access to the span that describes the overall
// request. Metadata may be added at the request level.
func (l *Log) Request() *Span {
	return l.request
}

// Request provides access to the current span
// Metadata may be added at the span level.
func (l *Log) Span() *Span {
	return &l.span
}

func (s *Span) TraceState() trace.State     { return s.seed.traceBundle.State }
func (s *Span) TraceBaggage() trace.Baggage { return s.seed.traceBundle.Baggage }
func (s *Span) TraceParent() trace.Trace    { return s.seed.traceBundle.TraceParent.Copy() }
func (s *Span) Trace() trace.Trace          { return s.seed.traceBundle.Trace.Copy() }
func (s *Span) Bundle() trace.Bundle        { return s.seed.traceBundle.Copy() }

func (s *Span) eft() *Span {
	s.log.enableFlushTimer()
	return s
}

func (s *Span) Int64(k *xopconst.IntAttribute, v int64) *Span {
	s.base.MetadataInt(k, v)
	return s.eft()
}
func (s *Span) Int8(k *xopconst.IntAttribute, v int8) *Span   { return s.Int64(k, int64(v)) }
func (s *Span) Int16(k *xopconst.IntAttribute, v int16) *Span { return s.Int64(k, int64(v)) }
func (s *Span) Int32(k *xopconst.IntAttribute, v int32) *Span { return s.Int64(k, int64(v)) }
func (s *Span) Int(k *xopconst.IntAttribute, v int) *Span     { return s.Int64(k, int64(v)) }

// AnyImmutable adds a key/value attribute to the Span.  The provided
// value must be immutable.
func (s *Span) AnyImmutable(k *xopconst.AnyAttribute, v interface{}) *Span {
	s.base.MetadataAny(k, v)
	return s.eft()
}

// Any adds a key/value attribute to the Span.  The provided
// value may be copied using github.com/mohae/deepcopy if any of the
// base loggers hold the value instead of immediately serializing it.
func (s *Span) Any(k *xopconst.AnyAttribute, v interface{}) *Span {
	if s.log.shared.ReferencesKept {
		v = deepcopy.Copy(v)
	}
	s.base.MetadataAny(k, v)
	return s.eft()
}

// Bool adds a bool key/value attribute to the Span
func (s *Span) Bool(k *xopconst.BoolAttribute, v bool) *Span {
	s.base.MetadataBool(k, v)
	return s.eft()
}

// Duration adds a time.Duration key/value attribute to the Span
func (s *Span) Duration(k *xopconst.DurationAttribute, v time.Duration) *Span {
	s.base.MetadataDuration(k, v)
	return s.eft()
}

// Link adds a trace.Trace key/value attribute to the Span
func (s *Span) Link(k *xopconst.LinkAttribute, v trace.Trace) *Span {
	s.base.MetadataLink(k, v)
	return s.eft()
}

// Str adds a string key/value attribute to the Span
func (s *Span) Str(k *xopconst.StrAttribute, v string) *Span {
	s.base.MetadataStr(k, v)
	return s.eft()
}

// Time adds a time.Time key/value attribute to the Span
func (s *Span) Time(k *xopconst.TimeAttribute, v time.Time) *Span {
	s.base.MetadataTime(k, v)
	return s.eft()
}
