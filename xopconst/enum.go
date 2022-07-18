package xopconst

import (
	"fmt"
	"reflect"
)

type EnumAttribute struct{ Int64Attribute }

type Enum interface {
	Int64Attribute() *Int64Attribute
	Value() int64
}

type enum struct {
	Int           int64
	EnumAttribute *EnumAttribute
	IntAttribute *Int64Attribute
}

var _ Enum = enum{}

func (e enum) Int64Attribute() *Int64Attribute {
	return e.IntAttribute
}

func (e enum) Value() int64 {
	return e.Int
}


func (s Make) EnumAttribute(exampleValue interface{}) *EnumAttribute {
	e, err := s.TryEnumAttribute(exampleValue)
	if err != nil {
		panic(err)
	}
	return e
}
func (s Make) TryEnumAttribute(exampleValue interface{}) (_ *EnumAttribute, err error) {
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

func (e *EnumAttribute) Add(v fmt.Stringer) Enum {
	enum, err := e.TryAdd(v)
	if err != nil {
		panic(err)
	}
	return enum
}

func (e *EnumAttribute) TryAdd(v fmt.Stringer) (Enum, error) {
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
	intValue := rv.Int()
	e.Attribute.names.Store(intValue, v.String())
	return enum{
		Int: intValue,
		EnumAttribute: e,
		IntAttribute: &e.Int64Attribute,
	}, nil
}

