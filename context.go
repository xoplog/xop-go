package xop

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/xoplog/xop-go/xopnum"
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

// LevelAdjuster returns a function that adjusts the level of
// a logger for local package-scoped defaults.  It is a sibling
// to AdjustedLevelLogger.
//
// The default behavior is to determine the name of the local
// package and then using that look at the environment variable
// "XOPLEVEL_<package_name>" (where <package_name> is the name of
// the current package) and set the minimum log level according
// to the value of that environment variable.
//
// 	package foo
//	var adjustLogger = xop.LevelAdjuster()
//
func LevelAdjuster(opts ...AdjusterOption) func(*Log) *Log {
	level := adjustConfig(opts)
	if level == 0 {
		return func(log *Log) *Log { return log }
	}
	return func(log *Log) *Log {
		return log.Sub().MinLevel(level).Log()
	}
}

// ContextLevelAdjuster returns a function that gets a logger from
// context.  The logger it returns will have a minimum logging level
// that is specific for the calling package. If starting from a log
// instead of a context, use LevelAdjuster
//
// The level used will be the level set in this order:
// (1) From the environment variable "XOPLEVEL_foo" (for package foo).  The level
// can either a string or the numeric level.  See xopnum/level.go.
// (2) From the level provided in the call to AdjustedLevelLoger assuming that
// the passed level is not zero.
// (3) The level that the logger already has.
//
// 	package foo
//	var getLogger = xop.AdjustedLevelLoger(xop.FromContextOrPanic)
//
func ContextLevelAdjuster(getLogFromContext func(context.Context) *Log, opts ...AdjusterOption) func(context.Context) *Log {
	level := adjustConfig(opts)
	if level == 0 {
		return getLogFromContext
	}
	return CustomFromContext(getLogFromContext, func(sub *Sub) *Sub {
		return sub.MinLevel(level)
	})
}

// WithPackage overrides how the package name is found.  The
// package name is combined with "XOPLEVEL_" to create the
// name of the environment variable to check for adjusting
// log levels with LevelAdjuster and ContextLevelAdjuster
func WithPackage(pkg string) AdjusterOption {
	return func(o *adjustOptions) {
		o.pkg = pkg
	}
}

// WithEnvironment overrides the name of the environment variable
// used to override log levels in LevelAdjuster and ContextLevelAdjuster.
// WithEnvironment makes the package name irrelevant and thus should
// not bue used in combination with WithPackage.
func WithEnvironment(environmentVariableName string) AdjusterOption {
	return func(o *adjustOptions) {
		o.env = environmentVariableName
	}
}

// WithDefault sets a default logger level for LevelAdjuster and
// ContextLevelAdjuster.  If the environment variable is found then
// that will override this default.  This default will override the
// existing minimum logging level.
func WithDefault(level xopnum.Level) AdjusterOption {
	return func(o *adjustOptions) {
		o.level = level
	}
}

// WithSkippedFrames is needed only if LevelAdjuster or ContextLevelAdjuster
// are called from withing another function that should not be used to
// derrive the package name.
func WithSkippedFrames(additionalFramesToSkip int) AdjusterOption {
	return func(o *adjustOptions) {
		o.skip = additionalFramesToSkip
	}
}

type AdjusterOption func(*adjustOptions)

type adjustOptions struct {
	pkg   string
	env   string
	level xopnum.Level
	skip  int
}

func adjustConfig(opts []AdjusterOption) xopnum.Level {
	var options adjustOptions
	for _, f := range opts {
		f(&options)
	}

	pc := make([]uintptr, 1)
	if options.pkg == "" && options.env == "" && runtime.Callers(3+options.skip, pc) > 0 {
		frames := runtime.CallersFrames(pc)
		frame, _ := frames.Next()
		if frame.Function != "" {
			// Example: github.com/xoplog/xop-go/xoptest/xoptestutil.foo.Context
			// Example: github.com/xoplog/xop-go/xoptest/xoptestutil.TestAdjusterContext
			// Example: github.com/xoplog/xop-go/xoptest/xoptestutil.init
			base := filepath.Base(frame.Function)
			parts := strings.SplitN(base, ".", 2)
			if len(parts) == 2 {
				options.pkg = parts[0]
			}
		}
		if options.pkg == "" {
			options.pkg = filepath.Base(filepath.Dir(frame.File))
			if options.pkg == "." || options.pkg == "/" {
				options.pkg = ""
			}
		}
	}

	if options.env == "" && options.pkg != "" {
		options.env = "XOPLEVEL_" + options.pkg
	}

	if options.env != "" {
		if ls, ok := os.LookupEnv(options.env); ok {
			lvl, err := xopnum.LevelString(ls)
			if err == nil {
				options.level = lvl
			} else if i, err := strconv.ParseInt(ls, 10, 64); err == nil {
				options.level = xopnum.Level(i)
			}
		}
	}

	return options.level
}
