
# Output formats

As a bi-level logger, there are multiple base-logger implementations for Xop.

Some of the base loggers retain the full information that is given to them.  That
include xoptest (models), xopjson, xopconsole, and xoppb.  Some lose information: 
xopjs, xopcon, xoptest (text output, xopotel).  All (except xoptest test output) can use
their output steam to call other base loggers and thus all formats can convert
to all other formats.

## xopjson

JSON is a common output format. Unfortunately, JSON does not distinquish between 
floats and integers. It also doesn't have type signfiers that would allow other
types to be reliably tracked.

When using JSON, there is a choice (possibly many choices) between natural expression
of values and preserving full information.

When preserving full information, there is a tradeoff between succinct representations
and verbose ones.

Xopjson format preserves full information at the cost of being more difficult to read.

## xopjs (NOT YET IMPLEMENTED)

Xopjs is an alternative JSON format that does not preserve all information.  Unlike
xopjson, xoptest, xopotel, and xoppb, you cannot convert from Xopjs to other formats.
Other formats can convert to Xopjs though, and that makes xopjs a useful format to
consume. It is the most "natural" encoding of the values.

## xoptest

Xoptest is meant for use inside tests.  It logs to a `testing.T` using `t.Log()`.  The
text output format is lossy (it cannot be converted to other Xop output formats).  In
addition to the text output, xoptest also builds a data structure that captures all of
the test output.  This structure is not lossy.

## xopotel

Xopotel exports Xop logs as Open Telemetry spans and traces.  Log lines become 
Span Events.

## xopcon (NOT YET IMPLEMENTED)

Console output will lose precision in multiple ways: the time format is at second-level
granularity rather than nanosecond.  Type information will be lost.  If the message has
things that look like key/value pairs, they cannot be distinquished from actual key/value
pairs.

## xopconsole (NOT YET IMPLEMENTED)

Xopconsole is a high-fidelity variant of xopcon.

## xoppb 

Xoppb is Xop's native protobuf base logger encoding. 

