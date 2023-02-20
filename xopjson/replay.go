package xopjson

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"

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

func ReplayFromStrings(ctx context.Context, data string, output xopbase.Logger) error {
	var lines []decodedLine
	var requests []*decodedRequest
	var spans []*decodedSpan
	spanMap := make(map[string]*decodedSpan)
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
		request := requests[i]
		if previous, ok := processed[request.SpanID]; ok && previous.ver > request.SpanVersion {

		} else {
		}
	}
	return nil
}
