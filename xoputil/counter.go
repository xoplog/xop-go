package xoputil

import (
	"sync"
	"sync/atomic"

	"github.com/xoplog/xop-go/xoptrace"

	"github.com/muir/gwrap"
)

type RequestCounter struct {
	traceCount int32
	traceMap   gwrap.SyncMap[[16]byte, *traceInfo]
}

type traceInfo struct {
	mu           sync.Mutex
	requestCount int32
	traceNum     int32
	requestMap   gwrap.SyncMap[[8]byte, *requestInfo]
}

type requestInfo struct {
	mu         sync.Mutex
	requestNum int32
}

func NewRequestCounter() *RequestCounter {
	return &RequestCounter{}
}

func (c *RequestCounter) GetNumber(trace xoptrace.Trace) (traceNum int, requestNum int, isNew bool) {
	traceID := trace.TraceID().Array()
	ti, ok := c.traceMap.Load(traceID)
	if !ok {
		var loaded bool
		n := &traceInfo{}
		n.mu.Lock() // unlocked only if !loaded (and thus this is the new value)
		ti, loaded = c.traceMap.LoadOrStore(traceID, n)
		if !loaded {
			// unfortunately, there is a race where a reader of this
			// value could get zero for a brand-new traceInfo. We
			// resolve that by releasing the lock to say the traceInfo
			// is now ready to use.
			atomic.StoreInt32(&n.traceNum, atomic.AddInt32(&c.traceCount, 1))
			n.mu.Unlock()
		}
	}
	if atomic.LoadInt32(&ti.traceNum) == 0 {
		// brand new traceInfo
		ti.mu.Lock()
		ti.mu.Unlock() //nolint:staticcheck // this empty lock/unlock is on purpose
		// no longer brand new, the request count will never change again
	}

	spanID := trace.SpanID().Array()
	ri, loaded := ti.requestMap.Load(spanID)
	if !loaded {
		n := &requestInfo{}
		n.mu.Lock()
		ri, loaded = ti.requestMap.LoadOrStore(spanID, n)
		if !loaded {
			atomic.StoreInt32(&ri.requestNum, atomic.AddInt32(&ti.requestCount, 1))
			ri.mu.Unlock()
		}
	}
	if atomic.LoadInt32(&ri.requestNum) == 0 {
		// brand new requestInfo
		ri.mu.Lock()
		ri.mu.Unlock() //nolint:staticcheck // this empty lock/unlock is on purpose
	}
	return int(atomic.LoadInt32(&ti.traceNum)), int(atomic.LoadInt32(&ri.requestNum)), !loaded
}
