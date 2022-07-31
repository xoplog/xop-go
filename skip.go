// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopconst"
)

var skipBase = skip{}

type skip struct{}

var _ xopbase.Line = skip{}

func (_ skip) AnyImmutable(string, interface{})                {}
func (_ skip) Enum(k *xopconst.EnumAttribute, v xopconst.Enum) {}

func (_ skip) Msg(string)      {}
func (_ skip) Static(string)   {}
func (_ skip) Template(string) {}

func (_ skip) Any(string, interface{})        {}
func (_ skip) Bool(string, bool)              {}
func (_ skip) Duration(string, time.Duration) {}
func (_ skip) Error(string, error)            {}
func (_ skip) Float64(string, float64)        {}
func (_ skip) Int(string, int64)              {}
func (_ skip) Link(string, trace.Trace)       {}
func (_ skip) Str(string, string)             {}
func (_ skip) Time(string, time.Time)         {}
func (_ skip) Uint(string, uint64)            {}
