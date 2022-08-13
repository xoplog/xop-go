package xoputil

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrealloc(t *testing.T) {
	var buf [100]byte

	p := NewPrealloc(buf[:])

	slices := strings.Split(`
		This is a bunch of short, and somewhat-longer words and
		strung-together-phrases that will get packed into buf
		and use up all of its pre-allocated space`, " ")

	refs := make([][]byte, len(slices))
	for i, s := range slices {
		_ = make([]int, 10)
		refs[i] = p.Pack([]byte(s))
	}
	for i, s := range slices {
		assert.Equal(t, string(s), string(refs[i]))
	}
	last := reflect.ValueOf(refs[0]).Pointer()
	totalLen := len(refs[0])
	for i := 1; i < len(slices); i++ {
		totalLen += len(refs[i])
		if totalLen > 100 {
			break
		}
		current := reflect.ValueOf(refs[i]).Pointer()
		assert.Equalf(t, last+uintptr(len(refs[i-1])), current, "pointer %d with total %d", i, totalLen)
		last = current
	}
}
