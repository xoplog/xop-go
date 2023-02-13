package xoptest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptest"

	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type selfRedactor struct {
	V string
}

func (r selfRedactor) Redact() interface{} {
	r.V = strings.ReplaceAll(r.V, "bribe", "consideration")
	return r
}

func (r selfRedactor) String() string {
	return r.V
}

type redactor interface {
	Redact() interface{}
}

func TestRedaction(t *testing.T) {
	tLog := xoptest.New(t)
	log := xop.NewSeed(
		xop.WithBase(tLog),
		xop.WithSettings(func(settings *xop.LogSettings) {
			settings.SetRedactStringFunc(func(baseLine xopbase.Line, k string, v string) {
				v = strings.ReplaceAll(v, "sunflower", "daisy")
				baseLine.String(k, v, xopbase.StringDataType)
			})
			settings.SetRedactAnyFunc(func(baseLine xopbase.Line, k string, v interface{}, alreadyImmutable bool) {
				if !alreadyImmutable {
					v = deepcopy.Copy(v)
				}
				if canRedact, ok := v.(redactor); ok {
					baseLine.Any(k, xopbase.ModelArg{
						Model: canRedact.Redact(),
					})
				} else {
					baseLine.Any(k, xopbase.ModelArg{
						Model: v,
					})
				}
			})
			settings.SetRedactErrorFunc(func(baseLine xopbase.Line, k string, v error) {
				baseLine.String(k, v.Error()+"(as string)", xopbase.ErrorDataType)
			})
		}),
	).Request(t.Name())

	a := selfRedactor{V: "I got the contract with a small bribe, just a sunflower cookie"}

	log.Info().
		String("garden", "nothing in my garden is taller than my sunflower!").
		Any("story", a).
		Any("tale", a).
		AnyWithoutRedaction("raw", a).
		Stringer("success", a).
		Error("oops", fmt.Errorf("outer: %w", fmt.Errorf("inner"))).
		Msg("foo")

	foos := tLog.FindLines(xoptest.MessageEquals("foo"))
	require.NotEmpty(t, foos, "foo line")

	assert.Equal(t, "nothing in my garden is taller than my daisy!", foos[0].Data["garden"], "garden")
	assert.Equal(t, "I got the contract with a small consideration, just a sunflower cookie", foos[0].Data["story"].(xopbase.ModelArg).Model.(selfRedactor).V, "story")
	assert.Equal(t, "I got the contract with a small consideration, just a sunflower cookie", foos[0].Data["tale"].(xopbase.ModelArg).Model.(selfRedactor).V, "tale")
	assert.Equal(t, "I got the contract with a small bribe, just a sunflower cookie", foos[0].Data["raw"].(xopbase.ModelArg).Model.(selfRedactor).V, "raw")
	assert.Equal(t, "I got the contract with a small bribe, just a daisy cookie", foos[0].Data["success"], "success")
	assert.Equal(t, "outer: inner(as string)", foos[0].Data["oops"], "oops")
}
