package xopbase

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"

	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoputil"
)

// ModelArg may be expanded in the future to supply: an Encoder; redaction
// information.
type ModelArg struct {
	// If specified, overrides what would be provided by reflect.TypeOf(obj).Name()
	// Encoding and Encoded can be set for models that are already encoded. If they're
	// set, then Model will be ignored.
	Encoding xopproto.Encoding `json:"encoding"`
	Encoded  []byte            `json:"v"`
	TypeName string            `json:"modelType"`
	Model    interface{}       `json:"-"`
	// TODO: extra fields for redacted models
}

// Calls to Encode are idempotent but not thread-safe
func (m *ModelArg) Encode() {
	if m.TypeName == "" && m.Model != nil {
		m.TypeName = reflect.TypeOf(m.Model).String()
	}
	if len(m.Encoded) != 0 {
		return
	}
	switch m.Encoding {
	case xopproto.Encoding_Unset:
		m.Encoding = xopproto.Encoding_JSON
		fallthrough
	case xopproto.Encoding_JSON:
		enc, err := json.Marshal(m.Model)
		if err != nil {
			m.Encoded = []byte(err.Error())
			m.Encoding = xopproto.Encoding_ErrorMessage
		} else {
			m.Encoded = enc
		}
	case xopproto.Encoding_XML:
		enc, err := xml.Marshal(m.Model)
		if err != nil {
			m.Encoded = []byte(err.Error())
			m.Encoding = xopproto.Encoding_ErrorMessage
		} else {
			m.Encoded = enc
		}
	case xopproto.Encoding_CSV:
		switch t := m.Model.(type) {
		case SimpleTable:
			var b bytes.Buffer
			w := csv.NewWriter(&b)
			err := w.Write(t.Header())
			if err != nil {
				m.Encoded = []byte(err.Error())
				m.Encoding = xopproto.Encoding_ErrorMessage
				return
			}
			err = w.WriteAll(t.Rows())
			if err != nil {
				m.Encoded = []byte(err.Error())
				m.Encoding = xopproto.Encoding_ErrorMessage
				return
			}
			m.Encoded = b.Bytes()
		default:
			m.Encoded = []byte(fmt.Sprintf("do not know how to turn %T into a CSV", m.Model))
			m.Encoding = xopproto.Encoding_ErrorMessage
		}
	case xopproto.Encoding_ErrorMessage:
		m.Encoded = []byte("Error is not a valid encoding")
	default:
		m.Encoded = []byte(fmt.Sprintf("unknown encoding %s", m.Encoding))
		m.Encoding = xopproto.Encoding_ErrorMessage
	}
}

type SimpleTable interface {
	Header() []string
	Rows() [][]string
}

var _ json.Marshaler = ModelArg{}
var _ json.Unmarshaler = &ModelArg{}

func (m ModelArg) MarshalJSON() ([]byte, error) {
	m.Encode()
	b := xoputil.JBuilder{
		B:        make([]byte, 0, len(m.Encoded)+len(m.TypeName)+30),
		FastKeys: true,
	}
	b.AppendBytes([]byte(`{"v":`))
	if m.Encoding == xopproto.Encoding_JSON {
		b.AppendBytes(m.Encoded)
	} else {
		b.AddString(string(m.Encoded))
		b.AppendBytes([]byte(`,"encoding":`))
		b.AddSafeString(m.Encoding.String())
	}
	b.AppendBytes([]byte(`,"modelType":`))
	b.AddString(m.TypeName)
	b.AppendByte('}')
	return b.B, nil
}

func (m *ModelArg) UnmarshalJSON(b []byte) error {
	var decode struct {
		Encoding xopproto.Encoding `json:"encoding"`
		Encoded  json.RawMessage   `json:"v"`
		TypeName string            `json:"modelType"`
	}
	err := json.Unmarshal(b, &decode)
	if err != nil {
		return err
	}
	m.Encoded = []byte(decode.Encoded)
	m.Model = nil
	if decode.Encoding == xopproto.Encoding_Unset {
		m.Encoding = xopproto.Encoding_JSON
	} else {
		m.Encoding = decode.Encoding
	}
	return nil
}
