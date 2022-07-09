package xoplog

import (
	"strconv"
)

func base26(v int) string {
	b26 := []byte(strconv.FormatInt(int64(v), 26))
	for i := len(b26) - 1; i >= 0; i-- {
		if b26[i] == '-' {
			// do not touch
		} else if b26[i] >= 'a' {
			b26[i] = byte(int(b26[i]) + 10 + int('A') - int('a'))
		} else {
			b26[i] += 'A' - '0'
		}
	}
	return string(b26)
}
