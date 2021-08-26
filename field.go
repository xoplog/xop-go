package xm

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

func copyFields(from []Field) []Field {
	n := make([]Field, len(from))
	copy(n, from)
	return n
}

// TODO: add the rest of the fields supported by zap
// TODO: how are we going to handle redacting Any?

func Any(key string, val interface{}) Field        { return zap.Any(key, val) }
func Binary(key string, val []byte) Field          { return zap.Binary(key, val) }
func Duration(key string, val time.Duration) Field { return zap.Duration(key, val) }
func Int(key string, val int) Field                { return zap.Int(key, val) }
func NamedError(key string, val error) Field       { return zap.NamedError(key, val) }
func String(key string, val string) Field          { return zap.String(key, val) }
