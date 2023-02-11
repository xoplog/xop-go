package xopbase

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"

	"github.com/xoplog/xop-go/xopproto"
)

// ModelArg may be expanded in the future to supply: an Encoder; redaction
// information.
type ModelArg struct {
	// If specified, overrides what would be provided by reflect.TypeOf(obj).Name()
	TypeName string
	Model    interface{}
	// Encoding and Encoded can be set for models that are already encoded. If they're
	// set, then Model will be ignored.
	Encoding xopproto.Encoding
	Encoded  []byte
	// TODO: extra fields for redacted models
}

// Calls to Encode are idempotent but not thread-safe
func (m *ModelArg) Encode() {
	if m.TypeName == "" && m.Model != nil {
		m.TypeName = reflect.TypeOf(m.Model).Name()
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
