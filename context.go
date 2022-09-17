package xop

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/muir/xop-go/xopnum"
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

// AdjustedLevelLogger returns a function that gets a logger from
// context.  The logger it returns will have a minimum logging level
// that is specific for the calling package.
//
// The level used will be the level set in this order:
// (1) From the environment variable "XOPLEVEL_foo" (for package foo).  The level
// can either a string or the numeric level.  See xopnum/level.go.
// (2) From the level provided in the call to AdjustedLevelLoger assuming that
// the passed level is not zero.
// (3) The level that the logger already has.
//
// 	package foo
//	var getLogger = xop.AdjustedLevelLoger(xop.FromContextOrPanic, xopnum.Info)
//
// 	package foo
//	var getLogger = xop.AdjustedLevelLoger(xop.FromContextOrDefault, 0)
//
func AdjustedLevelLoger(getLogFromContext func(context.Context) *Log, level xopnum.Level) func(context.Context) *Log {
	pc := make([]uintptr, 1)
	if runtime.Callers(1, pc) > 0 {
		frames := runtime.CallersFrames(pc)
		frame, _ := frames.Next()
		var pkg string
		if frame.Function != "" {
			p := strings.Split(frame.Function, ".")
			if p[0] != "" {
				pkg = p[0]
			}
		}
		if pkg == "" {
			pkg = filepath.Base(filepath.Dir(frame.File))
			if pkg == "." || pkg == "/" {
				pkg = ""
			}
		}
		if pkg != "" {
			if ls, ok := os.LookupEnv("XOPLEVEL_" + pkg); ok {
				lvl, err := xopnum.LevelString(ls)
				if err == nil {
					level = lvl
				} else if i, err := strconv.ParseInt(ls, 10, 64); err == nil {
					level = xopnum.Level(i)
				}
			}
		}
	}
	if level == 0 {
		return getLogFromContext
	}
	return CustomFromContext(getLogFromContext, func(sub *Sub) *Sub {
		return sub.MinLevel(level)
	})
}
