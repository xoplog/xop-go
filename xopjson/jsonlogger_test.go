package xopjson_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopjson"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checkConfig struct {
	minVersions         int
	hasAttributesObject bool
}

func TestASingleLineJSON(t *testing.T) {
	var buffer xoputil.Buffer
	jlog := xopjson.New(
		xopbytes.WriteToIOWriter(&buffer),
		xopjson.WithSpanStarts(false),
	)
	log := xop.NewSeed(
		xop.WithBase(jlog),
		xop.WithSettings(
			func(settings *xop.LogSettings) {
				settings.StackFilenameRewrite(func(s string) string {
					return "FOO-" + s
				})
			}),
	).Request(t.Name())
	log.Alert().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Done()
	s := buffer.String()
	t.Log(s)
	lines := strings.Split(buffer.String(), "\n")
	if !assert.Equal(t, 3, len(lines), "line count") {
		for i, line := range lines {
			t.Logf("#%d: %s", i+1, line)
		}
		return
	}
	assert.Contains(t, lines[0], `"span.id":`)
	assert.Contains(t, lines[0], `"attributes":{`)   // }
	assert.Contains(t, lines[0], `"foo":{"v":"bar"`) // }
	assert.Contains(t, lines[0], `"lvl":"alert"`)
	assert.Contains(t, lines[0], `"ts":`)
	assert.Contains(t, lines[0], `"blast":{"v":99`) // }
	assert.Contains(t, lines[0], `"stack":["FOO-`)
	assert.NotContains(t, lines[0], `"trace.id":`)
	assert.NotContains(t, lines[1], `"stack":[`)
	assert.Contains(t, lines[1], `"span.id":`)
	assert.Contains(t, lines[1], `"dur":`)
	assert.Contains(t, lines[1], `"span.ver":0`)
	assert.Contains(t, lines[1], `"type":"request"`)
	assert.Contains(t, lines[1], `"span.name":"TestASingleLineJSON"`)
}

func TestReplayJSON(t *testing.T) {
	jsonCases := []struct {
		name         string
		joptions     []xopjson.Option
		settings     func(settings *xop.LogSettings)
		waitForFlush bool
		checkConfig  checkConfig
		extraFlushes int
	}{
		{
			name: "unbuffered/attributes",
			joptions: []xopjson.Option{
				xopjson.WithSpanStarts(true),
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
						xopjson.WithSpanStarts(true),
						xopjson.WithAttributeDefinitions(xopjson.AttributesDefinedEachRequest),
					}
					joptions = append(joptions, tc.joptions...)

					jlog := xopjson.New(
						xopbytes.WriteToIOWriter(&buffer),
						joptions...)
					tLog := xoptest.New(t)
					settings := func(settings *xop.LogSettings) {
						settings.SynchronousFlush(true)
					}
					if tc.settings != nil {
						settings = tc.settings
					}
					seed := xop.NewSeed(
						xop.WithBase(jlog),
						xop.WithSettings(settings),
					).Copy(xop.WithBase(tLog))

					if len(mc.SeedMods) != 0 {
						t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
						seed = seed.Copy(mc.SeedMods...)
					}

					log := seed.Request(t.Name())

					mc.Do(t, log, tLog)

					expectedFlushes := 1 + tc.extraFlushes + mc.ExtraFlushes
					if tc.waitForFlush {
						assert.Eventually(t, func() bool {
							return xoptestutil.EventCount(tLog, xoptest.FlushEvent) >= expectedFlushes
						}, time.Second, time.Millisecond*3)
					}
					t.Log("\n", buffer.String())
					xoptestutil.DumpEvents(t, tLog)
					assert.Equal(t, expectedFlushes, xoptestutil.EventCount(tLog, xoptest.FlushEvent), "count of flush")

					t.Log("verify generated JSON decodes as JSON")
					for _, inputText := range strings.Split(buffer.String(), "\n") {
						if inputText == "" {
							continue
						}
						var generic map[string]interface{}
						err := json.Unmarshal([]byte(inputText), &generic)
						require.NoErrorf(t, err, "unmarshal to generic '%s': %s", inputText, err)
					}

					t.Log("Replay")
					rLog := xoptest.New(t)
					err := xopjson.ReplayFromStrings(context.Background(), buffer.String(), rLog)
					require.NoError(t, err, "replay")

					t.Log("verify replay equals original")
					xoptestutil.VerifyReplay(t, tLog, rLog)
				})
			}
		})
	}
}
