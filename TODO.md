
# Ideas to ponder

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

- Use ZSON for main transport?  https://www.brimdata.io/blog/unmarshal-interface/
 Use BSON for main transport?

- Hmmm, if entire trace is a bytes string, it can be treated as  StrAttribute...

- Allow base loggers to share resources (span attribute tracking, JSON formatting)

- Figure out a way to enforce limited vocabulary on select Attributes and Data elements

# Just do it (build ready)

- For enums from OTEL, generate them from the protobuf constants.

- Drop the makefile in favor of more sophisticated use of go:generate See example of enumer in zitadel/oidc

- Round out the kinds of things that can be logged:

 - Tables
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

- Rename "xoplog" -> "xop"

- All todos in code

- Add tests:
	Span attributes
	Structued lines

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

# Not build ready 

1. Base loggers:

 - Console to io.Writer
 - JSON to io.Writer
 - Gateway into Jaeger
 - Gateway into Open Telementry
 - Stream send to server (need to write server !)
 - Gateways to other loggers so that xoplog can be used by libraries:

1. Metrics

 - Re-use attribute naming?
 - Allow "tags" or some other multi-dimensional naming

1. Events

 - Gets counted
 -  Re-use attributes?
 - Attach arbitrary data

