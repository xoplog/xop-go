package xoputil

import (
	"io"
	"strconv"
)

type JBuilder struct {
	B        []byte
	FastKeys bool
}

var _ io.Writer = &JBuilder{}

// Comma adds a comma if a comma is needed based
// on what's already in the JBuilder: if the previous
// character is '{', '[', or ':' then it does not add a
// comma.  Otherwise it does.
func (b *JBuilder) Comma() {
	if len(b.B) == 0 {
		return
	}
	switch b.B[len(b.B)-1] {
	case '[', '{', ':':
		return
	}
	b.B = append(b.B, ',')
}

func (b *JBuilder) Byte(v byte) { // XXX rename: AppendByte
	b.B = append(b.B, v)
}

func (b *JBuilder) Append(v []byte) { // XXX rename: AppendBytes
	b.B = append(b.B, v...)
}

func (b *JBuilder) AppendString(v string) {
	b.B = append(b.B, v...)
}

func (b *JBuilder) Write(v []byte) (int, error) {
	b.B = append(b.B, v...)
	return len(v), nil
}

func (b *JBuilder) Reset() {
	b.B = b.B[:0]
}

// String adds a JSON-encoded string
func (b *JBuilder) String(v string) {
	b.B = append(b.B, '"')
	b.string(v)
	b.B = append(b.B, '"')
}

func (b *JBuilder) Uint64(i uint64) {
	b.B = strconv.AppendUint(b.B, i, 64)
}

func (b *JBuilder) Float64(f float64) {
	b.B = strconv.AppendFloat(b.B, f, 'f', -1, 64)
}

func (b *JBuilder) Int64(i int64) {
	b.B = strconv.AppendInt(b.B, i, 64)
}

func (b *JBuilder) Bool(v bool) {
	b.B = strconv.AppendBool(b.B, v)
}

// Key() calls Comma() and then adds " k " :
// It skips checking if the key needs escape if FastKeys
// is true.
func (b *JBuilder) Key(v string) {
	if b.FastKeys {
		b.UncheckedKey(v)
	} else {
		b.Comma()
		b.String(v)
		b.B = append(b.B, ':')
	}
}

func (b *JBuilder) UncheckedKey(v string) {
	b.Comma()
	b.B = append(b.B, '"')
	b.B = append(b.B, v...)
	b.B = append(b.B, '"', ':')
}

// UncheckedString does not check to see if the string
// has any characters in it that would need JSON escaping.
func (b *JBuilder) UncheckedString(v string) {
	b.B = append(b.B, '"')
	b.B = append(b.B, v...)
	b.B = append(b.B, '"')
}
