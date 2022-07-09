package xop

import (
	"time"
)

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

func Int64(k string, v int64) Thing              { return Thing{Key: k, Type: IntType, Int: v} }
func Int32(k string, v int32) Thing              { return Thing{Key: k, Type: IntType, Int: int64(v)} }
func Int16(k string, v int16) Thing              { return Thing{Key: k, Type: IntType, Int: int64(v)} }
func Int8(k string, v int8) Thing                { return Thing{Key: k, Type: IntType, Int: int64(v)} }
func Int(k string, v int) Thing                  { return Thing{Key: k, Type: IntType, Int: int64(v)} }
func Uint64(k string, v uint64) Thing            { return Thing{Key: k, Type: UintType, Any: v} }
func Uint32(k string, v uint32) Thing            { return Thing{Key: k, Type: UintType, Any: uint64(v)} }
func Uint16(k string, v uint16) Thing            { return Thing{Key: k, Type: UintType, Any: uint64(v)} }
func Uint8(k string, v uint8) Thing              { return Thing{Key: k, Type: UintType, Any: uint64(v)} }
func Uint(k string, v uint) Thing                { return Thing{Key: k, Type: UintType, Any: uint64(v)} }
func Bool(k string, v bool) Thing                { return Thing{Key: k, Type: BoolType, Any: v} }
func Str(k string, v string) Thing               { return Thing{Key: k, Type: StringType, String: v} }
func Time(k string, v time.Time) Thing           { return Thing{Key: k, Type: TimeType, Any: v} }
func AnyImmutable(k string, v interface{}) Thing { return Thing{Key: k, Type: AnyType, Any: v} }
func Error(k string, v error) Thing              { return Thing{Key: k, Type: ErrorType, Any: v} }

// TODO: make a copy
func Any(k string, v interface{}) Thing { return Thing{Key: k, Type: AnyType, Any: v} }

// TODO: make a copy
func MutableError(k string, v error) Thing { return Thing{Key: k, Type: ErrorType, Any: v} }

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

type Builder struct {
	T []Thing
}

func NewBuilder() *Builder                             { return &Builder{} }
func (b Builder) Things() []Thing                      { return b.T }
func (b *Builder) Int(k string, v int) *Builder        { b.T = append(b.T, Int(k, v)); return b }
func (b *Builder) Int8(k string, v int8) *Builder      { b.T = append(b.T, Int8(k, v)); return b }
func (b *Builder) Int16(k string, v int16) *Builder    { b.T = append(b.T, Int16(k, v)); return b }
func (b *Builder) Int32(k string, v int32) *Builder    { b.T = append(b.T, Int32(k, v)); return b }
func (b *Builder) Int64(k string, v int64) *Builder    { b.T = append(b.T, Int64(k, v)); return b }
func (b *Builder) Uint(k string, v uint) *Builder      { b.T = append(b.T, Uint(k, v)); return b }
func (b *Builder) Uint8(k string, v uint8) *Builder    { b.T = append(b.T, Uint8(k, v)); return b }
func (b *Builder) Uint16(k string, v uint16) *Builder  { b.T = append(b.T, Uint16(k, v)); return b }
func (b *Builder) Uint32(k string, v uint32) *Builder  { b.T = append(b.T, Uint32(k, v)); return b }
func (b *Builder) Uint64(k string, v uint64) *Builder  { b.T = append(b.T, Uint64(k, v)); return b }
func (b *Builder) Bool(k string, v bool) *Builder      { b.T = append(b.T, Bool(k, v)); return b }
func (b *Builder) Str(k string, v string) *Builder     { b.T = append(b.T, Str(k, v)); return b }
func (b *Builder) Time(k string, v time.Time) *Builder { b.T = append(b.T, Time(k, v)); return b }
func (b *Builder) Error(k string, v error) *Builder    { b.T = append(b.T, Error(k, v)); return b }
func (b *Builder) AnyImmutable(k string, v interface{}) *Builder {
	b.T = append(b.T, AnyImmutable(k, v))
	return b
}
