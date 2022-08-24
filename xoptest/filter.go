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

func (log *TestLogger) FindLines(predicates ...LinePredicate) []*Line {
	log.lock.Lock()
	defer log.lock.Unlock()
	var found []*Line
Line:
	for _, line := range log.Lines {
		for _, predicate := range predicates {
			if !predicate(line) {
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
