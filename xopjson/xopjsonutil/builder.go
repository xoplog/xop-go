package xopjsonutil

import (
	"encoding/json"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"
)

type Builder struct {
	xoputil.JBuilder
	encoder *json.Encoder
}

func (b *Builder) Init() {
	b.encoder = json.NewEncoder(&b.JBuilder)
	b.encoder.SetEscapeHTML(false)
}

func (b *Builder) AnyCommon(v xopbase.ModelArg) {
	v.Encode()
	if v.Encoding == xopproto.Encoding_JSON {
		b.AppendBytes(v.Encoded)
	} else {
		b.AddString(string(v.Encoded))
		b.AppendBytes([]byte(`,"encoding":`))
		b.AddSafeString(v.Encoding.String())
	}
	b.AppendBytes([]byte(`,"modelType":`))
	b.AddString(v.ModelType)
}

func (b *Builder) AttributeAny(v xopbase.ModelArg) {
	b.AppendBytes([]byte(`{"v":`)) // }
	b.AnyCommon(v)
	// {
	b.AppendByte('}')
}

func (b *Builder) AttributeEnum(v xopat.Enum) {
	b.AppendBytes([]byte(`{"s":`))
	b.AddString(v.String())
	b.AppendBytes([]byte(`,"i":`))
	b.AddInt64(v.Int64())
	b.AppendByte('}')
}

func (b *Builder) AttributeTime(t time.Time) {
	b.B = DefaultTimeFormatter(b.B, t)
}

func (b *Builder) AttributeBool(v bool) {
	b.AddBool(v)
}

func (b *Builder) AttributeInt64(v int64) {
	b.AddInt64(v)
}

func (b *Builder) AttributeString(v string) {
	b.AddString(v)
}

func (b *Builder) AttributeFloat64(f float64) {
	b.AddFloat64(f)
}

func (b *Builder) AttributeDuration(v time.Duration) {
	b.AddInt64(int64(v / time.Nanosecond))
}

func (b *Builder) AttributeLink(v xoptrace.Trace) {
	b.AddSafeString(v.String())
}

/* TODO
func DefaultTimeFormatter2(b []byte, t time.Time) []byte {
	b = append(b, '"')
	b = fasttime.AppendStrftime(b, fasttime.RFC3339Nano, t)
	b = append(b, '"')
	return b
}
*/

func DefaultTimeFormatter(b []byte, t time.Time) []byte {
	b = append(b, '"')
	b = append(b, []byte(t.Format(time.RFC3339Nano))...)
	b = append(b, '"')
	return b
}
