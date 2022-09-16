
Package xopotel provides gateways/connections between 
[xop](https://github.com/muir/xop-go) 
[OpenTelemetry](https://opentelemetry.io/).

## `SpanLog`

`SpanLog()` allows xop to be used as a logger to add "Events" to an 
existing OTEL span.

## `BaseLogger`

`BaseLogger()` creates a `xopbase.Logger` that can be used as a
to gateway xop logs and spans into OTEL.

## Compatability

### Data types

OTEL supports far fewer data types than xop.  Mostly, xop types
can be converted cleanly, but links are a special case: links can
only be added to OTEL spans when the span is created.  Since xop
allows links to be made at any time, links will be added as
ephemeral sub-spans.  Distinct, Multiple, and Locked attributes will
be ignored for links.

OTEL does not support unsigned ints. `uint64` will be converted to a
string and smaller unsigned ints will convert to `int64`.

