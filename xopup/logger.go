package xopup

import (
	"context"

	"github.com/xoplog/xop-go/xopjson"
)

type UploadLogger struct {
	*xopjson.Logger
	Uploader *Uploader
}

func New(ctx context.Context, config Config) UploadLogger {
	uploader := newUploader(ctx, config)
	jsonLogger := xopjson.New(uploader,
		xopjson.WithAttributesObject(true),
		xopjson.WithSpanStarts(false),
		xopjson.WithDuration("dur", xopjson.AsMicros),
	)
	return UploadLogger{
		Uploader: uploader,
		Logger:   jsonLogger,
	}
}
