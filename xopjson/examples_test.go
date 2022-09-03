// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson_test

import "github.com/muir/xop-go/xopat"

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
	ExampleMetadataSingleEnum   = xopat.Make{Key: "s-enum", Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataLockedEnum   = xopat.Make{Key: "l-enum", Locked: true, Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataMultipleEnum = xopat.Make{Key: "m-enum", Multiple: true, Namespace: "test"}.IotaEnumAttribute()
	ExampleMetadataDistinctEnum = xopat.Make{Key: "d-enum", Multiple: true, Distinct: true, Namespace: "test"}.IotaEnumAttribute()
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

// TODO: why the skips?
var ExampleMetadataSingleBool = xopat.Make{Key: "s-bool", Namespace: "test"}.BoolAttribute()

var (
	ExampleMetadataLockedBool       = xopat.Make{Key: "l-bool", Locked: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataMultipleBool     = xopat.Make{Key: "m-bool", Multiple: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataDistinctBool     = xopat.Make{Key: "d-bool", Multiple: true, Distinct: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataSingleDuration   = xopat.Make{Key: "s-time.Duration", Namespace: "test"}.DurationAttribute()
	ExampleMetadataLockedDuration   = xopat.Make{Key: "l-time.Duration", Locked: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataMultipleDuration = xopat.Make{Key: "m-time.Duration", Multiple: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataDistinctDuration = xopat.Make{Key: "d-time.Duration", Multiple: true, Distinct: true, Namespace: "test"}.DurationAttribute()
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
	ExampleMetadataSingleLink       = xopat.Make{Key: "s-trace.Trace", Namespace: "test"}.LinkAttribute()
	ExampleMetadataLockedLink       = xopat.Make{Key: "l-trace.Trace", Locked: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataMultipleLink     = xopat.Make{Key: "m-trace.Trace", Multiple: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataDistinctLink     = xopat.Make{Key: "d-trace.Trace", Multiple: true, Distinct: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataSingleString     = xopat.Make{Key: "s-string", Namespace: "test"}.StringAttribute()
	ExampleMetadataLockedString     = xopat.Make{Key: "l-string", Locked: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataMultipleString   = xopat.Make{Key: "m-string", Multiple: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataDistinctString   = xopat.Make{Key: "d-string", Multiple: true, Distinct: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataSingleTime       = xopat.Make{Key: "s-time.Time", Namespace: "test"}.TimeAttribute()
	ExampleMetadataLockedTime       = xopat.Make{Key: "l-time.Time", Locked: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataMultipleTime     = xopat.Make{Key: "m-time.Time", Multiple: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataDistinctTime     = xopat.Make{Key: "d-time.Time", Multiple: true, Distinct: true, Namespace: "test"}.TimeAttribute()
)
