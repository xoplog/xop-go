// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/muir/gwrap"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopbase/xopbaseutil"
	"github.com/xoplog/xop-go/xoprecorder"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil/pointer"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type BufferedReplayExporterWrapper interface {
	// WrapExporter augments the data that the wrapped exporter recevies so that if
	// the ReadOnlySpan came from a BufferedReplayLogger that is replaying data that
	// originally came from OTEL, then it can fill data that otherwise has no way
	// to be propagated through the OTEL APIs.  For example, instrumentation.Scope.Name.
	WrapExporter(sdktrace.SpanExporter) sdktrace.SpanExporter

	// BufferedReplayLogger creates a Logger that can be used when replaying from other
	// xopbase.Logger implementations into xopotel. It buffers all the logged data until
	// Done() is called on a per-request basis. Additional logging after Done() is discarded.
	//
	// A TracerProvider and Tracer are constructed for each Request and discarded afterwards.
	//
	// VERY IMPORTANT: every exporter wrapped with WrapExporter must be
	// passed to BufferedReplayLogger. If not, memory will leak.
	//
	// Also import: if the exporter's ExportSpans() isn't called with all
	// spans, memory will leak.  The amount of leaked memory is not large, maybe
	// 100 bytes per span, but it isn't zero.
	BufferedReplayLogger(...sdktrace.TracerProviderOption) xopbase.Logger
}

// BufferedReplayLogger creates a Logger that can be used when replaying from other
// xopbase.Logger implementations into xopotel. It buffers all the logged data until
// Done() is called on a per-request basis. Additional logging after Done() is discarded.
//
// A TracerProvider and Tracer are constructed for each Request and discarded afterwards.
//
// For improved fideltity of OTEL -> XOP -> OTEL replay, use
// BufferedReplayExporterWrapper.BufferedReplayLogger instead.
func BufferedReplayLogger(tracerProviderOpts ...sdktrace.TracerProviderOption) xopbase.Logger {
	return bufferedReplayLogger(&exporterWrapper{}, tracerProviderOpts)
}

func bufferedReplayLogger(exporterWrapper *exporterWrapper, tracerProviderOpts []sdktrace.TracerProviderOption) xopbase.Logger {
	return &bufferedLogger{
		tracerProviderOpts: tracerProviderOpts,
		id:                 "otelbuf-" + uuid.New().String(),
		exporterWrapper:    exporterWrapper,
	}
}

type bufferedLogger struct {
	id                 string
	tracerProviderOpts []sdktrace.TracerProviderOption
	exporterWrapper    *exporterWrapper
}

func (logger *bufferedLogger) ID() string           { return logger.id }
func (logger *bufferedLogger) ReferencesKept() bool { return true }
func (logger *bufferedLogger) Buffered() bool       { return false }

// bufferedRequest implements Request for a buferedLogger.  The bufferedLogger is a wrapper
// wrapper around request{}. It waits until the request is complete before it invokes
// logger.Request and in the meantime buffers all data into an ephemeral xoprecorder.Logger.
//
// It uses xoprecorder's replay functionality to dump the buffered logs.
//
// It creates a logger{} for each request data can be passed directly from the bufferedRequest
// to the logger{} and from the logger{} to the request{} because there is only one request{}
// per logger{} in this situation.
type bufferedRequest struct {
	xopbase.Request
	recorder      *xoprecorder.Logger
	finalized     bool
	logger        *bufferedLogger
	ctx           context.Context
	bundle        xoptrace.Bundle
	errorReporter func(error)
	isXOP         bool // request originally came from XOP
}

func (logger *bufferedLogger) Request(ctx context.Context, ts time.Time, bundle xoptrace.Bundle, description string, sourceInfo xopbase.SourceInfo) xopbase.Request {
	recorder := xoprecorder.New()
	return &bufferedRequest{
		recorder: recorder,
		Request:  recorder.Request(ctx, ts, bundle, description, sourceInfo),
		logger:   logger,
		ctx:      ctx,
		bundle:   bundle,
		isXOP:    sourceInfo.Source != otelDataSource,
	}
}

func (request *bufferedRequest) SetErrorReporter(f func(error)) { request.errorReporter = f }

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

	otelStuff := request.getStuff(request.bundle, false)
	// we do not call augment() here because that would result in
	// a duplicate call when it is also called from Reqeust()
	tpOpts = append(tpOpts, otelStuff.TracerProviderOptions()...)

	tracerProvider := sdktrace.NewTracerProvider(tpOpts...)
	defer tracerProvider.Shutdown(request.ctx)

	var tOpts []oteltrace.TracerOption
	if request.isXOP {
		tOpts = append(tOpts,
			oteltrace.WithInstrumentationAttributes(
				xopOTELVersion.String(xopotelVersionValue),
				xopVersion.String(xopVersionValue),
			),
			oteltrace.WithInstrumentationVersion(xopotelVersionValue),
		)
	}
	tOpts = append(tOpts, otelStuff.TracerOptions()...)
	tracer := tracerProvider.Tracer("xopotel", tOpts...)
	otel := &logger{
		id:              "bufotel-" + uuid.New().String(),
		doLogging:       true,
		tracer:          tracer,
		bufferedRequest: request,
	}
	request.recorder.Replay(request.ctx, otel)
	err := tracerProvider.ForceFlush(request.ctx)
	if err != nil && request.errorReporter != nil {
		request.errorReporter(err)
	}
}

func (req *bufferedRequest) getStuff(bundle xoptrace.Bundle, augment bool) (stuff *otelStuff) {
	if req == nil || req.recorder == nil {
		return
	}
	spanID := bundle.Trace.SpanID().Array()
	_ = req.recorder.WithLock(func(r *xoprecorder.Logger) error {
		span, ok := r.SpanIndex[spanID]
		if !ok {
			return nil
		}
		var otelStuff otelStuff
		stuff = &otelStuff

		addMetadataLink := func(key string, link xoptrace.Trace) {
			otelStuff.links = append(otelStuff.links, oteltrace.Link{
				SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
					TraceID:    link.TraceID().Array(),
					SpanID:     link.SpanID().Array(),
					TraceFlags: oteltrace.TraceFlags(link.Flags().Array()[0]),
					TraceState: emptyTraceState,
					Remote:     true,
				}),
				Attributes: []attribute.KeyValue{
					xopLinkMetadataKey.String(key),
				},
			})
		}
		span.SpanMetadata.Map.Range(func(key string, mt *xopbaseutil.MetadataTracker) bool {
			switch mt.Type {
			case xopbase.LinkDataType:
				addMetadataLink(key, mt.Value.(xoptrace.Trace))
			case xopbase.LinkArrayDataType:
				for _, link := range mt.Value.([]interface{}) {
					addMetadataLink(key, link.(xoptrace.Trace))
				}
			}
			return true
		})

		var missingData missingSpanData

		if len(span.Links) > 0 {
			for _, linkLine := range span.Links {
				plainBuilder := builder{
					attributes: make([]attribute.KeyValue, 0, len(linkLine.Data)),
				}
				linkLine = pointer.To(linkLine.Copy())

				var ts oteltrace.TraceState

				tsRaw, ok := extractValue[string](linkLine, xopOTELLinkTranceState, xopbase.StringDataType, xopLinkTraceStateError)
				if ok {
					var err error
					ts, err = oteltrace.ParseTraceState(tsRaw)
					if err != nil {
						linkLine.Data[xopLinkTraceStateError] = err.Error()
						linkLine.DataType[xopLinkTraceStateError] = xopbase.ErrorDataType
					}
				}

				isRemote, _ := extractValue[bool](linkLine, xopOTELLinkIsRemote, xopbase.BoolDataType, xopLinkRemoteError)

				linkSpanID := linkLine.AsLink.SpanID().Array()

				if augment {
					droppedCount, _ := extractValue[int64](linkLine, xopOTELLinkDroppedAttributeCount, xopbase.IntDataType, xopLinkeDroppedError)
					if droppedCount != 0 {
						if missingData.droppedLinkAttributeCount == nil {
							missingData.droppedLinkAttributeCount = make(map[[8]byte]int)
						}
						missingData.droppedLinkAttributeCount[linkSpanID] = int(droppedCount)
					}
				}

				xoprecorder.ReplayLineData(linkLine, &plainBuilder)

				otelStuff.links = append(otelStuff.links, oteltrace.Link{
					SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
						TraceID:    linkLine.AsLink.TraceID().Array(),
						SpanID:     linkSpanID,
						TraceFlags: oteltrace.TraceFlags(linkLine.AsLink.Flags().Array()[0]),
						TraceState: ts,
						Remote:     isRemote,
					}),
					Attributes: plainBuilder.attributes,
				})
			}
		}

		if failure := func() string {
			md := span.SpanMetadata.Get(otelReplayStuff.Key().String())
			if md == nil {
				return "span metdata missing replay key expected for BufferedReplayLogger " + otelReplayStuff.Key().String()
			}
			ma, ok := md.Value.(xopbase.ModelArg)
			if !ok {
				return fmt.Sprintf("cast of %s data to ModelArg failed, is %T", otelReplayStuff.Key(), md.Value)
			}
			err := ma.DecodeTo(&otelStuff)
			if err != nil {
				return fmt.Sprintf("failed to decode encoded data in %s: %s", otelReplayStuff.Key(), err)
			}
			return ""
		}(); failure != "" {
			missingData.error = failure
		} else {
			missingData.spanCounters = otelStuff.spanCounters
			missingData.scopeName = otelStuff.InstrumentationScope.Name
		}
		var zeroMissing missingSpanDataComparable
		if augment && atomic.LoadInt32(&req.logger.exporterWrapper.exporterCount) > 0 &&
			(missingData.missingSpanDataComparable != zeroMissing || missingData.droppedLinkAttributeCount != nil) {
			req.logger.exporterWrapper.augmentMap.Store(spanID, &missingData)
		}
		return nil
	})
	return
}

func extractValue[T any](linkLine *xoprecorder.Line, name xopat.K, dataType xopbase.DataType, errorKey xopat.K) (T, bool) {
	var zero T
	if raw, ok := linkLine.Data[name]; ok {
		if linkLine.DataType[name] == dataType {
			cooked, ok := raw.(T)
			if ok {
				delete(linkLine.Data, name)
				delete(linkLine.DataType, name)
				return cooked, true
			} else {
				linkLine.Data[errorKey] = fmt.Sprintf("%s should be a %T, but is a %T",
					name, zero, raw)
				linkLine.DataType[errorKey] = xopbase.ErrorDataType
			}
		} else {
			linkLine.Data[errorKey] = fmt.Sprintf("%s should be a %s, but is a %s",
				name, dataType, linkLine.DataType[name])
			linkLine.DataType[errorKey] = xopbase.ErrorDataType
		}
	}
	return zero, false
}

// NewBufferedReplayExportWrapper creates an export wrapper that can be used to
// improve the accuracy of data exported after replay. This comes into play
// when data is run from OTEL to XOP and then back to OTEL. When it comes back
// to OTEL, the export wrapper improves what a SpanExporter exports.
//
// Use the BufferedReplayExporterWrapper to both wrap SpanExporter and to
// create the BufferedReplayLogger that is used to export from XOP to OTEL.
func NewBufferedReplayExporterWrapper() BufferedReplayExporterWrapper {
	return &exporterWrapper{}
}

type exporterWrapper struct {
	exporterCount int32
	augmentMap    gwrap.SyncMap[[8]byte, *missingSpanData]
}

func (ew *exporterWrapper) WrapExporter(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	_ = atomic.AddInt32(&ew.exporterCount, 1)
	return &wrappedExporter{
		exporter: exporter,
		wrapper:  ew,
	}
}

func (ew *exporterWrapper) BufferedReplayLogger(tracerProviderOpts ...sdktrace.TracerProviderOption) xopbase.Logger {
	return bufferedReplayLogger(ew, tracerProviderOpts)
}

type wrappedExporter struct {
	wrapper  *exporterWrapper
	exporter sdktrace.SpanExporter
	shutdown int32
}

func (w *wrappedExporter) Shutdown(ctx context.Context) error {
	if atomic.AddInt32(&w.shutdown, 1) == 1 {
		atomic.AddInt32(&w.wrapper.exporterCount, -1)
	}
	return w.exporter.Shutdown(ctx)
}

func (w wrappedExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	n := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		missingSpanData, ok := w.wrapper.augmentMap.Load([8]byte(span.SpanContext().SpanID()))
		if ok {
			if atomic.AddInt32(&missingSpanData.consumptionCount, 1) >=
				atomic.LoadInt32(&w.wrapper.exporterCount) {
				w.wrapper.augmentMap.Delete([8]byte(span.SpanContext().SpanID()))
			}
			n[i] = wrappedSpan{
				ReadOnlySpan:    span,
				missingSpanData: missingSpanData,
			}
		} else {
			n[i] = span
		}
	}
	return w.exporter.ExportSpans(ctx, n)
}

// We delete the missingSpanData from the augmentMap when consumptionCount
// equals or exceeds the number of wrapped exporters.
type missingSpanData struct {
	missingSpanDataComparable
	droppedLinkAttributeCount map[[8]byte]int
}

type missingSpanDataComparable struct {
	spanCounters
	consumptionCount int32
	scopeName        string
	error            string // for when then round-trip fails
}

type wrappedSpan struct {
	sdktrace.ReadOnlySpan
	missingSpanData *missingSpanData
}

func (s wrappedSpan) InstrumentationScope() instrumentation.Scope {
	scope := s.ReadOnlySpan.InstrumentationScope()
	scope.Name = s.missingSpanData.scopeName
	return scope
}

func (s wrappedSpan) InstrumentationLibrary() instrumentation.Library {
	library := s.ReadOnlySpan.InstrumentationLibrary()
	library.Name = s.missingSpanData.scopeName
	return library
}

func (s wrappedSpan) DroppedAttributes() int { return s.missingSpanData.DroppedAttributes }
func (s wrappedSpan) DroppedLinks() int      { return s.missingSpanData.DroppedLinks }
func (s wrappedSpan) DroppedEvents() int     { return s.missingSpanData.DroppedEvents }
func (s wrappedSpan) ChildSpanCount() int    { return s.missingSpanData.ChildSpanCount }

func (s wrappedSpan) Links() []sdktrace.Link {
	links := s.ReadOnlySpan.Links()
	if s.missingSpanData != nil {
		for i, link := range links {
			if dropped, ok := s.missingSpanData.droppedLinkAttributeCount[([8]byte)(link.SpanContext.SpanID())]; ok {
				links[i].DroppedAttributeCount = dropped
			}
		}
	}
	return links
}

type bufferedResource struct {
	*resource.Resource
}

var _ json.Unmarshaler = &bufferedResource{}

func (r *bufferedResource) UnmarshalJSON(b []byte) error {
	var bufferedAttributes bufferedAttributes
	err := json.Unmarshal(b, &bufferedAttributes)
	if err != nil {
		return err
	}
	r.Resource = resource.NewWithAttributes("", bufferedAttributes.attributes...)
	return nil
}

func (o *otelStuff) TracerOptions() []oteltrace.TracerOption {
	if o == nil {
		return nil
	}
	return []oteltrace.TracerOption{
		oteltrace.WithSchemaURL(o.InstrumentationScope.SchemaURL),
		oteltrace.WithInstrumentationVersion(o.InstrumentationScope.Version),
	}
}

func (o *otelStuff) SpanOptions() []oteltrace.SpanStartOption {
	if o == nil {
		return nil
	}
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(oteltrace.SpanKind(o.SpanKind)),
	}
	if len(o.links) != 0 {
		opts = append(opts, oteltrace.WithLinks(o.links...))
	}
	return opts
}

func (o *otelStuff) Set(otelSpan oteltrace.Span) {
	if o == nil {
		return
	}
	otelSpan.SetStatus(o.Status.Code, o.Status.Description)
}

func (o *otelStuff) TracerProviderOptions() []sdktrace.TracerProviderOption {
	return []sdktrace.TracerProviderOption{
		sdktrace.WithResource(o.Resource.Resource),
	}
}

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
