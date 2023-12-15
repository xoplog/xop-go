package xoputil

import (
	"io"
	"regexp"
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

// AddString adds a JSON-encoded string, adding quotes and escaping any characters
// that need escaping
func (b *JBuilder) AddString(v string) {
	b.B = append(b.B, '"')
	b.AddStringBody(v)
	b.B = append(b.B, '"')
}

// Quoting required for:
//
//	,	- for separating multiple values in metadata
//	()	- for type annotations
//	' '	- for separating key/value pairs
//	"	- for quoting strings
var punct = "`" + `_~!@#$%^&*\[\]{};'<>.?/`
var safe = `^[-` + punct + `\w](?:[-:` + punct + `\w]*[-` + punct + `\w])?`
var safeRE = regexp.MustCompile(safe + `$`)
var UnquotedConsoleStringRE = regexp.MustCompile(safe)

// excluded:
//   - used for ints
//     / - used in enum
//     () - used for type signatures and lengths
//     " - used for quoted strings
//     space - used to separate attributes
//     : cannot end with :

// AddConsoleString adds a string that may or may not be quoted. Unquoted
// strings are not "t", "f", or have any "/", "(", ")", quotes ("), or spaces.
// These strings are used as both keys and values in xopconsole.
//
// If quoted, quoting is done by strconv.AppendQuote
func (b *JBuilder) AddConsoleString(s string) {
	if safeRE.MatchString(s) && s != "t" && s != "f" {
		b.B = append(b.B, []byte(s)...)
		return
	}
	b.B = strconv.AppendQuote(b.B, s)
}

func (b *JBuilder) AddUint64(i uint64) {
	b.B = strconv.AppendUint(b.B, i, 10)
}

func (b *JBuilder) AddFloat64(f float64) {
	b.B = strconv.AppendFloat(b.B, f, 'f', -1, 64)
}

// AddInt64 may change in the future to obey flags to encode
// large ints as strings
func (b *JBuilder) AddInt64(i int64) {
	b.B = strconv.AppendInt(b.B, i, 10)
}

func (b *JBuilder) AddInt32(i int32) {
	b.B = strconv.AppendInt(b.B, int64(i), 10)
}

func (b *JBuilder) AddBool(v bool) {
	b.B = strconv.AppendBool(b.B, v)
}

// Key() calls Comma() and then adds " k " :
// It skips checking if the key needs escape if FastKeys
// is true.
func (b *JBuilder) AddKey(v string) {
	if b.FastKeys {
		b.AddSafeKey(v)
	} else {
		b.Comma()
		b.AddString(v)
		b.B = append(b.B, ':')
	}
}

func (b *JBuilder) AddSafeKey(v string) {
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
