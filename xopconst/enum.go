package xopconst

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

const noCounter = -10000000

type EnumAttribute struct{ Attribute }

// EmbeddedEnumAttribute is a type of enum set that can be added
// onto from multiple places.  For example, SpanType is a
// IotaEnumAttribute and consumers of the xop can add additional
// values to the enum.  Values must be kept distinct.
type EmbeddedEnumAttribute struct{ EnumAttribute }

// IotaEnumAttribute is a type of enum set that can be added
// onto from multiple places.  For example, SpanType is a
// IotaEnumAttribute and consumers of the xop can add additional
// values to the enum.  Values are automatically kept distinct.
type IotaEnumAttribute struct {
	EnumAttribute
	counter int64
}

type EmbeddedEnum interface {
	EnumAttribute() *EnumAttribute
	Enum
}

type Enum interface {
	Int64() int64
	String() string
}

type enum struct {
	value         int64
	enumAttribute *EnumAttribute
	str           string
}

var _ Enum = enum{}

func (e enum) EnumAttribute() *EnumAttribute {
	return e.enumAttribute
}

func (e enum) Int64() int64 {
	return e.value
}

func (e enum) String() string {
	return e.str
}

func (s Make) EnumAttribute(exampleValue Enum) *EnumAttribute {
	return &EnumAttribute{Attribute: s.attribute(exampleValue, nil, AttributeTypeEnum)}
}

func (s Make) TryEnumAttribute(exampleValue Enum) (_ *EnumAttribute, err error) {
	return &EnumAttribute{Attribute: s.attribute(exampleValue, &err, AttributeTypeEnum)}, err
}

func (s Make) TypedEnumAttribute(exampleValue interface{}) *EmbeddedEnumAttribute {
	e, err := s.TryTypedEnumAttribute(exampleValue)
	if err != nil {
		panic(err)
	}
	return e
}
func (s Make) TryTypedEnumAttribute(exampleValue interface{}) (_ *EmbeddedEnumAttribute, err error) {
	attribute := s.attribute(exampleValue, &err, AttributeTypeEnum)
	if err != nil {
		return nil, err
	}
	if attribute.reflectType == nil {
		return nil, fmt.Errorf("cannot make enum attribute with a nil value")
	}
	return &EmbeddedEnumAttribute{
		EnumAttribute: EnumAttribute{
			Attribute: attribute,
		},
	}, nil
}

// Iota creates new enums.  It cannotnot be combined with
// Add, Add64, or TryAddStringer() etc.
func (e *IotaEnumAttribute) Iota(s string) EmbeddedEnum {
	old := atomic.AddInt64(&e.counter, 1)
	e.Attribute.names.Store(old+1, s)
	return enum{
		value:         old + 1,
		enumAttribute: &e.EnumAttribute,
		str:           s,
	}
}

func (s Make) EmbeddedEnumAttribute() *IotaEnumAttribute {
	e, err := s.TryEmbeddedEnumAttribute()
	if err != nil {
		panic(err)
	}
	return e
}

func (s Make) TryEmbeddedEnumAttribute() (_ *IotaEnumAttribute, err error) {
	attribute := s.attribute(EmbeddedEnum(enum{}), &err, AttributeTypeEnum)
	if err != nil {
		return nil, err
	}
	return &IotaEnumAttribute{
		EnumAttribute: EnumAttribute{
			Attribute: attribute,
		},
	}, nil
}

func (e *EmbeddedEnumAttribute) AddStringer(v fmt.Stringer) EmbeddedEnum {
	enum, err := e.TryAddStringer(v)
	if err != nil {
		panic(err)
	}
	return enum
}

func (e *EmbeddedEnumAttribute) TryAddStringer(v fmt.Stringer) (EmbeddedEnum, error) {
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
	return e.Add64(rv.Int(), v.String()), nil
}

func (e *EmbeddedEnumAttribute) Add(i int, s string) EmbeddedEnum { return e.Add64(int64(i), s) }

func (e *EmbeddedEnumAttribute) Add64(i int64, s string) EmbeddedEnum {
	e.Attribute.names.Store(i, s)
	return enum{
		value:         i,
		enumAttribute: &e.EnumAttribute,
		str:           s,
	}
}
