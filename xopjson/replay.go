// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"context"
	"encoding/json"
	"regexp"
	"runtime"
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
	"github.com/xoplog/xop-go/xoputil"
	"github.com/xoplog/xop-go/xoputil/replayutil"

	"github.com/pkg/errors"
)

var knownKeys = []string{
	"ts",
	"attributes",
	"span.id",
	"trace.id",
	"trace.parent",
	"span.seq",
	"lvl",
	"stack",
	"msg",
	"fmt",
	"impl",
	"trace.state",
	"trace.baggage",
	"ns",
	"source",
	"dur",
	"type",
	"span.name",
	"span.ver",
	"span.parent_span",
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
	Type         string                     `json:"type"`
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
	Level     xopnum.Level `json:"lvl"`
	Stack     []string     `json:"stack"`
	Msg       string       `json:"msg"`
	Format    string       `json:"fmt"`
	ModelType string       `json:"modelType"` // model only
	Encoding  string       `json:"encoding"`  // model only
	Encoded   interface{}  `json:"encoded"`   // model only
	Link      string       `json:"link"`      // link only
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
	Name        string `json:"name"`
	Duration    *int64 `json:"dur"`
	SpanVersion int    `json:"span.ver"`
	subSpans    []string
}

type decodedRequest struct {
	decodedSpanShared
	*decodeRequestExclusive
	request xopbase.Request
}

type decodeRequestExclusive struct {
	Implmentation string `json:"impl"`
	ParentID      string `json:"trace.parent"`
	State         string `json:"trace.state"`
	Baggage       string `json:"trace.baggage"`
	Namespace     string `json:"ns"`
	Source        string `json:"source"`
}

func (logger *Logger) Replay(ctx context.Context, input any, output xopbase.Logger) error {
	data, ok := input.(string)
	if ok {
		return ReplayFromStrings(ctx, data, output)
	}
	return errors.Errorf("format not supported")
}

type baseReplay struct {
	logger               xopbase.Logger
	spans                map[string]xopbase.Span
	lines                []decodedLine
	request              xopbase.Request
	requestID            string
	spanMap              map[string]*decodedSpan
	traceID              xoptrace.HexBytes16
	attributeRegistry    *xopat.Registry
	attributeDefinitions *replayutil.GlobalAttributeDefinitions
}

func ReplayFromStrings(ctx context.Context, data string, logger xopbase.Logger) error {
	xopat.ResetCachedKeys() // prevent memory exhaustion
	var requests []*decodedRequest
	var spans []*decodedSpan
	x := baseReplay{
		logger:               logger,
		spans:                make(map[string]xopbase.Span),
		spanMap:              make(map[string]*decodedSpan),
		attributeRegistry:    xopat.NewRegistry(false),
		attributeDefinitions: replayutil.NewGlobalAttributeDefinitions(),
	}
	var defineKeys []string
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
		case "", "line", "model", "link":
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
			if _, ok := x.spanMap[super.SpanID]; !ok {
				_ = x.attributeDefinitions.NewRequestAttributeDefinitions(super.SpanID)
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
			defineKeys = append(defineKeys, inputText)
		default:
			return errors.Errorf("unknown line type (%s) for input (%s)", super.Type, inputText)
		}
	}
	for _, inputText := range defineKeys {
		err := x.attributeDefinitions.Decode(inputText)
		if err != nil {
			return err
		}
	}
	// Attach child spans to parent spans so that they can be processed in order
	for _, span := range spans {
		if span.ParentSpanID == "" {
			return errors.Errorf("span (%s) is missing a span.parent_span", span.unparsed)
		}
		parent, ok := x.spanMap[span.ParentSpanID]
		if !ok {
			return errors.Errorf("parent span (%s) of span (%s) does not exist", span.ParentSpanID, span.unparsed)
		}
		parent.subSpans = append(parent.subSpans, span.SpanID)
	}

	last := make(map[string]int)
	for i := len(requests) - 1; i >= 0; i-- {
		requestInput := requests[i]
		if _, ok := last[requestInput.SpanID]; !ok {
			last[requestInput.SpanID] = i
		}
	}

	for i := range requests {
		requestInput := requests[i]
		if lastOccurence, ok := last[requestInput.SpanID]; ok {
			delete(last, requestInput.SpanID)
			requestInput = requests[lastOccurence]
		} else {
			continue
		}

		// example: {"type":"request","span.ver":0,"trace.id":"29ee2638726b8ef34fa2f51fa2c7f82e","span.id":"9a6bc944044578c6","span.name":"TestParameters/unbuffered/no-attributes/one-span","ts":"2023-02-20T14:01:28.343114-06:00","source":"xopjson.test 0.0.0","ns":"xopjson.test 0.0.0"}
		var bundle xoptrace.Bundle
		x.traceID = xoptrace.NewHexBytes16FromString(requestInput.TraceID)
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.SpanID().SetString(requestInput.SpanID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		if requestInput.ParentID != "" {
			if !bundle.Parent.SetString(requestInput.ParentID) {
				return errors.Errorf("invalid parent id (%s) in request (%s)", requestInput.ParentID, requestInput.unparsed)
			}
		} else {
			bundle.Parent.Flags().SetBytes([]byte{1})
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
			return errors.Errorf("invalid source (%s) in request (%s)", requestInput.Source, requestInput.unparsed)
		}
		sourceInfo.Namespace, sourceInfo.NamespaceVersion, err = version.SplitVersionWithError(requestInput.Namespace)
		if err != nil {
			return errors.Errorf("invalid namespace (%s) in request (%s)", requestInput.Namespace, requestInput.unparsed)
		}

		x.requestID = requestInput.SpanID
		x.request = x.logger.Request(ctx,
			requestInput.Timestamp.Time,
			bundle,
			requestInput.Name,
			sourceInfo)
		requestInput.request = x.request
		x.spans[requestInput.SpanID] = x.request
		err = spanReplayData{
			baseReplay:  x,
			span:        x.request,
			spanInput:   requestInput.decodedSpanShared,
			parentTrace: bundle.Trace,
			skipDone:    true,
		}.Replay(ctx)
		if err != nil {
			return errors.Wrapf(err, "in span (%s)", requestInput.decodedSpanShared.unparsed)
		}
	}
	err := x.ReplayLines()
	if err != nil {
		return err
	}
	for i := len(requests) - 1; i >= 0; i-- {
		requestInput := requests[i]
		if requestInput.request != nil {
			if requestInput.Duration != nil {
				requestInput.request.Done(requestInput.Timestamp.Time.Add(time.Duration(*requestInput.Duration)), false)
				requestInput.request = nil
			}
		}
	}
	return nil
}

type spanReplayData struct {
	baseReplay
	span        xopbase.Span
	spanInput   decodedSpanShared
	parentTrace xoptrace.Trace
	skipDone    bool
}

func (x spanReplayData) Replay(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "span '%s'", x.spanInput.unparsed)
		}
	}()
	attributes := x.spanInput.Attributes
	for k, v := range attributes {
		err := replaySpanAttribute{
			spanReplayData: x,
		}.Replay(k, v)
		if err != nil {
			return errors.Wrapf(err, "in span attribute (%s: %s)", k, string(v))
		}
	}
	for _, subSpanID := range x.spanInput.subSpans {
		if _, ok := x.spans[subSpanID]; ok {
			continue
		}
		spanInput, ok := x.spanMap[subSpanID]
		if !ok {
			return errors.Errorf("internal error, could not find span %s", subSpanID)
		}
		var bundle xoptrace.Bundle
		bundle.Trace.TraceID().Set(x.traceID)
		bundle.Trace.Flags().SetBytes([]byte{1})
		bundle.Trace.SpanID().SetString(subSpanID)
		bundle.Parent = x.parentTrace
		// {"type":"span","span.name":"a fork one span","ts":"2023-02-21T20:14:35.987376-05:00","span.parent_span":"3a215dd6311084d5","span.id":"3a215dd6311084d5","span.seq":".A","span.ver":1,"dur":135000,"http.route":"/some/thing"}
		span := x.span.Span(
			ctx,
			spanInput.Timestamp.Time,
			bundle,
			spanInput.Name,
			spanInput.SequenceCode,
		)
		x.spans[subSpanID] = span
		err := spanReplayData{
			baseReplay:  x.baseReplay,
			span:        span,
			spanInput:   spanInput.decodedSpanShared,
			parentTrace: bundle.Trace,
		}.Replay(ctx)
		if err != nil {
			return errors.Wrapf(err, "in subspan (%s)", spanInput.decodedSpanShared.unparsed)
		}
	}
	if !x.skipDone && x.spanInput.Duration != nil {
		x.span.Done(x.spanInput.Timestamp.Time.Add(time.Duration(*x.spanInput.Duration)), false)
	}
	return nil
}

type replaySpanAttribute struct {
	spanReplayData
}

func (x replaySpanAttribute) Replay(k string, v []byte) error {
	aDef := x.attributeDefinitions.Lookup(x.requestID, k)
	if aDef == nil {
		return errors.Errorf("No attribute definition in span (%s) for key (%s)", x.spanInput.SpanID, k)
	}

	switch aDef.AttributeType {
	case xopproto.AttributeType_Any:
		registeredAttribute, err := x.attributeRegistry.ConstructAnyAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v xopbase.ModelArg) {
			x.span.MetadataAny(ra, v)
		}
		if aDef.Multiple {
			var va []xopbase.ModelArg
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []xopbase.ModelArg")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e xopbase.ModelArg
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata xopbase.ModelArg")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Bool:
		registeredAttribute, err := x.attributeRegistry.ConstructBoolAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v bool) {
			x.span.MetadataBool(ra, v)
		}
		if aDef.Multiple {
			var va []bool
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []bool")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e bool
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata bool")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Duration:
		registeredAttribute, err := x.attributeRegistry.ConstructDurationAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v time.Duration) {
			x.span.MetadataInt64(&ra.Int64Attribute, int64(v))
		}
		if aDef.Multiple {
			var va []time.Duration
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []time.Duration")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e time.Duration
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata time.Duration")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Enum:
		registeredAttribute, err := x.attributeRegistry.ConstructEnumAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := &registeredAttribute.EnumAttribute
		addMetadata := func(v xopat.Enum) {
			x.span.MetadataEnum(ra, v)
		}
		if aDef.Multiple {
			var va []xoputil.DecodeEnum
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []xopat.Enum")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e xoputil.DecodeEnum
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata xopat.Enum")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Float64:
		registeredAttribute, err := x.attributeRegistry.ConstructFloat64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v float64) {
			x.span.MetadataFloat64(ra, v)
		}
		if aDef.Multiple {
			var va []float64
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []float64")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e float64
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata float64")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Int:
		registeredAttribute, err := x.attributeRegistry.ConstructIntAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v int) {
			x.span.MetadataInt64(&ra.Int64Attribute, int64(v))
		}
		if aDef.Multiple {
			var va []int
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []int")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e int
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata int")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Int16:
		registeredAttribute, err := x.attributeRegistry.ConstructInt16Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v int16) {
			x.span.MetadataInt64(&ra.Int64Attribute, int64(v))
		}
		if aDef.Multiple {
			var va []int16
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []int16")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e int16
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata int16")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Int32:
		registeredAttribute, err := x.attributeRegistry.ConstructInt32Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v int32) {
			x.span.MetadataInt64(&ra.Int64Attribute, int64(v))
		}
		if aDef.Multiple {
			var va []int32
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []int32")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e int32
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata int32")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Int64:
		registeredAttribute, err := x.attributeRegistry.ConstructInt64Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v int64) {
			x.span.MetadataInt64(ra, v)
		}
		if aDef.Multiple {
			var va []int64
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []int64")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e int64
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata int64")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Int8:
		registeredAttribute, err := x.attributeRegistry.ConstructInt8Attribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v int8) {
			x.span.MetadataInt64(&ra.Int64Attribute, int64(v))
		}
		if aDef.Multiple {
			var va []int8
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []int8")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e int8
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata int8")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Link:
		registeredAttribute, err := x.attributeRegistry.ConstructLinkAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v xoptrace.Trace) {
			x.span.MetadataLink(ra, v)
		}
		if aDef.Multiple {
			var va []xoptrace.Trace
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []xoptrace.Trace")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e xoptrace.Trace
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata xoptrace.Trace")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_String:
		registeredAttribute, err := x.attributeRegistry.ConstructStringAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v string) {
			x.span.MetadataString(ra, v)
		}
		if aDef.Multiple {
			var va []string
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []string")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e string
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata string")
			}
			addMetadata(e)
		}
	case xopproto.AttributeType_Time:
		registeredAttribute, err := x.attributeRegistry.ConstructTimeAttribute(aDef.Make, xopat.AttributeType(aDef.AttributeType))
		if err != nil {
			return err
		}
		ra := registeredAttribute
		addMetadata := func(v time.Time) {
			x.span.MetadataTime(ra, v)
		}
		if aDef.Multiple {
			var va []time.Time
			err := json.Unmarshal(v, &va)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata []time.Time")
			}
			for _, e := range va {
				addMetadata(e)
			}
		} else {
			var e time.Time
			err := json.Unmarshal(v, &e)
			if err != nil {
				return errors.Wrap(err, "could not unmarshal metadata time.Time")
			}
			addMetadata(e)
		}

	default:
		return errors.Errorf("unexpected attribute type (%s)", aDef.AttributeType)
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

var lineRE = regexp.MustCompile(`^(.+):(\d+)$`)

func (x baseReplay) ReplayLine(lineInput decodedLine) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "in line (%s)", lineInput.unparsed)
		}
	}()
	span, ok := x.spans[lineInput.SpanID]
	if !ok {
		return errors.Errorf("unknown span (%s)", lineInput.SpanID)
	}
	frames := make([]runtime.Frame, len(lineInput.Stack))
	for i, s := range lineInput.Stack {
		m := lineRE.FindStringSubmatch(s)
		if m == nil {
			return errors.Errorf("could not match stack line '%s'", s)
		}
		frames[i].File = m[1]
		num, _ := strconv.ParseInt(m[2], 10, 64)
		frames[i].Line = int(num)
	}
	line := span.NoPrefill().Line(
		lineInput.Level,
		lineInput.Timestamp.Time,
		frames,
	)

	for k, enc := range lineInput.Attributes {
		err := replayLineAttribute{
			baseReplay: x,
		}.Replay(line, lineInput, k, enc)
		if err != nil {
			return errors.Wrapf(err, "in attribute (%s: %s)", k, string(enc))
		}

	}
	switch lineInput.Type {
	case "", "line":
		switch lineInput.Format {
		case "tmpl":
			line.Template(lineInput.Msg)
		case "":
			line.Msg(lineInput.Msg)
		default:
			return errors.Errorf("unexpected line fmt: %s", lineInput.Format)
		}
	case "link":
		if link, ok := xoptrace.TraceFromString(lineInput.Link); ok {
			line.Link(lineInput.Msg, link)
		} else {
			return errors.Errorf("invalid link in line: '%s'", lineInput.Link)
		}
	case "model":
		// We are constrained by the fact that we decode all types at once
		// and thus cannot use a json.RawMessage for the encoded type.
		var ma xopbase.ModelArg
		switch lineInput.Encoding {
		case "", "JSON":
			ma.Model = lineInput.Encoded
			ma.Encoding = xopproto.Encoding_JSON
		default:
			vs, ok := lineInput.Encoded.(string)
			if !ok {
				return errors.Errorf("invalid model arg (%T) when decoding attribute", lineInput.Encoded)
			}
			ma.Encoded = []byte(vs)
			if en, ok := xopproto.Encoding_value[lineInput.Encoding]; ok {
				ma.Encoding = xopproto.Encoding(en)
			} else {
				return errors.Errorf("invalid encoding (%s) when decoding attribute", lineInput.Encoding)
			}
		}
		ma.ModelType = lineInput.ModelType
		line.Model(lineInput.Msg, ma)
	default:
		return errors.Errorf("unexpected type for line: %s", lineInput.Type)
	}
	return nil
}

type replayLineAttribute struct {
	baseReplay
}

func (x replayLineAttribute) Replay(line xopbase.Line, lineInput decodedLine, ks string, enc []byte) error {
	k := xopat.K(ks)
	var la lineAttribute
	err := json.Unmarshal(enc, &la)
	if err != nil {
		errors.Wrap(err, "could not decode")
	}
	dataType, ok := xopbase.StringToDataType[la.Type]
	if !ok {
		return errors.Errorf("unknown data type (%s)", la.Type)
	}
	switch dataType {
	case xopbase.EnumDataType:
		s, ok := la.Value.(string)
		if !ok {
			return errors.Errorf("invalid enum string (%T)", la.Value)
		}
		m := xopat.Make{
			Key: ks,
		}
		ea, err := x.attributeRegistry.ConstructEnumAttribute(m, xopat.AttributeTypeEnum)
		if err != nil {
			return errors.Wrap(err, "build enum attribute")
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
				return errors.Errorf("invalid model arg (%T)", la.Value)
			}
			ma.Encoded = []byte(vs)
			if en, ok := xopproto.Encoding_value[la.Encoding]; ok {
				ma.Encoding = xopproto.Encoding(en)
			} else {
				return errors.Errorf("invalid encoding (%s)", la.Encoding)
			}
		}
		ma.ModelType = la.TypeName
		line.Any(k, ma)
	case xopbase.StringDataType, xopbase.StringerDataType, xopbase.ErrorDataType:
		s, ok := la.Value.(string)
		if !ok {
			return errors.Errorf("invalid bool (%T)", la.Value)
		}
		line.String(k, s, dataType)
	case xopbase.BoolDataType:
		b, ok := la.Value.(bool)
		if !ok {
			return errors.Errorf("invalid bool (%T)", la.Value)
		}
		line.Bool(k, b)
	case xopbase.DurationDataType:
		if s, ok := la.Value.(string); ok {
			d, err := time.ParseDuration(s)
			if err != nil {
				return errors.Wrap(err, "parse duration")
			}
			line.Duration(k, d)
		} else {
			return errors.Wrapf(err, "invalid duration (%T)", la.Value)
		}
	case xopbase.TimeDataType:
		if s, ok := la.Value.(string); ok {
			d, err := time.Parse(time.RFC3339Nano, s)
			if err != nil {
				return errors.Wrap(err, "parse time")
			}
			line.Time(k, d)
		} else {
			return errors.Wrapf(err, "invalid duration (%T)", la.Value)
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
				return errors.Wrap(err, "invalid int encoded as string")
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
				return errors.Wrap(err, "invalid uint encoded as string")
			}
		default:
			return errors.Wrapf(err, "invalid uint (%T)", la.Value)
		}
		line.Uint64(k, i, dataType)
	default:
		return errors.Errorf("unexpected data type (%s)", la.Type)
	}
	return nil
}
