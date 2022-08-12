package xopjson_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xopjson"
	"github.com/muir/xop-go/xoptest"
	"github.com/muir/xop-go/xoptest/xoptestutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	debugTlog  = true
	debugTspan = true
)

type supersetObject struct {
	// lines, spans, and requests

	Attributes map[string]interface{} `json:"attributes"`
	Timestamp  xoptestutil.TS         `json:"ts"`

	// lines

	Level  int      `json:"lvl"`
	SpanID string   `json:"span.id"`
	Stack  []string `json:"stack"`
	Msg    string   `json:"msg"`
	Format string   `json:"fmt"`

	// requests & spans

	Type string `json:"type"`
	Name string `json:"name"`

	// requests

	Implmentation string `json:"impl"`
	ParentID      string `json:"parent.id"`
	RequestID     string `json:"request.id"`
	State         string `json:"trace.state"`
	Baggage       string `json:"trace.baggage"`

	// spans

	Update bool `json:"span.update"`
}

type checker struct {
	tlog                *xoptest.TestLogger
	hasAttributesObject bool
	spansSeen           []bool
	requestsSeen        []bool
	messagesNotSeen     map[string][]int
	spanIndex           map[string]int
	requestIndex        map[string]int
}

func TestASingleLine(t *testing.T) {
	var buffer bytes.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithEpochTime(time.Nanosecond),
		xopjson.WithDurationFormat(xopjson.AsNanos),
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
	var buffer bytes.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithEpochTime(time.Nanosecond),
		xopjson.WithDurationFormat(xopjson.AsNanos),
		xopjson.WithSpanTags(xopjson.SpanIDTagOption),
		xopjson.WithBufferedLines(8*1024*1024),
		xopjson.WithAttributesObject(true),
	)
	tlog := xoptest.New(t)
	t.Log(buffer.String())
	log := xop.NewSeed(xop.WithBase(jlog), xop.WithBase(tlog)).Request(t.Name())
	log.Info().Msg("basic info message")
	log.Error().Msg("basic error message")
	log.Alert().Msg("basic alert message")
	log.Debug().Msg("basic debug message")
	log.Trace().Msg("basic trace message")
	log.Info().String("foo", "bar").Int("num", 38).Template("a test {foo} with {num}")
	ss := log.Sub().Fork("a fork").Wait()
	ss.Alert().String("frightening", "stuff").Static("like a rock")
	log.Done()
	ss.Debug().Msg("sub-span debug message")
	ss.Done()

	t.Log(buffer.String())

	newChecker(t, tlog, true).check(t, &buffer)
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

func (c *checker) check(t *testing.T, stream io.Reader) {
	dec := json.NewDecoder(stream)
	for {
		var generic map[string]interface{}
		err := dec.Decode(&generic)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		enc, err := json.Marshal(generic)
		require.NoError(t, err)

		t.Logf("check: %s", string(enc))

		var super supersetObject
		err = json.Unmarshal(enc, &super)
		require.NoErrorf(t, err, "decode re-encoded")

		switch super.Type {
		case "", "line":
			c.line(t, super)
		case "span":
			c.span(t, super)
		case "request":
			// c.request(t, super)
		}
	}
	for _, ia := range c.messagesNotSeen {
		for _, li := range ia {
			line := c.tlog.Lines[li]
			t.Errorf("line '%s' not found in JSON output", line.Text)
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
}

func (c *checker) span(t *testing.T, super supersetObject) {
}
