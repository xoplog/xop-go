package basewrap

import (
	"github.com/muir/xm"
)

type nonBlockingBase struct {
	under       xm.BaseLogger
	level       xm.Level
	spanBuffer  chan spanMessage
	flushBuffer chan flushMessage
	logBuffer   chan logMessage
}

type bufferedNonBlocking struct {
	buffered xm.BufferedBase
	base     *nonBlockingBase
}

type prefilledNonBlocking struct {
	prefilled xm.Prefilled
	base      *nonBlockingBase
}

type spanMessage struct {
	buffered    xm.BufferedBase
	description string
	traceId     xm.HexBytes
	spanId      xm.HexBytes
	searchTerms []xm.Field
	data        []xm.Field
}

type flushMessage struct {
	buffered xm.BufferedBase
}

type logMessage struct {
	prefilled xm.Prefilled
	level     xm.Level
	msg       string
	traceId   xm.HexBytes
	spanId    xm.HexBytes
	values    []xm.Field
}

// NonBlocking wraps a BaseLogger so that nearly all operations are
// asynchronous including Flush().  The bufferSize argument controls
// how many operations can be in-flight at once.  Anything above that
// limit will be dropped.
//
// NonBlocking bufferSize must be at least 10 and 500 is suggested.
func NonBlocking(underlying xm.BaseLogger, bufferSize int) xm.BaseLogger {
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
			msg.buffered.Span(msg.description, msg.traceId, msg.spanId, msg.searchTerms, msg.data)
		case msg := <-n.logBuffer:
			msg.prefilled.Log(msg.level, msg.msg, msg.traceId, msg.spanId, msg.values)
		}
	}
}

func (n *nonBlockingBase) SetLevel(level xm.Level) {
	n.under.SetLevel(level)
}

func (n *nonBlockingBase) WantDurable() bool {
	return true
}

func (n *nonBlockingBase) StartBuffer() xm.BufferedBase {
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

func (b *bufferedNonBlocking) Span(description string, traceId xm.HexBytes, spanId xm.HexBytes, searchTerms []xm.Field, data []xm.Field) {
	select {
	case b.base.spanBuffer <- spanMessage{
		buffered:    b.buffered,
		description: description,
		traceId:     traceId,
		spanId:      spanId,
		searchTerms: searchTerms,
		data:        data,
	}:
	default:
	}
}

func (b *bufferedNonBlocking) Prefill(f []xm.Field) xm.Prefilled {
	return prefilledNonBlocking{
		prefilled: b.buffered.Prefill(f),
		base:      b.base,
	}
}

func (p prefilledNonBlocking) Log(level xm.Level, msg string, traceId xm.HexBytes, spanId xm.HexBytes, values []xm.Field) {
	select {
	case p.base.logBuffer <- logMessage{
		prefilled: p.prefilled,
		level:     level,
		msg:       msg,
		traceId:   traceId,
		spanId:    spanId,
		values:    values,
	}:
	default:
	}
}
