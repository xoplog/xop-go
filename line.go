// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"fmt"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopat"
	"github.com/muir/xop-go/xopbase"

	"github.com/mohae/deepcopy"
)

// AnyImmutable can be used to log something that is not going to be further modified
// after this call.
func (line *Line) AnyImmutable(k string, v interface{}) *Line {
	if line.skip {
		return line
	}
	if line.log.settings.redactAny != nil {
		line.log.settings.redactAny(line.line, k, v, true)
		return line
	}
	line.line.Any(k, v)
	return line
}

// AnyWithoutRedaction
func (line *Line) AnyWithoutRedaction(k string, v interface{}) *Line {
	line.line.Any(k, v)
	return line
}

// Any can be used to log something that might be modified after this call.  If any base
// logger does not immediately serialize, then the object will be copied using
// https://github.com/mohae/deepcopy 's Copy().
func (line *Line) Any(k string, v interface{}) *Line {
	if line.skip {
		return line
	}
	if line.log.settings.redactAny != nil {
		line.log.settings.redactAny(line.line, k, v, !line.log.span.referencesKept)
		return line
	}
	if line.log.span.referencesKept {
		// TODO: make copy function configurable
		v = deepcopy.Copy(v)
	}
	line.line.Any(k, v)
	return line
}

// TODO: func (l *Log) Guage(name string, value float64, )
// TODO: func (l *Log) AdjustCounter(name string, value float64, )
// TODO: func (l *Log) Event

func (line *Line) Float32(k string, v float32) *Line {
	line.line.Float64(k, float64(v), xopbase.Float32DataType)
	return line
}

func (line *Line) EmbeddedEnum(k xopat.EmbeddedEnum) *Line {
	return line.Enum(k.EnumAttribute(), k)
}

func (line *Line) Enum(k *xopat.EnumAttribute, v xopat.Enum) *Line {
	line.line.Enum(k, v)
	return line
}

func (line *Line) Stringer(k string, v fmt.Stringer) *Line {
	if line.skip {
		return line
	}
	if line.log.settings.redactString != nil {
		line.log.settings.redactString(line.line, k, v.String())
		return line
	}
	line.line.String(k, v.String())
	return line
}

func (line *Line) String(k string, v string) *Line {
	if line.skip {
		return line
	}
	if line.log.settings.redactString != nil {
		line.log.settings.redactString(line.line, k, v)
		return line
	}
	line.line.String(k, v)
	return line
}

func (line *Line) Bool(k string, v bool) *Line              { line.line.Bool(k, v); return line }
func (line *Line) Duration(k string, v time.Duration) *Line { line.line.Duration(k, v); return line }
func (line *Line) Error(k string, v error) *Line            { line.line.Error(k, v); return line }
func (line *Line) Link(k string, v trace.Trace) *Line       { line.line.Link(k, v); return line }
func (line *Line) Time(k string, v time.Time) *Line         { line.line.Time(k, v); return line }

func (line *Line) Float64(k string, v float64) *Line {
	line.line.Float64(k, v, xopbase.Float64DataType)
	return line
}

func (line *Line) Int64(k string, v int64) *Line {
	line.line.Int64(k, v, xopbase.Int64DataType)
	return line
}

func (line *Line) Uint64(k string, v uint64) *Line {
	line.line.Uint64(k, v, xopbase.Uint64DataType)
	return line
}

func (line *Line) Int(k string, v int) *Line {
	line.line.Int64(k, int64(v), xopbase.IntDataType)
	return line
}

func (line *Line) Int16(k string, v int16) *Line {
	line.line.Int64(k, int64(v), xopbase.Int16DataType)
	return line
}

func (line *Line) Int32(k string, v int32) *Line {
	line.line.Int64(k, int64(v), xopbase.Int32DataType)
	return line
}

func (line *Line) Int8(k string, v int8) *Line {
	line.line.Int64(k, int64(v), xopbase.Int8DataType)
	return line
}

func (line *Line) Uint(k string, v uint) *Line {
	line.line.Uint64(k, uint64(v), xopbase.UintDataType)
	return line
}

func (line *Line) Uint16(k string, v uint16) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint16DataType)
	return line
}

func (line *Line) Uint32(k string, v uint32) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint32DataType)
	return line
}

func (line *Line) Uint8(k string, v uint8) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint8DataType)
	return line
}
