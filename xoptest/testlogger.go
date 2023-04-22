package xoptest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopcon"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"
)

type testingT interface {
	Log(...interface{})
	Name() string
	Cleanup(func())
}

// Logger is a xopbase.Logger
type Logger struct {
	recorder *xoprecorder.Logger
	console  *xopcon.Logger
	combo    xopbase.Logger
	t        testingT
	id       string
}

var _ xopbase.Logger = &Logger{}

type tPassthrough struct{ t testingT }

func (t tPassthrough) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if b[len(b)-1] == '\n' {
		t.t.Log(string(b[0 : len(b)-1]))
	} else {
		t.t.Log(string(b))
	}
	return len(b), nil
}

func New(t testingT) *Logger {
	requestCounter := xoputil.NewRequestCounter()
	console := xopcon.New(
		xopcon.WithRequestCounter(requestCounter),
		xopcon.WithWriter(tPassthrough{t}),
	)
	recorder := xoprecorder.New(xoprecorder.WithRequestCounter(requestCounter))
	return &Logger{
		t:        t,
		recorder: recorder,
		console:  console,
		combo:    xop.CombineBaseLoggers(recorder, console),
		id:       t.Name() + "-xoptest-" + uuid.New().String(),
	}
}

func (log *Logger) Log() *xop.Log {
	return xop.NewSeed(xop.WithBase(log)).Request(log.t.Name())
}

func (log *Logger) Recorder() *xoprecorder.Logger { return log.recorder }

func (log *Logger) SetPrefix(p string) {
	log.console.SetPrefix(p)
}

// ID is a required method for xopbase.Logger
func (log *Logger) ID() string { return log.id }

// Buffered is a required method for xopbase.Logger
func (log *Logger) Buffered() bool { return false }

// ReferencesKept is a required method for xopbase.Logger
func (log *Logger) ReferencesKept() bool { return true }

// SetErrorReporter is a required method for xopbase.Logger
func (log *Logger) SetErrorReporter(func(error)) {}

// Request is a required method for xopbase.Logger
func (log *Logger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, name string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	log.t.Log(fmt.Sprintf("XXX TEST START REQUEST %s: %s%s", name, xop.LogLinkPrefix, bundle.Trace.String()))
	log.t.Cleanup(func() {
		log.t.Log(fmt.Sprintf("%s: %s%s", name, xop.LogLinkPrefix, bundle.Trace.String()))
	})
	return log.combo.Request(ctx, ts, bundle, name, sourceInfo)
}

func (log *Logger) CustomEvent(msg string, args ...interface{}) {
	log.t.Log(fmt.Sprintf(msg, args...))
	log.recorder.CustomEvent(msg, args...)
}
