// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoputil

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopat"
	"github.com/muir/xop-go/xopbase"
)

var SkipLine = skipLine{}

type skipLine struct{}

var _ xopbase.Line = skipLine{}

func (_ skipLine) AnyImmutable(string, interface{})          {}
func (_ skipLine) Enum(k *xopat.EnumAttribute, v xopat.Enum) {}

func (_ skipLine) Msg(string)      {}
func (_ skipLine) Static(string)   {}
func (_ skipLine) Template(string) {}

func (_ skipLine) Any(string, interface{})        {}
func (_ skipLine) Bool(string, bool)              {}
func (_ skipLine) Duration(string, time.Duration) {}
func (_ skipLine) Error(string, error)            {}
func (_ skipLine) Float64(string, float64)        {}
func (_ skipLine) Int(string, int64)              {}
func (_ skipLine) Link(string, trace.Trace)       {}
func (_ skipLine) String(string, string)          {}
func (_ skipLine) Time(string, time.Time)         {}
func (_ skipLine) Uint(string, uint64)            {}
