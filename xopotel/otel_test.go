package xopotel_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopjson"
	"github.com/xoplog/xop-go/xopotel"
	"github.com/xoplog/xop-go/xopotel/xopoteltest"
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
	log := xopotel.SpanToLog(ctx, "test-span")
	log.Alert().String(xopat.K("foo"), "bar").Int(xopat.K("blast"), 99).Msg("a test line")
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
		name          string
		idGen         bool
		useUnhacker   bool
		useBaseLogger bool
	}{
		{
			name:  "seedModifier-with-id",
			idGen: true,
		},
		{
			name:  "seedModifier-without-id",
			idGen: false,
		},
		{
			name:        "seedModifier-with-unhacker-and-id",
			idGen:       true,
			useUnhacker: true,
		},
		{
			name:          "baselogger",
			idGen:         true,
			useUnhacker:   false,
			useBaseLogger: true,
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

						jExporter := xopotel.ExportToXOP(jLog)
						tpo = append(tpo, sdktrace.WithBatcher(jExporter))
					}

					rLog := xoptest.New(t)
					rLog.SetPrefix("REPLAY ")
					exporter := xopotel.ExportToXOP(rLog)
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

					var seed xop.Seed
					if tc.useBaseLogger {
						seed = xop.NewSeed(
							xop.WithBase(tLog),
							xop.WithBase(xopotel.BaseLogger(tracerProvider)),
						)
					} else {
						seed = xop.NewSeed(
							xop.WithBase(tLog),
							xopotel.SeedModifier(ctx, tracerProvider),
						)
					}
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
//		test OTEL log actions
//		|
//		v
//	   	OTEL	-> JSON -> unpack "origin"
//		|
//		v
//		ExportToXOP
//		|
//		v
//		combinedBaseLogger -> xoptest.Logger -> xopcon.Logger -> "O"
//		|
//		v
//		xopotel.BufferedBaseLogger
//		|
//		v
//		OTEL	-> JSON -> unpack "replay"
//		|
//		v
//		ExportToXOP
//		|
//		v
//		xoptest.Logger -> xopcon.Logger "R"
//
// Do we get the same JSON?
func TestOTELRoundTrip(t *testing.T) {
	exampleResource := resource.NewWithAttributes("http://test/foo",
		attribute.String("environment", "demo"),
	)

	enc, e := json.Marshal(exampleResource)
	require.NoError(t, e)
	t.Log("example resource", string(enc))

	// JSON created directly by OTEL (no XOP involved) lands here
	var origin bytes.Buffer
	originExporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&origin),
	)
	require.NoError(t, err)

	// The final resulting JSON lands here after going to XOP and back to OTEL
	var replay bytes.Buffer
	replayExporter, err := stdouttrace.New(
		stdouttrace.WithWriter(&replay),
	)
	require.NoError(t, err)

	exporterWrapper := xopotel.NewBufferedReplayExporterWrapper()

	wrappedReplayExporter := exporterWrapper.WrapExporter(replayExporter)

	reExportXoptest := xoptest.New(t)
	reExportXoptest.SetPrefix("R:")

	reExportToXop := xopotel.ExportToXOP(reExportXoptest)

	// This is the base Logger that writes to OTEL
	replayBaseLogger := exporterWrapper.BufferedReplayLogger(
		// Notice: WithResource() is not used here since
		// this will come from the replay logger.
		sdktrace.WithBatcher(wrappedReplayExporter),
		sdktrace.WithBatcher(reExportToXop),
	)

	// This provides some text output
	testLogger := xoptest.New(t)
	testLogger.SetPrefix("O:")

	// This combines the replayBaseLogger with an xoptest.Logger so we can see the output
	combinedReplayBaseLogger := xop.CombineBaseLoggers(replayBaseLogger, testLogger)

	// This is the OTEL exporter that writes to the replay base logger
	xopotelExporter := xopotel.ExportToXOP(combinedReplayBaseLogger)

	// This is the TracerProvider that writes to the unmodified JSON exporter
	// and also to the exporter that writes to XOP and back to OTEL
	tpOrigin := sdktrace.NewTracerProvider(
		sdktrace.WithResource(exampleResource),
		sdktrace.WithBatcher(originExporter),
		sdktrace.WithBatcher(xopotelExporter),
	)

	// This is the OTEL tracer we'll use to generate some OTEL logs.
	tracer := tpOrigin.Tracer("round-trip",
		oteltrace.WithSchemaURL("http://something"),
		oteltrace.WithInstrumentationAttributes(kvExamples("ia")...),
		oteltrace.WithInstrumentationVersion("0.3.0-test4"),
	)
	ctx := context.Background()

	// Now we will generate some rich OTEl logs.
	span1Ctx, span1 := tracer.Start(ctx, "span1 first-name",
		oteltrace.WithNewRoot(),
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
		oteltrace.WithAttributes(kvExamples("s1start")...),
	)
	span1.AddEvent("span1-event",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s1event")...),
	)
	span1.SetAttributes(kvExamples("s1set")...)
	span1.SetStatus(codes.Error, "a-okay here")
	span1.SetName("span1 new-name")
	span1.RecordError(fmt.Errorf("an error"),
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s1error")...),
	)
	var bundle1 xoptrace.Bundle
	bundle1.Parent.TraceID().SetRandom()
	bundle1.Parent.SpanID().SetRandom()
	var traceState oteltrace.TraceState
	traceState, err = traceState.Insert("foo", "bar")
	require.NoError(t, err)
	traceState, err = traceState.Insert("abc", "xyz")
	require.NoError(t, err)
	spanConfig1 := oteltrace.SpanContextConfig{
		TraceID:    bundle1.Parent.TraceID().Array(),
		SpanID:     bundle1.Parent.SpanID().Array(),
		Remote:     true,
		TraceFlags: oteltrace.TraceFlags(bundle1.Parent.Flags().Array()[0]),
		TraceState: traceState,
	}
	var bundle2 xoptrace.Bundle
	bundle2.Parent.TraceID().SetRandom()
	bundle2.Parent.SpanID().SetRandom()
	spanConfig2 := oteltrace.SpanContextConfig{
		TraceID:    bundle2.Parent.TraceID().Array(),
		SpanID:     bundle2.Parent.SpanID().Array(),
		Remote:     false,
		TraceFlags: oteltrace.TraceFlags(bundle2.Parent.Flags().Array()[0]),
	}
	_, span2 := tracer.Start(span1Ctx, "span2",
		oteltrace.WithLinks(
			oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(spanConfig1),
				Attributes:  kvExamples("la"),
			},
			oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(spanConfig2),
			},
		),
	)
	span1.AddEvent("span2-event",
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithAttributes(kvExamples("s2event")...),
	)
	span2.End()
	span1.End()

	// We've finished generating logs. Flush them.
	require.NoError(t, tpOrigin.ForceFlush(ctx), "flush origin")

	// Now we verify the end result, looking for differences
	originSpans := unpack(t, "origin", origin.Bytes())
	replaySpans := unpack(t, "replay", replay.Bytes())
	assert.NotEmpty(t, originSpans, "some spans")
	assert.Equal(t, len(originSpans), len(replaySpans), "count of spans")
	diffs := xopoteltest.CompareSpanStubSlice("", originSpans, replaySpans)
	filtered := make([]xopoteltest.Diff, 0, len(diffs))
	for _, diff := range diffs {
		t.Log("diff", diff)
		filtered = append(filtered, diff)
	}
	assert.Equal(t, 0, len(filtered), "count of unfiltered diffs")
}

func unpack(t *testing.T, what string, data []byte) []xopoteltest.SpanStub {
	var spans []xopoteltest.SpanStub
	for _, chunk := range bytes.Split(data, []byte{'\n'}) {
		if len(chunk) == 0 {
			continue
		}
		var span xopoteltest.SpanStub
		err := json.Unmarshal(chunk, &span)
		require.NoErrorf(t, err, "unmarshal '%s'", string(chunk))
		t.Logf("%s unpacking %s", what, string(chunk))
		t.Logf("%s unpacked %s %s", what, span.SpanContext.TraceID().String(), span.SpanContext.SpanID().String())
		spans = append(spans, span)
	}
	return spans
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
		attribute.Int64Slice(p+"int64-slice", []int64{-7}),
		attribute.Float64(p+"one-float", 299943),
		attribute.Float64Slice(p+"float-slice", []float64{-7.3, 19.2}),
	}
}
