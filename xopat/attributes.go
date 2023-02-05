// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/internal/util/version"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/Masterminds/semver/v3"
)

var attributeCount int32 = 1

// Attribute is the base type for the keys that are used to add
// key/value metadata to spans.  The actual keys are matched to the
// values to provide compile-time type checking on the metadata calls.
// For example:
//
//	func (span *Span) String(k *xopconst.StringAttribute, v string) *Span
type Attribute struct {
	namespace    string
	version      string
	properties   Make
	jsonKey      string
	exampleValue interface{}
	reflectType  reflect.Type
	typeName     string
	subType      AttributeType
	names        sync.Map // key:int64 values:string used for enums
	values       sync.Map // name:enumValue used for enums
	defSize      int32
	semver       *semver.Version
	number       int32
}

// DefaultNamespace sets the namespace for attribute names
// used to describe spans. Within a namespace, the use of
// attribute names should be consistent.
//
// DefaultNamespace is also used as the default namespace
// that can be set by xop.WithNamespace() which is provided
// to base level loggers.
//
// DefaultNamespace defauls to filepath.Base(os.Args[0]).  Override it by
// setting DefaultNamespaceOverride using linker flags.
// Or (lower priority) override by setting the XOP_DEFAULT_NAMESPACE
// environment variable.
var DefaultNamespace = func() string {
	if DefaultNamespaceOverride == "" {
		if ns, ok := os.LookupEnv("XOP_DEFAULT_NAMESPACE"); ok {
			return ns
		}
		return filepath.Base(os.Args[0])
	}
	return DefaultNamespaceOverride
}()

// DefaultNamespaceOverride is meant to be set with compile options.
// go build -ldflags "-x xopat.DefaultNamespaceOverride myproject-1.0.0"
var DefaultNamespaceOverride string

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

// Can't use MACRO for these since default values are needed

func (s Make) AnyAttribute(exampleValue interface{}) *AnyAttribute {
	return &AnyAttribute{Attribute: s.attribute(defaultRegistry, exampleValue, nil, AttributeTypeAny)}
}

func (s Make) TryAnyAttribute(exampleValue interface{}) (_ *AnyAttribute, err error) {
	return &AnyAttribute{Attribute: s.attribute(defaultRegistry, exampleValue, &err, AttributeTypeAny)}, err
}

func (r *Registry) ContructAnyAttribute(s Make) (_ *AnyAttribute, err error) {
	return &AnyAttribute{Attribute: s.attribute(r, 0, &err, AttributeTypeAny)}, err
}

func (s Make) attribute(registry *Registry, exampleValue interface{}, ep *error, subType AttributeType) Attribute {
	a, err := s.make(registry, exampleValue, subType)
	if err != nil {
		if ep == nil {
			panic(err)
		}
		*ep = err
	}
	return a
}

func (s Make) make(registry *Registry, exampleValue interface{}, subType AttributeType) (Attribute, error) {
	registry.lock.Lock()
	defer registry.lock.Unlock()
	if prior, ok := registry.registeredNames[s.Key]; ok {
		if registry.errOnDuplicate {
			return *prior, fmt.Errorf("duplicate attribute registration for '%s'", s.Key)
		} else {
			return *prior, nil
		}
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

	namespace, sver, err := version.SplitVersionWithError(namespace)
	if err != nil {
		return Attribute{}, err
	}

	ra := Attribute{
		namespace:    namespace,
		version:      sver.String(),
		properties:   s,
		exampleValue: exampleValue,
		reflectType:  reflect.TypeOf(exampleValue),
		typeName:     fmt.Sprintf("%T", exampleValue),
		subType:      subType,
		jsonKey:      string(jsonKey) + ":",
		defSize:      int32(len(namespace) + len(s.Key) + len(s.Description) + len(sver.String())),
		semver:       sver,
		number:       atomic.AddInt32(&attributeCount, 1),
	}
	registry.registeredNames[s.Key] = &ra
	registry.allAttributes = append(registry.allAttributes, &ra)
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
func (r Attribute) Ranged() bool                      { return r.properties.Ranged } // XXX add to proto
func (r Attribute) Locked() bool                      { return r.properties.Locked }
func (r Attribute) Distinct() bool                    { return r.properties.Distinct }
func (r Attribute) Prominence() int                   { return r.properties.Prominence }
func (r Attribute) RegistrationNumber() int32         { return r.number }
func (r Attribute) ExampleValue() interface{}         { return r.exampleValue }
func (r Attribute) TypeName() string                  { return r.typeName }
func (r Attribute) SubType() AttributeType            { return r.subType }
func (r Attribute) ProtoType() xopproto.AttributeType { return xopproto.AttributeType(r.subType) }
func (r Attribute) DefinitionSize() int32             { return r.defSize }
func (r Attribute) Semver() *semver.Version           { return r.semver }
func (r Attribute) SemverString() string              { return r.version }
func (r *Attribute) Ptr() *Attribute                  { return r }

type AttributeInterface interface {
	JSONKey() JSONKey
	ReflectType() reflect.Type
	Key() string
	Description() string
	Namespace() string
	Indexed() bool
	Multiple() bool
	Ranged() bool
	Locked() bool
	Distinct() bool
	Prominence() int
	RegistrationNumber() int32
	ExampleValue() interface{}
	TypeName() string
	SubType() AttributeType
	ProtoType() xopproto.AttributeType
	DefinitionSize() int32
	Semver() *semver.Version
	SemverString() string
	Ptr() *Attribute
	EnumName(v int64) string
	GetEnum(n string) (Enum, bool)
}

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

func (s Make) BoolAttribute() *BoolAttribute {
	return &BoolAttribute{Attribute: s.attribute(defaultRegistry, true, nil, AttributeTypeBool)}
}

func (s Make) TryBoolAttribute() (*BoolAttribute, error) {
	return defaultRegistry.ConstructBoolAttribute(s)
}

func (r *Registry) ConstructBoolAttribute(s Make) (_ *BoolAttribute, err error) {
	return &BoolAttribute{Attribute: s.attribute(r, true, &err, AttributeTypeBool)}, err
}

func (s Make) Float32Attribute() *Float32Attribute {
	return &Float32Attribute{Attribute: s.attribute(defaultRegistry, float32(0.0), nil, AttributeTypeFloat32)}
}

func (s Make) TryFloat32Attribute() (*Float32Attribute, error) {
	return defaultRegistry.ConstructFloat32Attribute(s)
}

func (r *Registry) ConstructFloat32Attribute(s Make) (_ *Float32Attribute, err error) {
	return &Float32Attribute{Attribute: s.attribute(r, float32(0.0), &err, AttributeTypeFloat32)}, err
}

func (s Make) Float64Attribute() *Float64Attribute {
	return &Float64Attribute{Attribute: s.attribute(defaultRegistry, float64(0.0), nil, AttributeTypeFloat64)}
}

func (s Make) TryFloat64Attribute() (*Float64Attribute, error) {
	return defaultRegistry.ConstructFloat64Attribute(s)
}

func (r *Registry) ConstructFloat64Attribute(s Make) (_ *Float64Attribute, err error) {
	return &Float64Attribute{Attribute: s.attribute(r, float64(0.0), &err, AttributeTypeFloat64)}, err
}

func (s Make) Int64Attribute() *Int64Attribute {
	return &Int64Attribute{Attribute: s.attribute(defaultRegistry, int64(0), nil, AttributeTypeInt64)}
}

func (s Make) TryInt64Attribute() (*Int64Attribute, error) {
	return defaultRegistry.ConstructInt64Attribute(s)
}

func (r *Registry) ConstructInt64Attribute(s Make) (_ *Int64Attribute, err error) {
	return &Int64Attribute{Attribute: s.attribute(r, int64(0), &err, AttributeTypeInt64)}, err
}

func (s Make) LinkAttribute() *LinkAttribute {
	return &LinkAttribute{Attribute: s.attribute(defaultRegistry, xoptrace.Trace{}, nil, AttributeTypeLink)}
}

func (s Make) TryLinkAttribute() (*LinkAttribute, error) {
	return defaultRegistry.ConstructLinkAttribute(s)
}

func (r *Registry) ConstructLinkAttribute(s Make) (_ *LinkAttribute, err error) {
	return &LinkAttribute{Attribute: s.attribute(r, xoptrace.Trace{}, &err, AttributeTypeLink)}, err
}

func (s Make) StringAttribute() *StringAttribute {
	return &StringAttribute{Attribute: s.attribute(defaultRegistry, "", nil, AttributeTypeString)}
}

func (s Make) TryStringAttribute() (*StringAttribute, error) {
	return defaultRegistry.ConstructStringAttribute(s)
}

func (r *Registry) ConstructStringAttribute(s Make) (_ *StringAttribute, err error) {
	return &StringAttribute{Attribute: s.attribute(r, "", &err, AttributeTypeString)}, err
}

func (s Make) TimeAttribute() *TimeAttribute {
	return &TimeAttribute{Attribute: s.attribute(defaultRegistry, time.Time{}, nil, AttributeTypeTime)}
}

func (s Make) TryTimeAttribute() (*TimeAttribute, error) {
	return defaultRegistry.ConstructTimeAttribute(s)
}

func (r *Registry) ConstructTimeAttribute(s Make) (_ *TimeAttribute, err error) {
	return &TimeAttribute{Attribute: s.attribute(r, time.Time{}, &err, AttributeTypeTime)}, err
}

func (s Make) DurationAttribute() *DurationAttribute {
	return &DurationAttribute{Int64Attribute{Attribute: s.attribute(defaultRegistry, time.Duration(0), nil, AttributeTypeDuration)}}
}

func (s Make) TryDurationAttribute() (*DurationAttribute, error) {
	return defaultRegistry.ConstructDurationAttribute(s)
}

func (r *Registry) ConstructDurationAttribute(s Make) (_ *DurationAttribute, err error) {
	return &DurationAttribute{Int64Attribute{Attribute: s.attribute(r, time.Duration(0), &err, AttributeTypeDuration)}}, err
}

func (s Make) IntAttribute() *IntAttribute {
	return &IntAttribute{Int64Attribute{Attribute: s.attribute(defaultRegistry, int(0), nil, AttributeTypeInt)}}
}

func (s Make) TryIntAttribute() (*IntAttribute, error) {
	return defaultRegistry.ConstructIntAttribute(s)
}

func (r *Registry) ConstructIntAttribute(s Make) (_ *IntAttribute, err error) {
	return &IntAttribute{Int64Attribute{Attribute: s.attribute(r, int(0), &err, AttributeTypeInt)}}, err
}

func (s Make) Int16Attribute() *Int16Attribute {
	return &Int16Attribute{Int64Attribute{Attribute: s.attribute(defaultRegistry, int16(0), nil, AttributeTypeInt16)}}
}

func (s Make) TryInt16Attribute() (*Int16Attribute, error) {
	return defaultRegistry.ConstructInt16Attribute(s)
}

func (r *Registry) ConstructInt16Attribute(s Make) (_ *Int16Attribute, err error) {
	return &Int16Attribute{Int64Attribute{Attribute: s.attribute(r, int16(0), &err, AttributeTypeInt16)}}, err
}

func (s Make) Int32Attribute() *Int32Attribute {
	return &Int32Attribute{Int64Attribute{Attribute: s.attribute(defaultRegistry, int32(0), nil, AttributeTypeInt32)}}
}

func (s Make) TryInt32Attribute() (*Int32Attribute, error) {
	return defaultRegistry.ConstructInt32Attribute(s)
}

func (r *Registry) ConstructInt32Attribute(s Make) (_ *Int32Attribute, err error) {
	return &Int32Attribute{Int64Attribute{Attribute: s.attribute(r, int32(0), &err, AttributeTypeInt32)}}, err
}

func (s Make) Int8Attribute() *Int8Attribute {
	return &Int8Attribute{Int64Attribute{Attribute: s.attribute(defaultRegistry, int8(0), nil, AttributeTypeInt8)}}
}

func (s Make) TryInt8Attribute() (*Int8Attribute, error) {
	return defaultRegistry.ConstructInt8Attribute(s)
}

func (r *Registry) ConstructInt8Attribute(s Make) (_ *Int8Attribute, err error) {
	return &Int8Attribute{Int64Attribute{Attribute: s.attribute(r, int8(0), &err, AttributeTypeInt8)}}, err
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
