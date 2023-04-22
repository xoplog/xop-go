package recorderutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xoplog/xop-go/xoprecorder"
)

func DumpEvents(t testing.TB, tlog *xoprecorder.Logger) {
	var o []string
	_ = tlog.WithLock(func(tlog *xoprecorder.Logger) error {
		for _, event := range tlog.Events {
			switch event.Type {
			case xoprecorder.LineEvent:
				o = append(o, fmt.Sprintf("line: %s", event.Line.Text()))
			case xoprecorder.SpanStart:
				o = append(o, fmt.Sprintf("spanstart: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoprecorder.SpanDone:
				o = append(o, fmt.Sprintf("spandone: %s", event.Span.Bundle.Trace.SpanID().String()))

			case xoprecorder.RequestStart:
				o = append(o, fmt.Sprintf("requeststart: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoprecorder.RequestDone:
				o = append(o, fmt.Sprintf("requestdone: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoprecorder.FlushEvent:
				o = append(o, "Flush!")
			case xoprecorder.CustomEvent:
				o = append(o, "Custom: "+event.Msg)
			case xoprecorder.MetadataSet:
				o = append(o, fmt.Sprintf("Metadata on %s: %s", event.Span.Bundle.Trace.SpanID().String(), event.Msg))
			default:
				o = append(o, "unknown event")
			}
		}
		return nil
	})
	t.Logf("log events:\n%s", strings.Join(o, "\n"))
}
