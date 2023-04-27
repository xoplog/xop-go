package xopproto

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var _ json.Marshaler = Encoding(0)
var _ json.Unmarshaler = (*Encoding)(nil)

var encodingBytes = func() map[Encoding][]byte {
	m := make(map[Encoding][]byte)
	for k, v := range Encoding_name {
		m[Encoding(k)] = []byte(`"` + v + `"`)
	}
	return m
}()

func (x Encoding) ToString() string {
	if s, ok := Encoding_name[x]; ok {
		return s
	}
	return strconv.Itoa(int(x))
}

func (x *Encoding) FromString(s string) error {
	switch s[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*x = Encoding(i)
	default:
		if i, ok := Encoding_value[s]; ok {
			*x = Encoding(i)
		} else {
			return fmt.Errorf("invalid Encoding (%s)", s)
		}
	}
	return nil
}

func (x Encoding) MarshalJSON() ([]byte, error) {
	if e, ok := encodingBytes[x]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("invalid Encoding value %d", x)
}
func (x *Encoding) UnmarshalJSON(b []byte) error {
	return unmarshal(x, b, Encoding_value, encodingBytes)
}

var _ json.Marshaler = AttributeType(0)
var _ json.Unmarshaler = (*AttributeType)(nil)

var attributeTypeBytes = func() map[AttributeType][]byte {
	m := make(map[AttributeType][]byte)
	for k, v := range AttributeType_name {
		m[AttributeType(k)] = []byte(`"` + v + `"`)
	}
	return m
}()

func (x AttributeType) MarshalJSON() ([]byte, error) {
	if e, ok := attributeTypeBytes[x]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("invalid AttributeType value %d", x)
}

func (x *AttributeType) UnmarshalJSON(b []byte) error {
	return unmarshal(x, b, AttributeType_value, attributeTypeBytes)
}

func unmarshal[T ~int32](x *T, b []byte, s2i map[string]int32, i2b map[T][]byte) error {
	if len(b) == 0 {
		return fmt.Errorf("invalid %T value, empty", *x)
	}
	switch b[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			return err
		}
		if _, ok := i2b[T(i)]; !ok {
			return fmt.Errorf("invalid %T value %d", *x, i)
		}
		*x = T(i)
	case '"':
		if b[len(b)-1] != '"' {
			return fmt.Errorf("invalid %T value (%s) mismatched quote", *x, string(b))
		}
		s := string(b[1 : len(b)-1])
		if i, ok := s2i[s]; ok {
			*x = T(i)
		} else {
			return fmt.Errorf("invalid %T value (%s)", *x, s)
		}
	default:
		return fmt.Errorf("invalid %T value (%s), not quoted", *x, string(b))
	}
	return nil
}
