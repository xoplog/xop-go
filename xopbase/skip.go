// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopbase

import (
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xoptrace"
)

// SkipLine is a utility type that implements Line.  It discards
// all data.
var SkipLine = skipLine{}

type skipLine struct{}

var _ Line = skipLine{}

func (_ skipLine) AnyImmutable(xopat.K, interface{})         {}
func (_ skipLine) Enum(k *xopat.EnumAttribute, v xopat.Enum) {}

func (_ skipLine) Msg(string)                  {}
func (_ skipLine) Link(string, xoptrace.Trace) {}
func (_ skipLine) Model(string, ModelArg)      {}
func (_ skipLine) Template(string)             {}

func (_ skipLine) Any(xopat.K, ModelArg)           {}
func (_ skipLine) Bool(xopat.K, bool)              {}
func (_ skipLine) Duration(xopat.K, time.Duration) {}
func (_ skipLine) Time(xopat.K, time.Time)         {}

func (_ skipLine) Float64(xopat.K, float64, DataType) {}
func (_ skipLine) Int64(xopat.K, int64, DataType)     {}
func (_ skipLine) String(xopat.K, string, DataType)   {}
func (_ skipLine) Uint64(xopat.K, uint64, DataType)   {}
