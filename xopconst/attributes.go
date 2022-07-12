package xopconst

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/pkg/errors"
)

// TODO: PERFORMANCE: pre-allocate blocks of 128 Attributes to provide better locality of reference when using these

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

type Make struct {
	Key      string // the attribute name
	Desc     string // the attribute description
	Indexed  bool   // hint: this attribute should be indexed
	Multiple bool   // keep all values if the attribute is given multiple times
	Ranged   bool   // hint: comparisons between values are meaningful (eg: time, integers)
	Locked   bool   // only keep the first value
}

var lock sync.RWLock
var registeredMakes = make(map[string]*Attribute)
var allMakes []*Attribute

func (s Make) MustAttribute(exampleValue interface{}) *Attribute {
	ra, err := s.Register(exampleValue)
	if err != nil {
		panic(err)
	}
	return ra
}

func (s Make) IntAttribute() (*IntAttribute, error) {
	a, err := s.attribute(0)
	return &IntAttribute{attribute: a}, err
}

func (s Make) StrAttribute() (*StrAttribute, error) {
	a, err := s.attribute("")
	return &StrAttribute{attribute: a}, err
}

func (s Make) LinkAttribute() (*LinkAttribute, error) {
	a, err := s.attribute(trace.Trace{})
	return &StrAttribute{attribute: a}, err
}

func (s Make) TimeAttribute() (*LinkAttribute, error) {
	a, err := s.attribute(time.Time{})
	return &TimeAttribute{attribute: a}, err
}

func (s Make) DurationAttribute() (*LinkAttribute, error) {
	a, err := s.attribute(time.Duration(0))
	return &DurationAttribute{attribute: a}, err
}

func (s Make) attribute(exampleValue interface{}) (Attribute, error) {
	lock.Lock()
	defer lock.Unlock()
	if prior, ok := registeredMakes[s.Key]; ok {
		return prior, errors.New("duplicate attribute registration for '%s'", s.Key)
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
func (r Attribute) Indexed() bool             { return r.properties.Indexed }
func (r Attribute) Multiple() bool            { return r.properties.Multiple }
func (r Attribute) Ranged() bool              { return r.properties.Ranged }
func (r Attribute) Locked() bool              { return r.properties.Locked }
func (r Attribute) RegistrationNumber() int   { return r.number }
func (r Attribute) ExampleValue() interface{} { return r.exampleValue }
func (r Attribute) TypeName() string          { return r.typeName }
