
# Output formats

As a bi-level logger, there are multiple base-logger implementations for Xop.

Except where noted, all output formats can be converted to other output formats.  There are some
exceptions in terms of the keys used: attribute keys that are not words may not convert between
formats.

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
| xopcon | `key="value"` or `"k e y"="value"` (if there are spaces in the key) |
| xopotel | added as `Event` to a [Span](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Span) |
| xoppb | see Line in [xopproto](https://github.com/muir/xopproto) |

### Stack traces

### Models

### Redacted values

### Links

## Spans

A primary difference between lines and spans is that the attribute keys
for spans must be pre-registered but for lines, any old string will do.

## Requests

| format | example |
| --- | ---- |
| xopjson | `{"type":"request","trace.id":"acc816d3be4a88fa70e294030e84c9bf","span.id":"38c3ce6b70fb2468","span.name":"TestParameters/unsynced/log_levels","ts":"2023-01-13T22:05:40.873395-08:00","span.ver":0,"dur":128000}` |
| xopjs | same as xopjson except for attributes |
| xoptest (text) | `testlogger.go:214: Start request T1.1=00-c23eb18b4a69db30a93ad5a450b562c2-97b3fac6fa50ec6b-01 TestRedaction` |
| xoptest (model) | see Span in [xoptest](https://github.com/xoplog/xop-go |
| xopcon | `2009/11/10 23:00:00 Start request T1.1=00-c23eb18b4a69db30a93ad5a450b562c2-97b3fac6fa50ec6b-01 GET /foo` |
| xopotel | [Span](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Span) |
| xoppb | see Request and Span in [xopproto](https://github.com/muir/xopproto) |

# Attribute encoding

## Strings

| format | example |
| --- | ---- |
| xopjson | `"key(JSON-escaping)":"S:value(JSON-escaping)"` with the `S:` being optional most of the time, required if matching `^(?:\d|[A-Z]:|[a-f][\da-f]-)` |
| xopjs | `"key(JSON-escaping)":"value(JSON-escaping)"` |
| xoptest (text) | `key=value` |
| xoptest (model) | `line.Data["key"] = "value"` and `line.DataType["key"] = xopbase.StringDataType` |
| xopcon | `key="value"` or `"k e y"="value"` (if there are spaces in the key) |
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

## Enums

