package xoptest

import (
	"strings"
)

type LinePredicate func(*Line) bool

func MessageEquals(msg string) LinePredicate {
	return func(line *Line) bool {
		return line.Message == msg
	}
}

func TextContains(msg string) LinePredicate {
	return func(line *Line) bool {
		return strings.Contains(line.Text, msg)
	}
}

func (l *TestLogger) FindLines(predicates ...LinePredicate) []*Line {
	l.lock.Lock()
	defer l.lock.Unlock()
	var found []*Line
Line:
	for _, line := range l.Lines {
		for _, predicate := range predicates {
			if !predicate(line) {
				continue Line
			}
		}
		found = append(found, line)
	}
	return found
}

func (l *TestLogger) CountLines(predicates ...LinePredicate) int {
	return len(l.FindLines(predicates...))
}
