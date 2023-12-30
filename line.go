// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"fmt"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"

	"github.com/mohae/deepcopy"
)

type Key = xopat.K

// AnyWithoutRedaction
// The return value must be consumed for the line to be logged.
func (line *Line) AnyWithoutRedaction(k xopat.K, v interface{}) *Line {
	line.line.Any(k, xopbase.ModelArg{Model: v})
	return line
}

// Any can be used to log something that might be modified after this call.  If any base
// logger does not immediately serialize, then the object will be copied using
// https://github.com/mohae/deepcopy 's Copy().
// The return value must be consumed for the line to be logged.
func (line *Line) Any(k xopat.K, v interface{}) *Line {
	if line.skip {
		return line
	}
	if line.logger.settings.redactAny != nil {
		line.logger.settings.redactAny(line.line, k, v, !line.logger.span.referencesKept)
		return line
	}
	if line.logger.span.referencesKept {
		// TODO: make copy function configurable
		v = deepcopy.Copy(v)
	}
	return line.AnyWithoutRedaction(k, v)
}

// Float32 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Float32(k xopat.K, v float32) *Line {
	line.line.Float64(k, float64(v), xopbase.Float32DataType)
	return line
}

// EmbeddedEnum adds a key/value pair to the current log line.
// The type of the value implies the key.
// The return value must be consumed for the line to be logged.
func (line *Line) EmbeddedEnum(k xopat.EmbeddedEnum) *Line {
	return line.Enum(k.EnumAttribute(), k)
}

// Enum adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Enum(k *xopat.EnumAttribute, v xopat.Enum) *Line {
	line.line.Enum(k, v)
	return line
}

// Error adds a key/value pair to the current log line.
// The default logging of an error is simply err.Error() to change
// that, set a redaction function.
// The return value must be consumed for the line to be logged.
func (line *Line) Error(k xopat.K, v error) *Line {
	if line.skip {
		return line
	}
	if line.logger.settings.redactError != nil {
		line.logger.settings.redactError(line.line, k, v)
		return line
	}
	line.line.String(k, v.Error(), xopbase.ErrorDataType)
	return line
}

// Stringer adds a key/value pair to the current log line.
// The string can be redacted if a redaction function has been set.
// The return value must be consumed for the line to be logged.
func (line *Line) Stringer(k xopat.K, v fmt.Stringer) *Line {
	if line.skip {
		return line
	}
	if line.logger.settings.redactString != nil {
		line.logger.settings.redactString(line.line, k, v.String())
		return line
	}
	line.line.String(k, v.String(), xopbase.StringerDataType)
	return line
}

// String adds a key/value pair to the current log line.
// The string can be redacted if a redaction function has been set.
// The return value must be consumed for the line to be logged.
func (line *Line) String(k xopat.K, v string) *Line {
	if line.skip {
		return line
	}
	if line.logger.settings.redactString != nil {
		line.logger.settings.redactString(line.line, k, v)
		return line
	}
	line.line.String(k, v, xopbase.StringDataType)
	return line
}

// Bool adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Bool(k xopat.K, v bool) *Line { line.line.Bool(k, v); return line }

// Duration adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Duration(k xopat.K, v time.Duration) *Line { line.line.Duration(k, v); return line }

// Time adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Time(k xopat.K, v time.Time) *Line { line.line.Time(k, v); return line }

// Float64 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Float64(k xopat.K, v float64) *Line {
	line.line.Float64(k, v, xopbase.Float64DataType)
	return line
}

// Int64 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Int64(k xopat.K, v int64) *Line {
	line.line.Int64(k, v, xopbase.Int64DataType)
	return line
}

// Uint64 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uint64(k xopat.K, v uint64) *Line {
	line.line.Uint64(k, v, xopbase.Uint64DataType)
	return line
}

// Int adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Int(k xopat.K, v int) *Line {
	line.line.Int64(k, int64(v), xopbase.IntDataType)
	return line
}

// Int16 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Int16(k xopat.K, v int16) *Line {
	line.line.Int64(k, int64(v), xopbase.Int16DataType)
	return line
}

// Int32 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Int32(k xopat.K, v int32) *Line {
	line.line.Int64(k, int64(v), xopbase.Int32DataType)
	return line
}

// Int8 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Int8(k xopat.K, v int8) *Line {
	line.line.Int64(k, int64(v), xopbase.Int8DataType)
	return line
}

// Uint adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uint(k xopat.K, v uint) *Line {
	line.line.Uint64(k, uint64(v), xopbase.UintDataType)
	return line
}

// Uint16 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uint16(k xopat.K, v uint16) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint16DataType)
	return line
}

// Uint32 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uint32(k xopat.K, v uint32) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint32DataType)
	return line
}

// Uint8 adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uint8(k xopat.K, v uint8) *Line {
	line.line.Uint64(k, uint64(v), xopbase.Uint8DataType)
	return line
}

// Uintptr adds a key/value pair to the current log line.
// The return value must be consumed for the line to be logged.
func (line *Line) Uintptr(k xopat.K, v uintptr) *Line {
	line.line.Uint64(k, uint64(v), xopbase.UintptrDataType)
	return line
}
