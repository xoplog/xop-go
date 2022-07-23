package xoputil

import (
	"strconv"
)

type JBuilder struct {
	B []byte
}

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

// String adds a JSON-encoded string
func (b *JBuilder) String(s string) {
	b.B = append(b.B, '"')
	b.string(s)
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
