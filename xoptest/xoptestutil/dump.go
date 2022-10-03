package xoptestutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/muir/xop-go/xoptest"
)

func DumpEvents(t testing.TB, tlog *xoptest.TestLogger) {
	var o []string
	_ = tlog.WithLock(func(tlog *xoptest.TestLogger) error {
		for _, event := range tlog.Events {
			switch event.Type {
			case xoptest.LineEvent:
				o = append(o, fmt.Sprintf("line: %s", event.Line.Text))
			case xoptest.SpanStart:
				o = append(o, fmt.Sprintf("spanstart: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoptest.SpanDone:
				o = append(o, fmt.Sprintf("spandone: %s", event.Span.Bundle.Trace.SpanID().String()))

			case xoptest.RequestStart:
				o = append(o, fmt.Sprintf("requeststart: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoptest.RequestDone:
				o = append(o, fmt.Sprintf("requestdone: %s", event.Span.Bundle.Trace.SpanID().String()))
			case xoptest.FlushEvent:
				o = append(o, "Flush!")
			case xoptest.CustomEvent:
				o = append(o, "Custom: "+event.Msg)
			case xoptest.MetadataSet:
				o = append(o, fmt.Sprintf("Metadata on %s: %s", event.Span.Bundle.Trace.SpanID().String(), event.Msg))
			default:
				o = append(o, "unknown event")
			}
		}
		return nil
	})
	t.Logf("log events:\n%s", strings.Join(o, "\n"))
}
