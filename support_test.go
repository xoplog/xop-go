package xm

import (
	"testing"
)

func TestBase26(t *testing.T) {
	assert.Equal(t, "A", base26(1), "0")
	assert.Equal(t, "Z", base26(26), "25")
	assert.Equal(t, "AA", base26(27), "26")
}
