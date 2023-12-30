package xopat_test

import (
	"testing"

	"github.com/xoplog/xop-go/xopat"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	k := xopat.K(`foo " bar`)
	jsBody := `foo \" bar`
	js := `"` + jsBody + `"`

	assert.Equal(t, string(k), k.String())
	assert.Equal(t, js, string(k.JSON()))
	assert.Equal(t, jsBody, string(k.JSONBody()))
}
