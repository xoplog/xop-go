package xm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase26(t *testing.T) {
	assert.Equal(t, "A", base26(0), "0")
	assert.Equal(t, "Z", base26(25), "25")
	assert.Equal(t, "BA", base26(26), "26")
	assert.Equal(t, "BB", base26(27), "27")
	assert.Equal(t, "CA", base26(52), "52")
}
