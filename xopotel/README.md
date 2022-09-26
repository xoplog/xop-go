
Package xopotel provides gateways/connections between 
[xop](https://github.com/muir/xop-go) 
[OpenTelemetry](https://opentelemetry.io/).

## `SpanLog`

`SpanLog()` allows xop to be used as a logger to add "Events" to an 
existing OTEL span.

## `BaseLogger`

`BaseLogger()` creates a `xopbase.Logger` that can be used as a
to gateway xop logs and spans into OTEL.

## Compatibility

### Data types

OTEL supports far fewer data types than xop.  Mostly, xop types
can be converted cleanly, but links are a special case: links can
only be added to OTEL spans when the span is created.  Since xop
allows links to be made at any time, links will be added as
ephemeral sub-spans.  Distinct, Multiple, and Locked attributes will
be ignored for links.

OTEL does not support unsigned ints. `uint64` will be converted to a
string and smaller unsigned ints will convert to `int64`.

OTEL provides no way to record links (trace_id/span_id) except as part
of the initial set of "Links" associated with a span.  Those links are
values only (no key, no name).  Xop supports arbitrary numbers of links
on a per-logline basis.  There is no way to record those as links using
the OTEL structures so if there is more than one of them, then they'll
be recorded as strings.  Most of the time, a log line won't have more
than one trace reference.

For both events with links, and links as span metadata, an extra OTEL
span will be created just to hold the link.  The extra span will be marked
with span.is-link-attribute:true for span metadata and 
span.is-link-event for events.

