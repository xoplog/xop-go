package xoplog

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

// TODO: have a default log that prints
var Default = NewSeed().Request("discard")

func (l *Log) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, l)
}

func FromContextOrDefault(ctx context.Context) *Log {
	log, ok := FromContext(ctx)
	if ok {
		return log
	}
	return Default
}

func FromContext(ctx context.Context) (*Log, bool) {
	v := ctx.Value(contextKey)
	return v.(*Log), v != nil
}

func FromContextOrPanic(ctx context.Context) *Log {
	log, ok := FromContext(ctx)
	if !ok {
		panic("Could not find logger in context")
	}
	return log
}
