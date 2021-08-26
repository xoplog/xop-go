package xm

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func (log *Logger) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, log)
}

func FromContext(ctx context.Context) (*Logger, bool) {
	v := ctx.Value(contextKey)
	return v.(*Logger), v != nil
}

func MustFromContext(ctx context.Context) *Logger {
	log, ok := FromContext(ctx)
	if !ok {
		panic("Could not find logger in context")
	}
	return log
}
