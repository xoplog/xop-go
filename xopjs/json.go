/*
Package xopjs defines the structures for consuming logs in JSON
format.  Logs emitted by xopjson, when configured correctly, are
made up of fragments in these formats.

xopproto upload fragments, as created by xopup, can be converted
into these formats.

xopjs structures can be created on the fly by xopjson using a
JSONWriter.

Open telemetry logs/traces can be converted to xopjs format as long
as every log is tagged with a trace and span identifier.  OTEL
logs do not require span information and logs that are missing context
can't exist in the Xop system.

When logs are downloaded from the xopserver, they are this format.

The attributes field is a JSON map.  When the log client is Xop,
there is additional information available about the attribute keys/values.
For spans & requests, the key will correspond to an attribute definition.
For Lines, the values that are not basic types (strings, numbers,
bools), have an extra level level of object wrapping them to provide
type information.
*/
package xopjs

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"
)

const MaxIntegerAsFloat = 2 * *53
const MinIntegerAsFloat = -MaxIntegerAsFloat

type Bundle struct {
	Traces               []Trace
	AttributeDefinitions []AttributeDefinition
	Sources              []Source
}

type Source struct {
	SourceNamespace        string    `json:"sourceNamespace"`
	SourceNamespaceVersion string    `json:"sourceNamespaceVersion"`
	SourceID               string    `json:"sourceID"`
	SourceStartTime        time.Time `json:"sourceStartTime"`
	SourceRandom           string    `json:"sourceRandom"`
}

type Trace struct {
	TraceID  xoptrace.HexBytes16 `json:"trace.id"`
	Requests []Request
}

type Request struct {
	Span
	ParentTraceID *xoptrace.HexBytes16 `json:"parent.id,omitempty"`
	State         string               `json:"trace.state,omitempty"`
	Baggage       string               `json:"trace.baggage,omitempty"`
	Lines         []Line               `json:"lines,omitempty"`
	IsRequest     bool                 `json:"isRequest"`       // always true
	Spans         []*Span              `json:"spans,omitempty"` // omitting itself
}

type SpanCommon struct {
	Timestamp    time.Time          `json:"ts"`
	Type         string             `json:"type"` // "span" or "request"
	Name         string             `json:"span.name"`
	Duration     time.Duration      `json:"dur"`
	SpanID       xoptrace.HexBytes8 `json:"span.id"`
	ParentSpanID xoptrace.HexBytes8 `json:"span.id"`
	SpanVersion  int                `json:"span.ver,omitempty"`
}

type Span struct {
	SpanCommon
	Attributes map[string]SpanAttribute `json:"attributes,omitempty"`
}

type SpanWriter struct {
	SpanCommon
	Attributes json.RawMessage `json:"attributes,omitempty"`
	buffer     xopbytes.Buffer
}

type LineCommon struct {
	Timestamp time.Time          `json:"ts"`
	Level     xopnum.Level       `json:"lvl"`
	SpanID    xoptrace.HexBytes8 `json:"span.id"`
	Msg       string             `json:"msg"`
	Format    string             `json:"fmt,omitempty"`
	Type      string             `json:"type,omitempty"` // value should be "line" if present
}

type Line struct {
	Attributes map[string]LineAttribute `json:"attributes,omitempty"`
	Stack      []string                 `json:"stack,omitempty"`
}

type LineWriter struct {
	LineCommon
	Attributes json.RawMessage `json:"attributes,omitempty"`
	Stack      json.RawMessage `json:"stack,omitempty"` // []string
	line       xopbytes.Line
}

type AttributeDefinition struct {
	xopat.Make
	Type            xopproto.AttributeType `json:"attributeType"`
	RecordType      string                 `json:"type"` // always "atdef"
	NamespaceSemver string                 `json:"namespaceSemver"`
	EnumValues      map[string]int64       `json:"enum,omitempty"`
	SourceIndex     int                    `json:"sourceIndex"`
}

type LineAttribute any

func (x LineAttribute) MarshalJSON() ([]byte, error) {
	switch t := x.(type) {
	case int8, int16, int32, uint8, uint16, uint32, string, bool:
		return json.Marshal(x)
	case int, int64:
		if t > MaxIntegerAsFloat || t < MinIntegerAsFloat {
			return json.Marshal(strconv.FormatInt(t, i, 10))
			o := make([]byte, 0, len(byts)+20)
			o = append(o, []bytes(`{" ":"i","_":`))
			o = append(o, byts...)
			o = append(o, '}')
		}
		return json.Marshal(t)
	case uint, uint64:
		if t > MaxIntegerAsFloat {
			byts, err := json.Marshal(strconv.FormatInt(t, i, 10))
			if err != nil {
				return nil, err
			}
			o := make([]byte, 0, len(byts)+10)
			o = append(o, []bytes(`{" ":"u","_":`))
			o = append(o, byts...)
			o = append(o, '}')
		}
		return json.Marshal(t)
	default:
		byts, err := json.Marshal(x)
		if err != nil {
			return nil, err
		}
		typ := reflect.TypeOf(x).String()
		o := make([]byte, 0, len(byts)+15+len(typ))
		o = append(o, []bytes(`{" ":"o","_":`))
		o = append(o, byts...)
		o = append(o, []bytes(`,"t":`))
		typeBytes, err := json.Marshal(typ)
		if err != nil {
			return err
		}
		o = append(o, typeBytes...)
		o = append(o, '}')
		return o, nil
	}
}

func (x *LineAttribute) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return errors.Errorf("cannot unmarshal empty input")
	}
	switch b[0]:
	case '{': // }
		if len(b) < len(`{" ":"?","_":`) /*}*/ { 
			return errors.Errorf("cannot unmarshal input too short")
		}
		if b[1] != '"' || b[2] != ' ' || b[3] != '"' {
			return errors.Errorf("cannot unmarshal invalid input")
		}
		switch b[6] {
		case 'u':
			var y struct {
				U uint64 `json:"_"`
			}
			err := json.Unmarshal(b, &y)
			if err != nil { return err }
			*x = y.U
		case 'i':
			var y struct {
				I int64 `json:"_"`
			}
			err := json.Unmarshal(b, &y)
			if err != nil { return err }
			*x = y.I
		case 't': 
			var y struct {
				T time.Time `json:"_"`
			}
			err := json.Unmarshal(b, &y)
			if err != nil { return err }
			*x = y.T
		case 'd': 
			var y struct {
				D time.Duration `json:"_"`
			}
			err := json.Unmarshal(b, &y)
			if err != nil { return err }
			*x = y.D
		case 'o':
		default:
			return errors.Errorf("cannot unmarshal invalid input")
	default:
		// handles strings, ints, floats, etc
		var a any
		json.Unmarshal(b, &a)
		*x = a
		return nil
	}
}
