// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson_test

import "github.com/muir/xop-go/xopconst"

// TODO: why the skips?
var ExampleMetadataBool = xopconst.Make{Key: "ebool", Namespace: "test"}.BoolAttribute()

var (
	ExampleMetadataDuration = xopconst.Make{Key: "etime.Duration", Namespace: "test"}.DurationAttribute()
	ExampleMetadataInt      = xopconst.Make{Key: "eint", Namespace: "test"}.IntAttribute()
	ExampleMetadataInt16    = xopconst.Make{Key: "eint16", Namespace: "test"}.Int16Attribute()
	ExampleMetadataInt32    = xopconst.Make{Key: "eint32", Namespace: "test"}.Int32Attribute()
	ExampleMetadataInt64    = xopconst.Make{Key: "eint64", Namespace: "test"}.Int64Attribute()
	ExampleMetadataInt8     = xopconst.Make{Key: "eint8", Namespace: "test"}.Int8Attribute()
	ExampleMetadataLink     = xopconst.Make{Key: "etrace.Trace", Namespace: "test"}.LinkAttribute()
	ExampleMetadataString   = xopconst.Make{Key: "estring", Namespace: "test"}.StringAttribute()
	ExampleMetadataTime     = xopconst.Make{Key: "etime.Time", Namespace: "test"}.TimeAttribute()
)
