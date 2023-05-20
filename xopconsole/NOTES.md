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
             1. "state:"
             1. state header, quoted if needed
          1. Optional:
             1. A space
             1. "bundle:"
             1. bundle header, quoted if needed
        - "v"
          1. a record version number that increments each time the request is printed
          1. A space
          1. spanID
          1. A space
   - Span
     1. "Start" or "v" + a version number
        - "Start"
          1. A space
          1. The SpanID (hex)
          1. A space
          1. A quoted-if-needed "name" for the span
          1. A space
          1. The span sequence abbreviation or "" if none
        - "v"
          1. a record version number that inrements each time the span record is printed
          1. A space
          1. The SpanID (hex)
     1. A space
   - Def
   - Trace/Debug/Info/Warn/Error/Alert

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
  An integer, type `int`

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

- (11){"a":"foo"}/JSON
  An "any", a model, encoding specified, length first in parens

## Metadata attributes

Mostly uses the same encodings as line attributes. Lists are in the form of
[value,value,value].

