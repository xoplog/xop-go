package xopat

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

// EnumAttributes are logged as strings or as integers depending
// on the base logger used.
type EnumAttribute struct{ Attribute }

// EmbeddedEnumAttribute is a type of enum set that can be added
// onto from multiple places.  For example, SpanType is a
// IotaEnumAttribute and consumers of xop can add additional
// values to the enum.  Values must be kept distinct.
//
// Enum attributes are logged as strings or as integers depending
// on the base logger used.
type EmbeddedEnumAttribute struct{ EnumAttribute }

// IotaEnumAttribute is a type of enum set that can be added
// onto from multiple places.  For example, SpanType is a
// IotaEnumAttribute and consumers of the xop can add additional
// values to the enum.  Values are automatically kept distinct.
//
// Enum attributes are logged as strings or as integers depending
// on the base logger used.
type IotaEnumAttribute struct {
	EnumAttribute
	counter int64
}

// EmbeddedEnum is a value that can be passed to the Span.EmbeddedEnum()
// method to add span-level metadata.  It encapsulates both the attribute
// name and the attribute value in a single argument.
//
// EmbeddedEnum values can be construted with a EmbeddedEnumAttribute or
// with an IotaEnumAttribute.
type EmbeddedEnum interface {
	EnumAttribute() *EnumAttribute
	Enum
}

// Enum is a value that can be paired with an EnumAttribute to provide
// a key/value metadata attribute for a span:
//
//   func (span *Span) Enum(k *xopconst.EnumAttribute, v xopconst.Enum)
//
// The key (*xopconst.EnumAttribute) and value (xopconst.Enum) are
// provided separately, unlike with an EmbeddedEnum.
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

// EnumAttribute makes a new enum from a type that implments .String()
//
//	//go:generate go get github.com/dmarkham/enumer
//	//go:generate go run github.com/dmarkham/enumer -linecomment -sql -json -text -yaml -gqlgen -type=SpanKindEnum
//	type SpanKindEnum int
//
//	const (
//		SpanKindServer   SpanKindEnum = iota // SERVER
//		SpanKindClient                       // CLIENT
//		SpanKindProducer                     // PRODUCER
//		SpanKindConsumer                     // CONSUMER
//		SpanKindInternal                     // INTERNAL
//	)
//
//	func (e SpanKindEnum) Int64() int64 { return int64(e) }
//
//	var SpanKind = xopconst.Make{Key: "span.kind", Namespace: "OTAP", Indexed: true, Prominence: 30,
// 		Description: "https://opentelemetry.io/docs/reference/specification/trace/api/#spankind" +
// 			" Use one of SpanKindServer, SpanKindClient, SpanKindProducer, SpanKindConsumer, SpanKindInternal"}.
// 		EnumAttribute(SpanKindServer)
//
//	log := xop.NewSeed().Request("an example")
//	log.Request().EmbeddedEnum(SpanTypeHTTPClientRequest)
//
func (s Make) EnumAttribute(exampleValue Enum) *EnumAttribute {
	return &EnumAttribute{Attribute: s.attribute(exampleValue, nil, AttributeTypeEnum)}
}

func (s Make) TryEnumAttribute(exampleValue Enum) (_ *EnumAttribute, err error) {
	return &EnumAttribute{Attribute: s.attribute(exampleValue, &err, AttributeTypeEnum)}, err
}

// Iota creates new enum values.
//
// For example:
//
//	var SpanType = xopconst.Make{Key: "span.type", Namespace: "xop", Indexed: true, Prominence: 11,
//		Description: "what kind of span this is.  Often added automatically.  eg: SpanTypeHTTPClientRequest"}.
//		EmbeddedEnumAttribute()
//
//	var (
//		SpanTypeHTTPServerEndpoint = SpanType.Iota("endpoint")
//		SpanTypeHTTPClientRequest  = SpanType.Iota("REST")
//		SpanTypeCronJob            = SpanType.Iota("cron_job")
//	)
//
//	log := xop.NewSeed().Request("an example")
//	log.Request().EmbeddedEnum(SpanTypeHTTPClientRequest)
//
func (e *IotaEnumAttribute) Iota(s string) EmbeddedEnum {
	old := atomic.AddInt64(&e.counter, 1)
	e.Attribute.names.Store(old+1, s)
	return enum{
		value:         old + 1,
		enumAttribute: &e.EnumAttribute,
		str:           s,
	}
}

// EmbeddedEnumAttribute creates a new enum that embeds it's key with
// it's value.
func (s Make) EmbeddedEnumAttribute() *IotaEnumAttribute {
	e, err := s.TryEmbeddedEnumAttribute()
	if err != nil {
		panic(err)
	}
	return e
}

func (s Make) TryEmbeddedEnumAttribute() (_ *IotaEnumAttribute, err error) {
	ie := &IotaEnumAttribute{
		EnumAttribute: EnumAttribute{
			Attribute: s.attribute(EmbeddedEnum(enum{}), &err, AttributeTypeEnum),
		},
	}
	if err != nil {
		return nil, err
	}
	return ie, nil
}

// AddStringer is another way to construct
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
