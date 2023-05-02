
# Big picture

- TIMES
  - drop precision to microsecond?
  - add timezone offset?

- rip out current redaction functions.  They aren't right but it's not time
  to fix them

- Change logging API

  - note that `.Table()` will also be an alternative to `.Msg()`

- (Re)define/document output formats

  - xopjs (focuses on human-readability)

- Document the above and also

  - xoptest
  - xopotel

- Build the above

- Rebuild xopup uploader to use the new xopproto

- Rebuild server to ingest the new xopproto

# Just do it (build ready)


- check for unchecked casts 

- xopresource
  combine line.Data & line.DataType to have a sub-object
  
- reshuffle levels to be more OTEL-like, add a couple more.
  - trace becomes lower level than debug (ugh!)
  - log? (and then change seed.Log() to seed.Logger()?)
  - verbose?
  - flow?

- all "Xop" -> "XOP"

- simplify code in xopcon 

- share implementation between xopcon and xoprecorder (opportunities for DRY)

- add stack filters / filename formatters

- rename reactive to "hook"

- missing tests (in set of cases)

  - incoming baggage is set
  - coming tracestate is set
  - incoming trace id is set, verify later in test

- add buildinfo from runtime/debug to SourceInfo

- rebuild attribute defintions to use generics `type Int32Attribute Attribute[int64][int32]`?

- remove example values from attribute defintions (also removes TypeName)

- add marshal functions to HexBytes and Trace: JSON & sql

- pass a pre-made encoder to model.Encode()

- switch over to using the new atomic types rather than pointers to int32/int64...

- Remove example values from attribute registrations

- Make sure there aren't races for defintion of span attributes at
  the request or span level.  Do base loggers need to be able to handle
  multiple requests at the same time?

- sampling can be based on Boring() in which case the flags need to
  change before the "traceresponse" is set.  That means top logger
  must know if base loggers honored the boring.  Change xopbase.Boring
  to return a bool.  Propagate upwards.  Add log.IsBoring()

- add flag to honor sampling flag by defaulting to Boring

- change Boring to be request-level only

- grab and modify the the zerolog linter to make sure that log lines don't get dropped

- make deepcopy function configurable

- move xoputil to internal/xoputil -- at least for now since
  the APIs in xoputil are less stable than the rest of the code

- For enums from OTEL, generate them from the protobuf constants.

- Drop the makefile in favor of more sophisticated use of go:generate See example of enumer in zitadel/oidc.  Well, keep the makefile since coverage is complicated.

- Round out the kinds of things that can be logged:

  - Tables 

    ```go
    type Table interface{
        Header() []string
        Rows() [][]string
    }
    ```

  - Pre-encoded data
  - Add Object(func(*BaseType))
  - Add Pairs(k, v, k, v, ...)
  - URLs

- Round out the types of attributes

  - uint
  - table
  - url

- Write panic trappers.  Should log the panic and flush.

  ```
  defer TrapPanic(log).ReturnError(&err)
  defer TrapPanic(log).Rethrow()
  ```

- gateway loggers:

  Each should have a single func that bundles the baselogger into
  a logger.  The idea being to make xop a good choice for libraries
  that may not know what the project logger is.

  - gateway to standard "log"
  - gateway to logur
  - gateway to onelog
  - gateway to zerolog
  - gateway to zap

- emulation

  To help the transition to xop, write emulation packages
  that provide an interface on top of zop that looks like other popular 
  loggers

  `log.Zap().Info("msg", xopzap.String("uri", uri))`

  - zap
  - zap sugar
  - onelog
  - zerolog
  - logur
  - standard "log"

- Base loggers

  - Console (emphasis on readable, but still retains full data)
  - an additional OTEL API that directly creates []sdktrace.ReadOnlySpan
  - split xoptest into a combination of the new introspective base logger (introspect?) and the human-friendly xopconsole?

- xopup

  - enforce memory limit
  - enforce outstanding packet limit

- xopjson

  - additional features:

    - change time of timestamp key
    - allow custom error formats
    - allow go-routine id to be logged
    - allow int64 to switch to string encoding when >2**50
    - do something useful with Boring

  - write to xop server

    - keep dictionary of static strings

      - Enum descriptions and types

  - xopbytes.BytesWriter

    - write to writing rotating files based time or size

- xopotel

  - add wrapper to use otel exporter directly as xopbase
  - add wrapper to use xopbase as otel exporter
  - do something useful with Boring

- Performmance

  - replace locks with atomics where possible
  - replace call to time.Format in xopjson with custom code
  - switch to a faster JSON encoder when encoding arbitrary data
  - too much of xopup is single-threaded

    - switch almost everything to atomics instead
    - shard the uploader on request id

  - mark all places in the code where an allocation happens `// allocate`
  - Use sync.Pool aggressively to reduce allocations

    - xop.Logger

  - Replace flushers map with a flushers slice

  - In xopjson

    - use a priority queue instead of multiple channels for sending stuff to the writer
    - improve performance of time.RFC3339Nano encoding

  - Improve upon sync.Pool

     Rebuild https://github.com/Workiva/go-datastructures/blob/master/queue/ring.go to be
     "LossyPool" that queues upto 32 items (tossing any extra) and returning immedately if
     there aren't any available.  Use generics so no casting is needed.

  - how to make protobuf faster (when building OTEL compatibility):
    [notes](https://blog.najaryan.net/posts/partial-protobuf-encoding/?s=09)

  - can *Sub be Sub instead?  Would that have better performance? 

  - preallocate blocks of Attributes

- Standard tracing

  - figure out a way to modify trace State
  - add methods to query trace State
  - figure out a way to modify trace Baggage
  - add methods to query trace Baggage

- Provide structtags-based redaction function

  - Make something based on github.com/mohae/deepcopy that returns two
    copies of the original.  One that is redacted and one that is not?

# Ideas to ponder

- new "LeveledLog" type that combines Log & Line, pre-baking the level, defined as
  interface so that documentation isn't crazy

- would it be okay to have a dependency on OTEL?  It would help for TraceState and it
  would also be nice to directly reference SpanKind()

- only deliver span metadata at Flush() time?  That could get rid of code to handle
  multiple and distinct in the various bottom loggers

  Alternatively, have a wrapper package that bottom loggers can use to manage these

- use [goccy json](https://github.com/goccy/go-json#benchmarks)?

- could we drop xopbase.Logger.ID() in favor of using pointers?   or change ID sequential integers?  Add Name()?

- is Attribute.ExampleValue() useful?  It's not enforcable at compile time for Any.

- rename LogLine/Line to Entry?

- Consider replacing SeedModifier with a seed modifier type...

  ```
	Instead of
		log.Step(...SeedModifier)
		and functions that produce SeedModifier
	have:
		log.Step(func(SeedModifier))
		and methods on SeedModifier that directly write to the seed
  ```

- Should xop attributes in a line be bundled into a sub-object?

- Should Template() be a line starter rather than an ender?

- Use ZSON for main transport?  
  https://www.brimdata.io/blog/unmarshal-interface/
  Use BSON for main transport?

- Hmmm, if entire trace is a bytes string, it can be treated as  StrAttribute...

- Allow base loggers to share resources (span attribute tracking, JSON formatting)

- Figure out a way to enforce limited vocabulary on select Attributes and Data elements

- Some way to mix-in data from code context.  The log object will be in the call stack
  either passed explicityly or passed within context.  If the call is to a method on an
  initialized library, that library may want to wrap logging to:

  - adjust the logging level that is being retained
  - add it's own line prefix attributes

  What's the right API for this?

- Logging levels.  Are they actually useful?  Let's pretend they are.  What's the right
  behavior.  Baseloggers could choose to ignore levels that are too low.  The logger itself
  could choose to discard when the level is too low.

- support microsoft correlation vectors?  https://github.com/Microsoft/CorrelationVector-Go

- How to do full-text (short-term) indexing of log lines

  - Elastic
  - [zinc](https://github.com/zinclabs/zinc)

- Buffer logs to be included or destroyed depending on if a later event
  is logged at higher priority

  - See https://github.com/golang/go/discussions/54763#discussioncomment-4119965
  - Maybe span level where the span is discarded if no event is >= level?

# Not build ready 

- find a better way to discover own version.  debug.BuildInfo doesn't seem to work

- benchmarking

  Aside from building basic benchmarking (separate item)...  

  - figure out how much time is spent for `time.Now()`
  - figure out how much time is spent formatting timestamps

- send the spanSequenceCode as a xop: attribute

- Base loggers:

  Not ready because APIs should be more solid first

- Bytes writers 

  - Gateway into Jaeger
  - Gateway into Open Telementry

    It can have a bunch more knobs and controls

    - add ability to maintain and send dictionaries to compress logs

      - keys for key/value

  - send to server (need to design server)

- Metrics

  - An "Event" is something that can be counted.

    - An event can make delta adjustments to multiple gauges 

    - An event is specified with :

      - indentifiers that are specific to the event (at least one "name" is required)
      - stable identifiers that come from the seed and are common to all events
      - variables that provide sub-categorization (specified with struct)

    - The gauge adjustment and the event-specific identifiers and variables are
      all specified with struct tags.  The event is a struct.

    - Events only happen within a span and include a link back to the span

  - A "Value" is something that is scraped from an external source or is computed

    - it can be auto-scraped on a timer
