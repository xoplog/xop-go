// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopat

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
)

// Attribute is the base type for the keys that are used to add
// key/value metadata to spans.  The actual keys are matched to the
// values to provide compile-time type checking on the metadata calls.
// For example:
//
//	func (span *Span) String(k *xopconst.StringAttribute, v string) *Span
//
type Attribute struct {
	namespace    string
	version      string
	properties   Make
	number       int
	jsonKey      string
	exampleValue interface{}
	reflectType  reflect.Type
	typeName     string
	subType      AttributeType
	names        sync.Map // key:int64 values:string used for enums
	values       sync.Map // name:enumValue used for enums
	defSize      int32
	semver       *semver.Version
}

// DefaultNamespace sets the namespace for attribute names
// used to describe spans.  Within a namespace, the use of
// attribute names should be consistent.  If not specified,
// the name of the running program (os.Args[0]) is used.
// A better value is to set the namespace to be name of the
// code repository.  DefaultNamespace can be overridden in
// an init() function.
var DefaultNamespace = os.Args[0]

// Make is used to construct attributes.
// Some keys are reserved.  See https://github.com/xoplog/xop-go/blob/main/xopconst/reserved.go
// for the list of reserved keys.  Some keys are already registered.
//
// The Namespace can embed a semver version: eg: "oltp-1.3.7".  If no version is
// provided, 0.0.0 is assumed.
type Make struct {
	Key         string // the attribute name
	Description string // the attribute description
	Namespace   string // Versioned namespace for this attribute (otherwise DefaultNamespace is used) name-0.0.0 if no version provided
	Indexed     bool   // hint: this attribute should be indexed
	Prominence  int    // hint: how important is this attribute (lower is more important)
	Multiple    bool   // keep all values if the attribute is given multiple times
	Distinct    bool   // when keeping all values, only keep distinct values (not supported for interface{})
	Ranged      bool   // hint: comparisons between values are meaningful (eg: time, integers)
	Locked      bool   // only keep the first value
}

var (
	lock            sync.RWMutex
	registeredNames = make(map[string]*Attribute)
	allAttributes   []*Attribute
)

// Can't use MACRO for these since default values are needed

func (s Make) LinkAttribute() *LinkAttribute {
	return &LinkAttribute{Attribute: s.attribute(xoptrace.Trace{}, nil, AttributeTypeLink)}
}

func (s Make) TryLinkAttribute() (_ *LinkAttribute, err error) {
	return &LinkAttribute{Attribute: s.attribute(xoptrace.Trace{}, &err, AttributeTypeLink)}, err
}

func (s Make) StringAttribute() *StringAttribute {
	return &StringAttribute{Attribute: s.attribute("", nil, AttributeTypeString)}
}

func (s Make) TryStringAttribute() (_ *StringAttribute, err error) {
	return &StringAttribute{Attribute: s.attribute("", &err, AttributeTypeString)}, err
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

func (s Make) Float64Attribute() *Float64Attribute {
	return &Float64Attribute{Attribute: s.attribute(float64(0), nil, AttributeTypeFloat64)}
}

func (s Make) TryFloat64Attribute() (_ *Float64Attribute, err error) {
	return &Float64Attribute{Attribute: s.attribute(float64(0), &err, AttributeTypeFloat64)}, err
}

func (s Make) Float32Attribute() *Float32Attribute {
	return &Float32Attribute{Attribute: s.attribute(float32(0), nil, AttributeTypeFloat32)}
}

func (s Make) TryFloat32Attribute() (_ *Float32Attribute, err error) {
	return &Float32Attribute{Attribute: s.attribute(float32(0), &err, AttributeTypeFloat32)}, err
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

var namespaceVersionRE = regexp.MustCompile(`^(.+)-([0-9].+)$`)

func (s Make) make(exampleValue interface{}, subType AttributeType) (Attribute, error) {
	lock.Lock()
	defer lock.Unlock()
	if prior, ok := registeredNames[s.Key]; ok {
		return *prior, fmt.Errorf("duplicate attribute registration for '%s'", s.Key)
	}
	if _, ok := reservedKeys[s.Key]; ok {
		return Attribute{}, fmt.Errorf("key is reserved for internal use '%s'", s.Key)
	}

	namespace := s.Namespace
	if namespace == "" {
		namespace = DefaultNamespace
	}

	jsonKey, err := json.Marshal(s.Key)
	if err != nil {
		return Attribute{}, fmt.Errorf("cannot marshal attribute name '%s': %w", s.Key, err)
	}

	var version string

	if m := namespaceVersionRE.FindStringSubmatch(namespace); m != nil {
		namespace = m[1]
		version = m[2]
	} else {
		version = "0.0.0"
	}

	sver, err := semver.StrictNewVersion(version)
	if err != nil {
		return Attribute{}, fmt.Errorf("semver '%s' is not valid: %w", version, err)
	}

	ra := Attribute{
		namespace:    namespace,
		version:      version,
		properties:   s,
		exampleValue: exampleValue,
		reflectType:  reflect.TypeOf(exampleValue),
		typeName:     fmt.Sprintf("%T", exampleValue),
		subType:      subType,
		jsonKey:      string(jsonKey) + ":",
		defSize:      int32(len(namespace) + len(s.Key) + len(s.Description) + len(version)),
		semver:       sver,
	}
	registeredNames[s.Key] = &ra
	allAttributes = append(allAttributes, &ra)
	return ra, nil
}

type JSONKey string

func (jk JSONKey) String() string { return string(jk) }

// JSONKey returns a JSON-quoted string that can be used as the key to the
// name of the attribute.  It is a string because []byte is mutable.  JSONKey
// includes a colon at the end since it's uses is as a key.
func (r Attribute) JSONKey() JSONKey { return JSONKey(r.jsonKey) }

// ReflectType can be nil if the example value was nil
func (r Attribute) ReflectType() reflect.Type { return r.reflectType }

func (r Attribute) Key() string                       { return r.properties.Key }
func (r Attribute) Description() string               { return r.properties.Description }
func (r Attribute) Namespace() string                 { return r.namespace }
func (r Attribute) Indexed() bool                     { return r.properties.Indexed }
func (r Attribute) Multiple() bool                    { return r.properties.Multiple }
func (r Attribute) Ranged() bool                      { return r.properties.Ranged }
func (r Attribute) Locked() bool                      { return r.properties.Locked }
func (r Attribute) Distinct() bool                    { return r.properties.Distinct }
func (r Attribute) Prominence() int                   { return r.properties.Prominence }
func (r Attribute) RegistrationNumber() int           { return r.number }
func (r Attribute) ExampleValue() interface{}         { return r.exampleValue }
func (r Attribute) TypeName() string                  { return r.typeName }
func (r Attribute) SubType() AttributeType            { return r.subType }
func (r Attribute) ProtoType() xopproto.AttributeType { return xopproto.AttributeType(r.subType) }
func (r Attribute) DefinitionSize() int32             { return r.defSize }
func (r Attribute) Semver() *semver.Version           { return r.semver }
func (r Attribute) SemverString() string              { return r.version }
func (r *Attribute) Ptr() *Attribute                  { return r }

// EnumName only provides non-empty answers when SubType() == AttributeTypeEnum
func (r Attribute) EnumName(v int64) string {
	if n, ok := r.names.Load(v); ok {
		return n.(string)
	}
	return ""
}

func (r Attribute) GetEnum(n string) (Enum, bool) {
	if e, ok := r.values.Load(n); ok {
		return e.(Enum), true
	}
	return nil, false
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
	AttributeTypeUnknown  = AttributeType(xopproto.AttributeType_Unknown)
	AttributeTypeAny      = AttributeType(xopproto.AttributeType_Any)
	AttributeTypeBool     = AttributeType(xopproto.AttributeType_Bool)
	AttributeTypeDuration = AttributeType(xopproto.AttributeType_Duration)
	AttributeTypeEnum     = AttributeType(xopproto.AttributeType_Enum)
	AttributeTypeFloat32  = AttributeType(xopproto.AttributeType_Float32)
	AttributeTypeFloat64  = AttributeType(xopproto.AttributeType_Float64)
	AttributeTypeInt      = AttributeType(xopproto.AttributeType_Int)
	AttributeTypeInt16    = AttributeType(xopproto.AttributeType_Int16)
	AttributeTypeInt32    = AttributeType(xopproto.AttributeType_Int32)
	AttributeTypeInt64    = AttributeType(xopproto.AttributeType_Int64)
	AttributeTypeInt8     = AttributeType(xopproto.AttributeType_Int8)
	AttributeTypeLink     = AttributeType(xopproto.AttributeType_Link)
	AttributeTypeString   = AttributeType(xopproto.AttributeType_String)
	AttributeTypeTime     = AttributeType(xopproto.AttributeType_Time)
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

// Float64Attribute represents an attribute key that can be used
// with float64 values.
type Float64Attribute struct{ Attribute }

// Int64Attribute represents an attribute key that can be used
// with int64 values.
type Int64Attribute struct{ Attribute }

// LinkAttribute represents an attribute key that can be used
// with xoptrace.Trace values.
type LinkAttribute struct{ Attribute }

// StringAttribute represents an attribute key that can be used
// with string values.
type StringAttribute struct{ Attribute }

// TimeAttribute represents an attribute key that can be used
// with time.Time values.
type TimeAttribute struct{ Attribute }
