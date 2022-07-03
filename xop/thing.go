package xop

import "time"

type Type int

const (
	UnsetType Type = iota
	IntType
	UintType
	BoolType
	StringType
	TimeType
	AnyType
	ErrorType
)

// Thing is heavily influcenced by Uber's zapcore.Field
type Thing struct {
	Key    string
	Type   Type
	Int    int64
	String string
	Any    interface{}
}

type Things struct {
	Things []Thing
}

func Int64(k string, v int64) Thing              { return Thing{Key: k, Type: xop.IntType, Int: v} }
func Int32(k string, v int32) Thing              { return Thing{Key: k, Type: xop.IntType, Int: int64(v)} }
func Int16(k string, v int16) Thing              { return Thing{Key: k, Type: xop.IntType, Int: int64(v)} }
func Int8(k string, v int8) Thing                { return Thing{Key: k, Type: xop.IntType, Int: int64(v)} }
func Int(k string, v int) Thing                  { return Thing{Key: k, Type: xop.IntType, Int: int64(v)} }
func Uint64(k string, v uint64) Thing            { return Thing{Key: k, Type: xop.UintType, Any: v} }
func Uint32(k string, v uint32) Thing            { return Thing{Key: k, Type: xop.UintType, Any: uint64(v)} }
func Uint16(k string, v uint16) Thing            { return Thing{Key: k, Type: xop.UintType, Any: uint64(v)} }
func Uint8(k string, v uint8) Thing              { return Thing{Key: k, Type: xop.UintType, Any: uint64(v)} }
func Uint(k string, v uint) Thing                { return Thing{Key: k, Type: xop.UintType, Any: uint64(v)} }
func Bool(k string, v bool) Thing                { return Thing{Key: k, Type: xop.BoolType, Any: v} }
func Str(k string, v bool) Thing                 { return Thing{Key: k, Type: xop.StringType, String: v} }
func Time(k string, v time.Time) Thing           { return Thing{Key: k, Type: xop.TimeType, Any: v} }
func AnyImmutable(k string, v interface{}) Thing { return Thing{Key: k, Type: xop.AnyType, Any: v} }
func Error(k string, v error) Thing              { return Thing{Key: k, Type: xop.ErrorType, Any: v} }

// TODO: make a copy
func Any(k string, v interface{}) Thing { return Thing{Key: k, Type: xop.AnyType, Any: v} }

// TODO: make a copy
func MutableError(k string, v error) Thing { return Thing{Key: k, Type: xop.ErrorType, Any: v} }

func (t *Things) Int(k string, v int64)      { t.Things = append(t.Things, Int64(k, v)) }
func (t *Things) Uint(k string, v uint64)    { t.Things = append(t.Things, Uint64(k, v)) }
func (t *Things) Bool(k string, v bool)      { t.Things = append(t.Things, Bool(k, v)) }
func (t *Things) Str(k string, v string)     { t.Things = append(t.Things, Str(k, v)) }
func (t *Things) Time(k string, v time.Time) { t.Things = append(t.Things, Time(k, v)) }
func (t *Things) Error(k string, v error)    { t.Things = append(t.Things, Error(k, v)) }
func (t *Things) AnyImmutable(k string, v interface{}) {
	t.Things = append(t.Things, AnyImmutable(k, v))
}

// TODO: func (t *Things) SubObject(string) SubObject
// TODO: func (t *Things) Encoded(name string, elementName string, encoder Encoder, data interface{})
// TODO: func (t *Things) PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
// TODO: func (t *Things) ExternalReference(name string, itemId string, storageId string)
// TODO: func (t *Things) PreEncodedText(name string, elementName string, mimeType string, data string)
