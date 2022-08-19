
# Ideas to ponder

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

# Just do it (build ready)

- move xoputil to internal/xoputil -- at least for now since
  the APIs in xoputil are less stable than the rest of the code

- Better receiver names

  - change `func (l *Log)` to `func (log *Log)` everywhere
  - change `func (s *Sub)` to `func (sub *Sub)` everywhere
  - change `func (ll *LogLine)` to `func (line *Line)` everywhere
  - change `func (s *Span)` to `func (span *Span)` everywhere
  - change `func (s Seed)` to `func (seed Seed)` everywhere

- For enums from OTEL, generate them from the protobuf constants.

- Drop the makefile in favor of more sophisticated use of go:generate See example of enumer in zitadel/oidc

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

- Write panic trappers.  Should log the panic and flush.

  ```
  defer TrapPanic().ReturnError(&err)
  defer TrapPanic().Rethrow()
  ```

- Use generics to make type-safe Enum span attributes: can't do it as a method on Span, but can
  do it as a function.

- All todos in code

- Add tests:

  - subspans
  - Span attributes
  - Structued lines
  - JSON logger

- gateway loggers:

  Each should have a single func that bundles the baselogger into
  a logger.  The idea being to make xop a good choice for libraries
  that may not know what the project logger is.

  - gateway to standard "log"
  - gateway to logur
  - gateway to onelog
  - gateway to zerolog
  - gateway to zap
  - gateway to phuslog

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

- Bytes writers 

  - to console
  - to io.Writer (same as console?)

- Add CI

  Gitlab actions.  

  - Example in libschema for combining coverage from multiple source

- Performmance

  - mark all places in the code where an allocation happens `// allocate`
  - Use sync.Pool agressively to reduce allocations

    - xop.Logger

  - Replace flushers map with a flushers slice
  - In xopjson

    - track which spans have been modified
    - pre-allocate span array

  - Improve upon sync.Pool

     Rebuild https://github.com/Workiva/go-datastructures/blob/master/queue/ring.go to be
     "LossyPool" that queues upto 32 items (tossing any extra) and returning immedately if
     there aren't any availble.  Use generics so no casting is needed.

  - AttributeBuilder needs a JSON-specific version

    - per-key buffers
    - 64 buffers?
    - 64 bytes each?

  - how to make protobuf faster (when building OTEL compatability):
    [notes](https://blog.najaryan.net/posts/partial-protobuf-encoding/?s=09)

  - can *Sub be Sub instead?  Would that have better performance? 

  - the Flush() and Done() code ends locking and releasing the same locks 
    multiple times.  Refactor to put some parts of the work in functions that
    must be called while locked so that fewer lock/release cycles. 200ns per
    lock/release...

- Standard tracing

  - figure out a way to modify trace State
  - add methods to query trace State
  - figure out a way to modify trace Baggage
  - add methods to query trace Baggage

# Not build ready 

- benchmarking

  Aside from building basic benchmarking (separate item)...  

  - figure out how much time is spent for `time.Now()`
  - figure out how much time is spent formatting timestamps

- send the spanSequenceCode as a xop: attribute

- Base loggers:

  Not ready because APIs should be more solid first

  - Gateway into Jaeger
  - Gateway into Open Telementry
  - Improve the JSON logger

    It can have a bunch more knobs and controls

    - add additional time formats
    - add ability to maintain and send dictionaries to compress logs

      - keys for key/value
      - span ids (use the spanSequenceCode)
      - log line prefix sets

- Bytes writers:

  - send to server (need to design server)

- Metrics

  - Re-use attribute naming?
  - Allow "tags" or some other multi-dimensional naming

- Events

  - Gets counted
  - Re-use attributes?
  - Attach arbitrary data


