// Code generated by "enumer -type=DataType -linecomment -json -sql"; DO NOT EDIT.

package xopbase

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

const _DataTypeName = "EnumDataTypeEnumArrayDataTypeAnyDataTypeBoolDataTypeDurationDataTypeErrorDataTypeFloat32DataTypeFloat64DataTypeIntDataTypeInt16DataTypeInt32DataTypeInt64DataTypeInt8DataTypeLinkDataTypeStringDataTypeStringerDataTypeTimeDataTypeUintDataTypeUint16DataTypeUint32DataTypeUint64DataTypeUint8DataTypeUintptrDataTypeAnyArrayDataTypeBoolArrayDataTypeDurationArrayDataTypeErrorArrayDataTypeFloat32ArrayDataTypeFloat64ArrayDataTypeIntArrayDataTypeInt16ArrayDataTypeInt32ArrayDataTypeInt64ArrayDataTypeInt8ArrayDataTypeLinkArrayDataTypeStringArrayDataTypeStringerArrayDataTypeTimeArrayDataTypeUintArrayDataTypeUint16ArrayDataTypeUint32ArrayDataTypeUint64ArrayDataTypeUint8ArrayDataTypeUintptrArrayDataType"

var _DataTypeIndex = [...]uint16{0, 12, 29, 40, 52, 68, 81, 96, 111, 122, 135, 148, 161, 173, 185, 199, 215, 227, 239, 253, 267, 281, 294, 309, 325, 342, 363, 381, 401, 421, 437, 455, 473, 491, 508, 525, 544, 565, 582, 599, 618, 637, 656, 674, 694}

const _DataTypeLowerName = "enumdatatypeenumarraydatatypeanydatatypebooldatatypedurationdatatypeerrordatatypefloat32datatypefloat64datatypeintdatatypeint16datatypeint32datatypeint64datatypeint8datatypelinkdatatypestringdatatypestringerdatatypetimedatatypeuintdatatypeuint16datatypeuint32datatypeuint64datatypeuint8datatypeuintptrdatatypeanyarraydatatypeboolarraydatatypedurationarraydatatypeerrorarraydatatypefloat32arraydatatypefloat64arraydatatypeintarraydatatypeint16arraydatatypeint32arraydatatypeint64arraydatatypeint8arraydatatypelinkarraydatatypestringarraydatatypestringerarraydatatypetimearraydatatypeuintarraydatatypeuint16arraydatatypeuint32arraydatatypeuint64arraydatatypeuint8arraydatatypeuintptrarraydatatype"

func (i DataType) String() string {
	if i < 0 || i >= DataType(len(_DataTypeIndex)-1) {
		return fmt.Sprintf("DataType(%d)", i)
	}
	return _DataTypeName[_DataTypeIndex[i]:_DataTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _DataTypeNoOp() {
	var x [1]struct{}
	_ = x[EnumDataType-(0)]
	_ = x[EnumArrayDataType-(1)]
	_ = x[AnyDataType-(2)]
	_ = x[BoolDataType-(3)]
	_ = x[DurationDataType-(4)]
	_ = x[ErrorDataType-(5)]
	_ = x[Float32DataType-(6)]
	_ = x[Float64DataType-(7)]
	_ = x[IntDataType-(8)]
	_ = x[Int16DataType-(9)]
	_ = x[Int32DataType-(10)]
	_ = x[Int64DataType-(11)]
	_ = x[Int8DataType-(12)]
	_ = x[LinkDataType-(13)]
	_ = x[StringDataType-(14)]
	_ = x[StringerDataType-(15)]
	_ = x[TimeDataType-(16)]
	_ = x[UintDataType-(17)]
	_ = x[Uint16DataType-(18)]
	_ = x[Uint32DataType-(19)]
	_ = x[Uint64DataType-(20)]
	_ = x[Uint8DataType-(21)]
	_ = x[UintptrDataType-(22)]
	_ = x[AnyArrayDataType-(23)]
	_ = x[BoolArrayDataType-(24)]
	_ = x[DurationArrayDataType-(25)]
	_ = x[ErrorArrayDataType-(26)]
	_ = x[Float32ArrayDataType-(27)]
	_ = x[Float64ArrayDataType-(28)]
	_ = x[IntArrayDataType-(29)]
	_ = x[Int16ArrayDataType-(30)]
	_ = x[Int32ArrayDataType-(31)]
	_ = x[Int64ArrayDataType-(32)]
	_ = x[Int8ArrayDataType-(33)]
	_ = x[LinkArrayDataType-(34)]
	_ = x[StringArrayDataType-(35)]
	_ = x[StringerArrayDataType-(36)]
	_ = x[TimeArrayDataType-(37)]
	_ = x[UintArrayDataType-(38)]
	_ = x[Uint16ArrayDataType-(39)]
	_ = x[Uint32ArrayDataType-(40)]
	_ = x[Uint64ArrayDataType-(41)]
	_ = x[Uint8ArrayDataType-(42)]
	_ = x[UintptrArrayDataType-(43)]
}

var _DataTypeValues = []DataType{EnumDataType, EnumArrayDataType, AnyDataType, BoolDataType, DurationDataType, ErrorDataType, Float32DataType, Float64DataType, IntDataType, Int16DataType, Int32DataType, Int64DataType, Int8DataType, LinkDataType, StringDataType, StringerDataType, TimeDataType, UintDataType, Uint16DataType, Uint32DataType, Uint64DataType, Uint8DataType, UintptrDataType, AnyArrayDataType, BoolArrayDataType, DurationArrayDataType, ErrorArrayDataType, Float32ArrayDataType, Float64ArrayDataType, IntArrayDataType, Int16ArrayDataType, Int32ArrayDataType, Int64ArrayDataType, Int8ArrayDataType, LinkArrayDataType, StringArrayDataType, StringerArrayDataType, TimeArrayDataType, UintArrayDataType, Uint16ArrayDataType, Uint32ArrayDataType, Uint64ArrayDataType, Uint8ArrayDataType, UintptrArrayDataType}

var _DataTypeNameToValueMap = map[string]DataType{
	_DataTypeName[0:12]:         EnumDataType,
	_DataTypeLowerName[0:12]:    EnumDataType,
	_DataTypeName[12:29]:        EnumArrayDataType,
	_DataTypeLowerName[12:29]:   EnumArrayDataType,
	_DataTypeName[29:40]:        AnyDataType,
	_DataTypeLowerName[29:40]:   AnyDataType,
	_DataTypeName[40:52]:        BoolDataType,
	_DataTypeLowerName[40:52]:   BoolDataType,
	_DataTypeName[52:68]:        DurationDataType,
	_DataTypeLowerName[52:68]:   DurationDataType,
	_DataTypeName[68:81]:        ErrorDataType,
	_DataTypeLowerName[68:81]:   ErrorDataType,
	_DataTypeName[81:96]:        Float32DataType,
	_DataTypeLowerName[81:96]:   Float32DataType,
	_DataTypeName[96:111]:       Float64DataType,
	_DataTypeLowerName[96:111]:  Float64DataType,
	_DataTypeName[111:122]:      IntDataType,
	_DataTypeLowerName[111:122]: IntDataType,
	_DataTypeName[122:135]:      Int16DataType,
	_DataTypeLowerName[122:135]: Int16DataType,
	_DataTypeName[135:148]:      Int32DataType,
	_DataTypeLowerName[135:148]: Int32DataType,
	_DataTypeName[148:161]:      Int64DataType,
	_DataTypeLowerName[148:161]: Int64DataType,
	_DataTypeName[161:173]:      Int8DataType,
	_DataTypeLowerName[161:173]: Int8DataType,
	_DataTypeName[173:185]:      LinkDataType,
	_DataTypeLowerName[173:185]: LinkDataType,
	_DataTypeName[185:199]:      StringDataType,
	_DataTypeLowerName[185:199]: StringDataType,
	_DataTypeName[199:215]:      StringerDataType,
	_DataTypeLowerName[199:215]: StringerDataType,
	_DataTypeName[215:227]:      TimeDataType,
	_DataTypeLowerName[215:227]: TimeDataType,
	_DataTypeName[227:239]:      UintDataType,
	_DataTypeLowerName[227:239]: UintDataType,
	_DataTypeName[239:253]:      Uint16DataType,
	_DataTypeLowerName[239:253]: Uint16DataType,
	_DataTypeName[253:267]:      Uint32DataType,
	_DataTypeLowerName[253:267]: Uint32DataType,
	_DataTypeName[267:281]:      Uint64DataType,
	_DataTypeLowerName[267:281]: Uint64DataType,
	_DataTypeName[281:294]:      Uint8DataType,
	_DataTypeLowerName[281:294]: Uint8DataType,
	_DataTypeName[294:309]:      UintptrDataType,
	_DataTypeLowerName[294:309]: UintptrDataType,
	_DataTypeName[309:325]:      AnyArrayDataType,
	_DataTypeLowerName[309:325]: AnyArrayDataType,
	_DataTypeName[325:342]:      BoolArrayDataType,
	_DataTypeLowerName[325:342]: BoolArrayDataType,
	_DataTypeName[342:363]:      DurationArrayDataType,
	_DataTypeLowerName[342:363]: DurationArrayDataType,
	_DataTypeName[363:381]:      ErrorArrayDataType,
	_DataTypeLowerName[363:381]: ErrorArrayDataType,
	_DataTypeName[381:401]:      Float32ArrayDataType,
	_DataTypeLowerName[381:401]: Float32ArrayDataType,
	_DataTypeName[401:421]:      Float64ArrayDataType,
	_DataTypeLowerName[401:421]: Float64ArrayDataType,
	_DataTypeName[421:437]:      IntArrayDataType,
	_DataTypeLowerName[421:437]: IntArrayDataType,
	_DataTypeName[437:455]:      Int16ArrayDataType,
	_DataTypeLowerName[437:455]: Int16ArrayDataType,
	_DataTypeName[455:473]:      Int32ArrayDataType,
	_DataTypeLowerName[455:473]: Int32ArrayDataType,
	_DataTypeName[473:491]:      Int64ArrayDataType,
	_DataTypeLowerName[473:491]: Int64ArrayDataType,
	_DataTypeName[491:508]:      Int8ArrayDataType,
	_DataTypeLowerName[491:508]: Int8ArrayDataType,
	_DataTypeName[508:525]:      LinkArrayDataType,
	_DataTypeLowerName[508:525]: LinkArrayDataType,
	_DataTypeName[525:544]:      StringArrayDataType,
	_DataTypeLowerName[525:544]: StringArrayDataType,
	_DataTypeName[544:565]:      StringerArrayDataType,
	_DataTypeLowerName[544:565]: StringerArrayDataType,
	_DataTypeName[565:582]:      TimeArrayDataType,
	_DataTypeLowerName[565:582]: TimeArrayDataType,
	_DataTypeName[582:599]:      UintArrayDataType,
	_DataTypeLowerName[582:599]: UintArrayDataType,
	_DataTypeName[599:618]:      Uint16ArrayDataType,
	_DataTypeLowerName[599:618]: Uint16ArrayDataType,
	_DataTypeName[618:637]:      Uint32ArrayDataType,
	_DataTypeLowerName[618:637]: Uint32ArrayDataType,
	_DataTypeName[637:656]:      Uint64ArrayDataType,
	_DataTypeLowerName[637:656]: Uint64ArrayDataType,
	_DataTypeName[656:674]:      Uint8ArrayDataType,
	_DataTypeLowerName[656:674]: Uint8ArrayDataType,
	_DataTypeName[674:694]:      UintptrArrayDataType,
	_DataTypeLowerName[674:694]: UintptrArrayDataType,
}

var _DataTypeNames = []string{
	_DataTypeName[0:12],
	_DataTypeName[12:29],
	_DataTypeName[29:40],
	_DataTypeName[40:52],
	_DataTypeName[52:68],
	_DataTypeName[68:81],
	_DataTypeName[81:96],
	_DataTypeName[96:111],
	_DataTypeName[111:122],
	_DataTypeName[122:135],
	_DataTypeName[135:148],
	_DataTypeName[148:161],
	_DataTypeName[161:173],
	_DataTypeName[173:185],
	_DataTypeName[185:199],
	_DataTypeName[199:215],
	_DataTypeName[215:227],
	_DataTypeName[227:239],
	_DataTypeName[239:253],
	_DataTypeName[253:267],
	_DataTypeName[267:281],
	_DataTypeName[281:294],
	_DataTypeName[294:309],
	_DataTypeName[309:325],
	_DataTypeName[325:342],
	_DataTypeName[342:363],
	_DataTypeName[363:381],
	_DataTypeName[381:401],
	_DataTypeName[401:421],
	_DataTypeName[421:437],
	_DataTypeName[437:455],
	_DataTypeName[455:473],
	_DataTypeName[473:491],
	_DataTypeName[491:508],
	_DataTypeName[508:525],
	_DataTypeName[525:544],
	_DataTypeName[544:565],
	_DataTypeName[565:582],
	_DataTypeName[582:599],
	_DataTypeName[599:618],
	_DataTypeName[618:637],
	_DataTypeName[637:656],
	_DataTypeName[656:674],
	_DataTypeName[674:694],
}

// DataTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func DataTypeString(s string) (DataType, error) {
	if val, ok := _DataTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _DataTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to DataType values", s)
}

// DataTypeValues returns all values of the enum
func DataTypeValues() []DataType {
	return _DataTypeValues
}

// DataTypeStrings returns a slice of all String values of the enum
func DataTypeStrings() []string {
	strs := make([]string, len(_DataTypeNames))
	copy(strs, _DataTypeNames)
	return strs
}

// IsADataType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i DataType) IsADataType() bool {
	for _, v := range _DataTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for DataType
func (i DataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for DataType
func (i *DataType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DataType should be a string, got %s", data)
	}

	var err error
	*i, err = DataTypeString(s)
	return err
}

func (i DataType) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *DataType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	case fmt.Stringer:
		str = v.String()
	default:
		return fmt.Errorf("invalid value of DataType: %[1]T(%[1]v)", value)
	}

	val, err := DataTypeString(str)
	if err != nil {
		return err
	}

	*i = val
	return nil
}
