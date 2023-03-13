package xopbase

// DataTypeToString provides mapping from DataType to short,
// human-readable type strings.
var DataTypeToString = func() map[DataType]string {
	m := make(map[DataType]string)
	for k, v := range StringToDataType {
		m[v] = k
	}
	return m
}()

// StringToDataType reverses DataTypeToString
var StringToDataType = map[string]DataType{
	"i":        IntDataType,
	"i8":       Int8DataType,
	"i16":      Int16DataType,
	"i32":      Int32DataType,
	"i64":      Int64DataType,
	"u":        UintDataType,
	"u8":       Uint8DataType,
	"u16":      Uint16DataType,
	"u32":      Uint32DataType,
	"u64":      Uint64DataType,
	"uintptr":  UintptrDataType,
	"f32":      Float32DataType,
	"f64":      Float64DataType,
	"any":      AnyDataType,
	"bool":     BoolDataType,
	"dur":      DurationDataType,
	"time":     TimeDataType,
	"s":        StringDataType,
	"stringer": StringerDataType,
	"enum":     EnumDataType,
	"error":    ErrorDataType,
}
