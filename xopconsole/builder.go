package xopconsole

import (
	"encoding/json"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
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
	b.AppendByte('(')
	b.AddInt64(int64(len(v.Encoded)))
	b.AppendByte(')')
	b.AppendBytes(v.Encoded)
	b.AddSafeString(v.Encoding.String())
	b.AppendByte('/')
	b.AddString(v.ModelType)
}

func (b *Builder) AttributeAny(v xopbase.ModelArg) {
	b.AnyCommon(v)
}

func (b *Builder) AttributeEnum(v xopat.Enum) {
	b.AddInt64(v.Int64())
	b.B = append(b.B, '/')
	b.AddConsoleString(v.String())
}

func (b *Builder) AttributeTime(t time.Time) {
	b.B = append(b.B, '"')
	b.B = DefaultTimeFormatter(b.B, t)
	b.B = append(b.B, '"')
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
	b = fasttime.AppendStrftime(b, fasttime.RFC3339Nano, t)
	return b
}
*/

func DefaultTimeFormatter(b []byte, t time.Time) []byte {
	b = append(b, []byte(t.Format(time.RFC3339Nano))...)
	return b
}
