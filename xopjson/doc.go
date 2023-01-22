/*
xopjson is a xop base logger (xopbase.Logger) that encodes in JSON format.

The format of the encoded JSON is a stream of mixed objects.

Depending on options, the format of lines can vary significantly.

Lines

The JSON format of a line with WithAttributesInObject(false) is like:

	{
		"lvl": 9,
		"ts": 49394393493,
		"span.id": "34ec0b8ac9d65e91",
		"stack": [
			"some/file.go:382",
			"some/other/file.go:102"
		],
		"prefilled": "prefilled attributes come first",
		"user_attribute": "all in the main object",
		"another_user_attribute": "line specific attributes come next",
		"msg": "text given to the .Msg() method prepended with PrefillText"
	}

The JSON format of a line with WithAttributesInObject(false) is like:

	{
		"lvl": 9,
		"ts": 49394393493,
		"span.id": "34ec0b8ac9d65e91",
		"stack": [
			"some/file.go:382",
			"some/other/file.go:102"
		],
		"attributes": {
			"prefilled": "prefilled attributes are part of the attributes block",
			"user_attribute": "all in the main object",
			"another_user_attribute": "line specific attributes come next"
		},
		"msg": "text given to the .Msg() method prepended with PrefillText"
	}

"lvl"`is the xopconst.Level number

"ts" is a timestamp and it can be formatted various ways depending
on the options used to create the xopjson.Logger.  The default
format is an integer representing microseconds since Jan 1 1970.

"stack" will only be included if the logger options include sending stack frames.  By
default stack frames are included with Error and Alert level logs.

"span.id" is included when WithSpanTags(SpanIDTagOption) is used.

"fmt":"tmpl" is included when the format of the line is a template (logged
with .Template())

Spans

	{
		"type": "span",
		"name": "name provided by user creating span",
		"trace_header": "01-8a84c99x8230x29d8a84c99x8230x29d-8a84c99x8230x29d-00"
		"span.id":
		"dur": 82239222902,
		"span.ver": 0,
	}

Span.ver starts at zero.  The same span can be included in output more than once.  Each
time the span is serialized, span.ver is incremented.  When a field is included in span
output, it replaces any previous value.  Only changed fields are guaranteed to be output
(with the exception of "type" and "span.id")

"dur" will be included whenever span.ver is not zero.

Requests

	{
		"type": "span",
		"impl": "zop-go",
		"name": "name provided by user creating span",
		"request.id": "01-8a84c99x8230x29d8a84c99x8230x29d-8a84c99x8230x29d-00",
		"parent.id": "01-8a84c99x8230x29d8a84c99x8230x29d-8a84c99x8230x29d-00",
		"trace.state": "vendor:key vendor2:key2",
		"trace.baggage": "key:values,value key2:value1,value2"
		"dur": 822392,
	}

Bufferes

If WithBufferedLines is non-zero, then each buffer will being with a
record like:

	{
		"zop": {
			"type": "buffer_header",
			"seq_no": 38
		}
	}

OversizeBuffers

If WithBufferedLines is non-zero and the an there is too much data
too send because the buffer overflowed for one reason or another than
overflow records will be written.  These look like:

	{
		"xop": {
			"type": "buffer_header",
			"seq_no": 38,
			"part": 2,
			"parts_in_buffer": 3
		}
	}

*/
package xopjson
