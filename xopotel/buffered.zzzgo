package xopotel

import (
	"context"
	"fmt"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// BufferedReplayLogger creates a Logger that can be used when replaying from other
// xopbase.Logger implementations into xopotel. It buffers all the logged data until
// Done() is called on a per-request basis. Additional logging after Done() is discarded.
//
// A TracerProvider and Tracer are constructed for each Request and discarded afterwards.
func BufferedReplayLogger(tracerProviderOpts ...sdktrace.TracerProviderOption) xopbase.Logger {
	return &bufferedLogger{
		tracerProviderOpts: tracerProviderOpts,
		id:                 "otelbuf-" + uuid.New().String(),
	}
}

type bufferedLogger struct {
	id                 string
	tracerProviderOpts []sdktrace.TracerProviderOption
}

func (logger *bufferedLogger) ID() string           { return logger.id }
func (logger *bufferedLogger) ReferencesKept() bool { return true }
func (logger *bufferedLogger) Buffered() bool       { return false }

type bufferedRequest struct {
	xopbase.Request
	recorder  *xoprecorder.Logger
	finalized bool
	logger    *bufferedLogger
	ctx       context.Context
}

func (logger *bufferedLogger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	recorder := xoprecorder.New()
	return &bufferedRequest{
		recorder: recorder,
		Request:  recorder.Request(ctx, ts, bundle, description, sourceInfo),
		logger:   logger,
		ctx:      ctx,
	}
}

func (request *bufferedRequest) Done(endTime time.Time, final bool) {
	if request.finalized {
		return
	}
	request.Request.Done(endTime, final)
	if !final {
		return
	}
	request.finalized = true
	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSpanLimits(sdktrace.SpanLimits{
			AttributeValueLengthLimit:   -1,
			AttributeCountLimit:         -1,
			EventCountLimit:             -1,
			LinkCountLimit:              -1,
			AttributePerEventCountLimit: -1,
			AttributePerLinkCountLimit:  -1,
		})}
	tpOpts = append(tpOpts, request.logger.tracerProviderOpts...)
	tpOpts = append(tpOpts, IDGenerator())
	// XXX WithResource
	// XXX WithSpanLimits
	tracerProvider := sdktrace.NewTracerProvider(tpOpts...)
	defer tracerProvider.Shutdown(request.ctx)

	var tOpts []oteltrace.TracerOption
	isZOP := true // XXX
	if isZOP {
		tOpts = append(tOpts,
			oteltrace.WithInstrumentationAttributes(
				xopOTELVersion.String(xopotelVersionValue),
				xopVersion.String(xopVersionValue),
			),
			oteltrace.WithInstrumentationVersion(xopotelVersionValue),
		)
	}
	tracer := tracerProvider.Tracer("xopotel", tOpts...)
	otel := &logger{
		id:        "bufotel-" + uuid.New().String(),
		doLogging: true,
		tracer:    tracer,
		recorder:  request.recorder,
	}
	request.recorder.Replay(request.ctx, otel)
	err := tracerProvider.ForceFlush(request.ctx)
	if err != nil {
		fmt.Println("XXX", err)
	}
}

func getStuff(recorder *xoprecorder.Logger, bundle xoptrace.Bundle) (stuff *otelStuff) {
	if recorder == nil {
		return
	}
	_ = recorder.WithLock(func(r *xoprecorder.Logger) error {
		span, ok := r.SpanIndex[bundle.Trace.SpanID().Array()]
		if !ok {
			return nil
		}
		if md := span.SpanMetadata.Get(replayFromOTEL.Key()); md != nil {
			ma, ok := md.Value.(xopbase.ModelArg)
			if ok {
				var otelStuff otelStuff
				err := ma.DecodeTo(&otelStuff)
				if err != nil {
					fmt.Println("XXX could not decode", err)
				} else {
					stuff = &otelStuff
					fmt.Println("XXX decoded")
				}
			} else {
				fmt.Println("XXX cast failed")
			}
		} else {
			fmt.Println("XXX key missing")
		}
		return nil
	})
	return
}

func (o *otelStuff) Options() []oteltrace.SpanStartOption {
	return nil
}

func (o *otelStuff) Set(otelSpan oteltrace.Span) {
	if o == nil {
		return
	}
	otelSpan.SetStatus(o.Status.Code, o.Status.Description)
}
