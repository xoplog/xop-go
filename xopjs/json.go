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
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"
)

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
	Spans         []*Span               `json:"spans,omitempty"` // omitting itself
}

type Span struct {
	Timestamp    time.Time          `json:"ts"`
	Attributes   json.RawMessage    `json:"attributes,omitempty"`
	Type         string             `json:"type"` // "span" or "request"
	Name         string             `json:"span.name"`
	Duration     time.Duration      `json:"dur"`
	SpanID       xoptrace.HexBytes8 `json:"span.id"`
	ParentSpanID xoptrace.HexBytes8 `json:"span.id"`
	SpanVersion  int                `json:"span.ver,omitempty"`
	buffer       xopbytes.Buffer
}

type Line struct {
	Timestamp  time.Time          `json:"ts"`
	Attributes json.RawMessage    `json:"attributes,omitempty"`
	Level      xopnum.Level       `json:"lvl"`
	SpanID     xoptrace.HexBytes8 `json:"span.id"`
	Stack      json.RawMessage    `json:"stack,omitempty"` // []string
	Msg        string             `json:"msg"`
	Format     string             `json:"fmt,omitempty"`
	Type       string             `json:"type,omitempty"` // value should be "line" if present
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
