package xoptestutil_test

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/xoplog/xop-go/xoptest/xoptestutil"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		name       string
		timeFormat string
		divide     time.Duration
		gmt        bool
		rounding   time.Duration
	}{
		{
			divide:   time.Nanosecond,
			rounding: time.Nanosecond,
		},
		{
			divide:   time.Microsecond,
			rounding: time.Microsecond,
		},
		{
			divide:   time.Millisecond,
			rounding: time.Millisecond,
		},
		{
			divide:   time.Second,
			rounding: time.Second,
		},
		{
			name:       "rfc3339 GMT",
			timeFormat: time.RFC3339,
			rounding:   time.Second,
			gmt:        true,
		},
		{
			name:       "rfc3339",
			timeFormat: time.RFC3339,
			rounding:   time.Second,
		},
		{
			name:       "rfc3339 nano",
			timeFormat: time.RFC3339Nano,
			rounding:   time.Microsecond,
		},
		{
			name:       "rfc3339 nano gmt",
			timeFormat: time.RFC3339Nano,
			rounding:   time.Microsecond,
			gmt:        true,
		},
		{
			name:       "ansic",
			timeFormat: time.ANSIC,
			gmt:        true,
			rounding:   time.Second,
		},
		{
			name:       "unixdate",
			timeFormat: time.UnixDate,
			rounding:   time.Second,
		},
		{
			name:       "rubydate",
			timeFormat: time.RubyDate,
			rounding:   time.Second,
		},
		{
			name:       "rfc822",
			timeFormat: time.RFC822,
			gmt:        true,
			rounding:   time.Minute,
		},
	}

	for _, tc := range cases {
		tc := tc
		name := tc.name
		if name == "" {
			name = tc.timeFormat
		}
		if name == "" {
			name = tc.divide.String()
		}
		t.Run(name, func(t *testing.T) {
			n := time.Now()
			if tc.rounding != 0 {
				n = n.Round(tc.rounding)
			}
			if tc.gmt {
				n = n.UTC()
			}
			var s string
			switch {
			case tc.timeFormat != "":
				s = n.Format(tc.timeFormat)
			case tc.divide != 0:
				s = strconv.FormatInt(n.UnixNano()/int64(tc.divide), 10)
			default:
				t.FailNow()
			}
			t.Log("s =", s)

			var x struct {
				T xoptestutil.TS
			}
			err := json.Unmarshal([]byte(`{"T":"`+s+`"}`), &x)
			if assert.NoError(t, err) {
				assert.Truef(t, n.Equal(x.T.Time), "%s vs %s", n, x.T.Time)
			}
		})
	}
}
