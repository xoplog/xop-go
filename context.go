package xop

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

// Default serves as a fallback logger if FromContextOrDefault
// does not find a logger.  Unless modified, it discards all logs.
var Default = NewSeed().Request("discard")

func (log *Log) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, log)
}

func FromContext(ctx context.Context) (*Log, bool) {
	v := ctx.Value(contextKey)
	if v == nil {
		return nil, false
	}
	return v.(*Log), true
}

func FromContextOrDefault(ctx context.Context) *Log {
	log, ok := FromContext(ctx)
	if ok {
		return log
	}
	return Default
}

func FromContextOrPanic(ctx context.Context) *Log {
	log, ok := FromContext(ctx)
	if !ok {
		panic("Could not find logger in context")
	}
	return log
}

// CustomFromContext returns a convenience function: it calls either
// FromContextOrPanic() or FromContextOrDefault() and then calls a
// function to adjust setting.
//
// Pass FromContextOrPanic or FromContextOrDefault as the first argument
// and a function to adjust settings as the second argument.
func CustomFromContext(getLogFromContext func(context.Context) *Log, adjustSettings func(*Sub) *Sub) func(context.Context) *Log {
	return func(ctx context.Context) *Log {
		log := getLogFromContext(ctx)
		return adjustSettings(log.Sub()).Log()
	}
}
