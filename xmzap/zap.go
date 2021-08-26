package xmzap

import (
	"github.com/muir/xm"
	"go.uber.org/zap"
)

type xmzap struct {
	zapLogger *zap.Logger
	level     xm.Level
}

// ZapBase wraps a zap logger so that it can function as
// a BaseLogger.  Note that the levels don't match exactly
// and Trace becomes Debug and Alert becomes Error.
func ZapBase(zapLogger *zap.Logger) xm.BaseLogger {
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
	trace xm.Trace,
	parent xm.Trace,
	searchTerms []xm.Field,
	data []xm.Field) {
	z.zapLogger.Info(
	// XXX
	)
}

func (z *xmzap) Prefill(trace xm.Trace, f []xm.Field) xm.Prefilled {
	return &xmzap{
		zapLogger: z.zapLogger.With(
			append([]xm.Field{
				xm.String("xm.trace", trace.GetTraceId().String()),
				xm.String("xm.span", trace.GetSpanId().String()),
			}, f...)...),
		level: z.level,
	}
}

func (z *xmzap) Log(level xm.Level, msg string, values []xm.Field) {
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
