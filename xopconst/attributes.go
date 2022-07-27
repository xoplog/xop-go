// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopconst

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/muir/xoplog/trace"
)

// TODO: PERFORMANCE: pre-allocate blocks of 128 Attributes to provide better locality of reference when using these
// TODO: UintAttribute?
// TODO: TableAttribute?
// TODO: URLAttribute?

// Attribute represents an "any" attribute for a span.
type Attribute struct {
	properties   Make
	number       int
	jsonKey      string
	exampleValue interface{}
	reflectType  reflect.Type
	typeName     string
	subType      AttributeType
	names        sync.Map // key:int64 values:string used for enums
}

// DefaultNamespace sets the namespace for attribute names
// used to describe spans.  Within a namespace, the use of
// attribute names should be consistent.  If not specified,
// the name of the running program (os.Args[0]) is used.
// A better value is to set the namespace to be name of the
// code repository.  DefaultNamespace can be overridden in
// an init() function.
var DefaultNamespace = os.Args[0]

type Make struct {
	Key          string // the attribute name
	Description  string // the attribute description
	Namespace    string // the namespace for this attribute (otherwise DefaultNamespace is used)
	Indexed      bool   // hint: this attribute should be indexed
	Prominence   int    // hint: how important is this attribute (lower is more important)
	Multiple     bool   // keep all values if the attribute is given multiple times
	Distinct     bool   // when keeping all values, only keep distinct values (not supported for interface{})
	Ranged       bool   // hint: comparisons between values are meaningful (eg: time, integers)
	Locked       bool   // only keep the first value
	AppenderFunc func() ArrayAppender
}

type ArrayAppender interface {
	AppendAny(interface{})
}

var (
	lock            sync.RWMutex
	registeredNames = make(map[string]*Attribute)
	allAttributes   []*Attribute
)

// Can't use MACRO for these since default values are needed

func (s Make) LinkAttribute() *LinkAttribute {
	return &LinkAttribute{Attribute: s.attribute(trace.Trace{}, nil, AttributeTypeLink)}
}

func (s Make) TryLinkAttribute() (_ *LinkAttribute, err error) {
	return &LinkAttribute{Attribute: s.attribute(trace.Trace{}, &err, AttributeTypeLink)}, err
}

func (s Make) StrAttribute() *StrAttribute {
	return &StrAttribute{Attribute: s.attribute("", nil, AttributeTypeStr)}
}

func (s Make) TryStrAttribute() (_ *StrAttribute, err error) {
	return &StrAttribute{Attribute: s.attribute("", &err, AttributeTypeStr)}, err
}

func (s Make) BoolAttribute() *BoolAttribute {
	return &BoolAttribute{Attribute: s.attribute(false, nil, AttributeTypeBool)}
}

func (s Make) TryBoolAttribute() (_ *BoolAttribute, err error) {
	return &BoolAttribute{Attribute: s.attribute(false, &err, AttributeTypeBool)}, err
}

func (s Make) TimeAttribute() *TimeAttribute {
	return &TimeAttribute{Attribute: s.attribute(time.Time{}, nil, AttributeTypeEnum)}
}

func (s Make) TryTimeAttribute() (_ *TimeAttribute, err error) {
	return &TimeAttribute{Attribute: s.attribute(time.Time{}, &err, AttributeTypeEnum)}, err
}

func (s Make) AnyAttribute(exampleValue interface{}) *AnyAttribute {
	return &AnyAttribute{Attribute: s.attribute(exampleValue, nil, AttributeTypeAny)}
}

func (s Make) TryAnyAttribute(exampleValue interface{}) (_ *AnyAttribute, err error) {
	return &AnyAttribute{Attribute: s.attribute(exampleValue, &err, AttributeTypeAny)}, err
}

func (s Make) Int64Attribute() *Int64Attribute {
	return &Int64Attribute{Attribute: s.attribute(int64(0), nil, AttributeTypeInt64)}
}

func (s Make) TryInt64Attribute() (_ *Int64Attribute, err error) {
	return &Int64Attribute{Attribute: s.attribute(int64(0), &err, AttributeTypeInt64)}, err
}

func (s Make) attribute(exampleValue interface{}, ep *error, subType AttributeType) Attribute {
	a, err := s.make(exampleValue, subType)
	if err != nil {
		if ep == nil {
			panic(err)
		}
		*ep = err
	}
	return a
}

func (s Make) make(exampleValue interface{}, subType AttributeType) (Attribute, error) {
	lock.Lock()
	defer lock.Unlock()
	if prior, ok := registeredNames[s.Key]; ok {
		return *prior, fmt.Errorf("duplicate attribute registration for '%s'", s.Key)
	}
	if s.Namespace == "" {
		s.Namespace = DefaultNamespace
	}
	ra := Attribute{
		properties:   s,
		exampleValue: exampleValue,
		reflectType:  reflect.TypeOf(exampleValue),
		typeName:     fmt.Sprintf("%T", exampleValue),
		subType:      subType,
	}
	jsonKey, err := json.Marshal(s.Key)
	if err != nil {
		return ra, fmt.Errorf("cannot marshal attribute name '%s': %w", s.Key, err)
	}
	ra.jsonKey = string(jsonKey)
	registeredNames[s.Key] = &ra
	allAttributes = append(allAttributes, &ra)
	return ra, nil
}

// JSONKey returns a JSON-quoted string that can be used as the key to the
// name of the attribute.  It is a string because []byte is mutable
func (r Attribute) JSONKey() string { return r.jsonKey }

// ReflectType can be nil if the example value was nil
func (r Attribute) ReflectType() reflect.Type { return r.reflectType }

func (r Attribute) Key() string               { return r.properties.Key }
func (r Attribute) Description() string       { return r.properties.Description }
func (r Attribute) Namespace() string         { return r.properties.Namespace }
func (r Attribute) Indexed() bool             { return r.properties.Indexed }
func (r Attribute) Multiple() bool            { return r.properties.Multiple }
func (r Attribute) Ranged() bool              { return r.properties.Ranged }
func (r Attribute) Locked() bool              { return r.properties.Locked }
func (r Attribute) Distinct() bool            { return r.properties.Distinct }
func (r Attribute) Prominence() int           { return r.properties.Prominence }
func (r Attribute) RegistrationNumber() int   { return r.number }
func (r Attribute) ExampleValue() interface{} { return r.exampleValue }
func (r Attribute) TypeName() string          { return r.typeName }
func (r Attribute) SubType() AttributeType    { return r.subType }

// EnumName only provides non-empty answers when SubType() == AttributeTypeEnum
func (r Attribute) EnumName(v int64) string {
	if n, ok := r.names.Load(v); ok {
		return n.(string)
	}
	return ""
}

func (s Make) DurationAttribute() *DurationAttribute {
	return &DurationAttribute{Int64Attribute{Attribute: s.attribute(time.Duration(0), nil, AttributeTypeDuration)}}
}

func (s Make) TryDurationAttribute() (_ *DurationAttribute, err error) {
	return &DurationAttribute{Int64Attribute{Attribute: s.attribute(time.Duration(0), &err, AttributeTypeDuration)}}, err
}

func (s Make) IntAttribute() *IntAttribute {
	return &IntAttribute{Int64Attribute{Attribute: s.attribute(int(0), nil, AttributeTypeInt)}}
}

func (s Make) TryIntAttribute() (_ *IntAttribute, err error) {
	return &IntAttribute{Int64Attribute{Attribute: s.attribute(int(0), &err, AttributeTypeInt)}}, err
}

func (s Make) Int16Attribute() *Int16Attribute {
	return &Int16Attribute{Int64Attribute{Attribute: s.attribute(int16(0), nil, AttributeTypeInt16)}}
}

func (s Make) TryInt16Attribute() (_ *Int16Attribute, err error) {
	return &Int16Attribute{Int64Attribute{Attribute: s.attribute(int16(0), &err, AttributeTypeInt16)}}, err
}

func (s Make) Int32Attribute() *Int32Attribute {
	return &Int32Attribute{Int64Attribute{Attribute: s.attribute(int32(0), nil, AttributeTypeInt32)}}
}

func (s Make) TryInt32Attribute() (_ *Int32Attribute, err error) {
	return &Int32Attribute{Int64Attribute{Attribute: s.attribute(int32(0), &err, AttributeTypeInt32)}}, err
}

func (s Make) Int8Attribute() *Int8Attribute {
	return &Int8Attribute{Int64Attribute{Attribute: s.attribute(int8(0), nil, AttributeTypeInt8)}}
}

func (s Make) TryInt8Attribute() (_ *Int8Attribute, err error) {
	return &Int8Attribute{Int64Attribute{Attribute: s.attribute(int8(0), &err, AttributeTypeInt8)}}, err
}

type AttributeType int

const (
	AttributeTypeUnknown AttributeType = iota
	AttributeTypeAny
	AttributeTypeBool
	AttributeTypeDuration
	AttributeTypeEnum
	AttributeTypeFloat32
	AttributeTypeInt
	AttributeTypeInt16
	AttributeTypeInt32
	AttributeTypeInt64
	AttributeTypeInt8
	AttributeTypeLink
	AttributeTypeNumber
	AttributeTypeStr
	AttributeTypeTime
)

// DurationAttribute is a just an Int64Attribute that with
// SubType() == AttributeTypeDuration.  A base logger may
// look at SubType() to provide specialized behavior.
type DurationAttribute struct{ Int64Attribute }

// IntAttribute is a just an Int64Attribute that with
// SubType() == AttributeTypeDuration.  A base logger may
// look at SubType() to provide specialized behavior.
type IntAttribute struct{ Int64Attribute }

// Int16Attribute is a just an Int64Attribute that with
// SubType() == AttributeTypeDuration.  A base logger may
// look at SubType() to provide specialized behavior.
type Int16Attribute struct{ Int64Attribute }

// Int32Attribute is a just an Int64Attribute that with
// SubType() == AttributeTypeDuration.  A base logger may
// look at SubType() to provide specialized behavior.
type Int32Attribute struct{ Int64Attribute }

// Int8Attribute is a just an Int64Attribute that with
// SubType() == AttributeTypeDuration.  A base logger may
// look at SubType() to provide specialized behavior.
type Int8Attribute struct{ Int64Attribute }

// AnyAttribute represents an attribute key that can be used
// with interface{} values.
type AnyAttribute struct{ Attribute }

// BoolAttribute represents an attribute key that can be used
// with bool values.
type BoolAttribute struct{ Attribute }

// Float32Attribute represents an attribute key that can be used
// with float32 values.
type Float32Attribute struct{ Attribute }

// Int64Attribute represents an attribute key that can be used
// with int64 values.
type Int64Attribute struct{ Attribute }

// LinkAttribute represents an attribute key that can be used
// with trace.Trace values.
type LinkAttribute struct{ Attribute }

// NumberAttribute represents an attribute key that can be used
// with float64 values.
type NumberAttribute struct{ Attribute }

// StrAttribute represents an attribute key that can be used
// with string values.
type StrAttribute struct{ Attribute }

// TimeAttribute represents an attribute key that can be used
// with time.Time values.
type TimeAttribute struct{ Attribute }
