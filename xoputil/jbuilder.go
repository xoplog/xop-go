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

func (b *JBuilder) AppendByte(v byte) {
	b.B = append(b.B, v)
}

// AppendBytes adds the bytes without wrapping or checking
func (b *JBuilder) AppendBytes(v []byte) {
	b.B = append(b.B, v...)
}

// AppendString adds the bytes without wrapping or checking
func (b *JBuilder) AppendString(v string) {
	b.B = append(b.B, v...)
}

// Write allows JBuilder to be an io.Writer
func (b *JBuilder) Write(v []byte) (int, error) {
	b.B = append(b.B, v...)
	return len(v), nil
}

func (b *JBuilder) Reset() {
	b.B = b.B[:0]
}

// AddString adds a JSON-encoded string that is known to not need escaping
func (b *JBuilder) AddSafeString(v string) {
	b.B = append(b.B, '"')
	b.AppendString(v)
	b.B = append(b.B, '"')
}

// AddString adds a JSON-encoded string
func (b *JBuilder) AddString(v string) {
	b.B = append(b.B, '"')
	b.AddStringBody(v)
	b.B = append(b.B, '"')
}

func (b *JBuilder) AddUint64(i uint64) {
	b.B = strconv.AppendUint(b.B, i, 10)
}

func (b *JBuilder) AddFloat64(f float64) {
	b.B = strconv.AppendFloat(b.B, f, 'f', -1, 64)
}

func (b *JBuilder) AddInt64(i int64) {
	b.B = strconv.AppendInt(b.B, i, 10)
}

func (b *JBuilder) AddBool(v bool) {
	b.B = strconv.AppendBool(b.B, v)
}

// Key() calls Comma() and then adds " k " :
// It skips checking if the key needs escape if FastKeys
// is true.
func (b *JBuilder) AddKey(v string) {
	if b.FastKeys {
		b.AddUncheckedKey(v)
	} else {
		b.Comma()
		b.AddString(v)
		b.B = append(b.B, ':')
	}
}

func (b *JBuilder) AddUncheckedKey(v string) {
	b.Comma()
	b.B = append(b.B, '"')
	b.B = append(b.B, v...)
	b.B = append(b.B, '"', ':')
}

func BuildKey(v string) []byte {
	b := &JBuilder{}
	b.B = append(b.B, ',')
	b.AddString(v)
	b.B = append(b.B, ':')
	return b.B
}
