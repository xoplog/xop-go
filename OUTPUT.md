
# Output formats

As a bi-level logger, there are multiple base-logger implementations for Xop.

Some of the base loggers retain the full information that is given to them.  That
include xoptest (models), xopjson, xopconsole, and xoppb.  Some lose information: 
xopjs, xopcon, xoptest (text output, xopotel).  All (except xoptest test output) can use
their output steam to call other base loggers and thus all formats can convert
to all other formats.

## xopjson

JSON is a common output format.  Unfortunately, JSON does not distinquish between 
floats and integers.  It also doesn't have type signfiers that would allow other
types to be reliably tracked.

When using JSON, there is a choice (possibly many choices) between natural expression
of values and preserving full information.

When preserving full information, there is a tradeoff between succinct representations
and verbose ones.

Xopjson format preserves full information at the cost of being more difficult to read.

## xopjs 

Xopjs is an alternative JSON format that does not preserve all information.  Unlike
xopjson, xoptest, xopotel, and xoppb, you cannot convert from Xopjs to other formats.
Other formats can convert to Xopjs though, and that makes xopjs a useful format to
consume.  It is the most "natural" encoding of the values.

## xoptest

Xoptest is meant for use inside tests.  It logs to a `testing.T` using `t.Log()`.  The
text output format is lossy (it cannot be converted to other Xop output formats).  In
addition to the text output, xoptest also builds a data structure that captures all of
the test output.  This structure is not lossy.

## xopotel

Xopotel exports Xop logs as Open Telemetry spans and traces.  Log lines become 
Span events.

## xopcon

TO-BE-BUILT

Xopcon has not yet been built.  It is the planned console format.

Console output will lose precision in multiple ways: the time format is at second-level
granularity rather than nanosecond.  Type information will be lost.  If the message has
things that look like key/value pairs, they cannot be distinquished from actual key/value
pairs.

## xopconsole

Xopconsole is a high-fidelity variant of xopcon.

## xoppb

TO-BE-BUILT

# Structure encoding

## Lines

| format | example |
| --- | ---- |
| xopjson | `{"lvl":8,"ts":"2023-01-13T22:05:40.873462-08:00","span.id":"38c3ce6b70fb2468","attributes":{"do":"foobar"},"msg":"yes, foobar"}` |
| xopjs | `{"lvl":"INFO","ts":"2023-01-13T22:05:40.873462-08:00","span.id":"38c3ce6b70fb2468","do":"foobar","msg":"yes, foobar"}` |
| xoptest (text) | `testlogger.go:420: T1.1: foo garden=nothing in my garden is taller than my daisy! story=I got the contract with a small consideration, just a sunflower cookie tale=I got the contract with a small consideration, just a sunflower cookie raw=I got the contract with a small bribe, just a sunflower cookie success=I got the contract with a small bribe, just a daisy cookie oops=outer: inner(as string) ` |
| xoptest (model) | see [Line](https://pkg.go.dev/github.com/xoplog/xop-go/xoptest#Line) |
| xopcon | `2009/11/10 23:00:00 INFO T1.1: message goes here with key=value after=text` |
| xopconsole | `2009/11/10T23:00:00.00329Z INFO 38c3ce6b70fb2468 "message goes here with" "key"="S:value" "after"="text"` |
| xopotel | added as `Event` to a [Span](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Span) |
| xoppb | see Line in [xopproto](https://github.com/muir/xopproto) |

### Stack traces

### Models

### Redacted values

### Links (line)

| format | example |
| --- | ---- |
| xopjson | |
| xopjs | |
| xoptest (text) | |
| xoptest (model) | |
| xopcon | |
| xopconsole | |
| xopotel | (1) Event with `attribute.String("xop.message", prefill+linkText), attribute.String("xop.type", "link")`; and (2) sub-span with attribute.String("xop.type", "link-event"); and (3) an Event in the sub-span with the line attributes repeated |

`attribute.StringSlice(key, []string{"enum", "namespace", "namespace-version", "string-value", "32"])` |
| xoppb | |
## Spans

A primary difference between lines and spans is that the attribute keys
for spans must be pre-registered but for lines, any old string will do.

## Span Attribute

| format | example |
| --- | ---- |
| xopjson | buffred for a resend of the span |
| xopjs | buffred for a resend of the span |
| xoptest (text) | |
| xoptest (model) | see [Span](https://pkg.go.dev/github.com/xoplog/xop-go/xoptest#Span) |
| xopcon | `2009/11/10 23:00:00 T1.1: key=value` |
| xopconsole | `2009/11/10T23:00:00.00329Z 38c3ce6b70fb2468 "key"="value" |
| xopotel | |
| xoppb | |

## Span flush  

## Requests

| format | example |
| --- | ---- |
| xopjson | `{"type":"request","trace.id":"acc816d3be4a88fa70e294030e84c9bf","span.id":"38c3ce6b70fb2468","span.name":"TestParameters/unsynced/log_levels","ts":"2023-01-13T22:05:40.873395-08:00","span.ver":0,"dur":128000}` |
| xopjs | same as xopjson except for attributes |
| xoptest (text) | `testlogger.go:214: Start request T1.1=00-c23eb18b4a69db30a93ad5a450b562c2-97b3fac6fa50ec6b-01 TestRedaction` |
| xoptest (model) | see Span in [xoptest](https://github.com/xoplog/xop-go |
| xopcon | `2009/11/10 23:00:00 Start request T1.1=00-c23eb18b4a69db30a93ad5a450b562c2-97b3fac6fa50ec6b-01 GET /foo` |
| xopconsole | `2009/11/10T23:00:00.2822Z Start request 00-c23eb18b4a69db30a93ad5a450b562c2-97b3fac6fa50ec6b-01 GET /foo` |
| xopotel | [Span](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Span) |
| xoppb | see Request and Span in [xopproto](https://github.com/muir/xopproto) |

# Attribute encoding

## Strings

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":"S:value(JSON-escaping)"` with the `S:` being optional most of the time, required if matching `^(?:\d\|[SUITD]:\|[a-f][\da-f]-)` |
| xopjs | `"key(JSON-escaping)":"value(JSON-escaping)"` |
| xoptest (text) | `key=value` |
| xoptest (model) | `line.Data["key"] = "value"` and `line.DataType["key"] = xopbase.StringDataType` |
| xopcon | `key="value"` or `"k e y"="value"` (if there are spaces in the key) |
| xopconsole | `"key"="S:value"` (same rules as xopjson for values) |
| xopotel | `attribute.String(key, value)` |
| xoppb | in `Line`, `map<string,string> stringAttributes = 5` |

## Bools

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":true` |
| xopjs | `"key(JSON-escaping)":true` |
| xoptest (text) | `key=true` |
| xoptest (model) | `line.Data["key"] = true` and `line.DataType["key"] = xopbase.BoolDataType` |
| xopcon | `key=true` or `"k e y"=true` (if there are spaces in the key) |
| xopotel | `attribute.Bool(key, value)` |
| xoppb | |

## Integers

JSON cannot round-trip integers > 2**52 or < -2**52.  There is no way, in JSON, to tell
the difference between floats that have no fractional part and integers.

If integers will stored in strings, there is no way to tell that they're integers.

| format | example |
| --- | ---- |
| xopjson (signed) | `"key(JSON-escaping)":"I:83823282923" |
| xopjson (unsigned) | `"key(JSON-escaping)":"U:83823282923"} |
| xopjs | `"key(JSON-escaping)":83823282923} |
| xoptest (text) | `key=83823282923` |
| xoptest (model) | `line.Data["key"] = 83823282923` and `line.DataType["key"] = xopbase.Int32DataType` |
| xopcon | `key=83823282923` or `"k e y"=83823282923` (if there are spaces in the key) |
| xopotel | `attribute.Int64(key, value)` |
| xoppb (signed) | | 
| xoppb (unsigned) | |

## Floats

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":328.4 |
| xopjs | `"key(JSON-escaping)":328.4 |
| xoptest (text) | `key=328.4` |
| xoptest (model) | `line.Data["key"] = 328.4` and `line.DataType["key"] = xopbase.Float64DataType` |
| xopcon | `key=328.4` or `"k e y"=328.4` (if there are spaces in the key) |
| xopotel | `attribute.Float64(key, value)` |
| xoppb (signed) | | 
| xoppb (unsigned) | |

## Time 

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":"T:2023-01-14T23:17:14.414054-08:00"` |
| xopjs | `"key(JSON-escaping)":"2023-01-14T23:17:14.414054-08:00"` |
| xoptest (text) | `key=2023-01-14T23:17:14.414054-08:00` |
| xoptest (model) | `line.Data["key"] = time.MustParse("2023-01-14T23:17:14.414054-08:00") and `line.DataType["key"] = xopbase.TimeDataType` |
| xopcon | `key=2023-01-14T23:17:14.414054-08:00` or `"k e y"=2023-01-14T23:17:14.414054-08:00` (if there are spaces in the key) |
| xopotel | `attribute.StringSlice(key, []string{"time", "2023-01-14T23:17:14.414054-08:00"])` |
| xoppb (signed) | | 
| xoppb (unsigned) | |

## Duration

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":"5m10s" |
| xopjs | `"key(JSON-escaping)":"5m10s" |
| xoptest (text) | `key=5m10s` |
| xoptest (model) | `line.Data["key"] = time.Minute*5+10*time.Second` and `line.DataType["key"] = xopbase.DurationDataType` |
| xopcon | `key=5m10s` or `"k e y"=5m10s` (if there are spaces in the key) |
| xopotel | `attribute.StringSlice(key, []string{"duration", "5m10s"])` |
| xoppb (signed) | | 
| xoppb (unsigned) | |

## Objects

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":{"type":"thing","value":{object here}}` |
| xopjs | `"key(JSON-escaping)":{object here}` |
| xoptest (text) | `key={object here}` |
| xoptest (model) | `line.Data["key"] = deepcopy.Copy(object)` and `line.DataType["key"] = deepcopy.Copy(object)` |
| xopcon | `key={object here}` or `"k e y"={object here}` (if there are spaces in the key) |
| xopotel | `attribute.StringSlice(key, []string{"object", "thing", "application/json", "{object here}"])` |
| xoppb (signed) | | 
| xoppb (unsigned) | |

## Enums

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":{"kind":"enum","namespace":"namespace-namspace-version","string":"string-value","int":328}` |
| xopjs | `"key(JSON-escaping)":"string-value"` |
| xoptest (text) | `key=enum_string_value` |
| xoptest (model) | `line.Data["key"] = XXX` and `line.DataType["key"] = XXX` |
| xopcon | `key=enum_string_value` or `"k e y"=enum_string_value` (if there are spaces in the key) |
| xopconsole | `key="string-value"(32,namespace,namespace-version)
| xopotel | `attribute.StringSlice(key, []string{"enum", "namespace", "namespace-version", "string-value", "32"])` |
| xoppb (signed) | |
| xoppb (unsigned) | |
