package xop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase26(t *testing.T) {
	tc := map[int]string{
		0:  "A",
		1:  "B",
		25: "Z",
		26: "BA",
		27: "BB",
		52: "CA",
		-1: "-B",
	}
	for num, want := range tc {
		assert.Equalf(t, want, base26(num), "base26(%d)", num)
	}
}
