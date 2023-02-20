package xopjson

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/xoplog/xop-go/internal/util/version"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/pkg/errors"
)

type decodeAll struct {
	decodeCommon
	decodeLineExclusive
	decodeSpanShared
	decodeSpanExclusive
	decodeRequestExclusive
}

type decodeCommon struct {
	// lines, spans, and requests
	Timestamp  xoptestutil.TS         `json:"ts"`
	Attributes map[string]interface{} `json:"attributes"`
	SpanID     string                 `json:"span.id"`
}

type decodedLine struct {
	*decodeCommon
	*decodeLineExclusive
}

type decodeLineExclusive struct {
	Level  int      `json:"lvl"`
	Stack  []string `json:"stack"`
	Msg    string   `json:"msg"`
	Format string   `json:"fmt"`
}

type decodedSpan struct {
	*decodeCommon
	*decodeSpanShared
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
	subspans    []string
}

type decodedRequest struct {
	*decodeCommon
	*decodeSpanShared
	*decodeRequestExclusive
}

type decodeRequestExclusive struct {
	Implmentation string `json:"impl"`
	TraceID       string `json:"trace.id"`
	ParentID      string `json:"parent.id"`
	State         string `json:"trace.state"`
	Baggage       string `json:"trace.baggage"`
	Namespace     string `json:"ns"`
	Source        string `json:"source"`
}

type replayTrace struct {
	logger    xopbase.Logger
	spansSeen map[string]spanData
}

type spanData struct {
	version int
	span    xopbase.Span
}

func (logger *Logger) Replay(ctx context.Context, input any, output xopbase.Logger) error {
	data, ok := input.(string)
	if ok {
		return ReplayFromStrings(ctx, data, output)
	}
	return errors.Errorf("format not supported")
}

func ReplayFromStrings(ctx context.Context, data string, logger xopbase.Logger) error {
	var lines []decodedLine
	var requests []*decodedRequest
	var spans []*decodedSpan
	spanMap := make(map[string]*decodedSpan)
	x := replayTrace{
		logger:    logger,
		spansSeen: make(map[string]spanData),
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

		switch super.Type {
		case "", "line":
			lines = append(lines, decodedLine{
				decodeCommon:        &super.decodeCommon,
				decodeLineExclusive: &super.decodeLineExclusive,
			})
		case "request":
			requests = append(requests, &decodedRequest{
				decodeCommon:           &super.decodeCommon,
				decodeSpanShared:       &super.decodeSpanShared,
				decodeRequestExclusive: &super.decodeRequestExclusive,
			})
			dc := &decodedSpan{
				decodeCommon:     &super.decodeCommon,
				decodeSpanShared: &super.decodeSpanShared,
				// decodeSpanExclusive: &super.decodeSpanExclusive,
			}
			spanMap[super.SpanID] = dc
		case "span":
			dc := &decodedSpan{
				decodeCommon:        &super.decodeCommon,
				decodeSpanShared:    &super.decodeSpanShared,
				decodeSpanExclusive: &super.decodeSpanExclusive,
			}
			spanMap[super.SpanID] = dc
			spans = append(spans, dc)
		}
	}
	// Attach child spans to parent spans so that they can be processed in order
	for _, span := range spans {
		if span.ParentSpanID == "" {
			return errors.Errorf("span (%s) is missing a span.parent_span", span.SpanID)
		}
		parent, ok := spanMap[span.ParentSpanID]
		if !ok {
			return errors.Errorf("parent span (%s) of span (%s) does not exist", span.ParentSpanID, span.SpanID)
		}
		parent.subspans = append(parent.subspans, span.SpanID)
	}

	for i := len(requests) - 1; i >= 0; i-- {
		// example: {"type":"request","span.ver":0,"trace.id":"29ee2638726b8ef34fa2f51fa2c7f82e","span.id":"9a6bc944044578c6","span.name":"TestParameters/unbuffered/no-attributes/one-span","ts":"2023-02-20T14:01:28.343114-06:00","source":"xopjson.test 0.0.0","ns":"xopjson.test 0.0.0"}
		requestInput := requests[i]
		if previous, ok := x.spansSeen[requestInput.SpanID]; ok && previous.version > requestInput.SpanVersion {
		} else {
			var bundle xoptrace.Bundle
			bundle.Trace.TraceID().SetString(requestInput.TraceID)
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

			x.request = x.logger.Request(ctx,
				requestInput.Timestamp.Time,
				bundle,
				requestInput.Name,
				sourceInfo)
		}
	}
	return nil
}
