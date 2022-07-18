package xopconst

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

const noCounter = -10000000

type EnumAttribute struct {
	Int64Attribute
	counter int64 // values for non-counters: 0, noCounter
}

type Enum interface {
	Int64Attribute() *Int64Attribute
	Value() int64
	String() string
}

type enum struct {
	Int           int64
	EnumAttribute *EnumAttribute
	IntAttribute  *Int64Attribute
	Str           string
}

var _ Enum = enum{}

func (e enum) Int64Attribute() *Int64Attribute {
	return e.IntAttribute
}

func (e enum) Value() int64 {
	return e.Int
}

func (e enum) String() string {
	return e.Str
}

func (s Make) TypedEnumAttribute(exampleValue interface{}) *EnumAttribute {
	e, err := s.TryTypedEnumAttribute(exampleValue)
	if err != nil {
		panic(err)
	}
	return e
}
func (s Make) TryTypedEnumAttribute(exampleValue interface{}) (_ *EnumAttribute, err error) {
	attribute := s.attribute(exampleValue, &err, AttributeTypeEnum)
	if err != nil {
		return nil, err
	}
	if attribute.reflectType == nil {
		return nil, fmt.Errorf("cannot make enum attribute with a nil value")
	}
	intAttribute := Int64Attribute{
		Attribute: attribute,
	}
	enumAttribute := EnumAttribute{
		Int64Attribute: intAttribute,
	}
	return &enumAttribute, nil
}

func (s Make) EnumAttribute() *EnumAttribute {
	e, err := s.TryEnumAttribute()
	if err != nil {
		panic(err)
	}
	return e
}
func (s Make) TryEnumAttribute() (_ *EnumAttribute, err error) {
	attribute := s.attribute(Enum(enum{}), &err, AttributeTypeEnum)
	if err != nil {
		return nil, err
	}
	intAttribute := Int64Attribute{
		Attribute: attribute,
	}
	enumAttribute := EnumAttribute{
		Int64Attribute: intAttribute,
	}
	return &enumAttribute, nil
}

func (e *EnumAttribute) AddStringer(v fmt.Stringer) Enum {
	enum, err := e.TryAddStringer(v)
	if err != nil {
		panic(err)
	}
	return enum
}

func (e *EnumAttribute) TryAddStringer(v fmt.Stringer) (Enum, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil, fmt.Errorf("cannot add enum with a value of nil")
	}
	if t != e.Attribute.reflectType {
		return nil, fmt.Errorf("cannot add enum, type %s does not match EnumAttribute's %s",
			t, e.Attribute.reflectType)
	}
	rv := reflect.ValueOf(v)
	if !rv.CanInt() {
		return nil, fmt.Errorf("cannot add enum, underlying type of %s is not 'int'", t)
	}
	return e.TryAdd64(rv.Int(), v.String())
}

func (e *EnumAttribute) Add(i int, s string) Enum             { return e.Add64(int64(i), s) }
func (e *EnumAttribute) TryAdd(i int, s string) (Enum, error) { return e.TryAdd64(int64(i), s) }

func (e *EnumAttribute) Add64(i int64, s string) Enum {
	enum, err := e.TryAdd64(i, s)
	if err != nil {
		panic(err)
	}
	return enum
}

func (e *EnumAttribute) TryAdd64(i int64, s string) (Enum, error) {
	old := atomic.SwapInt64(&e.counter, noCounter)
	if old != 0 && old != noCounter {
		return nil, fmt.Errorf("invalid combination of Iota and Add, found when adding %d/%s", i, s)
	}
	e.Attribute.names.Store(i, s)
	return enum{
		Int:           i,
		EnumAttribute: e,
		IntAttribute:  &e.Int64Attribute,
		Str:           s,
	}, nil
}

func (e *EnumAttribute) TryIota(s string) (Enum, error) {
	old := atomic.AddInt64(&e.counter, 1)
	if old < 0 {
		return nil, fmt.Errorf("invalid combination of Iota and Add, found when adding %s", s)
	}
	e.Attribute.names.Store(old+1, s)
	return enum{
		Int:           old + 1,
		EnumAttribute: e,
		IntAttribute:  &e.Int64Attribute,
		Str:           s,
	}, nil
}

// Iota creates new enums.  It cannotnot be combined with
// Add, Add64, or TryAddStringer() etc.
func (e *EnumAttribute) Iota(s string) Enum {
	enum, err := e.TryIota(s)
	if err != nil {
		panic(err)
	}
	return enum
}
