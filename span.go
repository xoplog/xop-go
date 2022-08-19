// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"

	"github.com/mohae/deepcopy"
)

// Request provides access to the span that describes the overall
// request. Metadata may be added at the request level.
func (log *Log) Request() *Span {
	return log.request.capSpan
}

// Request provides access to the current span
// Metadata may be added at the span level.
func (log *Log) Span() *Span {
	return log.capSpan
}

func (span *Span) TraceState() trace.State     { return span.seed.traceBundle.State }
func (span *Span) TraceBaggage() trace.Baggage { return span.seed.traceBundle.Baggage }
func (span *Span) TraceParent() trace.Trace    { return span.seed.traceBundle.TraceParent.Copy() }
func (span *Span) Trace() trace.Trace          { return span.seed.traceBundle.Trace.Copy() }
func (span *Span) Bundle() trace.Bundle        { return span.seed.traceBundle.Copy() }

func (span *Span) eft() *Span {
	span.log.hasActivity(true)
	return span
}

// Int64 adds a int64 key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Int64(k *xopconst.Int64Attribute, v int64) *Span {
	span.base.MetadataInt64(k, v)
	return span.eft()
}

// EmbeddedEnum adds a kev/value attribute to the Span.  The key and the value
// are bundled together: the key is derrived from the type of the Enum.
// Alternatively, use xopconst.KeyedEnumAttribute() to create functions
// to add enum key/value pairs where the key and value are specified
// separately.
// The return value does not need to be used.
func (span *Span) EmbeddedEnum(kv xopconst.EmbeddedEnum) *Span {
	return span.Enum(kv.EnumAttribute(), kv)
}

// AnyImmutable adds a key/value attribute to the current Span.  The provided
// value must be immutable.  If it is not, then there could be race conditions
// or the value that ends up logged could be different from the value at the
// time when AnyImmutible was called.
//
// While the AnyAttribute has an expectation
// for the type of the value, that type may or may not be checked depending
// on the base logger being used.
// The return value does not need to be used.
func (span *Span) AnyImmutable(k *xopconst.AnyAttribute, v interface{}) *Span {
	span.base.MetadataAny(k, v)
	return span.eft()
}

// Any adds a key/value attribute to the current Span.  The provided
// value may be copied using github.com/mohae/deepcopy if any of the
// base loggers hold the value instead of immediately serializing it.
// While the AnyAttribute has an expectation
// for the type of the value, that type may or may not be checked depending
// on the base logger being used.
// The return value does not need to be used.
func (span *Span) Any(k *xopconst.AnyAttribute, v interface{}) *Span {
	if span.log.span.referencesKept {
		v = deepcopy.Copy(v)
	}
	span.base.MetadataAny(k, v)
	return span.eft()
}

// Bool adds a bool key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Bool(k *xopconst.BoolAttribute, v bool) *Span {
	span.base.MetadataBool(k, v)
	return span.eft()
}

// Enum adds a xopconst.Enum key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) *Span {
	span.base.MetadataEnum(k, v)
	return span.eft()
}

// Float64 adds a float64 key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Float64(k *xopconst.Float64Attribute, v float64) *Span {
	span.base.MetadataFloat64(k, v)
	return span.eft()
}

// Link adds a trace.Trace key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Link(k *xopconst.LinkAttribute, v trace.Trace) *Span {
	span.base.MetadataLink(k, v)
	return span.eft()
}

// String adds a string key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) String(k *xopconst.StringAttribute, v string) *Span {
	span.base.MetadataString(k, v)
	return span.eft()
}

// Time adds a time.Time key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Time(k *xopconst.TimeAttribute, v time.Time) *Span {
	span.base.MetadataTime(k, v)
	return span.eft()
}

// should skip Int64
// Duration adds a time.Duration key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Duration(k *xopconst.DurationAttribute, v time.Duration) *Span {
	span.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return span.eft()
}

// Int adds a int key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Int(k *xopconst.IntAttribute, v int) *Span {
	span.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return span.eft()
}

// Int16 adds a int16 key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Int16(k *xopconst.Int16Attribute, v int16) *Span {
	span.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return span.eft()
}

// Int32 adds a int32 key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Int32(k *xopconst.Int32Attribute, v int32) *Span {
	span.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return span.eft()
}

// Int8 adds a int8 key/value attribute to the current Span.
// The return value does not need to be used.
func (span *Span) Int8(k *xopconst.Int8Attribute, v int8) *Span {
	span.base.MetadataInt64(&k.Int64Attribute, int64(v))
	return span.eft()
}
