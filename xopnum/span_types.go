package xopnum

type SpanType int

const (
	Request SpanType = iota
	Event
	Cron
	Watcher
	Stream
)
