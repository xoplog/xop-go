// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopbase

// DataTypeToString provides mapping from DataType to short,
// human-readable type strings.
var DataTypeToString = map[DataType]string{
	AnyDataType:      "any",
	BoolDataType:     "bool",
	DurationDataType: "dur",
	EnumDataType:     "enum",
	ErrorDataType:    "error",
	Float32DataType:  "f32",
	Float64DataType:  "f64",
	IntDataType:      "i",
	Int16DataType:    "i16",
	Int32DataType:    "i32",
	Int64DataType:    "i64",
	Int8DataType:     "i8",
	StringDataType:   "s",
	StringerDataType: "stringer",
	TimeDataType:     "time",
	UintDataType:     "u",
	Uint16DataType:   "u16",
	Uint32DataType:   "u32",
	Uint64DataType:   "u64",
	Uint8DataType:    "u8",
	UintptrDataType:  "uintptr",
}

// StringToDataType reverses DataTypeToString
var StringToDataType = map[string]DataType{
	"any":      AnyDataType,
	"bool":     BoolDataType,
	"dur":      DurationDataType,
	"enum":     EnumDataType,
	"error":    ErrorDataType,
	"f32":      Float32DataType,
	"f64":      Float64DataType,
	"i":        IntDataType,
	"i16":      Int16DataType,
	"i32":      Int32DataType,
	"i64":      Int64DataType,
	"i8":       Int8DataType,
	"s":        StringDataType,
	"stringer": StringerDataType,
	"time":     TimeDataType,
	"u":        UintDataType,
	"u16":      Uint16DataType,
	"u32":      Uint32DataType,
	"u64":      Uint64DataType,
	"u8":       Uint8DataType,
	"uintptr":  UintptrDataType,
}

const (
	AnyDataTypeAbbr      = "any"
	BoolDataTypeAbbr     = "bool"
	DurationDataTypeAbbr = "dur"
	EnumDataTypeAbbr     = "enum"
	ErrorDataTypeAbbr    = "error"
	Float32DataTypeAbbr  = "f32"
	Float64DataTypeAbbr  = "f64"
	IntDataTypeAbbr      = "i"
	Int16DataTypeAbbr    = "i16"
	Int32DataTypeAbbr    = "i32"
	Int64DataTypeAbbr    = "i64"
	Int8DataTypeAbbr     = "i8"
	StringDataTypeAbbr   = "s"
	StringerDataTypeAbbr = "stringer"
	TimeDataTypeAbbr     = "time"
	UintDataTypeAbbr     = "u"
	Uint16DataTypeAbbr   = "u16"
	Uint32DataTypeAbbr   = "u32"
	Uint64DataTypeAbbr   = "u64"
	Uint8DataTypeAbbr    = "u8"
	UintptrDataTypeAbbr  = "uintptr"
)
