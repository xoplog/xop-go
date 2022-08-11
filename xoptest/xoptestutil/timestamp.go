package xoptestutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
)

// TS encodes time in many formats.  As as number, it looks for
// timestamps that are in reasonable ranges.
type TS struct {
	time.Time
}

var (
	tNano   = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	tMicro  = tNano / int64(time.Microsecond)
	tMilli  = tNano / int64(time.Millisecond)
	tSecond = tNano / int64(time.Second)
)

var lastIndexFound int32

var formats = []struct {
	fmt      string
	re       *regexp.Regexp
	isInt    bool
	isFloat  bool
	compiled func(string) (time.Time, error)
}{
	{
		re:    regexp.MustCompile(`^\d+$`),
		isInt: true,
	},
	{
		re:      regexp.MustCompile(`^\d\.\d+$`),
		isFloat: true,
	},
	{
		fmt: time.RFC3339,
		re:  regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d[T ]\d\d:\d\d:\d\dZ\d\d:\d\d`),
	},
	{
		fmt: time.RFC3339Nano,
		re:  regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d[T ]\d\d:\d\d:\d\d\.\d+Z\d\d:\d\d`),
	},
	{
		fmt: time.ANSIC,
		re:  regexp.MustCompile(`^(?:Mon|Tue|Wed|Thr|Fri|Sat|Sun) (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d? \d\d:\d\d:\d\d \d\d\d\d$`),
	},
	{
		fmt: time.UnixDate,
		re:  regexp.MustCompile(`^(?:Mon|Tue|Wed|Thr|Fri|Sat|Sun) (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d? \d\d:\d\d:\d\d [A-Z]{3} \d\d\d\d$`),
	},
	{
		fmt: time.RubyDate,
		re:  regexp.MustCompile(`^(?:Mon|Tue|Wed|Thr|Fri|Sat|Sun) (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d? \d\d:\d\d:\d\d [-+]\d\d\d\d \d\d\d\d$`),
	},
	{
		fmt: time.RFC822,
		re:  regexp.MustCompile(`^\d\d (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d \d\d:\d\d [A-Z]{3}$`),
	},
	{
		fmt: time.RFC822Z,
		re:  regexp.MustCompile(`^\d\d (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d \d\d:\d\d [-+]\d\d\d\d$`),
	},
	{
		fmt: time.RFC850,
		re:  regexp.MustCompile(`^(?:Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday), \d\d-(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-\d\d \d\d:\d\d:\d\d [A-Z]{3}$`),
	},
	{
		fmt: time.RFC1123,
		re:  regexp.MustCompile(`^(?:Mon|Tue|Wed|Thr|Fri|Sat|Sun), \d\d (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d\d\d \d\d:\d\d:\d\d [A-Z]{3}$`),
	},
	{
		fmt: time.RFC1123Z,
		re:  regexp.MustCompile(`^(?:Mon|Tue|Wed|Thr|Fri|Sat|Sun), \d\d (?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d\d\d\d \d\d:\d\d:\d\d [-+]\d\d\d\d$`),
	},
}

func init() {
	for i, format := range formats {
		format := format
		switch {
		case format.isInt:
			formats[i].compiled = func(s string) (time.Time, error) {
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return time.Time{}, err
				}
				return handleInt(i), nil
			}
		case format.isFloat:
			formats[i].compiled = func(s string) (time.Time, error) {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return time.Time{}, err
				}
				t, err := handleFloat(f)
				return t, err
			}
		default:
			formats[i].compiled = func(s string) (time.Time, error) {
				return time.Parse(s, format.fmt)
			}
		}
	}
}

func handleFloat(f float64) (time.Time, error) {
	if f > float64(math.MaxInt64) {
		return time.Time{}, fmt.Errorf("invalid timestamp (too big)")
	}
	if f < float64(math.MinInt64) {
		return time.Time{}, fmt.Errorf("invalid timestamp (too small)")
	}
	return handleInt(int64(f)), nil
}

func handleInt(i int64) time.Time {
	switch {
	case i > 0 && i < tSecond:
		i *= int64(time.Second)
	case i > 0 && i < tMilli:
		i *= int64(time.Millisecond)
	case i > 0 && i < tMicro:
		i *= int64(time.Microsecond)
	}
	return time.Unix(0, i)
}

func (ts *TS) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("invalid timestamp, no content")
	}
	switch b[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if bytes.IndexByte(b, '.') == -1 {
			var i int64
			if err := json.Unmarshal(b, &i); err != nil {
				return err
			}
			*ts = TS{handleInt(i)}
			return nil
		}
		var f float64
		if err := json.Unmarshal(b, &f); err != nil {
			return err
		}
		t, err := handleFloat(f)
		*ts = TS{t}
		return err
	case '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		last := atomic.LoadInt32(&lastIndexFound)
		if t, err := formats[last].compiled(s); err != nil {
			*ts = TS{t}
			return nil
		}
		for i, f := range formats {
			if f.re.MatchString(s) {
				t, err := f.compiled(s)
				if err != nil {
					return err
				}
				*ts = TS{t}
				atomic.StoreInt32(&lastIndexFound, int32(i))
			}
		}
		return fmt.Errorf("no for time format")
	default:
		return fmt.Errorf("invalid time format")
	}
}
