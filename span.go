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

func (s *Span) Int64(k *xopconst.Int64Attribute, v int64) *Span {
	s.base.MetadataInt64(k, v)
	return s.eft()
}

// EmbeddedEnum adds a kev/value attribute to the Span.  The key and the value
// are bundled together: the key is derrived from the type of the Enum.
// Alternatively, use xopconst.KeyedEnumAttribute() to create functions
// to add enum key/value pairs where the key and value are specified
// separately.
func (s *Span) EmbeddedEnum(kv xopconst.EmbeddedEnum) *Span {
	return s.Enum(kv.EnumAttribute(), kv)
}

// AnyImmutable adds a key/value attribute to the current Span.  The provided
// value must be immutable.  If it is not, then there could be race conditions
// or the value that ends up logged could be different from the value at the
// time when AnyImmutible was called.
//
// While the AnyAttribute has an expectation
// for the type of the value, that type may or may not be checked depending
// on the base logger being used.
func (s *Span) AnyImmutable(k *xopconst.AnyAttribute, v interface{}) *Span {
	s.base.MetadataAny(k, v)
	return s.eft()
}

// Any adds a key/value attribute to the current Span.  The provided
// value may be copied using github.com/mohae/deepcopy if any of the
// base loggers hold the value instead of immediately serializing it.
// While the AnyAttribute has an expectation
// for the type of the value, that type may or may not be checked depending
// on the base logger being used.
func (s *Span) Any(k *xopconst.AnyAttribute, v interface{}) *Span {
	if s.log.shared.ReferencesKept {
		v = deepcopy.Copy(v)
	}
	s.base.MetadataAny(k, v)
	return s.eft()
}

// Bool adds a bool key/value attribute to the current Span
func (s *Span) Bool(k *xopconst.BoolAttribute, v bool) *Span {
	s.base.MetadataBool(k, v)
	return s.eft()
}

// Enum adds a xopconst.Enum key/value attribute to the current Span
func (s *Span) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) *Span {
	s.base.MetadataEnum(k, v)
	return s.eft()
}

// Link adds a trace.Trace key/value attribute to the current Span
func (s *Span) Link(k *xopconst.LinkAttribute, v trace.Trace) *Span {
	s.base.MetadataLink(k, v)
	return s.eft()
}

// Number adds a float64 key/value attribute to the current Span
func (s *Span) Number(k *xopconst.NumberAttribute, v float64) *Span {
	s.base.MetadataNumber(k, v)
	return s.eft()
}

// Str adds a string key/value attribute to the current Span
func (s *Span) Str(k *xopconst.StrAttribute, v string) *Span {
	s.base.MetadataStr(k, v)
	return s.eft()
}

// Time adds a time.Time key/value attribute to the current Span
func (s *Span) Time(k *xopconst.TimeAttribute, v time.Time) *Span {
	s.base.MetadataTime(k, v)
	return s.eft()
}

// should skip Int64
// Duration adds a time.Duration key/value attribute to the current Span
func (s *Span) Duration(k *xopconst.DurationAttribute, v time.Duration) *Span {
	s.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return s.eft()
}

// Int adds a int key/value attribute to the current Span
func (s *Span) Int(k *xopconst.IntAttribute, v int) *Span {
	s.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return s.eft()
}

// Int16 adds a int16 key/value attribute to the current Span
func (s *Span) Int16(k *xopconst.Int16Attribute, v int16) *Span {
	s.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return s.eft()
}

// Int32 adds a int32 key/value attribute to the current Span
func (s *Span) Int32(k *xopconst.Int32Attribute, v int32) *Span {
	s.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return s.eft()
}

// Int8 adds a int8 key/value attribute to the current Span
func (s *Span) Int8(k *xopconst.Int8Attribute, v int8) *Span {
	s.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return s.eft()
}
