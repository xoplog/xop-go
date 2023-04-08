/*
Packabe basebuffer provides a wrapper around other xopbase.Loggers.
Unlike regular the xopbase.Logger, all aspects of a Request may be
modifed at any time before Flush() is called.  All data is held
until Done() and Flush() have been called.

Further, the Context passed to the underlying xopbase.Logger provides
early access to the buffered data so that it can be consumed out-of-order.

package basebuffer

import (
	"context"
	"time"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"
)

type log struct {
	base xopbase.Logger
}

// XXX var _ xopbase.Logger = &buf{}

func New(base xopbase.Logger) xopbase.Logger {
	return log{base: base}
}

func (log log) ID() string {
	return log.base.ID() + "/buffered"
}
func (log log) ReferencesKept() bool { return true }
func (log log) Buffered() bool       { return true }

type request struct {
	log         *log
	baseRequest xopbase.Request
	span        *span
	sourceInfo  xopbase.SourceInfo
}

type span struct {
	ctx         context.Context
	bundle      xoptrace.Bundle
	startTime   time.Time
	description string
	baseSpan    xopbase.Span
	parent      *span
}

func (log *log) Request(ctx context.Context, ts time.Time, span xoptrace.Bundle, description string, source xopbase.SourceInfo) xopbase.Request {
	req := &request{
		log: log,
		span: &span{
			ctx:         ctx,
			startTime:   ts,
			bundle:      span,
			description: description,
		},
		sourceInfo: source,
	}
	req.span.parent = req
	return req
}

func (s *span) Span(ctx context.Context, ts time.Time, span xoptrace.Bundle, descriptionOrName string, spanSequenceCode string) xopbase.Span {
	return &span{
		parent:      s,
		ctx:         ctx,
		startTime:   ts,
		bundle:      span,
		description: description,
	}
}
*/
package XXX
