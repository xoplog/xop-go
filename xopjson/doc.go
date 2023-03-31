/*
xopjson is a xop base logger (xopbase.Logger) that encodes in JSON format.

The format of the encoded JSON is a stream of mixed objects.

Depending on options, the format of lines can vary significantly.

# Lines

The JSON format of a line with WithAttributesInObject(false) is like (actual encoding w/o whitespace):

	{
		"lvl": "alert",
		"ts": "2023-03-30T21:27:36.901822-07:00",
		"stack": [
			"/Users/sharnoff/src/github.com/muir/xop-go/xopjson/jsonlogger_test.go:79",
			"/usr/local/Cellar/go/1.20.1/libexec/src/testing/testing.go:1576"
		],
		"span.id":"e006cc70e2453480",
		"attributes": {
			"foo": {"v":"bar","t":"s"},
			"blast": {"v":99,"t":"i"}
		},
		"msg":"a test line"
	}

"lvl"`is the xopconst.Level logging level

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
			"span.name": "a fork one span",
			"ts": "2023-03-30T21:27:36.902446-07:00",
			"span.parent_span": "70adac21637a869d",
			"span.id": "193586833ecbd336",
			"span.seq": ".A",
			"span.ver":1,
			"dur": 216000,
			"attributes": {
				"http.route":"/some/thing"
			}
		}

Span.ver starts at zero.  The same span can be included in output more than once.  Each
time the span is serialized, span.ver is incremented.  When a field is included in span
output, it replaces any previous value.  Only changed fields are guaranteed to be output
(with the exception of "type" and "span.id")

"dur" will be included whenever span.ver is not zero.

Requests

	        {
			"type": "request",
			"trace.id": "045fbbb27fab63e80bdef127c35e9abe",
			"span.id": "70adac21637a869d",
			"span.name": "TestReplayJSON/unbuffered/attributes/one-span",
			"ts": "2023-03-30T21:27:36.902242-07:00",
			"source": "xopjson.test 0.0.0",
			"ns": "xopjson.test 0.0.0",
			"span.ver": 1,
			"dur": 403000
		}

# Attribute Definitions

Attributes on spans and requests are defined before first use.

	        {
			"type": "defineKey",
			"key": "some-boolean-value",
			"desc": "an example used in a test",
			"ns": "test 0.0.0",
			"prom": 0,
			"locked": true,
			"vtype": "Bool",
			"span.id":"9291b0d415db0d3c"
		}
*/
package xopjson
