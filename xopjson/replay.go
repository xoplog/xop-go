// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xoplog/xop-go/internal/util/version"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/pkg/errors"
)

var knownKeys = []string{
	"ts",
	"attributes",
	"span.id",
	"trace.id",
	"span.seq",
	"lvl",
	"stack",
	"msg",
	"fmt",
	"impl",
	"parent.id",
	"trace.state",
	"trace.baggage",
	"ns",
	"source",
	"dur",
	"type",
	"span.name",
	"span.ver",
}

type decodeAll struct {
	decodeCommon
	decodeLineExclusive
	decodeSpanShared
	decodeSpanExclusive
	decodeRequestExclusive
}

type decodeCommon struct {
	// lines, spans, and requests
	Timestamp    xoptestutil.TS             `json:"ts"`
	Attributes   map[string]json.RawMessage `json:"attributes"`
	SpanID       string                     `json:"span.id"`
	TraceID      string                     `json:"trace.id"` // optional for spans and lines, required for traces
	SequenceCode string                     `json:"span.seq"` // optional
	unparsed     string                     `json:"-"`
}

type decodedLine struct {
	*decodeCommon
	*decodeLineExclusive
}

type decodeLineExclusive struct {
	Level  xopnum.Level `json:"lvl"`
	Stack  []string     `json:"stack"`
	Msg    string       `json:"msg"`
	Format string       `json:"fmt"`
}

type decodedSpanShared struct {
	*decodeCommon
	*decodeSpanShared
}

type decodedSpan struct {
	decodedSpanShared
	*decodeSpanExclusive
}

type decodeSpanExclusive struct {
	ParentSpanID string `json:"span.parent_span"`
}

type decodeSpanShared struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Duration    int64  `json:"dur"`
	SpanVersion int    `json:"span.ver"`
	subSpans    []string
}

type decodedRequest struct {
	decodedSpanShared
	*decodeRequestExclusive
}

type decodeRequestExclusive struct {
	Implmentation string `json:"impl"`
	ParentID      string `json:"parent.id"`
	State         string `json:"trace.state"`
	Baggage       string `json:"trace.baggage"`
	Namespace     string `json:"ns"`
	Source        string `json:"source"`
}

type decodeAttributeDefinition struct {
	xopat.Make
	AttributeType xopproto.AttributeType `json:"vtype"`
	SpanID        string                 `json:"span.id"`
}

func (logger *Logger) Replay(ctx context.Context, input any, output xopbase.Logger) error {
	data, ok := input.(string)
	if ok {
		return ReplayFromStrings(ctx, data, output)
	}
	return errors.Errorf("format not supported")
}

type baseReplay struct {
	logger                        xopbase.Logger
	spans                         map[string]xopbase.Span
	lines                         []decodedLine
	request                       xopbase.Request
	requestID                     string
	spanMap                       map[string]*decodedSpan
	traceID                       xoptrace.HexBytes16
	aDefs                         map[string]decodeAttributeDefinition
	perRequestAttributeDefintions map[string]map[string]*decodeAttributeDefinition
	attributeDefinitions          map[string]*decodeAttributeDefinition
	attributeRegistry             *xopat.Registry
}

func ReplayFromStrings(ctx context.Context, data string, logger xopbase.Logger) error {
	var requests []*decodedRequest
	var spans []*decodedSpan
	x := baseReplay{
		logger:                        logger,
		spans:                         make(map[string]xopbase.Span),
		spanMap:                       make(map[string]*decodedSpan),
		attributeRegistry:             xopat.NewRegistry(false),
		attributeDefinitions:          make(map[string]*decodeAttributeDefinition),
		perRequestAttributeDefintions: make(map[string]map[string]*decodeAttributeDefinition),
	}
	for _, inputText := range strings.Split(data, "\n") {
		if inputText == "" {
			continue
		}

		var super decodeAll
		err := json.Unmarshal([]byte(inputText), &super)
		if err != nil {
			return errors.Wrapf(err, "decode to super: %s", inputText)
		}

		super.unparsed = inputText

		switch super.Type {
		case "", "line":
			x.lines = append(x.lines, decodedLine{
				decodeCommon:        &super.decodeCommon,
				decodeLineExclusive: &super.decodeLineExclusive,
			})
		case "request":
			requests = append(requests, &decodedRequest{
				decodedSpanShared: decodedSpanShared{
					decodeCommon:     &super.decodeCommon,
					decodeSpanShared: &super.decodeSpanShared,
				},
				decodeRequestExclusive: &super.decodeRequestExclusive,
			})
			dc := &decodedSpan{
				decodedSpanShared: decodedSpanShared{
					decodeCommon:     &super.decodeCommon,
					decodeSpanShared: &super.decodeSpanShared,
				},
				// decodeSpanExclusive: &super.decodeSpanExclusive,
			}
			x.spanMap[super.SpanID] = dc // this may overwrite previous versions
		case "span":
			dc := &decodedSpan{
				decodedSpanShared: decodedSpanShared{
					decodeCommon:     &super.decodeCommon,
					decodeSpanShared: &super.decodeSpanShared,
				},
				decodeSpanExclusive: &super.decodeSpanExclusive,
			}
			x.spanMap[super.SpanID] = dc // this may overwrite previous versions
			spans = append(spans, dc)
		case "defineKey":
			var aDef decodeAttributeDefinition
			err := json.Unmarshal([]byte(inputText), &aDef)
			if err != nil {
				return errors.Wrapf(err, "decode attribute defintion (%s)", inputText)
			}
			if aDef.SpanID != "" {
				if x.perRequestAttributeDefintions[aDef.SpanID] == nil {
					x.perRequestAttributeDefintions[aDef.SpanID] = make(map[string]*decodeAttributeDefinition)
				}
				x.perRequestAttributeDefintions[aDef.SpanID][aDef.Key] = &aDef
			} else {
				x.attributeDefinitions[aDef.Key] = &aDef
			}
		default:
			return errors.Errorf("unknown line type (%s)", super.Type)
		}
	}
	// Attach child spans to parent spans so that they can be processed in order
	for _, span := range spans {
		if span.ParentSpanID == "" {
			return errors.Errorf("span (%s) is missing a span.parent_span", span.SpanID)
		}
		parent, ok := x.spanMap[span.ParentSpanID]
		if !ok {
			return errors.Errorf("parent span (%s) of span (%s) does not exist", span.ParentSpanID, span.SpanID)
		}
		parent.subSpans = append(parent.subSpans, span.SpanID)
	}

	for i := len(requests) - 1; i >= 0; i-- {
		// example: {"type":"request","span.ver":0,"trace.id":"29ee2638726b8ef34fa2f51fa2c7f82e","span.id":"9a6bc944044578c6","span.name":"TestParameters/unbuffered/no-attributes/one-span","ts":"2023-02-20T14:01:28.343114-06:00","source":"xopjson.test 0.0.0","ns":"xopjson.test 0.0.0"}
		requestInput := requests[i]
		if _, ok := x.spans[requestInput.SpanID]; ok {
			continue
		}
		var bundle xoptrace.Bundle
		x.traceID = xoptrace.NewHexBytes16FromString(requestInput.TraceID)
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.SpanID().SetString(requestInput.SpanID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		if requestInput.ParentID != "" {
			if !bundle.Parent.SetString(requestInput.ParentID) {
				return errors.Errorf("invalid parent id (%s) in request (%s)", requestInput.ParentID, bundle.Trace)
			}
		}
		if requestInput.Baggage != "" {
			bundle.Baggage.SetString(requestInput.Baggage)
		}
		if requestInput.State != "" {
			bundle.State.SetString(requestInput.State)
		}
		var err error
		var sourceInfo xopbase.SourceInfo
		sourceInfo.Source, sourceInfo.SourceVersion, err = version.SplitVersionWithError(requestInput.Source)
		if err != nil {
			return errors.Errorf("invalid source (%s) in request (%s)", requestInput.Source, bundle.Trace)
		}
		sourceInfo.Namespace, sourceInfo.NamespaceVersion, err = version.SplitVersionWithError(requestInput.Namespace)
		if err != nil {
			return errors.Errorf("invalid namespace (%s) in request (%s)", requestInput.Namespace, bundle.Trace)
		}

		if len(x.perRequestAttributeDefintions) != 0 {
			x.attributeDefinitions = make(map[string]*decodeAttributeDefinition)
		}
		x.requestID = requestInput.SpanID
		x.request = x.logger.Request(ctx,
			requestInput.Timestamp.Time,
			bundle,
			requestInput.Name,
			sourceInfo)
		x.spans[requestInput.SpanID] = x.request
		fmt.Println("XXX add request", requestInput.SpanID)
		err = spanReplayData{
			baseReplay:  x,
			span:        x.request,
			spanInput:   requestInput.decodedSpanShared,
			parentTrace: bundle.Trace,
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return x.ReplayLines()
}

type spanReplayData struct {
	baseReplay
	span        xopbase.Span
	spanInput   decodedSpanShared
	parentTrace xoptrace.Trace
}

func (x spanReplayData) Replay(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "span '%s'", x.spanInput.unparsed)
		}
	}()
	attributes := x.spanInput.Attributes
	if attributes == nil {
		err := json.Unmarshal([]byte(x.spanInput.unparsed), &attributes)
		if err != nil {
			return errors.Errorf("cannot unmarshal to generic for span %s", x.spanInput.SpanID)
		}
		for _, k := range knownKeys {
			delete(attributes, k)
		}
	}
	for k, v := range attributes {
		var aDef *decodeAttributeDefinition
		if m, ok := x.perRequestAttributeDefintions[x.requestID]; ok {
			aDef = m[k]
		} else {
			aDef = x.attributeDefinitions[k]
		}
		if aDef == nil {
			return errors.Errorf("No attribute definition in span (%s) for key (%s)", x.spanInput.SpanID, k)
		}

		switch aDef.AttributeType {
		case xopproto.AttributeType_Any:
			registeredAttribute, err := x.attributeRegistry.ConstructAnyAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []xopbase.ModelArg
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []xopbase.ModelArg")
				}
				for _, e := range va {
					x.span.MetadataAny(ra, e)
				}
			} else {
				var e xopbase.ModelArg
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata xopbase.ModelArg")
				}
				x.span.MetadataAny(ra, e)
			}
		case xopproto.AttributeType_Bool:
			registeredAttribute, err := x.attributeRegistry.ConstructBoolAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []bool
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []bool")
				}
				for _, e := range va {
					x.span.MetadataBool(ra, e)
				}
			} else {
				var e bool
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata bool")
				}
				x.span.MetadataBool(ra, e)
			}
		case xopproto.AttributeType_Enum:
			registeredAttribute, err := x.attributeRegistry.ConstructEnumAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := &registeredAttribute.EnumAttribute
			if aDef.Multiple {
				var va []xopat.Enum
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []xopat.Enum")
				}
				for _, e := range va {
					x.span.MetadataEnum(ra, e)
				}
			} else {
				var e xopat.Enum
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata xopat.Enum")
				}
				x.span.MetadataEnum(ra, e)
			}
		case xopproto.AttributeType_Float64:
			registeredAttribute, err := x.attributeRegistry.ConstructFloat64Attribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []float64
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []float64")
				}
				for _, e := range va {
					x.span.MetadataFloat64(ra, e)
				}
			} else {
				var e float64
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata float64")
				}
				x.span.MetadataFloat64(ra, e)
			}
		case xopproto.AttributeType_Int64:
			registeredAttribute, err := x.attributeRegistry.ConstructInt64Attribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []int64
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []int64")
				}
				for _, e := range va {
					x.span.MetadataInt64(ra, e)
				}
			} else {
				var e int64
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata int64")
				}
				x.span.MetadataInt64(ra, e)
			}
		case xopproto.AttributeType_Link:
			registeredAttribute, err := x.attributeRegistry.ConstructLinkAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []xoptrace.Trace
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []xoptrace.Trace")
				}
				for _, e := range va {
					x.span.MetadataLink(ra, e)
				}
			} else {
				var e xoptrace.Trace
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata xoptrace.Trace")
				}
				x.span.MetadataLink(ra, e)
			}
		case xopproto.AttributeType_String:
			registeredAttribute, err := x.attributeRegistry.ConstructStringAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []string
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []string")
				}
				for _, e := range va {
					x.span.MetadataString(ra, e)
				}
			} else {
				var e string
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata string")
				}
				x.span.MetadataString(ra, e)
			}
		case xopproto.AttributeType_Time:
			registeredAttribute, err := x.attributeRegistry.ConstructTimeAttribute(aDef.Make)
			if err != nil {
				return err
			}
			ra := registeredAttribute
			if aDef.Multiple {
				var va []time.Time
				err := json.Unmarshal(v, &va)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata []time.Time")
				}
				for _, e := range va {
					x.span.MetadataTime(ra, e)
				}
			} else {
				var e time.Time
				err := json.Unmarshal(v, &e)
				if err != nil {
					return errors.Wrap(err, "could not unmarshal metadata time.Time")
				}
				x.span.MetadataTime(ra, e)
			}

		default:
			return errors.Errorf("unexpected attribute type (%s)", aDef.AttributeType)
		}
	}
	for _, subSpanID := range x.spanInput.subSpans {
		spanInput, ok := x.spanMap[subSpanID]
		if !ok {
			return errors.Errorf("internal error")
		}
		var bundle xoptrace.Bundle
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		bundle.Trace.SpanID().SetString(subSpanID)
		bundle.Parent = x.parentTrace
		// {"type":"span","span.name":"a fork one span","ts":"2023-02-21T20:14:35.987376-05:00","span.parent_span":"3a215dd6311084d5","span.id":"3a215dd6311084d5","span.seq":".A","span.ver":1,"dur":135000,"http.route":"/some/thing"}
		span := x.span.Span(
			ctx,
			x.spanInput.Timestamp.Time,
			bundle,
			x.spanInput.Name,
			x.spanInput.SequenceCode,
		)
		x.spans[subSpanID] = span
		fmt.Println("XXX add span", subSpanID)
		err := spanReplayData{
			baseReplay:  x.baseReplay,
			span:        span,
			spanInput:   spanInput.decodedSpanShared,
			parentTrace: bundle.Trace,
		}.Replay(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type lineAttribute struct {
	Type     string      `json:"t"`
	Value    interface{} `json:"v"`
	Encoding string      `json:"encoding,omitempty"`  // for models
	TypeName string      `json:"modelType,omitempty"` // for models
	IntValue int64       `json:"i,omitempty"`         // for enums
}

func (x baseReplay) ReplayLines() error {
	for _, lineInput := range x.lines {
		err := x.ReplayLine(lineInput)
		if err != nil {
			return err
		}
	}
	return nil
}

func (x baseReplay) ReplayLine(lineInput decodedLine) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "line '%s'", lineInput.unparsed)
		}
	}()
	span, ok := x.spans[lineInput.SpanID]
	if !ok {
		return errors.Errorf("unknown span for line %s", lineInput.unparsed)
	}
	line := span.NoPrefill().Line(
		lineInput.Level,
		lineInput.Timestamp.Time,
		nil, // XXX TODO
	)
	for k, enc := range lineInput.Attributes {
		var la lineAttribute
		err := json.Unmarshal(enc, &la)
		if err != nil {
			errors.Wrap(err, "could not decode line") // XXX
		}
		dataType, ok := stringToDataType[la.Type]
		if !ok {
			return errors.Errorf("unknown data type (%s) in line attribute", la.Type)
		}
		switch dataType {
		case xopbase.EnumDataType:
			s, ok := la.Value.(string)
			if !ok {
				return errors.Errorf("invalid enum string (%T) when decoding attribute", la.Value)
			}
			m := xopat.Make{
				Key: k,
			}
			ea, err := x.attributeRegistry.ConstructEnumAttribute(m)
			if err != nil {
				return errors.Wrapf(err, "build enum attribute for line attribute (%s)", k)
			}
			enum := ea.Add64(la.IntValue, s)
			line.Enum(&ea.EnumAttribute, enum)
		case xopbase.AnyDataType:
			var ma xopbase.ModelArg
			switch la.Encoding {
			case "", "JSON":
				ma.Model = la.Value
				ma.Encoding = xopproto.Encoding_JSON
			default:
				vs, ok := la.Value.(string)
				if !ok {
					return errors.Errorf("invalid model arg (%T) when decoding attribute", la.Value)
				}
				ma.Encoded = []byte(vs)
				if en, ok := xopproto.Encoding_value[la.Encoding]; ok {
					ma.Encoding = xopproto.Encoding(en)
				} else {
					return errors.Errorf("invalid encoding (%s) when decoding attribute", la.Encoding)
				}
			}
			ma.TypeName = la.TypeName
			line.Any(k, ma)
		case xopbase.StringDataType, xopbase.StringerDataType:
			s, ok := la.Value.(string)
			if !ok {
				return errors.Errorf("invalid bool (%T) when decoding attribute", la.Value)
			}
			line.String(k, s, dataType)
		case xopbase.BoolDataType:
			b, ok := la.Value.(bool)
			if !ok {
				return errors.Errorf("invalid bool (%T) when decoding attribute", la.Value)
			}
			line.Bool(k, b)
		case xopbase.DurationDataType:
			if s, ok := la.Value.(string); ok {
				d, err := time.ParseDuration(s)
				if err != nil {
					return errors.Wrap(err, "parse duration when decoding attribute")
				}
				line.Duration(k, d)
			} else {
				return errors.Wrapf(err, "invalid duration (%T) in attribute", la.Value)
			}
		case xopbase.TimeDataType:
			if s, ok := la.Value.(string); ok {
				d, err := time.Parse(time.RFC3339Nano, s)
				if err != nil {
					return errors.Wrap(err, "parse time when decoding attribute")
				}
				line.Time(k, d)
			} else {
				return errors.Wrapf(err, "invalid duration (%T) in attribute", la.Value)
			}
		case xopbase.Float32DataType, xopbase.Float64DataType:
			if f, ok := la.Value.(float64); ok {
				line.Float64(k, f, dataType)
			}
		case xopbase.IntDataType, xopbase.Int8DataType, xopbase.Int16DataType, xopbase.Int32DataType, xopbase.Int64DataType:
			var i int64
			switch t := la.Value.(type) {
			case float64:
				i = int64(t)
			case string:
				var err error
				i, err = strconv.ParseInt(t, 10, 64)
				if err != nil {
					return errors.Wrap(err, "invalid int encoded as string in attribute")
				}
			default:
				return errors.Wrapf(err, "invalid int (%T) in attribute", la.Value)
			}
			line.Int64(k, i, dataType)
		case xopbase.UintDataType, xopbase.Uint8DataType, xopbase.Uint16DataType, xopbase.Uint32DataType, xopbase.Uint64DataType, xopbase.UintptrDataType:
			var i uint64
			switch t := la.Value.(type) {
			case float64:
				i = uint64(t)
			case string:
				var err error
				i, err = strconv.ParseUint(t, 10, 64)
				if err != nil {
					return errors.Wrap(err, "invalid uint encoded as string in attribute")
				}
			default:
				return errors.Wrapf(err, "invalid uint (%T) in attribute", la.Value)
			}
			line.Uint64(k, i, dataType)
		default:
			return errors.Errorf("unexpected data type (%s) in line attribute", la.Type)
		}
	}
	return nil
}

// XXX shorten and generate reverse
var stringToDataType = map[string]xopbase.DataType{
	"i":        xopbase.IntDataType,
	"i8":       xopbase.Int8DataType,
	"i16":      xopbase.Int16DataType,
	"i32":      xopbase.Int32DataType,
	"i64":      xopbase.Int64DataType,
	"u":        xopbase.UintDataType,
	"u8":       xopbase.Uint8DataType,
	"u16":      xopbase.Uint16DataType,
	"u32":      xopbase.Uint32DataType,
	"u64":      xopbase.Uint64DataType,
	"uintptr":  xopbase.UintptrDataType,
	"f32":      xopbase.Float32DataType,
	"f64":      xopbase.Float64DataType,
	"any":      xopbase.AnyDataType,
	"bool":     xopbase.BoolDataType,
	"dur":      xopbase.DurationDataType,
	"time":     xopbase.TimeDataType,
	"s":        xopbase.StringDataType,
	"stringer": xopbase.StringerDataType,
	"enum":     xopbase.EnumDataType,
}

var dataTypeToString = func() map[xopbase.DataType]string {
	m := make(map[xopbase.DataType]string)
	for k, v := range stringToDataType {
		m[v] = k
	}
	return m
}()
