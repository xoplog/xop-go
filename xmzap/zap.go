package xmzap

import (
	"github.com/muir/xm"
	"github.com/muir/xm/trace"
	"github.com/muir/xm/zap"

	uberzap "go.uber.org/zap"
)

type xmzap struct {
	zapLogger *uberzap.Logger
	level     xm.Level
}

// ZapBase wraps a zap logger so that it can function as
// a BaseLogger.  Note that the levels don't match exactly
// and Trace becomes Debug and Alert becomes Error.
func ZapBase(zapLogger *uberzap.Logger) xm.BaseLogger {
	return &xmzap{
		zapLogger: zapLogger,
		level:     xm.DebugLevel,
	}
}

func (z *xmzap) WantDurable() bool { return false }

func (z *xmzap) StartBuffer() xm.BufferedBase {
	return &xmzap{
		zapLogger: z.zapLogger.WithOptions(),
		level:     z.level.AtomicLoad(),
	}
}

func (z *xmzap) SetLevel(level xm.Level) {
	z.level.AtomicStore(level)
}

func (z *xmzap) Flush() {
	_ = z.zapLogger.Sync() // what are you supposed to do with an error anyway?
}

func (z *xmzap) Span(
	description string,
	trace trace.Trace,
	parent trace.Trace,
	searchTerms map[string][]string,
	data map[string]interface{}) {
	fields := make([]zap.Field, 0, 20)
	fields = append(fields,
		zap.String("xm.type", "span"),
		zap.String("xm.trace", trace.GetTraceId().String()),
		zap.String("xm.span", trace.GetSpanId().String()),
	)
	if !parent.IsZero() {
		fields = append(fields,
			zap.String("xm.parent_span", parent.GetSpanId().String()),
			zap.String("xm.parent_trace", parent.GetSpanId().String()),
		)
	}
	if len(searchTerms) != 0 {
		fields = append(fields, zap.Any("xm.index", searchTerms))
	}
	if len(data) != 0 {
		fields = append(fields, zap.Any("xm.data", data))
	}
	z.zapLogger.Warn(description, fields...)
}

func (z *xmzap) Prefill(trace trace.Trace, f []zap.Field) xm.Prefilled {
	return &xmzap{
		zapLogger: z.zapLogger.With(
			append([]zap.Field{
				zap.String("xm.trace", trace.GetTraceId().String()),
				zap.String("xm.span", trace.GetSpanId().String()),
			}, f...)...),
		level: z.level,
	}
}

func (z *xmzap) Log(level xm.Level, msg string, values []zap.Field) {
	if level < z.level {
		return
	}
	switch level {
	case xm.DebugLevel, xm.TraceLevel:
		z.zapLogger.Debug(msg, values...)
	case xm.InfoLevel:
		z.zapLogger.Info(msg, values...)
	case xm.WarnLevel:
		z.zapLogger.Warn(msg, values...)
	case xm.ErrorLevel:
		z.zapLogger.Error(msg, values...)
	case xm.AlertLevel:
		z.zapLogger.Error("Alert: "+msg, values...)
	}
}
