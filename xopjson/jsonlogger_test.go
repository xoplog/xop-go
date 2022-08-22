package xopjson_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopjson"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoptest/xoptestutil"
	"github.com/muir/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	debugTlog  = true
	debugTspan = true
)

type supersetObject struct {
	// lines, spans, and requests

	Timestamp  xoptestutil.TS         `json:"ts"`
	Attributes map[string]interface{} `json:"attributes"`

	// lines

	Level  int      `json:"lvl"`
	SpanID string   `json:"span.id"`
	Stack  []string `json:"stack"`
	Msg    string   `json:"msg"`
	Format string   `json:"fmt"`

	// requests & spans

	Type     string `json:"type"`
	Name     string `json:"name"`
	Duration int64  `json:"dur"`

	// requests

	Implmentation  string `json:"impl"`
	TraceID        string `json:"trace.id"`
	ParentID       string `json:"parent.id"`
	RequestID      string `json:"request.id"`
	State          string `json:"trace.state"`
	Baggage        string `json:"trace.baggage"`
	RequestVersion int    `json:"request.ver"` // TODO: change to span.ver?

	// spans

	SpanVersion int `json:"span.ver"`
}

type checker struct {
	tlog                *xoptest.TestLogger
	hasAttributesObject bool
	spansSeen           []bool
	requestsSeen        []bool
	messagesNotSeen     map[string][]int
	spanIndex           map[string]int
	requestIndex        map[string]int
	accumulatedSpans    map[string]map[string]interface{}
}

func TestASingleLine(t *testing.T) {
	var buffer xoputil.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithEpochTime(time.Microsecond),
		xopjson.WithDuration("dur", xopjson.AsString),
		xopjson.WithSpanTags(xopjson.SpanIDTagOption),
		xopjson.WithBufferedLines(8*1024*1024),
		xopjson.WithAttributesObject(true),
	)
	log := xop.NewSeed(xop.WithBase(jlog)).Request(t.Name())
	log.Info().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Error().Msg("basic error message")
	log.Flush()
	log.Done()
	s := buffer.String()
	t.Log(s)
}

func TestNoBuffer(t *testing.T) {
	var buffer xoputil.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithEpochTime(time.Microsecond),
		xopjson.WithDuration("dur", xopjson.AsMicros),
		xopjson.WithSpanTags(xopjson.SpanIDTagOption),
		xopjson.WithSpanStarts(true),
		xopjson.WithBufferedLines(8*1024*1024),
		xopjson.WithAttributesObject(true),
	)
	tlog := xoptest.New(t)
	t.Log(buffer.String())
	log := xop.NewSeed(
		xop.WithBase(jlog),
		xop.WithBase(tlog),
		xop.WithSettings(func(settings *xop.LogSettings) {
			settings.SynchronousFlush(true)
		}),
	).Request(t.Name())
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")

	ss := log.Sub().Detach().Fork("a fork")
	xoptestutil.MicroNap()
	ss.Alert().String("frightening", "stuff").Static("like a rock")
	ss.Span().String(xopconst.EndpointRoute, "/some/thing")

	xoptestutil.MicroNap()
	tlog.CustomEvent("before log.Done")
	log.Done()
	tlog.CustomEvent("after log.Done")
	ss.Debug().Msg("sub-span debug message")
	xoptestutil.MicroNap()
	tlog.CustomEvent("before ss.Done")
	ss.Done()
	tlog.CustomEvent("after ss.Done")

	t.Log("\n", buffer.String())
	xoptestutil.DumpEvents(t, tlog)
	xoptestutil.ExpectEventCount(t, tlog, xoptest.FlushEvent, 1)

	newChecker(t, tlog, true).check(t, buffer.String())
}

func newChecker(t *testing.T, tlog *xoptest.TestLogger, hasAttributesObject bool) *checker {
	c := &checker{
		tlog:                tlog,
		hasAttributesObject: hasAttributesObject,
		spansSeen:           make([]bool, len(tlog.Spans)),
		requestsSeen:        make([]bool, len(tlog.Requests)),
		messagesNotSeen:     make(map[string][]int),
		spanIndex:           make(map[string]int),
		requestIndex:        make(map[string]int),
		accumulatedSpans:    make(map[string]map[string]interface{}),
	}
	for i, line := range tlog.Lines {
		if debugTlog {
			t.Logf("recorded line: '%s'", line.Message)
		}
		c.messagesNotSeen[line.Message] = append(c.messagesNotSeen[line.Message], i)
	}
	for i, span := range tlog.Spans {
		if debugTspan {
			t.Logf("recorded span: %s - %s", span.Trace.Trace.SpanIDString(), span.Name)
		}
		_, ok := c.spanIndex[span.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate span id %s", span.Trace.Trace.SpanIDString())
		c.spanIndex[span.Trace.Trace.SpanIDString()] = i
	}
	for i, request := range tlog.Requests {
		if debugTspan {
			t.Logf("recorded request: %s - %s", request.Trace.Trace.SpanIDString(), request.Name)
		}
		_, ok := c.spanIndex[request.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate span/request id %s", request.Trace.Trace.SpanIDString())
		_, ok = c.requestIndex[request.Trace.Trace.SpanIDString()]
		assert.Falsef(t, ok, "duplicate request id %s", request.Trace.Trace.SpanIDString())
		c.requestIndex[request.Trace.Trace.SpanIDString()] = i
	}
	return c
}

func (c *checker) check(t *testing.T, data string) {
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		var generic map[string]interface{}
		err := json.Unmarshal([]byte(line), &generic)
		require.NoErrorf(t, err, "decode to generic '%s'", line)

		var super supersetObject
		err = json.Unmarshal([]byte(line), &super)
		require.NoErrorf(t, err, "decode to super: %s", line)

		switch super.Type {
		case "", "line":
			t.Logf("check line: %s", line)
			c.line(t, super)
		case "span":
			t.Logf("check span: %s", line)
			c.span(t, super)
		case "request":
			t.Logf("check request: %s", line)
			c.request(t, super)
		}
	}
	for _, ia := range c.messagesNotSeen {
		for _, li := range ia {
			line := c.tlog.Lines[li]
			t.Errorf("line '%s' not found in JSON output", line.Text)
		}
	}
	for _, span := range c.tlog.Spans {
		if len(span.Metadata) != 0 || len(c.accumulatedSpans[span.Trace.Trace.SpanID().String()]) != 0 {
			assert.Equalf(t, span.Metadata, c.accumulatedSpans[span.Trace.Trace.SpanID().String()], "attributes for span %s", span.Trace.Trace.SpanID())
		}
	}
}

func (c *checker) line(t *testing.T, super supersetObject) {
	assert.NotEqual(t, xopconst.Level(0), super.Level, "level")
	assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
	assert.NotEmpty(t, super.Msg, "message")
	mns := c.messagesNotSeen[super.Msg]
	if !assert.NotNilf(t, mns, "test line with message '%s'", super.Msg) {
		return
	}
	line := c.tlog.Lines[mns[0]]
	c.messagesNotSeen[super.Msg] = c.messagesNotSeen[super.Msg][1:]
	assert.Truef(t, super.Timestamp.Round(time.Millisecond).Equal(line.Timestamp.Round(time.Millisecond)), "timestamps %s vs %s", line.Timestamp, super.Timestamp)
	assert.Equal(t, int(line.Level), super.Level, "level")
	compareData(t, line.Data, "xoptest.Data", super.Attributes, "xopjson.Attributes")
}

func (c *checker) span(t *testing.T, super supersetObject) {
	assert.Empty(t, super.Level, "no level expected")
	if super.SpanVersion > 0 {
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
		combineAttributes(super, c.accumulatedSpans[super.SpanID])
	} else {
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	assert.Less(t, super.Duration, int64(time.Second*10/time.Microsecond), "duration")
}

func (c *checker) request(t *testing.T, super supersetObject) {
	assert.Empty(t, super.Level, "no level expected")
	assert.NotEmpty(t, super.SpanID, "has span id")
	if super.RequestVersion > 0 {
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
		combineAttributes(super, c.accumulatedSpans[super.SpanID])
	} else {
		assert.NotEmpty(t, super.TraceID, "has trace id")
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	assert.Less(t, super.Duration, int64(time.Second*10/time.Microsecond), "duration")
}

func combineAttributes(super supersetObject, attributes map[string]interface{}) {
	for k, v := range super.Attributes {
		attributes[k] = v
	}
}

func compareData(t *testing.T, a map[string]interface{}, aDesc string, b map[string]interface{}, bDesc string) {
	if len(a) == 0 && len(b) == 0 {
		return
	}
	aEnc, err := json.Marshal(a)
	if !assert.NoErrorf(t, err, "marshal %s", aDesc) {
		return
	}
	var aRedone map[string]interface{}
	if !assert.NoErrorf(t, json.Unmarshal(aEnc, &aRedone), "remarshal %s", aDesc) {
		return
	}
	bEnc, err := json.Marshal(b)
	if !assert.NoErrorf(t, err, "marshal %s", bDesc) {
		return
	}
	var bRedone map[string]interface{}
	if !assert.NoErrorf(t, json.Unmarshal(bEnc, &bRedone), "remarshal %s", bDesc) {
		return
	}
	assert.Equalf(t, aRedone, bRedone, "%s vs %s", aDesc, bDesc)
}
