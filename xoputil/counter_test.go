package xoputil_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"
)

func TestRequestCounter(t *testing.T) {
	// The basic idea is to hit it hard with unique values and duplicates.
	seed := time.Now().UnixNano()
	const traceCount = 20
	const requestCount = 200
	const dupCount = 5
	t.Logf("seed = %d", seed)
	testRequestCounter(t, seed, traceCount, requestCount, dupCount)
}

func testRequestCounter(t *testing.T, seed int64, traceCount int, requestCount int, dupCount int) {
	rand := rand.New(rand.NewSource(seed))
	traces := make([]xoptrace.HexBytes16, traceCount)
	for i := range traces {
		traces[i] = xoptrace.NewRandomTraceID()
	}
	// for i := range traces {
	// t.Logf("%d %s", i, traces[i])
	// }
	requests := make([]xoptrace.Trace, requestCount)
	for i := range requests {
		requests[i].TraceID().Set(traces[rand.Intn(traceCount)])
		requests[i].SpanID().SetRandom()
	}
	// for i := range requests {
	// t.Logf("%d %s", i, requests[i])
	// }
	starter := make(chan struct{})
	counter := xoputil.NewRequestCounter()
	var wg sync.WaitGroup
	traceNums := make(map[int]int)
	requestNums := make(map[[2]int]int)
	var threadCount int
	var mu sync.Mutex
	for _, req := range requests {
		req := req
		for i := 0; i < dupCount; i++ {
			wg.Add(1)
			go func() {
				<-starter
				traceNum, requestNum := counter.GetNumber(req)
				mu.Lock()
				traceNums[traceNum]++
				requestNums[[2]int{traceNum, requestNum}]++
				threadCount++
				mu.Unlock()
				wg.Done()
			}()
		}
	}
	close(starter)
	wg.Wait()
	assert.Equal(t, traceCount, len(traceNums), "unique trace numbers")
	assert.Equal(t, requestCount, len(requestNums), "unique request numbers")
	t.Logf("raced %d threads", threadCount)
}
