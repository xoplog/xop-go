package xopconst

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/pkg/errors"
)

// TODO: PERFORMANCE: pre-allocate blocks of 128 Attributes to provide better locality of reference when using these
// TODO: UintAttribute?
// TODO: TableAttribute?
// TODO: URLAttribute?

// TODO: BoolAttribute

type IntAttribute struct{ Attribute }
type StrAttribute struct{ Attribute }
type LinkAttribute struct{ Attribute }
type TimeAttribute struct{ Attribute }
type DurationAttribute struct{ Attribute }

// Attribute represents an "any" attribute for a span.
type Attribute struct {
	properties   Make
	number       int
	jsonKey      string
	exampleValue interface{}
	reflectType  reflect.Type
	typeName     string
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
	Key       string // the attribute name
	Desc      string // the attribute description
	Namespace string // the namespace for this attribute (otherwise DefaultNamespace is used)
	Indexed   bool   // hint: this attribute should be indexed
	Multiple  bool   // keep all values if the attribute is given multiple times
	Ranged    bool   // hint: comparisons between values are meaningful (eg: time, integers)
	Locked    bool   // only keep the first value
}

var lock sync.RWLock
var registeredMakes = make(map[string]*Attribute)
var allMakes []*Attribute

func (s Make) Attribute(exampleValue interface{}) *Attribute {
	a, err := s.TryAttribute(exampleValue)
	if err != nil {
		panic(err)
	}
	return a
}

func (s Make) Attribute(exampleValue interface{}) *Attribute {
	return &Attribute{attribute: s.attribute(exampleValue, nil)}
}
func (s Make) TryAttribute(exampleValue interface{}) (_ *Attribute, err error) {
	return &Attribute{attribute: s.attribute(exampleValue, &err)}, err
}

func (s Make) LinkAttribute() *LinkAttribute {
	return &LinkAttribute{attribute: s.attribute(trace.Trace{}, nil)}
}
func (s Make) TryLinkAttribute() (_ *LinkAttribute, err error) {
	return &LinkAttribute{attribute: s.attribute(trace.Trace{}, &err)}, err
}

func (s Make) StrAttribute() *StrAttribute {
	return &StrAttribute{attribute: s.attribute("", nil)}
}
func (s Make) TryStrAttribute() (_ *StrAttribute, err error) {
	return &StrAttribute{attribute: s.attribute("", &err)}, err
}

func (s Make) IntAttribute() *IntAttribute {
	return &IntAttribute{attribute: s.attribute(int(0), nil)}
}
func (s Make) TryIntAttribute() (_ *IntAttribute, err error) {
	return &IntAttribute{attribute: s.attribute(int(0), &err)}, err
}

func (s Make) TimeAttribute() *TimeAttribute {
	return &TimeAttribute{attribute: s.attribute(time.Time{}, nil)}
}
func (s Make) TryTimeAttribute() (_ *TimeAttribute, err error) {
	return &TimeAttribute{attribute: s.attribute(time.Time{}, &err)}, err
}

func (s Make) DurationAttribute() *DurationAttribute {
	return &DurationAttribute{attribute: s.attribute(time.Duration(0), nil)}
}
func (s Make) TryDurationAttribute() (_ *DurationAttribute, err error) {
	return &DurationAttribute{attribute: s.attribute(time.Duration(0), &err)}, err
}

func (s Make) attribute(exampleValue interface{}, ep *error) Attribute {
	a, err := doAddAttribute(exampleValue)
	if err != nil {
		if ep == nil {
			panic(err)
		}
		*ep = err
	}
	return a
}

func (s Make) doAddAttribute(exampleValue interface{}) (Attribute, error) {
	lock.Lock()
	defer lock.Unlock()
	if prior, ok := registeredMakes[s.Key]; ok {
		return prior, errors.New("duplicate attribute registration for '%s'", s.Key)
	}
	if s.Namespace == "" {
		s.Namespace = DefaultNamespace
	}
	ra := &RegisteredAtrribute{
		properties:   s,
		exampleValue: exampleValue,
		reflectType:  reflect.TypeOf(exampleValue),
		typeName:     fmt.Sprintf("%T", exampleValue),
	}
	jsonKey, err = json.Marshal(s.Key)
	if err != nil {
		return ra, errors.Errorf("cannot marshal attribute name '%s': %w", s.Key, err)
	}
	ra.jsonKey = string(jsonKey)
	registeredMakes[s.Key] = &s
	allMakes = append(allMakes, &s)
	return &s, nil
}

// JSONKey returns a JSON-quoted string that can be used as the key to the
// name of the attribute.  It is a string because []byte is mutable
func (r Attribute) JSONKey() string { return r.jsonKey }

// ReflectType can be nil if the example value was nil
func (r Attribute) ReflectType() reflect.Type { return r.reflectType }

func (r Attribute) Key() string               { return r.properties.Key }
func (r Attribute) Desc() string              { return r.properties.Desc }
func (r Attribute) Namespace() string         { return r.properties.Namespace }
func (r Attribute) Indexed() bool             { return r.properties.Indexed }
func (r Attribute) Multiple() bool            { return r.properties.Multiple }
func (r Attribute) Ranged() bool              { return r.properties.Ranged }
func (r Attribute) Locked() bool              { return r.properties.Locked }
func (r Attribute) RegistrationNumber() int   { return r.number }
func (r Attribute) ExampleValue() interface{} { return r.exampleValue }
func (r Attribute) TypeName() string          { return r.typeName }
