// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"context"
	"encoding/json"
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
				return errors.Wrap(err, "decode attribute defintion")
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

func (x spanReplayData) Replay(ctx context.Context) error {
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
	Encoding string      `json:"encoding,omitempty"`
	TypeName string      `json:"modelType,omitempty"`
	Value    interface{} `json:"v"`
}

func (x baseReplay) ReplayLines() error {
	for _, lineInput := range x.lines {
		span, ok := x.spans[lineInput.SpanID]
		if !ok {
			return errors.Errorf("unknown span for line %s", lineInput.unparsed)
		}
		line := span.NoPrefill().Line(
			x.lineInput.Level,
			x.lineInput.Timestamp.Time,
			nil, // XXX TODO
		)
		for k, enc := range x.lineInput.Attributes {
			var la lineAttribute
			err := json.Unmarshal(enc, &la)
			if err != nil {
				errors.Wrap(err, "could not decode line") // XXX
			}
			switch la.Type {
			case "model":
				var ma xopbase.ModelArg
				switch la.Encoding {
				case "", "JSON":
					ma.Model = la.Value
					ma.Encoding = xopproto.Encoding_JSON
				default:
					ma.Encoded = []byte(la.Value)
					var ok bool
					ma.Encoding, ok = xoproto.Encoding_value[la.Encoding]
					if !ok {
						return errors.Errorf("invalid encoding (%s) when decoding attribute", la.Encoding)
					}
				}
				ma.TypeName = la.TypeName
				line.Any(k, ma)
			case "bool":
				line.Any(k, la.Value.(bool))
			case "Int", "Int8", "Int32", "Int64":
			default:
			}
		}
	}
}
