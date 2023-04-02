package xopotel_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopjson"
	"github.com/xoplog/xop-go/xopotel"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
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

// TestOTELRoundTrip does a round trip of logging:
//
//	OTEL -> xopotel.Exporter -> xopotel.BaseLogger -> OTEL
//	   \--> JSON                                       \--> JSON
//
// Do we get the same JSON?
func TestOTELRoundTrip(t *testing.T) {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)

	var replay bytes.Buffer
	replayExporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&replay),
		stdouttrace.WithPrettyPrint(),
	)
	require.NoError(t, err)

	var origin bytes.Buffer
	originExporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&origin),
		stdouttrace.WithPrettyPrint(),
	)
	require.NoError(t, err)

	tpReplay := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(replayExporter),
		sdktrace.WithResource(r),
	)

	ctx := context.Background()

	xopotelExporter := xopotel.NewExporter(xopotel.BaseLogger(ctx, tpReplay))

	tpOrigin := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(originExporter),
		sdktrace.WithBatcher(xopotelExporter),
		sdktrace.WithResource(r),
	)

	tracer := tpOrigin.Tracer("round-trip",
		oteltrace.WithSchemaURL("http://something"),
		oteltrace.WithInstrumentationAttributes(kvExamples("ia")...),
		oteltrace.WithInstrumentationVersion("0.3.0-test4"),
	)
	span1Ctx, span1 := tracer.Start(ctx, "span1 first-name",
		oteltrace.WithNewRoot(),
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
	)
	span1.AddEvent("span1-event",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s1event")...),
	)
	span1.SetAttributes("span1-attributes", oteltrace.WithAttributes(kvExamples("s1a")...))
	span1.SetStatus(codes.Ok, "a-okay here")
	span1.SetName("span1 new-name")
	span1.RecordError(fmt.Errorf("an error"),
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s1error")...),
	)
	var bundle xoptrace.Bundle
	bundle.Parent.TraceID().SetRandom()
	bundle.Parent.SpanID().SetRandom()
	var traceState oteltrace.TraceState
	traceState, err = traceState.Insert("foo", "bar")
	require.NoError(t, err)
	traceState, err = traceState.Insert("abc", "xyz")
	require.NoError(t, err)
	spanConfig := oteltrace.SpanContextConfig{
		TraceID:    bundle.Parent.TraceID().Array(),
		SpanID:     bundle.Parent.SpanID().Array(),
		Remote:     true,
		TraceFlags: oteltrace.TraceFlags(bundle.Parent.Flags().Array()[0]),
		TraceState: traceState,
	}
	span2Ctx, span2 := tracer.Start(span1Ctx, "span2",
		oteltrace.WithLinks(oteltrace.Link{
			SpanContext: oteltrace.NewSpanContext(spanConfig),
			Attributes:  kvExamples("la"),
		}),
	)
	span1.AddEvent("span2-event",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s2event")...),
	)
	span2.End()
	span1.End()
	require.NoError(t, tpOrigin.ForceFlush(ctx), "flush origin")
	require.NoError(t, tpReplay.ForceFlush(ctx), "flush replay")
	assert.Equal(t, origin.String(), replay.String())
}

func kvExamples(p string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Bool(p+"one-bool", true),
		attribute.BoolSlice(p+"bool-slice", []bool{false, true, false}),
		attribute.String(p+"one-string", "slilly stuff"),
		attribute.StringSlice(p+"string-slice", []string{"one", "two", "three"}),
		attribute.Int(p+"one-int", 389),
		attribute.IntSlice(p+"int-slice", []int{93, -4}),
		attribute.Int64(p+"one-int64", 299943),
		attribute.Int64Slice(p+"int64-slice", []int{-7}),
		attribute.Float64(p+"one-float", 299943),
		attribute.Float64Slice(p+"float-slice", []int{-7}),
	}
}
