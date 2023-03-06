package xopproto_test

import (
	"encoding/json"
	"testing"

	"github.com/xoplog/xop-go/xopproto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decode[T any](b []byte) (interface{}, error) {
	var t T
	err := json.Unmarshal(b, &t)
	return t, err
}

func TestEncoding(t *testing.T) {
	cases := []struct {
		bad    string
		value  interface{}
		want   interface{}
		decode func([]byte) (interface{}, error)
	}{
		{
			value:  xopproto.Encoding_JSON,
			want:   xopproto.Encoding_JSON,
			decode: decode[xopproto.Encoding],
		},
		{
			value:  int(xopproto.Encoding_JSON),
			want:   xopproto.Encoding_JSON,
			decode: decode[xopproto.Encoding],
		},
		{
			value:  xopproto.AttributeType_String,
			want:   xopproto.AttributeType_String,
			decode: decode[xopproto.AttributeType],
		},
		{
			value:  int(xopproto.AttributeType_Int32),
			want:   xopproto.AttributeType_Int32,
			decode: decode[xopproto.AttributeType],
		},
		{
			bad:    "392023292",
			decode: decode[xopproto.AttributeType],
		},
		{
			bad:    `"salj"`,
			decode: decode[xopproto.AttributeType],
		},
		{
			bad:    "true",
			decode: decode[xopproto.AttributeType],
		},
	}

	for _, tc := range cases {
		if tc.bad != "" {
			_, err := tc.decode([]byte(tc.bad))
			t.Logf("error from '%s': %s", tc.bad, err)
			assert.Error(t, err, tc.bad)
			continue
		}
		enc, err := json.Marshal(tc.value)
		require.NoErrorf(t, err, "encode %+v", tc.value)
		t.Logf("encoded %v = %s", tc.value, string(enc))
		got, err := tc.decode(enc)
		require.NoErrorf(t, err, "decode '%s' to %T", string(enc), got)
		assert.Equal(t, tc.want, got)
	}
}
