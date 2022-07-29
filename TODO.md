
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

- Base loggers

  - Console (emphasis on readable, but still retains full data)

- Bytes writers 

  - to console
  - to io.Writer (same as console?)

# Not build ready 

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


