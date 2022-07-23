package xoputil

/*

This file contains code that is direclty copied or derrived from
https://github.com/phuslu/log

The original is subject to the following license.

MIT License

Copyright (c) 2022 Phus Lu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

var escapes = [256]bool{
	'"':  true,
	'<':  true,
	'\'': true,
	'\\': true,
	'\b': true,
	'\f': true,
	'\n': true,
	'\r': true,
	'\t': true,
}

func (b *JBuilder) escapeb(n []byte) {
	l := len(n)
	j := 0
	if l > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = n[l-1]
	}
	for i := 0; i < l; i++ {
		switch n[i] {
		case '"':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', '"')
			j = i + 1
		case '\\':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', '\\')
			j = i + 1
		case '\n':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'n')
			j = i + 1
		case '\r':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'r')
			j = i + 1
		case '\t':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 't')
			j = i + 1
		case '\f':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			b.B = append(b.B, n[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	b.B = append(b.B, n[j:]...)
}

func (b *JBuilder) escapes(s string) {
	n := len(s)
	j := 0
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = s[n-1]
	}
	for i := 0; i < n; i++ {
		switch s[i] {
		case '"':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', '"')
			j = i + 1
		case '\\':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', '\\')
			j = i + 1
		case '\n':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'n')
			j = i + 1
		case '\r':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'r')
			j = i + 1
		case '\t':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 't')
			j = i + 1
		case '\f':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			b.B = append(b.B, s[j:i]...)
			b.B = append(b.B, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	b.B = append(b.B, s[j:]...)
}

func (b *JBuilder) string(s string) {
	for _, c := range []byte(s) {
		if escapes[c] {
			b.escapes(s)
			return
		}
	}
	b.B = append(b.B, s...)
}

func (b *JBuilder) bytes(n []byte) {
	for _, c := range n {
		if escapes[c] {
			b.escapeb(n)
			return
		}
	}
	b.B = append(b.B, n...)
}

/*
// Interface adds the field key with i marshaled using reflection.
func (e *Entry) Interface(key string, i interface{}) *Entry {
	if e == nil {
		return nil
	}

	b.B = append(b.B, ',', '"')
	b.B = append(b.B, key...)
	b.B = append(b.B, '"', ':', '"')
	b := bbpool.Get().(*bb)
	b.B = b.B[:0]
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(i)
	if err != nil {
		b.B = b.B[:0]
		fmt.Fprintf(b, "marshaling error: %+v", err)
	} else {
		b.B = b.B[:len(b.B)-1]
	}
	b.bytes(b.B)
	b.B = append(b.B, '"')
	if cap(b.B) <= bbcap {
		bbpool.Put(b)
	}

	return e
}

*/
