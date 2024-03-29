// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xop

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopconst"
	"github.com/xoplog/xop-go/xopnum"
)

// Sub holds an ephemeral state of a log being tranformed to a new log.
type Sub struct {
	detached bool
	settings LogSettings
	logger   *Logger
}

// Detaching is a ephemeral type used in the chain
//
//	child := logger.Sub().Detach().Fork()
//	child := logger.Sub().Detach().Step()
//
// to indicate that the new logger/span has an independent lifetime
// from it's parent so a call to Done() on the parent does not imply
// the child is done.
type Detaching struct {
	sub *Sub
}

// RedactAnyFunc is used to redact models as they're being logged.
// It is RedactAnyFunc's responsibility to call
//
//	baseLine.Any(k, xopbase.ModelArg{
//		Model: v,
//	})
//
// if it wants the value to be logged.  If it does make that call, it
// must pass an immutable value.  Perhaps use "github.com/mohae/deepcopy"
// to make a copy?
//
// The provided xopbase.Line may not be retained beyond the duration of
// the function call.
type RedactAnyFunc func(baseLine xopbase.Line, k xopat.K, v interface{}, alreadyImmutable bool)

// RedactStringFunc is used to redact strings as they're being logged.
// It is RedactStringFunc's responsiblity to call
//
//	baseLine.String(k, v, xopbase.StringDataType)
//
// if it wants the value to be logged.
//
// RedactStringFunc is applied only to String(), and Stringer() attributes.
// It is not applied to Msg(), Msgf(), or Msgs().
//
// The provided xopbase.Line may not be retained beyond the duration of
// the function call.
type RedactStringFunc func(baseLine xopbase.Line, k xopat.K, v string)

// RedactErrorFunc is used to redact or format errors as they're being
// logged.  It is RedactErrorFunc's responsibility to call
//
//	baseLine.String(k, v.Error(), xopbase.ErrorDataType)
//
// if it wants the value to be logged.  Alternatively, it could log the
// error as a model:
//
//	baseLine.Any(k, v)
//
// The provided xopbase.Line may not be retained beyond the duration of
// the function call.
type RedactErrorFunc func(baseLine xopbase.Line, k xopat.K, v error)

type LogSettings struct {
	prefillMsg               string
	prefillData              []func(xopbase.Prefilling)
	minimumLogLevel          xopnum.Level
	stackFramesWanted        [xopnum.AlertLevel + 1]int // indexed
	tagLinesWithSpanSequence bool
	synchronousFlushWhenDone bool
	redactAny                RedactAnyFunc
	redactString             RedactStringFunc
	redactError              RedactErrorFunc
	stackFilenameRewrite     func(string) string
}

// String is for debugging purposes. It is not complete or preformant.
func (settings LogSettings) String() string {
	var str string
	if settings.prefillMsg != "" {
		str += " prefill:" + settings.prefillMsg
	}
	if len(settings.prefillData) > 0 {
		str += " prefillDataCount:" + strconv.Itoa(len(settings.prefillData))
	}
	if settings.minimumLogLevel != 0 {
		str += " minLevel:" + settings.minimumLogLevel.String()
	}
	if settings.synchronousFlushWhenDone {
		str += " flush-when-done"
	}
	return str
}

// DefaultSettings are the settings that are used if no setting changes
// are made. Trace logs are excluded. Alert and Error level log lines
// get stack traces.
var DefaultSettings = func() LogSettings {
	var settings LogSettings
	settings.stackFramesWanted[xopnum.AlertLevel] = 20
	settings.stackFramesWanted[xopnum.ErrorLevel] = 10
	settings.minimumLogLevel = xopnum.DebugLevel
	settings.synchronousFlushWhenDone = true
	settings.stackFilenameRewrite = func(s string) string { return s }
	return settings
}()

func (settings LogSettings) Copy() LogSettings {
	if settings.prefillData != nil {
		n := make([]func(xopbase.Prefilling), len(settings.prefillData))
		copy(n, settings.prefillData)
		settings.prefillData = n
	}
	return settings
}

func (logger *Logger) Settings() LogSettings {
	return logger.settings.Copy()
}

// Sub is a first step in creating a sub-Log from the current logger.
// Sub allows log settings to be modified.  The returned value must
// be used.  It is used by a call to sub.Logger(), sub.Fork(), or
// sub.Step().
//
// Logs created from Sub() are done when their parent is done.
func (logger *Logger) Sub() *Sub {
	return &Sub{
		settings: logger.settings.Copy(),
		logger:   logger,
	}
}

// Detach followed by Fork() or Step() creates a sub-span/logger that is detached from
// it's parent.  A Done() on the parent does not imply Done() on the detached
// logger.
func (sub Sub) Detach() *Detaching {
	sub.detached = true
	return &Detaching{
		sub: &sub,
	}
}

func (d *Detaching) Step(msg string, mods ...SeedModifier) *Logger { return d.sub.Step(msg, mods...) }
func (d *Detaching) Fork(msg string, mods ...SeedModifier) *Logger { return d.sub.Fork(msg, mods...) }

// Fork creates a new logger that does not need to be terminated because
// it is assumed to be done with the current logger is finished.  The new logger
// has its own span.
func (sub *Sub) Fork(msg string, mods ...SeedModifier) *Logger {
	seed := sub.logger.capSpan.SubSeed(mods...)
	counter := int(atomic.AddInt32(&sub.logger.span.forkCounter, 1))
	seed.spanSequenceCode += "." + base26(counter-1)
	seed.settings = sub.settings
	return sub.logger.newChildLog(seed, msg, sub.detached)
}

// Step creates a new logger that does not need to be terminated -- it
// represents the continued execution of the current logger but doing
// something that is different and should be in a fresh span. The expectation
// is that there is a parent logger that is creating various sub-logs using
// Step over and over as it does different things.
func (sub *Sub) Step(msg string, mods ...SeedModifier) *Logger {
	seed := sub.logger.capSpan.SubSeed(mods...)
	counter := int(atomic.AddInt32(&sub.logger.span.stepCounter, 1))
	seed.spanSequenceCode += "." + strconv.Itoa(counter)
	seed.settings = sub.settings
	return sub.logger.newChildLog(seed, msg, sub.detached)
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (sub *Sub) StackFrames(level xopnum.Level, count int) *Sub {
	sub.settings.StackFrames(level, count)
	return sub
}

// StackFrames sets the number of stack frames to include at
// a logging level.  Levels above the given level will be set to
// get least this many.  Levels below the given level will be set
// to receive at most this many.
func (settings *LogSettings) StackFrames(level xopnum.Level, frameCount int) {
	for _, l := range xopnum.LevelValues() {
		current := settings.stackFramesWanted[l]
		if l <= level && current > frameCount {
			settings.stackFramesWanted[l] = frameCount
		}
		if l >= level && current < frameCount {
			settings.stackFramesWanted[l] = frameCount
		}
	}
}

// StackFilenameRewrite is used to rewrite filenames in stack
// traces. Generally this is used to eliminate path prefixes.
// An empty return value indicates that the rest of the stack
// trace should be discarded.
func (sub *Sub) StackFilenameRewrite(f func(string) string) *Sub {
	sub.settings.StackFilenameRewrite(f)
	return sub
}

// StackFilenameRewrite is used to rewrite filenames in stack
// traces. Generally this is used to eliminate path prefixes.
// An empty return value indicates that the rest of the stack
// trace should be discarded.
func (settings *LogSettings) StackFilenameRewrite(f func(string) string) {
	settings.stackFilenameRewrite = f
}

// SynchronousFlush sets the behavior for any Flush()
// triggered by a call to Done().  When true, the
// call to Done() will not return until the Flush() is
// complete.
func (sub *Sub) SynchronousFlush(b bool) *Sub {
	sub.settings.SynchronousFlush(b)
	return sub
}

// SynchronousFlush sets the behavior for any Flush()
// triggered by a call to Done().  When true, the
// call to Done() will not return until the Flush() is
// complete.
func (settings *LogSettings) SynchronousFlush(b bool) {
	settings.synchronousFlushWhenDone = b
}

// MinLevel sets the minimum logging level below which logs will
// be discarded. The default minimum level comes from DefaultSettings.
func (sub *Sub) MinLevel(level xopnum.Level) *Sub {
	sub.settings.MinLevel(level)
	return sub
}

// MinLevel sets the minimum logging level below which logs will
// be discarded. The default minimum level comes from DefaultSettings.
func (settings *LogSettings) MinLevel(level xopnum.Level) {
	settings.minimumLogLevel = level
}

func (settings LogSettings) GetMinLevel() xopnum.Level {
	return settings.minimumLogLevel
}

// TagLinesWithSpanSequence controls if the span sequence
// indicator (see Fork() and Step()) should be included in
// the prefill data on each line.
func (sub *Sub) TagLinesWithSpanSequence(b bool) *Sub {
	sub.settings.TagLinesWithSpanSequence(b)
	return sub
}

// TagLinesWithSpanSequence controls if the span sequence
// indicator (see Fork() and Step()) should be included in
// the prefill data on each line.
func (settings *LogSettings) TagLinesWithSpanSequence(b bool) {
	settings.tagLinesWithSpanSequence = b
}

// PrefillText is prepended to any eventual Line.Msg() or Line.Template().
// PrefillText will be ignored for Line.Model() and Line.Link().
func (sub *Sub) PrefillText(m string) *Sub {
	sub.settings.PrefillText(m)
	return sub
}

// PrefillText is prepended to any eventual Line.Msg() or Line.Template()
// PrefillText will be ignored for Line.Model() and Line.Link().
func (settings *LogSettings) PrefillText(m string) {
	settings.prefillMsg = m
}

func (sub *Sub) NoPrefill() *Sub {
	sub.settings.NoPrefill()
	return sub
}

func (settings *LogSettings) NoPrefill() {
	settings.prefillData = nil
	settings.prefillMsg = ""
}

func (logger *Logger) sendPrefill() {
	if logger.settings.prefillData == nil && logger.settings.prefillMsg == "" && !logger.settings.tagLinesWithSpanSequence {
		logger.prefilled = logger.span.base.NoPrefill()
		return
	}
	prefilling := logger.span.base.StartPrefill()
	for _, f := range logger.settings.prefillData {
		f(prefilling)
	}
	if logger.settings.tagLinesWithSpanSequence {
		prefilling.String(xopconst.SpanSequenceCode.Key(), logger.span.seed.spanSequenceCode, xopbase.StringDataType)
	}
	logger.prefilled = prefilling.PrefillComplete(logger.settings.prefillMsg)
}

// PrefillEmbeddedEnum is used to set a data element that is included on every log
// line.
// PrefillEmbeddedEnum is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillEmbeddedEnum(k xopat.EmbeddedEnum) *Sub {
	sub.settings.PrefillEmbeddedEnum(k)
	return sub
}

func (settings *LogSettings) PrefillEmbeddedEnum(k xopat.EmbeddedEnum) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Enum(k.EnumAttribute(), k)
	})
}

// PrefillEnum is used to set a data element that is included on every log
// line.
// PrefillEnum is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillEnum(k *xopat.EnumAttribute, v xopat.Enum) *Sub {
	sub.settings.PrefillEnum(k, v)
	return sub
}

func (settings *LogSettings) PrefillEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Enum(k, v)
	})
}

// PrefillError is used to set a data element that is included on every log
// line.  Errors will always be formatted with v.Error().  Redaction is
// not supported.
func (sub *Sub) PrefillError(k xopat.K, v error) *Sub {
	sub.settings.PrefillError(k, v)
	return sub
}

// PrefillError is used to set a data element that is included on every log
// line.  Errors will always be formatted with v.Error().  Redaction is
// not supported.
func (settings *LogSettings) PrefillError(k xopat.K, v error) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.String(k, v.Error(), xopbase.ErrorDataType)
	})
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
// Redaction is not supported.
func (sub *Sub) PrefillAny(k xopat.K, v interface{}) *Sub {
	sub.settings.PrefillAny(k, v)
	return sub
}

// PrefillAny is used to set a data element that is included on every log
// line.  Values provided with PrefillAny will be copied
// using https://github.com/mohae/deepcopy 's Copy().
// PrefillAny is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
// Redaction is not supported.
func (settings *LogSettings) PrefillAny(k xopat.K, v interface{}) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Any(k, xopbase.ModelArg{Model: v})
	})
}

// PrefillFloat32 is used to set a data element that is included on every log
// line.
// PrefillFloat32 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillFloat32(k xopat.K, v float32) *Sub {
	sub.settings.PrefillFloat32(k, v)
	return sub
}

func (settings *LogSettings) PrefillFloat32(k xopat.K, v float32) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Float64(k, float64(v), xopbase.Float32DataType)
	})
}

// PrefillBool is used to set a data element that is included on every log
// line.
// PrefillBool is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillBool(k xopat.K, v bool) *Sub {
	sub.settings.PrefillBool(k, v)
	return sub
}

func (settings *LogSettings) PrefillBool(k xopat.K, v bool) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Bool(k, v)
	})
}

// PrefillDuration is used to set a data element that is included on every log
// line.
// PrefillDuration is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillDuration(k xopat.K, v time.Duration) *Sub {
	sub.settings.PrefillDuration(k, v)
	return sub
}

func (settings *LogSettings) PrefillDuration(k xopat.K, v time.Duration) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Duration(k, v)
	})
}

// PrefillTime is used to set a data element that is included on every log
// line.
// PrefillTime is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillTime(k xopat.K, v time.Time) *Sub {
	sub.settings.PrefillTime(k, v)
	return sub
}

func (settings *LogSettings) PrefillTime(k xopat.K, v time.Time) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Time(k, v)
	})
}

// PrefillFloat64 is used to set a data element that is included on every log
// line.
// PrefillFloat64 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillFloat64(k xopat.K, v float64) *Sub {
	sub.settings.PrefillFloat64(k, v)
	return sub
}

func (settings *LogSettings) PrefillFloat64(k xopat.K, v float64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Float64(k, v, xopbase.Float64DataType)
	})
}

// PrefillInt64 is used to set a data element that is included on every log
// line.
// PrefillInt64 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillInt64(k xopat.K, v int64) *Sub {
	sub.settings.PrefillInt64(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt64(k xopat.K, v int64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int64(k, v, xopbase.Int64DataType)
	})
}

// PrefillString is used to set a data element that is included on every log
// line.
// PrefillString is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillString(k xopat.K, v string) *Sub {
	sub.settings.PrefillString(k, v)
	return sub
}

func (settings *LogSettings) PrefillString(k xopat.K, v string) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.String(k, v, xopbase.StringDataType)
	})
}

// PrefillUint64 is used to set a data element that is included on every log
// line.
// PrefillUint64 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUint64(k xopat.K, v uint64) *Sub {
	sub.settings.PrefillUint64(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint64(k xopat.K, v uint64) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, v, xopbase.Uint64DataType)
	})
}

// PrefillInt is used to set a data element that is included on every log
// line.
// PrefillInt is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillInt(k xopat.K, v int) *Sub {
	sub.settings.PrefillInt(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt(k xopat.K, v int) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int64(k, int64(v), xopbase.IntDataType)
	})
}

// PrefillInt16 is used to set a data element that is included on every log
// line.
// PrefillInt16 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillInt16(k xopat.K, v int16) *Sub {
	sub.settings.PrefillInt16(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt16(k xopat.K, v int16) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int64(k, int64(v), xopbase.Int16DataType)
	})
}

// PrefillInt32 is used to set a data element that is included on every log
// line.
// PrefillInt32 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillInt32(k xopat.K, v int32) *Sub {
	sub.settings.PrefillInt32(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt32(k xopat.K, v int32) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int64(k, int64(v), xopbase.Int32DataType)
	})
}

// PrefillInt8 is used to set a data element that is included on every log
// line.
// PrefillInt8 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillInt8(k xopat.K, v int8) *Sub {
	sub.settings.PrefillInt8(k, v)
	return sub
}

func (settings *LogSettings) PrefillInt8(k xopat.K, v int8) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Int64(k, int64(v), xopbase.Int8DataType)
	})
}

// PrefillUint is used to set a data element that is included on every log
// line.
// PrefillUint is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUint(k xopat.K, v uint) *Sub {
	sub.settings.PrefillUint(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint(k xopat.K, v uint) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, uint64(v), xopbase.UintDataType)
	})
}

// PrefillUint16 is used to set a data element that is included on every log
// line.
// PrefillUint16 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUint16(k xopat.K, v uint16) *Sub {
	sub.settings.PrefillUint16(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint16(k xopat.K, v uint16) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, uint64(v), xopbase.Uint16DataType)
	})
}

// PrefillUint32 is used to set a data element that is included on every log
// line.
// PrefillUint32 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUint32(k xopat.K, v uint32) *Sub {
	sub.settings.PrefillUint32(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint32(k xopat.K, v uint32) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, uint64(v), xopbase.Uint32DataType)
	})
}

// PrefillUint8 is used to set a data element that is included on every log
// line.
// PrefillUint8 is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUint8(k xopat.K, v uint8) *Sub {
	sub.settings.PrefillUint8(k, v)
	return sub
}

func (settings *LogSettings) PrefillUint8(k xopat.K, v uint8) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, uint64(v), xopbase.Uint8DataType)
	})
}

// PrefillUintptr is used to set a data element that is included on every log
// line.
// PrefillUintptr is not threadsafe with respect to other calls on the same *Sub.
// Should not be used after Step(), Fork(), or Logger() is called.
func (sub *Sub) PrefillUintptr(k xopat.K, v uintptr) *Sub {
	sub.settings.PrefillUintptr(k, v)
	return sub
}

func (settings *LogSettings) PrefillUintptr(k xopat.K, v uintptr) {
	settings.prefillData = append(settings.prefillData, func(line xopbase.Prefilling) {
		line.Uint64(k, uint64(v), xopbase.UintptrDataType)
	})
}

// SetRedactAnyFunc sets a redaction function to be used
// when Line.Any() is called.
func (sub *Sub) SetRedactAnyFunc(f RedactAnyFunc) *Sub {
	sub.settings.SetRedactAnyFunc(f)
	return sub
}

// SetRedactErrorFunc sets a redaction function to be used
// when Line.Error() is called.
func (sub *Sub) SetRedactErrorFunc(f RedactErrorFunc) *Sub {
	sub.settings.SetRedactErrorFunc(f)
	return sub
}

// SetRedactStringFunc sets a redaction function to be used
// when Line.String() is called.
func (sub *Sub) SetRedactStringFunc(f RedactStringFunc) *Sub {
	sub.settings.SetRedactStringFunc(f)
	return sub
}

// SetRedactAnyFunc sets a redaction function to be used
// when Line.Any() is called.
func (settings *LogSettings) SetRedactAnyFunc(f RedactAnyFunc) {
	settings.redactAny = f
}

// SetRedactErrorFunc sets a redaction function to be used
// when Line.Error() is called.
func (settings *LogSettings) SetRedactErrorFunc(f RedactErrorFunc) {
	settings.redactError = f
}

// SetRedactStringFunc sets a redaction function to be used
// when Line.String() is called.
func (settings *LogSettings) SetRedactStringFunc(f RedactStringFunc) {
	settings.redactString = f
}
