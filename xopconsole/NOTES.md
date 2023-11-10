## General format

```
xop Request 2009-11-10T23:00:00.032832823Z 
```

1. Everything starts with `"xop"`.
1. A space
1. An indicator of what kind of record it is. For lines, this is the logging level, for other things, it is what they are.
   - Request
   - Span
   - Def
   - Trace/Debug/Info/Warn/Error/Alert
1. A space
1. For all but Def:
   1. a timestamp in RFC3339Nano format. 
   1. A space
1. Details, which vary based on what kind of thing it is
   - Request
     1. "Start" or "v"
        - Start
          1. Next up is a console format version number, "1" for now.
          1. A space
          1. trace-header
          1. A space
          1. A quoted-if-needed "name" for the request
          1. A space
          1. A quoted-if-needed Source+version 
          1. A space
          1. A quoted-if-needed Namespace+version
          1. Optional:
             1. A space
             1. "parent:"
             1. parent trace header OR parent span id
          1. Optional:
             1. A space
             1. "state:"
             1. state header, quoted if needed
          1. Optional:
             1. A space
             1. "baggage:"
             1. baggage header, quoted if needed
        - "v"
          1. a record version number that increments each time the request is printed
          1. A space
          1. spanID
          1. Repeating:
             1. A space
             1. A quoted-if-needed key
             1. "="
             1. A quoted-if-needed value
   - Span
     1. "Start" or "v" + a version number
        - "Start"
          1. A space
          1. The SpanID (hex)
          1. A space
          1. The parent SpanID (hex)
          1. A space
          1. A quoted-if-needed "name" for the span
          1. A space
          1. The span sequence abbreviation or "" if none
        - "v"
          1. same as request version updates
     1. A space
   - Def
     1. A space
     1. JSON-encoded attribute definition
   - Trace/Debug/Info/Warn/Error/Alert
     1. A space
     1. The SpanID (hex)
     1. A space
     1. Optional kind indicator
        - Model:
          1. " MODEL:"
        - Link:
          1. " LINK:"
        - Message:
          1. Optional format indicator:
             - Template:
               1. "TEMPLATE:"
             - No template:
     1. quoted-if-needed line message (or template)
     1. Repeating:
        1. A space
        1. a line attribute (see below)
     1. Optional stack:
        1. A space
        1. " STACK:"
        1. Repeating frames
           1. A space
           1. Line
           1. ":"
           1. Number

## Line attributes

The general form for attributes is 

```
key=value(type)
```

Key can either be an unquoted word or it can be a "quoted" string.

Not all values are followed with a type.

Values(type) can be one of the following:

- 3m10s(dur)
  A duration

- 3.72e20(f32)
  A float32

- "somthing with a space"
  A string

- "somthing with a space"(stringer)
  A string, orignally a stringer

- Hi38
  A string without any spaces (and not a duration). It must not
  start with a number, parenthesis, or slash.

- 82392
  An integer, type `int`.  Negative too.

- -328204(i32)
  An integer, any type other than `int`

- f
  A bool, false

- t
  A bool, true

- 2009-11-10T23:00:04.843394034Z(time)
  A time, RFC3339Nano format

- 1/JSON
  An enum, numeric and string values provided

- (11){"a":"foo"}JSON/foobarType
  An "any", a model, encoding specified, length first in parens

## Metadata attributes

key=value, no type indicators

