package xopbase

import (
	"fmt"
	"sync"
	"time"

	"github.com/muir/xoplog/xop"
	"github.com/muir/xoplog/xopconst"
)

type Requests []Request
type Spans []Span
type Lines []Line

func (r Requests) Flush() {
	if len(r) == 1 {
		r[0].Flush()
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(r))
	for _, request := range r {
		go func() {
			defer wg.Done()
			request.Flush()
		}()
	}
	wg.Wait()
}

func (r Requests) Spans() Spans {
	spans := make(Spans, len(r))
	for i, request := range r {
		spans[i] = request.(Span)
	}
	return spans
}

func (s Spans) Line(level xopconst.Level, t time.Time) Lines {
	lines := make(Lines, len(s))
	for i, span := range s {
		lines[i] = span.Line(level, t)
	}
	return lines
}

func (l Lines) Int(k string, v int64) {
	for _, line := range l {
		line.Int(k, v)
	}
}
func (l Lines) Str(k string, v string) {
	for _, line := range l {
		line.Str(k, v)
	}
}
func (l Lines) Bool(k string, v bool) {
	for _, line := range l {
		line.Bool(k, v)
	}
}
func (l Lines) Uint(k string, v uint64) {
	for _, line := range l {
		line.Uint(k, v)
	}
}
func (l Lines) Time(k string, v time.Time) {
	for _, line := range l {
		line.Time(k, v)
	}
}
func (l Lines) Any(k string, v interface{}) {
	for _, line := range l {
		line.Any(k, v)
	}
}
func (l Lines) Error(k string, v error) {
	for _, line := range l {
		line.Error(k, v)
	}
}
func (l Lines) Msg(m string) {
	for _, line := range l {
		line.Msg(m)
	}
}

func (l Lines) Things(things []xop.Thing) {
	for _, line := range l {
		LineThings(line, things)
	}
}

func LineThings(line Line, things []xop.Thing) {
	for _, thing := range things {
		switch thing.Type {
		case xop.IntType:
			line.Int(thing.Key, thing.Int)
		case xop.UintType:
			line.Uint(thing.Key, thing.Any.(uint64))
		case xop.BoolType:
			line.Bool(thing.Key, thing.Any.(bool))
		case xop.StringType:
			line.Str(thing.Key, thing.String)
		case xop.TimeType:
			line.Time(thing.Key, thing.Any.(time.Time))
		case xop.AnyType:
			line.Any(thing.Key, thing.Any)
		case xop.ErrorType:
			line.Error(thing.Key, thing.Any.(error))
		case xop.UnsetType:
			fallthrough
		default:
			panic(fmt.Sprintf("malformed xop.Thing, type is %d", thing.Type))
		}
	}
}
