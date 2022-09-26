package xopjson_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/muir/xop-go"
	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopbase"
	"github.com/muir/xop-go/xopbytes"
	"github.com/muir/xop-go/xopjson"
	"github.com/muir/xop-go/xopnum"
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

	Type        string `json:"type"`
	Name        string `json:"name"`
	Duration    int64  `json:"dur"`
	SpanVersion int    `json:"span.ver"`

	// requests

	Implmentation string `json:"impl"`
	TraceID       string `json:"trace.id"`
	ParentID      string `json:"parent.id"`
	RequestID     string `json:"request.id"`
	State         string `json:"trace.state"`
	Baggage       string `json:"trace.baggage"`
}

type checkConfig struct {
	minVersions         int
	maxVersions         int
	hasAttributesObject bool
}

type checker struct {
	tlog             *xoptest.TestLogger
	config           checkConfig
	spansSeen        []bool
	requestsSeen     []bool
	messagesNotSeen  map[string][]int
	spanIndex        map[string]int
	requestIndex     map[string]int
	accumulatedSpans map[string]map[string]interface{}
	sequencing       map[string]int
}

func TestASingleLine(t *testing.T) {
	var buffer xoputil.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithDuration("dur", xopjson.AsString),
		xopjson.WithSpanTags(xopjson.SpanIDTagOption),
		xopjson.WithAttributesObject(true),
		xopjson.WithStackLineRewrite(func(s string) string {
			return "FOO-" + s
		}),
	)
	log := xop.NewSeed(xop.WithBase(jlog)).Request(t.Name())
	log.Alert().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Done()
	s := buffer.String()
	t.Log(s)
	lines := strings.Split(buffer.String(), "\n")
	require.Equal(t, 3, len(lines), "three lines")
	assert.Contains(t, lines[0], `"span.id":`)
	assert.Contains(t, lines[0], `"attributes":{`) // }
	assert.Contains(t, lines[0], `"foo":"bar"`)
	assert.Contains(t, lines[0], `"lvl":20`)
	assert.Contains(t, lines[0], `"ts":`)
	assert.Contains(t, lines[0], `"blast":99`)
	assert.Contains(t, lines[0], `"stack":["FOO-`)
	assert.NotContains(t, lines[0], `"trace.id":`)
	assert.NotContains(t, lines[1], `"stack":[`)
	assert.Contains(t, lines[1], `"span.id":`)
	assert.Contains(t, lines[1], `"dur":"`)
	assert.Contains(t, lines[1], `"span.ver":0`)
	assert.Contains(t, lines[1], `"type":"request"`)
	assert.Contains(t, lines[1], `"span.name":"TestASingleLine"`)
}

func TestParameters(t *testing.T) {
	jsonCases := []struct {
		name         string
		joptions     []xopjson.Option
		settings     func(settings *xop.LogSettings)
		waitForFlush bool
		checkConfig  checkConfig
		extraFlushes int
	}{
		{
			name: "buffered",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(true),
				xopjson.WithBufferedLines(8 * 1024 * 1024),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			},
			checkConfig: checkConfig{
				minVersions:         2,
				hasAttributesObject: true,
			},
		},
		{
			name: "unbuffered/no-attributes",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(true),
				xopjson.WithBufferedLines(8 * 1024 * 1024),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
				xopjson.WithAttributesObject(false),
			},
			checkConfig: checkConfig{
				minVersions:         2,
				hasAttributesObject: false,
			},
		},
		{
			name: "unsynced",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(false),
				xopjson.WithSpanTags(xopjson.SpanIDTagOption),
			},
			settings: func(settings *xop.LogSettings) {
				settings.SynchronousFlush(false)
			},
			// with sync=false, we don't know when .Done will trigger a flush.
			waitForFlush: true,
			checkConfig: checkConfig{
				minVersions:         1,
				hasAttributesObject: true,
			},
		},
	}

	for _, tc := range jsonCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, mc := range xoptestutil.MessageCases {
				mc := mc
				t.Run(mc.Name, func(t *testing.T) {
					var buffer xoputil.Buffer
					joptions := []xopjson.Option{
						xopjson.WithDuration("dur", xopjson.AsNanos),
						xopjson.WithSpanStarts(true),
						xopjson.WithBufferedLines(8 * 1024 * 1024),
						xopjson.WithAttributesObject(true),
					}
					joptions = append(joptions, tc.joptions...)

					jlog := xopjson.New(
						xopbytes.WriteToIOWriter(&buffer),
						joptions...)
					tlog := xoptest.New(t)
					settings := func(settings *xop.LogSettings) {
						settings.SynchronousFlush(true)
					}
					if tc.settings != nil {
						settings = tc.settings
					}
					log := xop.NewSeed(
						xop.WithBase(jlog),
						xop.WithBase(tlog),
						xop.WithSettings(settings),
					).Request(t.Name())

					mc.Do(t, log, tlog)

					expectedFlushes := 1 + tc.extraFlushes + mc.ExtraFlushes
					if tc.waitForFlush {
						assert.Eventually(t, func() bool {
							return xoptestutil.EventCount(tlog, xoptest.FlushEvent) >= expectedFlushes
						}, time.Second, time.Millisecond*3)
					}
					t.Log("\n", buffer.String())
					xoptestutil.DumpEvents(t, tlog)
					assert.Equal(t, expectedFlushes, xoptestutil.EventCount(tlog, xoptest.FlushEvent), "count of flush")
					newChecker(t, tlog, tc.checkConfig).check(t, buffer.String())
				})
			}
		})
	}
}

func newChecker(t *testing.T, tlog *xoptest.TestLogger, config checkConfig) *checker {
	if config.maxVersions < config.minVersions {
		config.maxVersions = config.minVersions
	}
	c := &checker{
		tlog:             tlog,
		config:           config,
		spansSeen:        make([]bool, len(tlog.Spans)),
		requestsSeen:     make([]bool, len(tlog.Requests)),
		messagesNotSeen:  make(map[string][]int),
		spanIndex:        make(map[string]int),
		requestIndex:     make(map[string]int),
		accumulatedSpans: make(map[string]map[string]interface{}),
		sequencing:       make(map[string]int),
	}
	for i, line := range tlog.Lines {
		m := line.TemplateOrMessage()
		if debugTlog {
			t.Logf("recorded line: '%s'", m)
		}
		c.messagesNotSeen[m] = append(c.messagesNotSeen[m], i)
	}
	for i, span := range tlog.Spans {
		if debugTspan {
			t.Logf("recorded span: %s - %s", span.Trace.Trace.SpanID().String(), span.Name)
		}
		_, ok := c.spanIndex[span.Trace.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate span id %s", span.Trace.Trace.SpanID().String())
		c.spanIndex[span.Trace.Trace.SpanID().String()] = i
	}
	for i, request := range tlog.Requests {
		if debugTspan {
			t.Logf("recorded request: %s - %s", request.Trace.Trace.SpanID().String(), request.Name)
		}
		_, ok := c.spanIndex[request.Trace.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate span/request id %s", request.Trace.Trace.SpanID().String())
		_, ok = c.requestIndex[request.Trace.Trace.SpanID().String()]
		assert.Falsef(t, ok, "duplicate request id %s", request.Trace.Trace.SpanID().String())
		c.requestIndex[request.Trace.Trace.SpanID().String()] = i
	}
	for spanID, versions := range c.sequencing {
		if c.config.minVersions == c.config.maxVersions {
			assert.Equal(t, versions, c.config.minVersions, "version count for span %s", spanID)
		} else {
			assert.GreaterOrEqualf(t, versions, c.config.minVersions, "version count for span %s", spanID)
			assert.LessOrEqualf(t, versions, c.config.maxVersions, "version count for span %s", spanID)
		}
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
			c.line(t, super, generic)
		case "span":
			t.Logf("check span: %s", line)
			c.span(t, super, generic)
		case "request":
			t.Logf("check request: %s", line)
			c.request(t, super, generic)
		}
	}
	for _, ia := range c.messagesNotSeen {
		for _, li := range ia {
			line := c.tlog.Lines[li]
			t.Errorf("line '%s' not found in JSON output", line.Text)
		}
	}
	for _, span := range c.tlog.Spans {
		spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
			if c.config.hasAttributesObject {
				t.Logf("comparing metadata: %+v vs %+v", span.Metadata, spanAttributes)
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.attributes", false)
			} else {
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
	}
	for _, span := range c.tlog.Requests {
		spanAttributes := c.accumulatedSpans[span.Trace.Trace.SpanID().String()]
		if len(span.Metadata) != 0 || len(spanAttributes) != 0 {
			if c.config.hasAttributesObject {
				t.Logf("comparing metadata: %+v vs %+v", span.Metadata, spanAttributes)
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.request.attributes", false)
			} else {
				compareData(t, span.Metadata, span.MetadataType, "xoptest.Metadata", spanAttributes, "xopjson.span.generic", true)
			}
		}
	}
}

func (c *checker) line(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.NotEqual(t, xopnum.Level(0), super.Level, "level")
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
	if c.config.hasAttributesObject {
		compareData(t, line.Data, line.DataType, "xoptest.Data", super.Attributes, "xopjson.Attributes", false)
	} else {
		assert.Empty(t, super.Attributes)
		compareData(t, line.Data, line.DataType, "xoptest.Data", generic, "xopjson.Generic", true)
	}
}

func (c *checker) span(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.Empty(t, super.Level, "no level expected")
	var prior int
	var ok bool
	if assert.NotEmpty(t, super.SpanID, "has span id") {
		prior, ok = c.sequencing[super.SpanID]
	}
	if super.SpanVersion > 0 {
		if assert.True(t, ok, "has prior version") {
			assert.Equal(t, prior+1, super.SpanVersion, "version is in sequence")
		}
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
	} else {
		assert.False(t, ok, "no prior version expected")
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	if c.config.hasAttributesObject {
		combineAttributes(super.Attributes, c.accumulatedSpans[super.SpanID])
	} else {
		combineAttributes(generic, c.accumulatedSpans[super.SpanID])
	}
	c.sequencing[super.SpanID] = super.SpanVersion
	assert.Less(t, super.Duration, int64(time.Second*10), "duration")
}

func (c *checker) request(t *testing.T, super supersetObject, generic map[string]interface{}) {
	assert.Empty(t, super.Level, "no level expected")
	var prior int
	var ok bool
	if assert.NotEmpty(t, super.SpanID, "has span id") {
		prior, ok = c.sequencing[super.SpanID]
	}
	if super.SpanVersion > 0 {
		if assert.True(t, ok, "has prior version") {
			assert.Equal(t, prior+1, super.SpanVersion, "version is in sequence")
		}
		assert.NotEmpty(t, super.Duration, "duration is set")
		assert.NotNil(t, c.accumulatedSpans[super.SpanID], "has prior")
	} else {
		assert.False(t, ok, "no prior version expected")
		assert.NotEmpty(t, super.TraceID, "has trace id")
		assert.False(t, super.Timestamp.IsZero(), "timestamp is set")
		assert.Nil(t, c.accumulatedSpans[super.SpanID], "has prior")
		c.accumulatedSpans[super.SpanID] = make(map[string]interface{})
	}
	if c.config.hasAttributesObject {
		combineAttributes(super.Attributes, c.accumulatedSpans[super.SpanID])
	} else {
		combineAttributes(generic, c.accumulatedSpans[super.SpanID])
	}
	c.sequencing[super.SpanID] = super.SpanVersion
	assert.Less(t, super.Duration, int64(time.Second*10), "duration")
}

func combineAttributes(from map[string]interface{}, attributes map[string]interface{}) {
	for k, v := range from {
		attributes[k] = v
	}
}

var xoptestConvert map[xopbase.DataType]func(interface{}) interface{}

func init() {
	xoptestConvert = map[xopbase.DataType]func(interface{}) interface{}{
		xopbase.LinkDataType: func(generic interface{}) interface{} {
			return map[string]interface{}{
				"xop.link": generic.(trace.Trace).String(),
			}
		},
		xopbase.LinkArrayDataType: genArrayConvert(xopbase.LinkDataType),
	}
}

func genArrayConvert(edt xopbase.DataType) func(interface{}) interface{} {
	return func(generic interface{}) interface{} {
		var sl []interface{}
		for _, element := range generic.([]interface{}) {
			sl = append(sl, xoptestConvert[edt](element))
		}
		return sl
	}
}

func compareData(t *testing.T, aOrig map[string]interface{}, types map[string]xopbase.DataType, aDesc string, b map[string]interface{}, bDesc string, ignoreExtra bool) {
	if len(aOrig) == 0 && len(b) == 0 {
		return
	}
	a := make(map[string]interface{})
	for k, v := range aOrig {
		if f, ok := xoptestConvert[types[k]]; ok {
			a[k] = f(v)
		} else {
			a[k] = v
		}
	}
	if ignoreExtra {
		tmp := make(map[string]interface{})
		for k := range a {
			if v, ok := b[k]; ok {
				tmp[k] = v
			}
		}
		b = tmp
	}
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
