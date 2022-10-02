package xoptest

import (
	"strings"
)

type LinePredicate struct {
	f    func(*Line) bool
	desc string
}

type Predicates []LinePredicate

func (p LinePredicate) String() string { return p.desc }

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
			return strings.Contains(line.Text, msg)
		},
		desc: "text contains " + msg,
	}
}

func (log *TestLogger) FindLines(predicates ...LinePredicate) []*Line {
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

func (log *TestLogger) CountLines(predicates ...LinePredicate) int {
	return len(log.FindLines(predicates...))
}

// FindSpanByLine returns nil unless there is exactly one span that
// has lines that match the predicate.
func (log *TestLogger) FindSpanByLine(predicates ...LinePredicate) *Span {
	matching := log.FindLines(predicates...)
	if len(matching) == 0 {
		log.t.Log("no lines matching", Predicates(predicates))
		return nil
	}
	span := matching[0].Span
	for _, m := range matching {
		if m.Span != span {
			log.t.Log("multiple lines matching", Predicates(predicates))
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

func ShortEquals(short string) SpanPredicate {
	return SpanPredicate{
		f: func(span *Span) bool {
			return span.Short == short
		},
		desc: "short equals " + short,
	}
}

func (log *TestLogger) FindSpan(predicates ...SpanPredicate) *Span {
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
	log.t.Log("no spans match", SpanPredicates(predicates))
	return nil
}
