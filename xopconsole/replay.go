// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopconsole

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/pkg/errors"
)

type replayData struct {
	lineCount   int
	currentLine string
	errors      []error
}

type replayRequest struct {
	replayData
	ts                  time.Time
	trace               xoptrace.Trace
	version             int64
	name                string
	sourceAndVersion    string
	namespaceAndVersion string
}

func (x replayData) replayLine1(ctx context.Context, level xopnum.Level, t string) error { return nil }
func (x replayData) replaySpan1(ctx context.Context, t string) error                     { return nil }
func (x replayData) replayDef(ctx context.Context, t string) error                       { return nil }

// so far: xop Request
// this func: timestamp "Start1" or "vNNN"
func (x replayData) replayRequest1(ctx context.Context, t string) error {
	ts, t, err := oneTime(t)
	if err != nil {
		return err
	}
	n, _, t := oneWord(t, " ")
	switch n {
	case "":
		return errors.Errorf("invalid request")
	case "Start1":
		return replayRequest{
			replayData: x,
			ts:         ts,
		}.replayRequestStart(ctx, t)
	default:
		if !strings.HasPrefix(t, "v") {
			return errors.Errorf("invalid request")
		}
		v, err := strconv.ParseInt(t[1:], 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid request, invalid version number")
		}
		return replayRequest{
			replayData: x,
			ts:         ts,
			version:    v,
		}.replayRequestUpdate(ctx, t)
	}
}

func (x replayRequest) replayRequestUpdate(ctx context.Context, t string) error { return nil } // XXX

// so far: xop Request timestamp Start1
// this func: trace-headder request-name source+version namespace+version
func (x replayRequest) replayRequestStart(ctx context.Context, t string) error {
	th, _, t := oneWord(t, " ")
	if th == "" {
		return errors.Errorf("missing trace header")
	}
	var ok bool
	x.trace, ok = xoptrace.TraceFromString(th)
	if !ok {
		return errors.Errorf("invalid trace header")
	}
	x.name, t = oneString(t)
	if x.name == "" {
		return errors.Errorf("missing request name")
	}
	x.sourceAndVersion, t = oneString(t)
	if x.sourceAndVersion == "" {
		return errors.Errorf("missing source+version")
	}
	x.namespaceAndVersion, t = oneString(t)
	if x.namespaceAndVersion == "" {
		return errors.Errorf("missing namespace+version")
	}
	// XXX
	return nil
}

// oneString reads a possibly-quoted string
func oneString(t string) (string, string) {
	if len(t) == 0 {
		return "", ""
	}
	if t[0] == '"' {
		for i := 1; i < len(t); i++ {
			switch t[i] {
			case '\\':
				if i < len(t) {
					i++
				}
			case '"':
				one, err := strconv.Unquote(t[0 : i+1])
				if err != nil {
					return "", t
				}
				return one, t[i+1:]
			}
		}
	}
	one := xoputil.UnquotedConsoleStringRE.FindString(t)
	if one != "" {
		return one, t[len(one):]
	}
	return "", t
}

func oneTime(t string) (time.Time, string, error) {
	w, _, t := oneWord(t, " ")
	ts, err := time.Parse(time.RFC3339, w)
	return ts, t, err
}

func oneWord(t string, boundary string) (string, byte, string) {
	i := strings.IndexAny(t, boundary)
	switch i {
	case -1:
		return "", '\000', t
	case 0:
		return "", t[0], t[1:]
	}
	return t[:i], t[i], t[i+1:]
}

func Replay(ctx context.Context, inputStream io.Reader, dest xopbase.Logger) error {
	scanner := bufio.NewScanner(inputStream)
	var x replayData
	for scanner.Scan() {
		x.lineCount++
		t := scanner.Text()
		if !strings.HasPrefix(t, "xop ") {
			continue
		}
		x.currentLine = t
		t = t[len("xop "):]
		kind, _, t := oneWord(t, " ")
		var err error
		switch kind {
		case "Request":
			err = x.replayRequest1(ctx, t)
		case "Span":
			err = x.replaySpan1(ctx, t)
		case "Def":
			err = x.replayDef(ctx, t)
		case "Alert":
			err = x.replayLine1(ctx, xopnum.AlertLevel, t)
		case "Debug":
			err = x.replayLine1(ctx, xopnum.DebugLevel, t)
		case "Error":
			err = x.replayLine1(ctx, xopnum.ErrorLevel, t)
		case "Info":
			err = x.replayLine1(ctx, xopnum.InfoLevel, t)
		case "Trace":
			err = x.replayLine1(ctx, xopnum.TraceLevel, t)
		case "Warn":
			err = x.replayLine1(ctx, xopnum.WarnLevel, t)

			// prior line must be blank
		default:
			err = fmt.Errorf("invalid kind designator '%s'", kind)
		}
		if err != nil {
			x.errors = append(x.errors, errors.Wrapf(err, "line %d: %s", x.lineCount, x.currentLine))
		}
	}
	if len(x.errors) != 0 {
		// TODO: use a multi-error
		return x.errors[0]
	}
	return nil
}
