
Package xopotel provides a gateway from xop into open telemetry
using OTEL's top-level APIs.

This gateway can be used either as a base layer for xop allowing
xop to output through OTEL; or it can be used to bridge the gap
between an application that is otherwise using OTEL and a library
that expects to be provided with a xop logger.

OTEL supports far fewer data types than xop.  Mostly, xop types
can be converted cleanly, but links are a special case: links can
only be added to OTEL spans when the span is created.  Since xop
allows links to be made at any time, links will be added as
ephemeral sub-spans.  Distinct, Multiple, and Locked attributes will
be ignored for links.

OTEL does not support unsigned ints. `uint64` will be converted to a
string and smaller unsigned ints will convert to `int64`.
