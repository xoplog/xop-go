package xopconsole

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneString(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		want      string
		remainder string
	}{
		{
			input: "",
			want:  "",
		},
		{
			input:     "foo xyz",
			want:      "foo",
			remainder: " xyz",
		},
		{
			input:     "foo=bar xyz",
			want:      "foo",
			remainder: "=bar xyz",
		},
		{
			input: "foo-bar=bar xyz",
			want:  "foo-bar",
		},
		{
			name:  "oddchars",
			input: "foo-'$#@[]bar=bar xyz",
			want:  "foo-'$#@[]bar",
		},
		{
			name:      "quoted",
			input:     strconv.Quote(`f"oo-'$#\@[]bar`) + "=bar xyz",
			want:      `f"oo-'$#\@[]bar`,
			remainder: "=bar xyz",
		},
	}
	for _, tc := range cases {
		name := tc.name
		if name == "" {
			name = tc.input
		}
		t.Run(name, func(t *testing.T) {
			got, gotRemainder := oneString(tc.input)
			var wantRemainder string
			if tc.remainder == "" {
				wantRemainder = tc.input[len(tc.want):]
			} else {
				wantRemainder = tc.remainder
			}
			if assert.Equal(t, tc.want, got, "string") {
				assert.Equal(t, wantRemainder, gotRemainder, "remainder")
			}
		})
	}
}

func TestOneWord(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		breakOn   string
		want      string
		remainder string
		wantSep   byte
	}{
		{
			name:      "regression",
			input:     `"a test {foo} with {num}" foo=bar num=38`,
			breakOn:   " ",
			want:      "a test {foo} with {num}",
			wantSep:   ' ',
			remainder: "foo=bar num=38",
		},
	}
	for _, tc := range cases {
		name := tc.name
		if name == "" {
			name = tc.input
		}
		t.Run(name, func(t *testing.T) {
			got, sep, gotRemainder := oneWordMaybeQuoted(tc.input, tc.breakOn)
			var wantRemainder string
			if tc.remainder == "" {
				wantRemainder = tc.input[len(tc.want):]
			} else {
				wantRemainder = tc.remainder
			}
			if assert.Equal(t, tc.want, got, "string") {
				assert.Equal(t, wantRemainder, gotRemainder, "remainder")
			}
			assert.Equal(t, string(tc.wantSep), string(sep), "sep")
		})
	}
}
