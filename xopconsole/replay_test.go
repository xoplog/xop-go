package xopconsole

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneString(t *testing.T) {
	cases := []struct {
		name string
		input string
		want  string
		remainder string
	}{
		{
			input: "",
			want:  "",
		},
		{
			input: "foo=bar xyz",
			want:  "foo",
		},
		{
			input: "foo-bar=bar xyz",
			want:  "foo-bar",
		},
		{
			name: "oddchars",
			input: "foo-'$#@[]bar=bar xyz",
			want:  "foo-'$#@[]bar",
		},
		{
			name: "quoted",
			input: strconv.Quote(`f"oo-'$#\@[]bar`) + "=bar xyz",
			want:  `f"oo-'$#\@[]bar`,
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
