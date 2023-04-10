package xoprecorder

import (
	"strings"
)

type LinePredicate struct {
	f    func(*Line) bool
	desc string
}

type Predicates []LinePredicate

func (p LinePredicate) String() string { return p.desc }

func (p SpanPredicate) LinePredicate() LinePredicate {
	return LinePredicate{
		f: func(line *Line) bool {
			return p.f(line.Span)
		},
		desc: "span " + p.String(),
	}
}

func MessageEquals(msg string) LinePredicate {
	return LinePredicate{
		f: func(line *Line) bool {
			return line.Message == msg
		},
		desc: "message equals " + msg,
	}
}

func TextContains(msg string) LinePredicate {
	return LinePredicate{
		f: func(line *Line) bool {
			return strings.Contains(line.Text(), msg)
		},
		desc: "text contains " + msg,
	}
}

func (log *Logger) FindLines(predicates ...LinePredicate) []*Line {
	log.lock.Lock()
	defer log.lock.Unlock()
	var found []*Line
Line:
	for _, line := range log.Lines {
		for _, predicate := range predicates {
			if !predicate.f(line) {
				continue Line
			}
		}
		found = append(found, line)
	}
	return found
}

func (log *Logger) CountLines(predicates ...LinePredicate) int {
	return len(log.FindLines(predicates...))
}

// FindSpanByLine returns nil unless there is exactly one span that
// has lines that match the predicate.
func (log *Logger) FindSpanByLine(predicates ...LinePredicate) *Span {
	matching := log.FindLines(predicates...)
	if len(matching) == 0 {
		return nil
	}
	span := matching[0].Span
	for _, m := range matching {
		if m.Span != span {
			return nil
		}
	}
	return span
}

type SpanPredicate struct {
	f    func(*Span) bool
	desc string
}

type SpanPredicates []SpanPredicate

func (p SpanPredicate) String() string { return p.desc }

func ShortEquals(name string) SpanPredicate {
	return SpanPredicate{
		f: func(span *Span) bool {
			return span.Short() == name
		},
		desc: "short equals " + name,
	}
}

func NameEquals(name string) SpanPredicate {
	return SpanPredicate{
		f: func(span *Span) bool {
			return span.Name == name
		},
		desc: "name equals " + name,
	}
}

func (log *Logger) FindSpan(predicates ...SpanPredicate) *Span {
Request:
	for _, span := range log.Requests {
		for _, predicate := range predicates {
			if !predicate.f(span) {
				continue Request
			}
		}
		return span
	}
Span:
	for _, span := range log.Spans {
		for _, predicate := range predicates {
			if !predicate.f(span) {
				continue Span
			}
		}
		return span
	}
	return nil
}
