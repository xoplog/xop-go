// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
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
	bundle    xoptrace.Bundle
}

func (logger *bufferedLogger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	recorder := xoprecorder.New()
	return &bufferedRequest{
		recorder: recorder,
		Request:  recorder.Request(ctx, ts, bundle, description, sourceInfo),
		logger:   logger,
		ctx:      ctx,
		bundle:   bundle,
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
		}),
	}
	tpOpts = append(tpOpts, request.logger.tracerProviderOpts...)
	tpOpts = append(tpOpts, IDGenerator())

	otelStuff := getStuff(request.recorder, request.bundle)
	tpOpts = append(tpOpts, otelStuff.TracerProviderOptions()...)

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
				fmt.Println("XXX otelStuff.Encoded", string(ma.Encoded))
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

type bufferedResource struct {
	*resource.Resource
}

var _ json.Unmarshaler = &bufferedResource{}

func (r *bufferedResource) UnmarshalJSON(b []byte) error {
	fmt.Println("XXX unmarshal resoruce", string(b))
	var bufferedAttributes bufferedAttributes
	err := json.Unmarshal(b, &bufferedAttributes)
	if err != nil {
		return err
	}
	fmt.Println("XXX attributes", len(bufferedAttributes.attributes), bufferedAttributes.attributes)
	r.Resource = resource.NewWithAttributes("", bufferedAttributes.attributes...)
	fmt.Println("XXX resource now", r.Resource)
	return nil
}

func (o *otelStuff) Options() []oteltrace.SpanStartOption {
	if o == nil {
		return nil
	}
	return []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(oteltrace.SpanKind(o.SpanKind)),
	}
}

func (o *otelStuff) Set(otelSpan oteltrace.Span) {
	if o == nil {
		return
	}
	otelSpan.SetStatus(o.Status.Code, o.Status.Description)
}

func (o *otelStuff) TracerProviderOptions() []sdktrace.TracerProviderOption {
	fmt.Println("XXX Resource=", o.Resource.Resource)
	return []sdktrace.TracerProviderOption{
		sdktrace.WithResource(o.Resource.Resource),
	}
}

// {"Key":"environment","Value":{"Type":"STRING","Value":"demo"}

type bufferedAttributes struct {
	attributes []attribute.KeyValue
}

var _ json.Unmarshaler = &bufferedAttributes{}

func (a *bufferedAttributes) UnmarshalJSON(b []byte) error {
	var standIn []bufferedKeyValue
	err := json.Unmarshal(b, &standIn)
	if err != nil {
		return err
	}
	a.attributes = make([]attribute.KeyValue, len(standIn))
	for i, si := range standIn {
		a.attributes[i] = si.KeyValue
	}
	return nil
}

type bufferedKeyValue struct {
	attribute.KeyValue
}

var _ json.Unmarshaler = &bufferedKeyValue{}

func (a *bufferedKeyValue) UnmarshalJSON(b []byte) error {
	var standIn struct {
		Key   string
		Value struct {
			Type  string
			Value any
		}
	}
	err := json.Unmarshal(b, &standIn)
	if err != nil {
		return err
	}
	switch standIn.Value.Type {
	case "BOOL":
		if c, ok := standIn.Value.Value.(bool); ok {
			a.KeyValue = attribute.Bool(standIn.Key, c)
		} else {
			var si2 struct {
				Value struct {
					Value bool
				}
			}
			err := json.Unmarshal(b, &si2)
			if err != nil {
				return err
			}
			a.KeyValue = attribute.Bool(standIn.Key, si2.Value.Value)
		}
	case "BOOLSLICE":
		var si2 struct {
			Value struct {
				Value []bool
			}
		}
		err := json.Unmarshal(b, &si2)
		if err != nil {
			return err
		}
		a.KeyValue = attribute.BoolSlice(standIn.Key, si2.Value.Value)
	// blank line required here
	case "FLOAT64":
		if c, ok := standIn.Value.Value.(float64); ok {
			a.KeyValue = attribute.Float64(standIn.Key, c)
		} else {
			var si2 struct {
				Value struct {
					Value float64
				}
			}
			err := json.Unmarshal(b, &si2)
			if err != nil {
				return err
			}
			a.KeyValue = attribute.Float64(standIn.Key, si2.Value.Value)
		}
	case "FLOAT64SLICE":
		var si2 struct {
			Value struct {
				Value []float64
			}
		}
		err := json.Unmarshal(b, &si2)
		if err != nil {
			return err
		}
		a.KeyValue = attribute.Float64Slice(standIn.Key, si2.Value.Value)
	// blank line required here
	case "INT64":
		if c, ok := standIn.Value.Value.(int64); ok {
			a.KeyValue = attribute.Int64(standIn.Key, c)
		} else {
			var si2 struct {
				Value struct {
					Value int64
				}
			}
			err := json.Unmarshal(b, &si2)
			if err != nil {
				return err
			}
			a.KeyValue = attribute.Int64(standIn.Key, si2.Value.Value)
		}
	case "INT64SLICE":
		var si2 struct {
			Value struct {
				Value []int64
			}
		}
		err := json.Unmarshal(b, &si2)
		if err != nil {
			return err
		}
		a.KeyValue = attribute.Int64Slice(standIn.Key, si2.Value.Value)
	// blank line required here
	case "STRING":
		if c, ok := standIn.Value.Value.(string); ok {
			a.KeyValue = attribute.String(standIn.Key, c)
		} else {
			var si2 struct {
				Value struct {
					Value string
				}
			}
			err := json.Unmarshal(b, &si2)
			if err != nil {
				return err
			}
			a.KeyValue = attribute.String(standIn.Key, si2.Value.Value)
		}
	case "STRINGSLICE":
		var si2 struct {
			Value struct {
				Value []string
			}
		}
		err := json.Unmarshal(b, &si2)
		if err != nil {
			return err
		}
		a.KeyValue = attribute.StringSlice(standIn.Key, si2.Value.Value)
	// blank line required here

	default:
		return fmt.Errorf("unknown attribute.KeyValue type '%s'", standIn.Value.Type)
	}
	return nil
}
