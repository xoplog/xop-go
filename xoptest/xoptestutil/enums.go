// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptestutil

import (
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopconst"
)

type AnyObject struct {
	I int
	S string
	A []string
	P *AnyObject
}

var (
	ExampleMetadataSingleAny   = xopat.Make{Key: "s-any", Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataLockedAny   = xopat.Make{Key: "l-any", Locked: true, Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataMultipleAny = xopat.Make{Key: "m-any", Multiple: true, Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataDistinctAny = xopat.Make{Key: "d-any", Multiple: true, Distinct: true, Namespace: "test"}.AnyAttribute(AnyObject{})
)

var (
	ExampleMetadataSingleEnum   = xopat.Make{Key: "s-ienum", Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataLockedEnum   = xopat.Make{Key: "l-ienum", Locked: true, Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataMultipleEnum = xopat.Make{Key: "m-ienum", Multiple: true, Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataDistinctEnum = xopat.Make{Key: "d-ienum", Multiple: true, Distinct: true, Namespace: "test"}.IotaEnumAttribute()
)

var (
	SingleEnumOne   = ExampleMetadataSingleEnum.Iota("one")
	SingleEnumTwo   = ExampleMetadataSingleEnum.Iota("two")
	SingleEnumThree = ExampleMetadataSingleEnum.Iota("Three")

	LockedEnumOne   = ExampleMetadataLockedEnum.Iota("one")
	LockedEnumTwo   = ExampleMetadataLockedEnum.Iota("two")
	LockedEnumThree = ExampleMetadataLockedEnum.Iota("Three")

	MultipleEnumOne   = ExampleMetadataMultipleEnum.Iota("one")
	MultipleEnumTwo   = ExampleMetadataMultipleEnum.Iota("two")
	MultipleEnumThree = ExampleMetadataMultipleEnum.Iota("Three")

	DistinctEnumOne   = ExampleMetadataDistinctEnum.Iota("one")
	DistinctEnumTwo   = ExampleMetadataDistinctEnum.Iota("two")
	DistinctEnumThree = ExampleMetadataDistinctEnum.Iota("Three")
)

var (
	ExampleMetadataSingleEEnum   = xopat.Make{Key: "s-eenum", Namespace: "test"}.EmbeddedEnumAttribute(threeType(1))
	ExampleMetadataLockedEEnum   = xopat.Make{Key: "l-eenum", Locked: true, Namespace: "test"}.EmbeddedEnumAttribute(1)
	ExampleMetadataMultipleEEnum = xopat.Make{Key: "m-eenum", Multiple: true, Namespace: "test"}.EmbeddedEnumAttribute(int64(1))
	ExampleMetadataDistinctEEnum = xopat.Make{Key: "d-eenum", Multiple: true, Distinct: true, Namespace: "test"}.EmbeddedEnumAttribute(threeType(1))
)

type threeType int

func (t threeType) String() string {
	switch t {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	default:
		return "other"
	}
}

var (
	SingleEEnumOne   = ExampleMetadataSingleEEnum.AddStringer(threeType(1))
	SingleEEnumTwo   = ExampleMetadataSingleEEnum.AddStringer(threeType(2))
	SingleEEnumThree = ExampleMetadataSingleEEnum.AddStringer(threeType(3))

	LockedEEnumOne   = ExampleMetadataLockedEEnum.Add(1, "one")
	LockedEEnumTwo   = ExampleMetadataLockedEEnum.Add(2, "two")
	LockedEEnumThree = ExampleMetadataLockedEEnum.Add(3, "three")

	MultipleEEnumOne   = ExampleMetadataMultipleEEnum.Add64(int64(1), "one")
	MultipleEEnumTwo   = ExampleMetadataMultipleEEnum.Add64(int64(2), "two")
	MultipleEEnumThree = ExampleMetadataMultipleEEnum.Add64(int64(3), "three")

	DistinctEEnumOne   = ExampleMetadataDistinctEEnum.AddStringer(threeType(1))
	DistinctEEnumTwo   = ExampleMetadataDistinctEEnum.AddStringer(threeType(2))
	DistinctEEnumThree = ExampleMetadataDistinctEEnum.AddStringer(threeType(3))
)

var (
	ExampleMetadataSingleXEnum   = xopat.Make{Key: "s-xenum", Namespace: "test"}.EnumAttribute(xopconst.SpanKindServer)
	ExampleMetadataLockedXEnum   = xopat.Make{Key: "l-xenum", Locked: true, Namespace: "test"}.EnumAttribute(xopconst.SpanKindServer)
	ExampleMetadataMultipleXEnum = xopat.Make{Key: "m-xenum", Multiple: true, Namespace: "test"}.EnumAttribute(xopconst.SpanKindServer)
	ExampleMetadataDistinctXEnum = xopat.Make{Key: "d-xenum", Multiple: true, Distinct: true, Namespace: "test"}.EnumAttribute(xopconst.SpanKindServer)
)

var (
	ExampleMetadataSingleBool       = xopat.Make{Key: "s-bool", Namespace: "test"}.BoolAttribute()
	ExampleMetadataLockedBool       = xopat.Make{Key: "l-bool", Locked: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataMultipleBool     = xopat.Make{Key: "m-bool", Multiple: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataDistinctBool     = xopat.Make{Key: "d-bool", Multiple: true, Distinct: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataSingleDuration   = xopat.Make{Key: "s-time.Duration", Namespace: "test"}.DurationAttribute()
	ExampleMetadataLockedDuration   = xopat.Make{Key: "l-time.Duration", Locked: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataMultipleDuration = xopat.Make{Key: "m-time.Duration", Multiple: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataDistinctDuration = xopat.Make{Key: "d-time.Duration", Multiple: true, Distinct: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataSingleFloat32    = xopat.Make{Key: "s-float32", Namespace: "test"}.Float32Attribute()
	ExampleMetadataLockedFloat32    = xopat.Make{Key: "l-float32", Locked: true, Namespace: "test"}.Float32Attribute()
	ExampleMetadataMultipleFloat32  = xopat.Make{Key: "m-float32", Multiple: true, Namespace: "test"}.Float32Attribute()
	ExampleMetadataDistinctFloat32  = xopat.Make{Key: "d-float32", Multiple: true, Distinct: true, Namespace: "test"}.Float32Attribute()
	ExampleMetadataSingleFloat64    = xopat.Make{Key: "s-float64", Namespace: "test"}.Float64Attribute()
	ExampleMetadataLockedFloat64    = xopat.Make{Key: "l-float64", Locked: true, Namespace: "test"}.Float64Attribute()
	ExampleMetadataMultipleFloat64  = xopat.Make{Key: "m-float64", Multiple: true, Namespace: "test"}.Float64Attribute()
	ExampleMetadataDistinctFloat64  = xopat.Make{Key: "d-float64", Multiple: true, Distinct: true, Namespace: "test"}.Float64Attribute()
	ExampleMetadataSingleInt        = xopat.Make{Key: "s-int", Namespace: "test"}.IntAttribute()
	ExampleMetadataLockedInt        = xopat.Make{Key: "l-int", Locked: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataMultipleInt      = xopat.Make{Key: "m-int", Multiple: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataDistinctInt      = xopat.Make{Key: "d-int", Multiple: true, Distinct: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataSingleInt16      = xopat.Make{Key: "s-int16", Namespace: "test"}.Int16Attribute()
	ExampleMetadataLockedInt16      = xopat.Make{Key: "l-int16", Locked: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataMultipleInt16    = xopat.Make{Key: "m-int16", Multiple: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataDistinctInt16    = xopat.Make{Key: "d-int16", Multiple: true, Distinct: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataSingleInt32      = xopat.Make{Key: "s-int32", Namespace: "test"}.Int32Attribute()
	ExampleMetadataLockedInt32      = xopat.Make{Key: "l-int32", Locked: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataMultipleInt32    = xopat.Make{Key: "m-int32", Multiple: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataDistinctInt32    = xopat.Make{Key: "d-int32", Multiple: true, Distinct: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataSingleInt64      = xopat.Make{Key: "s-int64", Namespace: "test"}.Int64Attribute()
	ExampleMetadataLockedInt64      = xopat.Make{Key: "l-int64", Locked: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataMultipleInt64    = xopat.Make{Key: "m-int64", Multiple: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataDistinctInt64    = xopat.Make{Key: "d-int64", Multiple: true, Distinct: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataSingleInt8       = xopat.Make{Key: "s-int8", Namespace: "test"}.Int8Attribute()
	ExampleMetadataLockedInt8       = xopat.Make{Key: "l-int8", Locked: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataMultipleInt8     = xopat.Make{Key: "m-int8", Multiple: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataDistinctInt8     = xopat.Make{Key: "d-int8", Multiple: true, Distinct: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataSingleString     = xopat.Make{Key: "s-string", Namespace: "test"}.StringAttribute()
	ExampleMetadataLockedString     = xopat.Make{Key: "l-string", Locked: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataMultipleString   = xopat.Make{Key: "m-string", Multiple: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataDistinctString   = xopat.Make{Key: "d-string", Multiple: true, Distinct: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataSingleTime       = xopat.Make{Key: "s-time.Time", Namespace: "test"}.TimeAttribute()
	ExampleMetadataLockedTime       = xopat.Make{Key: "l-time.Time", Locked: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataMultipleTime     = xopat.Make{Key: "m-time.Time", Multiple: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataDistinctTime     = xopat.Make{Key: "d-time.Time", Multiple: true, Distinct: true, Namespace: "test"}.TimeAttribute()
)
