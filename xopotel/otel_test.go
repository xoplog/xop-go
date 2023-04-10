package xopotel_test

import (
	"context"
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopjson"
	"github.com/xoplog/xop-go/xopotel"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestSingleLineOTEL(t *testing.T) {
	var buffer xoputil.Buffer

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&buffer),
		stdouttrace.WithPrettyPrint(),
	)
	require.NoError(t, err, "exporter")

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	ctx := context.Background()
	defer func() {
		err := tracerProvider.Shutdown(ctx)
		assert.NoError(t, err, "shutdown")
	}()

	tracer := tracerProvider.Tracer("")

	ctx, span := tracer.Start(ctx, "test-span")
	log := xopotel.SpanLog(ctx, "test-span")
	log.Alert().String("foo", "bar").Int("blast", 99).Msg("a test line")
	log.Done()
	span.End()
	tracerProvider.ForceFlush(context.Background())
	t.Log("logged:", buffer.String())
	assert.NotEmpty(t, buffer.String())
}

const jsonToo = false
const otelToo = true

func TestOTELBaseLoggerReplay(t *testing.T) {
	cases := []struct {
		name        string
		idGen       bool
		useUnhacker bool
	}{
		{
			name:  "baselogger-with-id",
			idGen: true,
		},
		{
			name:  "baselogger-without-id",
			idGen: false,
		},
		{
			name:        "baselogger-with-unhacker-and-id",
			idGen:       true,
			useUnhacker: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, mc := range xoptestutil.MessageCases {
				mc := mc
				if mc.SkipOTEL {
					continue
				}
				t.Run(mc.Name, func(t *testing.T) {

					var tpo []sdktrace.TracerProviderOption

					var jBuffer xoputil.Buffer
					if jsonToo {
						jLog := xopjson.New(
							xopbytes.WriteToIOWriter(&jBuffer),
						)

						jExporter := xopotel.NewExporter(jLog)
						tpo = append(tpo, sdktrace.WithBatcher(jExporter))
					}

					rLog := xoptest.New(t)
					rLog.SetPrefix("REPLAY ")
					exporter := xopotel.NewExporter(rLog)
					if tc.useUnhacker {
						unhacker := xopotel.NewUnhacker(exporter)
						tpo = append(tpo, sdktrace.WithBatcher(unhacker))
					} else {
						tpo = append(tpo, sdktrace.WithBatcher(exporter))
					}

					var buffer xoputil.Buffer
					if otelToo {
						otelExporter, err := stdouttrace.New(
							stdouttrace.WithWriter(&buffer),
							stdouttrace.WithPrettyPrint(),
						)
						require.NoError(t, err, "exporter")
						tpo = append(tpo, sdktrace.WithBatcher(otelExporter))
					}

					if tc.idGen {
						tpo = append(tpo, xopotel.IDGenerator())
					}

					tracerProvider := sdktrace.NewTracerProvider(tpo...)
					ctx, cancel := context.WithCancel(context.Background())
					defer func() {
						err := tracerProvider.Shutdown(context.Background())
						assert.NoError(t, err, "shutdown")
					}()

					tLog := xoptest.New(t)

					seed := xop.NewSeed(
						xop.WithBase(tLog),
						xopotel.BaseLogger(ctx, tracerProvider),
					)
					if len(mc.SeedMods) != 0 {
						t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
						seed = seed.Copy(mc.SeedMods...)
					}
					log := seed.Request(t.Name())
					mc.Do(t, log, tLog)

					cancel()
					tracerProvider.ForceFlush(context.Background())

					if otelToo {
						t.Log("logged:", buffer.String())
					}
					if jsonToo {
						t.Log("Jlogged:", jBuffer.String())
					}

					t.Log("verify replay equals original")
					xoptestutil.VerifyTestReplay(t, tLog, rLog)
				})
			}
		})
	}
}
