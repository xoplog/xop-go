package xopotel

import (
	"context"

	"github.com/xoplog/xop-go/xopbase"
)

func (logger *logger) Replay(ctx context.Context, input any, output xopbase.Logger) error {
	return logger.LosslessReplay(ctx, input, output)
}

func (_ *logger) LosslessReplay(ctx context.Context, input any, output xopbase.Logger) error {
	// XXX
	return nil
}
