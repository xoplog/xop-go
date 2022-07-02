package xopthing

import "time"

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

func (t *Things) Int(k string, v int64) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.IntType,
		Int:  v,
	})
}
func (t *Things) Uint(k string, v uint64) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.UintType,
		Any:  v,
	})
}
func (t *Things) Bool(string, bool) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.BoolType,
		Any:  v,
	})
}
func (t *Things) Str(string, string) {
	t.Things = append(t.Things, Thing{
		Key:    k,
		Type:   zap.StringType,
		String: v,
	})
}
func (t *Things) Time(string, time.Time) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.TimeType,
		Any:  v,
	})
}
func (t *Things) Any(string, interface{}) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.AnyType,
		Any:  v,
	})
}
func (t *Things) Error(string, error) {
	t.Things = append(t.Things, Thing{
		Key:  k,
		Type: zap.ErrorType,
		Any:  v,
	})
}

// TODO: func (t *Things) SubObject(string) SubObject
// TODO: func (t *Things) Encoded(name string, elementName string, encoder Encoder, data interface{})
// TODO: func (t *Things) PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
// TODO: func (t *Things) ExternalReference(name string, itemId string, storageId string)
// TODO: func (t *Things) PreEncodedText(name string, elementName string, mimeType string, data string)
