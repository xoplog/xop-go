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

const debugTlog = true

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
	log.Done()

	t.Log(buffer.String())

	newChecker(t, tlog, true).check(t, &buffer)
}

type IntTime struct {
}

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
}

func newChecker(t *testing.T, tlog *xoptest.TestLogger, hasAttributesObject bool) *checker {
	c := &checker{
		tlog:                tlog,
		hasAttributesObject: hasAttributesObject,
		spansSeen:           make([]bool, len(tlog.Spans)),
		requestsSeen:        make([]bool, len(tlog.Requests)),
		messagesNotSeen:     make(map[string][]int),
	}
	for i, line := range tlog.Lines {
		if debugTlog {
			t.Logf("test line message '%s'", line.Message)
		}
		c.messagesNotSeen[line.Message] = append(c.messagesNotSeen[line.Message], i)
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
			// c.span(t, super)
		case "request":
			// c.request(t, super)
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
	assert.Truef(t, super.Timestamp.Equal(line.Timestamp), "timestamps %s vs %s", line.Timestamp, super.Timestamp)
	assert.Equal(t, int(line.Level), super.Level, "level")
}
