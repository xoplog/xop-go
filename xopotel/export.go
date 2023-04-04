// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopotel

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xoplog/xop-go/internal/util/version"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ sdktrace.SpanExporter = &spanExporter{}

var ErrShutdown = fmt.Errorf("Shutdown called")

type spanExporter struct {
	base           xopbase.Logger
	orderedFinish  []orderedFinish
	sequenceNumber int32
	done           int32
}

type spanReplay struct {
	*spanExporter
	id2Index map[oteltrace.SpanID]int
	spans    []ReadOnlySpan
	subSpans [][]int
	data     []*datum
}

type datum struct {
	baseSpan             xopbase.Span
	requestIndex         int // index of request ancestor
	attributeDefinitions map[string]*decodeAttributeDefinition
	xopSpan              bool
	registry             *xopat.Registry
}

func (x *spanExporter) addOrdered(seq int32, f func()) {
	x.orderedFinish = append(x.orderedFinish, orderedFinish{
		num: seq,
		f:   f,
	})
}

type orderedFinish struct {
	num int32
	f   func()
}

type baseSpanReplay struct {
	spanReplay
	*datum
	span ReadOnlySpan
}

type decodeAttributeDefinition struct {
	xopat.Make
	AttributeType xopproto.AttributeType `json:"vtype"`
}

// ExportToXOP allows open telementry spans to be exported through
// a xopbase.Logger. If the open telementry spans were generated
// originally using xoputil, then the exported data should almost
// exactly match the original inputs.
func ExportToXOP(base xopbase.Logger) sdktrace.SpanExporter {
	return &spanExporter{base: base}
}

func (e *spanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) (err error) {
	s := make([]ReadOnlySpan, len(spans))
	for i, span := range spans {
		s[i] = span
	}
	return e.exportSpans(ctx, s)
}

func (e *spanExporter) exportSpans(ctx context.Context, spans []ReadOnlySpan) (err error) {
	id2Index := makeIndex(spans)
	subSpans, todo := makeSubspans(id2Index, spans)
	x := spanReplay{
		spanExporter: e,
		id2Index:     id2Index,
		spans:        spans,
		subSpans:     subSpans,
		data:         make([]*datum, len(spans)),
	}
	done := make([]bool, len(spans))
	for len(todo) > 0 {
		i := todo[0]
		if done[i] {
			return errors.Errorf("attempt to re-do span %d", i)
		}
		done[i] = true
		todo = todo[1:]
		x.data[i] = &datum{}
		err := x.Replay(ctx, spans[i], x.data[i], i)
		if err != nil {
			return err
		}
		todo = append(todo, subSpans[i]...)
		if atomic.LoadInt32(&e.done) == 1 {
			return ErrShutdown
		}
	}
	sort.Slice(e.orderedFinish, func(i, j int) bool {
		return e.orderedFinish[i].num < e.orderedFinish[j].num
	})
	for _, o := range x.orderedFinish {
		o.f()
	}
	return nil
}

func (x spanReplay) Replay(ctx context.Context, span ReadOnlySpan, data *datum, myIndex int) error {
	var bundle xoptrace.Bundle
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		bundle.Trace.TraceID().SetArray(spanContext.TraceID())
	}
	if spanContext.HasSpanID() {
		bundle.Trace.SpanID().SetArray(spanContext.SpanID())
	}
	if spanContext.IsSampled() {
		bundle.Trace.Flags().SetArray([1]byte{1})
	}
	if spanContext.TraceState().Len() != 0 {
		bundle.State.SetString(spanContext.TraceState().String())
	}
	parentIndex, hasParent := lookupParent(x.id2Index, span)
	var xopParent *datum
	if hasParent {
		parentContext := x.spans[parentIndex].SpanContext()
		xopParent = x.data[parentIndex]
		if parentContext.HasTraceID() {
			bundle.Parent.TraceID().SetArray(parentContext.TraceID())
			if bundle.Trace.TraceID().IsZero() {
				bundle.Trace.TraceID().Set(bundle.Parent.GetTraceID())
			}
		}
		if parentContext.HasSpanID() {
			bundle.Parent.SpanID().SetArray(parentContext.SpanID())
		}
		if parentContext.IsSampled() {
			bundle.Parent.Flags().SetArray([1]byte{1})
		}
	} else if span.Parent().HasTraceID() {
		bundle.Parent.TraceID().SetArray(span.Parent().TraceID())
		if span.Parent().HasSpanID() {
			bundle.Parent.SpanID().SetArray(span.Parent().SpanID())
		}
		if span.Parent().IsSampled() {
			bundle.Parent.Flags().SetArray([1]byte{1})
		}
	}

	bundle.Parent.Flags().SetBytes([]byte{1})
	bundle.Trace.Flags().SetBytes([]byte{1})
	spanKind := span.SpanKind()
	attributeMap := mapAttributes(span.Attributes())
	if b := attributeMap.GetString(xopBaggage); b != "" {
		bundle.Baggage.SetString(b)
	}
	if spanKind == oteltrace.SpanKindUnspecified {
		spanKind = oteltrace.SpanKind(defaulted(attributeMap.GetInt(otelSpanKind), int64(oteltrace.SpanKindUnspecified)))
	}
	if attributeMap.GetBool(spanIsLinkEventKey) {
		// span is extra just for link
		return nil
	}
	switch spanKind {
	case oteltrace.SpanKindUnspecified, oteltrace.SpanKindInternal:
		if hasParent {
			spanSeq := defaulted(attributeMap.GetString(xopSpanSequence), "")
			data.xopSpan = xopParent.xopSpan
			data.baseSpan = xopParent.baseSpan.Span(ctx, span.StartTime(), bundle, span.Name(), spanSeq)
			data.requestIndex = xopParent.requestIndex
			data.attributeDefinitions = xopParent.attributeDefinitions
			data.registry = xopParent.registry
		} else {
			// This is a difficult sitatuion. We have an internal/unspecified span
			// that does not have a parent present. There is no right answer for what
			// to do. In the Xop world, such a span isn't allowed to exist. We'll treat
			// this span as a request, but mark it as promoted.
			data.xopSpan = attributeMap.GetString(xopVersion) != ""
			data.baseSpan = x.base.Request(ctx, span.StartTime(), bundle, span.Name(), buildSourceInfo(span, attributeMap))
			data.baseSpan.MetadataBool(xopPromotedMetadata, true)
			data.requestIndex = myIndex
			data.attributeDefinitions = make(map[string]*decodeAttributeDefinition)
			data.registry = xopat.NewRegistry(false)
		}
	default:
		data.baseSpan = x.base.Request(ctx, span.StartTime(), bundle, span.Name(), buildSourceInfo(span, attributeMap))
		data.requestIndex = myIndex
		data.attributeDefinitions = make(map[string]*decodeAttributeDefinition)
		data.xopSpan = attributeMap.GetString(xopVersion) != ""
		data.registry = xopat.NewRegistry(false)
	}
	y := baseSpanReplay{
		spanReplay: x,
		span:       span,
		datum:      data,
	}
	for _, attribute := range span.Attributes() {
		err := y.AddSpanAttribute(ctx, attribute)
		if err != nil {
			return err
		}
	}
	var maxNumber int32
	for _, event := range span.Events() {
		lastNumber, err := y.AddEvent(ctx, event)
		if err != nil {
			return err
		}
		if lastNumber > maxNumber {
			maxNumber = lastNumber
		}
	}
	if endTime := span.EndTime(); !endTime.IsZero() {
		x.addOrdered(maxNumber+1, func() {
			data.baseSpan.Done(endTime, true)
		})
	}
	return nil
}

type lineType int

const (
	lineTypeLine lineType = iota
	lineTypeLink
	lineTypeLinkEvent
	lineTypeModel
)

type lineFormat int

const (
	lineFormatDefault lineFormat = iota
	lineFormatTemplate
)

var lineRE = regexp.MustCompile(`^(.+):(\d+)$`)

func (x baseSpanReplay) AddEvent(ctx context.Context, event sdktrace.Event) (int32, error) {
	level, err := xopnum.LevelString(event.Name)
	var nameMessage string
	if err != nil {
		level = xopnum.InfoLevel
		nameMessage = event.Name
	}
	lineType := lineTypeLine
	lineFormat := lineFormatDefault
	var message string
	var template string
	var link xoptrace.Trace
	var modelArg xopbase.ModelArg
	var frames []runtime.Frame
	x.sequenceNumber++
	lineNumber := x.sequenceNumber
	nonSpecial := make([]attribute.KeyValue, 0, len(event.Attributes))
	for _, a := range event.Attributes {
		switch a.Key {
		case xopLineNumber:
			if a.Value.Type() == attribute.INT64 {
				lineNumber = int32(a.Value.AsInt64())
			} else {
				return 0, errors.Errorf("invalid line number attribute type %s", a.Value.Type())
			}
		case logMessageKey:
			if a.Value.Type() == attribute.STRING {
				message = a.Value.AsString()
			} else {
				return 0, errors.Errorf("invalid line message attribute type %s", a.Value.Type())
			}
		case typeKey:
			if a.Value.Type() == attribute.STRING {
				switch a.Value.AsString() {
				case "link":
					lineType = lineTypeLink
				case "link-event":
					lineType = lineTypeLinkEvent
				case "model":
					lineType = lineTypeModel
				case "line":
					// defaulted
				default:
					return 0, errors.Errorf("invalid line type attribute value %s", a.Value.AsString())
				}
			} else {
				return 0, errors.Errorf("invalid line type attribute type %s", a.Value.Type())
			}
		case xopModelType:
			if a.Value.Type() == attribute.STRING {
				modelArg.ModelType = a.Value.AsString()
			} else {
				return 0, errors.Errorf("invalid model type attribute type %s", a.Value.Type())
			}
		case xopEncoding:
			if a.Value.Type() == attribute.STRING {
				e, ok := xopproto.Encoding_value[a.Value.AsString()]
				if !ok {
					return 0, errors.Errorf("invalid model encoding '%s'", a.Value.AsString())
				}
				modelArg.Encoding = xopproto.Encoding(e)
			} else {
				return 0, errors.Errorf("invalid model encoding attribute type %s", a.Value.Type())
			}
		case xopModel:
			if a.Value.Type() == attribute.STRING {
				modelArg.Encoded = []byte(a.Value.AsString())
			} else {
				return 0, errors.Errorf("invalid model encoding attribute type %s", a.Value.Type())
			}
		case xopLineFormat:
			if a.Value.Type() == attribute.STRING {
				switch a.Value.AsString() {
				case "tmpl":
					lineFormat = lineFormatTemplate
				default:
					return 0, errors.Errorf("invalid line format attribute value %s", a.Value.AsString())
				}
			} else {
				return 0, errors.Errorf("invalid line format attribute type %s", a.Value.Type())
			}
		case xopTemplate:
			if a.Value.Type() == attribute.STRING {
				template = a.Value.AsString()
			} else {
				return 0, errors.Errorf("invalid line template attribute type %s", a.Value.Type())
			}
		case xopLinkData:
			if a.Value.Type() == attribute.STRING {
				var ok bool
				link, ok = xoptrace.TraceFromString(a.Value.AsString())
				if !ok {
					return 0, errors.Errorf("invalid link data attribute value %s", a.Value.AsString())
				}
			} else {
				return 0, errors.Errorf("invalid link data attribute type %s", a.Value.Type())
			}
		case xopStackTrace:
			if a.Value.Type() == attribute.STRINGSLICE {
				raw := a.Value.AsStringSlice()
				frames = make([]runtime.Frame, len(raw))
				for i, s := range raw {
					m := lineRE.FindStringSubmatch(s)
					if m == nil {
						return 0, errors.Errorf("could not match stack line '%s'", s)
					}
					frames[i].File = m[1]
					num, _ := strconv.ParseInt(m[2], 10, 64)
					frames[i].Line = int(num)
				}
			} else {
				return 0, errors.Errorf("invalid stack trace attribute type %s", a.Value.Type())
			}
		default:
			nonSpecial = append(nonSpecial, a)
		}
	}
	line := x.baseSpan.NoPrefill().Line(
		level,
		event.Time,
		frames,
	)
	for _, a := range nonSpecial {
		if x.xopSpan {
			err := x.AddXopEventAttribute(ctx, event, a, line)
			if err != nil {
				return 0, errors.Wrapf(err, "add xop event attribute %s", string(a.Key))
			}
		} else {
			err := x.AddEventAttribute(ctx, event, a, line)
			if err != nil {
				return 0, errors.Wrapf(err, "add event attribute %s with type %s", string(a.Key), a.Value.Type())
			}
		}
	}
	if nameMessage != "" {
		if message != "" {
			message = nameMessage + " " + message
		} else {
			message = nameMessage
		}
	}
	switch lineType {
	case lineTypeLine:
		switch lineFormat {
		case lineFormatDefault:
			x.addOrdered(lineNumber, func() {
				line.Msg(message)
			})
		case lineFormatTemplate:
			x.addOrdered(lineNumber, func() {
				line.Template(template)
			})
		default:
			return 0, errors.Errorf("unexpected lineType %d", lineType)
		}
	case lineTypeLink:
		x.addOrdered(lineNumber, func() {
			line.Link(message, link)
		})
	case lineTypeLinkEvent:
		return 0, errors.Errorf("unexpected lineType: link event")
	case lineTypeModel:
		x.addOrdered(lineNumber, func() {
			line.Model(message, modelArg)
		})
	default:
		return 0, errors.Errorf("unexpected lineType %d", lineType)
	}
	return lineNumber, nil
}

func (e *spanExporter) Shutdown(ctx context.Context) error {
	atomic.StoreInt32(&e.done, 1)
	return nil
}

type unhack struct {
	next sdktrace.SpanExporter
}

func makeIndex(spans []ReadOnlySpan) map[oteltrace.SpanID]int {
	id2Index := make(map[oteltrace.SpanID]int)
	for i, span := range spans {
		spanContext := span.SpanContext()
		if spanContext.HasSpanID() {
			id2Index[spanContext.SpanID()] = i
		}
	}
	return id2Index
}

func lookupParent(id2Index map[oteltrace.SpanID]int, span ReadOnlySpan) (int, bool) {
	parent := span.Parent()
	if !parent.HasSpanID() {
		return 0, false
	}
	parentIndex, ok := id2Index[parent.SpanID()]
	if !ok {
		return 0, false
	}
	return parentIndex, true
}

func makeSubspans(id2Index map[oteltrace.SpanID]int, spans []ReadOnlySpan) ([][]int, []int) {
	ss := make([][]int, len(spans))
	noParent := make([]int, 0, len(spans))
	for i, span := range spans {
		parentIndex, ok := lookupParent(id2Index, span)
		if !ok {
			noParent = append(noParent, i)
			continue
		}
		ss[parentIndex] = append(ss[parentIndex], i)
	}
	return ss, noParent
}

func buildSourceInfo(span ReadOnlySpan, attributeMap aMap) xopbase.SourceInfo {
	var si xopbase.SourceInfo
	var source string
	if s := attributeMap.GetString(xopSource); s != "" {
		source = s
	} else if n := span.InstrumentationScope().Name; n != "" {
		if v := span.InstrumentationScope().Version; v != "" {
			source = n + " " + v
		} else {
			source = n
		}
	} else {
		source = "OTEL"
	}
	namespace := defaulted(attributeMap.GetString(xopNamespace), source)
	si.Source, si.SourceVersion = version.SplitVersion(source)
	si.Namespace, si.NamespaceVersion = version.SplitVersion(namespace)
	return si
}

type aMap struct {
	strings map[attribute.Key]string
	ints    map[attribute.Key]int64
	bools   map[attribute.Key]bool
}

func mapAttributes(list []attribute.KeyValue) aMap {
	m := aMap{
		strings: make(map[attribute.Key]string),
		ints:    make(map[attribute.Key]int64),
		bools:   make(map[attribute.Key]bool),
	}
	for _, a := range list {
		switch a.Value.Type() {
		case attribute.STRING:
			m.strings[a.Key] = a.Value.AsString()
		case attribute.INT64:
			m.ints[a.Key] = a.Value.AsInt64()
		case attribute.BOOL:
			m.bools[a.Key] = a.Value.AsBool()
		}
	}
	return m
}

func (m aMap) GetString(k attribute.Key) string { return m.strings[k] }
func (m aMap) GetInt(k attribute.Key) int64     { return m.ints[k] }
func (m aMap) GetBool(k attribute.Key) bool     { return m.bools[k] }

func defaulted[T comparable](a, b T) T {
	var zero T
	if a == zero {
		return b
	}
	return a
}

func (x baseSpanReplay) AddXopEventAttribute(ctx context.Context, event sdktrace.Event, a attribute.KeyValue, line xopbase.Line) error {
	key := string(a.Key)
	switch a.Value.Type() {
	case attribute.STRINGSLICE:
		slice := a.Value.AsStringSlice()
		if len(slice) < 2 {
			return errors.Errorf("invalid xop attribute encoding slice is too short")
		}
		switch slice[1] {
		case "any":
			if len(slice) != 4 {
				return errors.Errorf("key %s invalid any encoding, slice too short", key)
			}
			var ma xopbase.ModelArg
			ma.Encoded = []byte(slice[0])
			e, ok := xopproto.Encoding_value[slice[2]]
			if !ok {
				return errors.Errorf("invalid model encoding '%s'", a.Value.AsString())
			}
			ma.Encoding = xopproto.Encoding(e)
			ma.ModelType = slice[3]
			line.Any(key, ma)
		case "bool":
		case "dur":
			dur, err := time.ParseDuration(slice[0])
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Duration(key, dur)
		case "enum":
			if len(slice) != 3 {
				return errors.Errorf("key %s invalid enum encoding, slice too short", key)
			}
			ea, err := x.registry.ConstructEnumAttribute(xopat.Make{
				Key: key,
			}, xopat.AttributeTypeEnum)
			if err != nil {
				return errors.Errorf("could not turn key %s into an enum", key)
			}
			i, err := strconv.ParseInt(slice[2], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "could not turn key %s into an enum", key)
			}
			enum := ea.Add64(i, slice[0])
			line.Enum(&ea.EnumAttribute, enum)
		case "error":
			line.String(key, slice[0], xopbase.StringToDataType["error"])
		case "f32":
			f, err := strconv.ParseFloat(slice[0], 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Float64(key, f, xopbase.StringToDataType["f32"])
		case "f64":
			f, err := strconv.ParseFloat(slice[0], 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Float64(key, f, xopbase.StringToDataType["f64"])
		case "i":
			i, err := strconv.ParseInt(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Int64(key, i, xopbase.StringToDataType["i"])
		case "i16":
			i, err := strconv.ParseInt(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Int64(key, i, xopbase.StringToDataType["i16"])
		case "i32":
			i, err := strconv.ParseInt(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Int64(key, i, xopbase.StringToDataType["i32"])
		case "i64":
			i, err := strconv.ParseInt(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Int64(key, i, xopbase.StringToDataType["i64"])
		case "i8":
			i, err := strconv.ParseInt(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Int64(key, i, xopbase.StringToDataType["i8"])
		case "s":
			line.String(key, slice[0], xopbase.StringToDataType["s"])
		case "stringer":
			line.String(key, slice[0], xopbase.StringToDataType["stringer"])
		case "time":
			ts, err := time.Parse(time.RFC3339Nano, slice[0])
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Time(key, ts)
		case "u":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["u"])
		case "u16":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["u16"])
		case "u32":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["u32"])
		case "u64":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["u64"])
		case "u8":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["u8"])
		case "uintptr":
			i, err := strconv.ParseUint(slice[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "key %s invalid %s", key, slice[1])
			}
			line.Uint64(key, i, xopbase.StringToDataType["uintptr"])

		}

	case attribute.BOOL:
		line.Bool(key, a.Value.AsBool())
	default:
		return errors.Errorf("unexpected event attribute type %s for xop-encoded line", a.Value.Type())
	}
	return nil
}

func (x baseSpanReplay) AddEventAttribute(ctx context.Context, event sdktrace.Event, a attribute.KeyValue, line xopbase.Line) error {
	switch a.Value.Type() {
	case attribute.BOOL:
		line.Bool(string(a.Key), a.Value.AsBool())
	case attribute.BOOLSLICE:
		var ma xopbase.ModelArg
		ma.Model = a.Value.AsBoolSlice()
		ma.ModelType = toTypeSliceName["BOOL"]
		line.Any(string(a.Key), ma)
	case attribute.FLOAT64:
		line.Float64(string(a.Key), a.Value.AsFloat64(), xopbase.Float64DataType)
	case attribute.FLOAT64SLICE:
		var ma xopbase.ModelArg
		ma.Model = a.Value.AsFloat64Slice()
		ma.ModelType = toTypeSliceName["FLOAT64"]
		line.Any(string(a.Key), ma)
	case attribute.INT64:
		line.Int64(string(a.Key), a.Value.AsInt64(), xopbase.Int64DataType)
	case attribute.INT64SLICE:
		var ma xopbase.ModelArg
		ma.Model = a.Value.AsInt64Slice()
		ma.ModelType = toTypeSliceName["INT64"]
		line.Any(string(a.Key), ma)
	case attribute.STRING:
		line.String(string(a.Key), a.Value.AsString(), xopbase.StringDataType)
	case attribute.STRINGSLICE:
		var ma xopbase.ModelArg
		ma.Model = a.Value.AsStringSlice()
		ma.ModelType = toTypeSliceName["STRING"]
		line.Any(string(a.Key), ma)

	case attribute.INVALID:
		fallthrough
	default:
		return errors.Errorf("invalid type")
	}
	return nil
}

var toTypeSliceName = map[string]string{
	"BOOL":    "[]bool",
	"STRING":  "[]string",
	"INT64":   "[]int64",
	"FLOAT64": "[]float64",
}

func (x baseSpanReplay) AddSpanAttribute(ctx context.Context, a attribute.KeyValue) (err error) {
	switch a.Key {
	case spanIsLinkAttributeKey,
		spanIsLinkEventKey,
		xopVersion,
		xopOTELVersion,
		xopSource,
		xopNamespace,
		xopBaggage,
		xopSpanSequence,
		typeKey,
		otelSpanKind:
		// special cases handled elsewhere
		return nil
	}
	key := string(a.Key)
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "add span attribute %s with type %s", key, a.Value.Type())
		}
	}()
	if strings.HasPrefix(key, attributeDefinitionPrefix) {
		key := strings.TrimPrefix(key, attributeDefinitionPrefix)
		if _, ok := x.data[x.requestIndex].attributeDefinitions[key]; ok {
			return nil
		}
		if a.Value.Type() != attribute.STRING {
			return errors.Errorf("expected type to be string")
		}
		var aDef decodeAttributeDefinition
		err := json.Unmarshal([]byte(a.Value.AsString()), &aDef)
		if err != nil {
			return errors.Wrapf(err, "could not unmarshal attribute defintion")
		}
		x.data[x.requestIndex].attributeDefinitions[key] = &aDef
		return nil
	}

	if aDef, ok := x.data[x.requestIndex].attributeDefinitions[key]; ok {
		return x.AddXopMetadataAttribute(ctx, a, aDef)
	}
	if x.xopSpan {
		return errors.Errorf("missing attribute defintion for key %s in xop span", key)
	}

	mkMake := func(key string, multiple bool) xopat.Make {
		return xopat.Make{
			Description: xopSynthesizedForOTEL,
			Key:         key,
			Multiple:    multiple,
		}
	}
	switch a.Value.Type() {
	case attribute.BOOL:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(mkMake(key, false), xopat.AttributeTypeBool)
		if err != nil {
			return err
		}
		x.baseSpan.MetadataBool(registeredAttribute, a.Value.AsBool())
	case attribute.BOOLSLICE:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(mkMake(key, true), xopat.AttributeTypeBool)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsBoolSlice() {
			x.baseSpan.MetadataBool(registeredAttribute, v)
		}
	case attribute.FLOAT64:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(mkMake(key, false), xopat.AttributeTypeFloat64)
		if err != nil {
			return err
		}
		x.baseSpan.MetadataFloat64(registeredAttribute, a.Value.AsFloat64())
	case attribute.FLOAT64SLICE:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(mkMake(key, true), xopat.AttributeTypeFloat64)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsFloat64Slice() {
			x.baseSpan.MetadataFloat64(registeredAttribute, v)
		}
	case attribute.INT64:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(mkMake(key, false), xopat.AttributeTypeInt64)
		if err != nil {
			return err
		}
		x.baseSpan.MetadataInt64(registeredAttribute, a.Value.AsInt64())
	case attribute.INT64SLICE:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(mkMake(key, true), xopat.AttributeTypeInt64)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsInt64Slice() {
			x.baseSpan.MetadataInt64(registeredAttribute, v)
		}
	case attribute.STRING:
		registeredAttribute, err := x.registry.ConstructStringAttribute(mkMake(key, false), xopat.AttributeTypeString)
		if err != nil {
			return err
		}
		x.baseSpan.MetadataString(registeredAttribute, a.Value.AsString())
	case attribute.STRINGSLICE:
		registeredAttribute, err := x.registry.ConstructStringAttribute(mkMake(key, true), xopat.AttributeTypeString)
		if err != nil {
			return err
		}
		for _, v := range a.Value.AsStringSlice() {
			x.baseSpan.MetadataString(registeredAttribute, v)
		}

	case attribute.INVALID:
		fallthrough
	default:
		return errors.Errorf("span attribute key (%s) has value type (%s) that is not expected", key, a.Value.Type())
	}
	return nil
}

func (x baseSpanReplay) AddXopMetadataAttribute(ctx context.Context, a attribute.KeyValue, aDef *decodeAttributeDefinition) error {
	switch aDef.AttributeType {
	case xopproto.AttributeType_Any:
		registeredAttribute, err := x.registry.ConstructAnyAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			var ma xopbase.ModelArg
			err := ma.UnmarshalJSON([]byte(v))
			if err != nil {
				return err
			}
			x.baseSpan.MetadataAny(registeredAttribute, ma)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Bool:
		registeredAttribute, err := x.registry.ConstructBoolAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.BOOL, attribute.BOOLSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v bool) error {
			x.baseSpan.MetadataBool(registeredAttribute, v)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsBoolSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsBool()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Duration:
		registeredAttribute, err := x.registry.ConstructDurationAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			d, err := time.ParseDuration(v)
			if err != nil {
				return err
			}
			x.baseSpan.MetadataInt64(&registeredAttribute.Int64Attribute, int64(d))
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Enum:
		registeredAttribute, err := x.registry.ConstructEnumAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			i := strings.LastIndexByte(v, '/')
			if i == -1 {
				return errors.Errorf("invalid enum %s", v)
			}
			if i == len(v)-1 {
				return errors.Errorf("invalid enum %s", v)
			}
			vi, err := strconv.ParseInt(v[i+1:], 10, 64)
			if err != nil {
				return errors.Wrap(err, "invalid enum")
			}
			enum := registeredAttribute.Add64(vi, v[:i])
			x.baseSpan.MetadataEnum(&registeredAttribute.EnumAttribute, enum)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Float64:
		registeredAttribute, err := x.registry.ConstructFloat64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.FLOAT64, attribute.FLOAT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v float64) error {
			x.baseSpan.MetadataFloat64(registeredAttribute, v)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsFloat64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsFloat64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Int:
		registeredAttribute, err := x.registry.ConstructIntAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v int64) error {
			x.baseSpan.MetadataInt64(&registeredAttribute.Int64Attribute, int64(v))
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsInt64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Int16:
		registeredAttribute, err := x.registry.ConstructInt16Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v int64) error {
			x.baseSpan.MetadataInt64(&registeredAttribute.Int64Attribute, int64(v))
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsInt64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Int32:
		registeredAttribute, err := x.registry.ConstructInt32Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v int64) error {
			x.baseSpan.MetadataInt64(&registeredAttribute.Int64Attribute, int64(v))
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsInt64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Int64:
		registeredAttribute, err := x.registry.ConstructInt64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v int64) error {
			x.baseSpan.MetadataInt64(registeredAttribute, v)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsInt64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Int8:
		registeredAttribute, err := x.registry.ConstructInt8Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.INT64, attribute.INT64SLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v int64) error {
			x.baseSpan.MetadataInt64(&registeredAttribute.Int64Attribute, int64(v))
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsInt64Slice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsInt64()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Link:
		registeredAttribute, err := x.registry.ConstructLinkAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			t, ok := xoptrace.TraceFromString(v)
			if !ok {
				return errors.Errorf("invalid trace string %s", v)
			}
			x.baseSpan.MetadataLink(registeredAttribute, t)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_String:
		registeredAttribute, err := x.registry.ConstructStringAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			x.baseSpan.MetadataString(registeredAttribute, v)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}
	case xopproto.AttributeType_Time:
		registeredAttribute, err := x.registry.ConstructTimeAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		expectedSingleType, expectedMultiType := attribute.STRING, attribute.STRINGSLICE
		expectedType := expectedSingleType
		if registeredAttribute.Multiple() {
			expectedType = expectedMultiType
		}
		if a.Value.Type() != expectedType {
			return errors.Errorf("expected type %s", expectedMultiType)
		}
		setter := func(v string) error {
			t, err := time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return err
			}
			x.baseSpan.MetadataTime(registeredAttribute, t)
			return nil
		}
		if registeredAttribute.Multiple() {
			values := a.Value.AsStringSlice()
			for _, value := range values {
				err := setter(value)
				if err != nil {
					return err
				}
			}
		} else {
			value := a.Value.AsString()
			err := setter(value)
			if err != nil {
				return err
			}
		}

	default:
		return errors.Errorf("unexpected attribute type %s", aDef.AttributeType)
	}
	return nil
}
