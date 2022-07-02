package basewrap

import (
	"github.com/muir/xop"
	"github.com/muir/xop/trace"
	"github.com/muir/xop/zap"
)

type nonBlockingBase struct {
	under       xop.BaseLogger
	level       xopconst.Level
	spanBuffer  chan spanMessage
	flushBuffer chan flushMessage
	logBuffer   chan logMessage
}

type bufferedNonBlocking struct {
	buffered xop.BufferedBase
	base     *nonBlockingBase
}

type prefilledNonBlocking struct {
	prefilled xop.Prefilled
	base      *nonBlockingBase
}

type spanMessage struct {
	buffered    xop.BufferedBase
	description string
	trace       trace.Trace
	parent      trace.Trace
	searchTerms map[string][]string
	data        map[string]interface{}
}

type flushMessage struct {
	buffered xop.BufferedBase
}

type logMessage struct {
	prefilled xop.Prefilled
	level     xopconst.Level
	msg       string
	values    []xopthing.Thing
}

// NonBlocking wraps a BaseLogger so that nearly all operations are
// asynchronous including Flush().  The bufferSize argument controls
// how many operations can be in-flight at once.  Anything above that
// limit will be dropped.
//
// NonBlocking bufferSize must be at least 10 and 500 is suggested.
func NonBlocking(underlying xop.BaseLogger, bufferSize int) xop.BaseLogger {
	if bufferSize < 10 {
		bufferSize = 10
	}
	n := &nonBlockingBase{
		under:       underlying,
		spanBuffer:  make(chan spanMessage, int(bufferSize/10)),
		flushBuffer: make(chan flushMessage, int(bufferSize/10)),
		logBuffer:   make(chan logMessage, bufferSize-2*int(bufferSize/10)),
	}
	go n.receive()
	return n
}

func (n *nonBlockingBase) receive() {
	for {
		select {
		case msg := <-n.flushBuffer:
			msg.buffered.Flush()
		case msg := <-n.spanBuffer:
			msg.buffered.Span(msg.description, msg.trace, msg.parent, msg.searchTerms, msg.data)
		case msg := <-n.logBuffer:
			msg.prefilled.Log(msg.level, msg.msg, msg.values)
		}
	}
}

func (n *nonBlockingBase) SetLevel(level xopconst.Level) {
	n.under.SetLevel(level)
}

func (n *nonBlockingBase) WantDurable() bool {
	return true
}

func (n *nonBlockingBase) StartBuffer() xop.BufferedBase {
	return &bufferedNonBlocking{
		base:     n,
		buffered: n.StartBuffer(),
	}
}

func (b *bufferedNonBlocking) Flush() {
	select {
	case b.base.flushBuffer <- flushMessage{
		buffered: b.buffered,
	}:
	default:
	}
}

func (b *bufferedNonBlocking) Span(
	description string,
	trace trace.Trace,
	parent trace.Trace,
	searchTerms map[string][]string,
	data map[string]interface{},
) {
	select {
	case b.base.spanBuffer <- spanMessage{
		buffered:    b.buffered,
		description: description,
		trace:       trace,
		parent:      parent,
		searchTerms: searchTerms,
		data:        data,
	}:
	default:
	}
}

func (b *bufferedNonBlocking) Prefill(trace trace.Trace, f []xopthing.Thing) xop.Prefilled {
	return prefilledNonBlocking{
		prefilled: b.buffered.Prefill(trace, f),
		base:      b.base,
	}
}

func (p prefilledNonBlocking) Log(level xopconst.Level, msg string, values []xopthing.Thing) {
	select {
	case p.base.logBuffer <- logMessage{
		prefilled: p.prefilled,
		level:     level,
		msg:       msg,
		values:    values,
	}:
	default:
	}
}
