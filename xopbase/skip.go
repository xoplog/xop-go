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

func (_ skipLine) AnyImmutable(string, interface{})          {}
func (_ skipLine) Enum(k *xopat.EnumAttribute, v xopat.Enum) {}

func (_ skipLine) Msg(string)                  {}
func (_ skipLine) Link(string, xoptrace.Trace) {}
func (_ skipLine) Model(string, ModelArg)      {}
func (_ skipLine) Template(string)             {}

func (_ skipLine) Any(string, ModelArg)           {}
func (_ skipLine) Bool(string, bool)              {}
func (_ skipLine) Duration(string, time.Duration) {}
func (_ skipLine) Time(string, time.Time)         {}

func (_ skipLine) Float64(string, float64, DataType) {}
func (_ skipLine) Int64(string, int64, DataType)     {}
func (_ skipLine) String(string, string, DataType)   {}
func (_ skipLine) Uint64(string, uint64, DataType)   {}
