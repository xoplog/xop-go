# xop - Golang structured log generation combined with tracing (Cross Obserability Platform)

[![Stability: Experimental](https://masterminds.github.io/stability/experimental.svg)](https://masterminds.github.io/stability/experimental.html)
[![GoDoc](https://godoc.org/github.com/xoplog/xop-go?status.png)](https://pkg.go.dev/github.com/xoplog/xop-go)
![unit tests](https://github.com/xoplog/xop-go/actions/workflows/go.yml/badge.svg)
[![report card](https://goreportcard.com/badge/github.com/xoplog/xop-go)](https://goreportcard.com/report/github.com/xoplog/xop-go)
[![codecov](https://codecov.io/gh/xoplog/xop-go/branch/main/graph/badge.svg)](https://codecov.io/gh/xoplog/xop-go)

## Vision

Logging that is chap, rich, in-context, searchable, and long-lived
creates a situation where bugs can be fixed quickly, especially when
combined with with supportive programming practices.

Logs that are in-context allow a story to be told. A story of what
happened and why. This story can be supported by logging data models,
tables, and other references that would allow bugs to be reproduced.

In-context means that you know what triggered something to happen. You
also know what trigggered that prior thing. And what triggered the 
thing before that. This requires logging a lot. It also requires being
able to see only the part of the logs that matter. All logs are within
spans. All spans are within requests. All requests are within traces.

Searchable doesn't mean full-text search, though that could be present.
When a software engineer learns that there is a problem with someting, 
that usually comes with identifiers attached to it. Customer numbers. 
Invoices. Some kinda of identity that lives within the system. These
identifiers can be attributes that are attached to requests. Searchable
logs means that given an identifier or two and some hint of what went wrong,
an engineer can quickly find the exact point, in a log, where the problem
first manifested, and by working backwards, following links in the logs,
find likely cause.

The main supportive coding practice that helps at this point is building
complex things with pure functions. The inputs to pure functions, or at
least ways to regenerate the inputs, can be included in the logs. Once
you have one regression test built from replaying inputs from logs, any
further problems with the same pure function can usually be reporoduced
in minutes.

Combine all these elements and the time to diagnose and reproduce a new
code issue can drop down to a handful of minutes.

The goal of XOP is to provide that richness in a way that is easy to
use, efficient, and easy to integrate into existing code.

## Development status

Ready to use, not yet stable.  Any incompatible changes will be clearly
marked in the release descriptions.  At this point, adoptors should make
themselves known and discuss anything that comes up.

Please open issues to start discussions around a subject.  Please feel free
to open pull requests, especially to added base loggers or propagators.

Expect the following changes as development continues:

- API changes as additional features are added

  Currently xop has no metrics support.  That will change and adding
  metrics will probably be the biggest API change

- Additional gateway base loggers will be written

  To make xop the best logging library for library writers, a full compliment
  of xop -> logger gateways will be written.

  - zap
  - logrus
  - zerolog
  - onelog 

- The full set of OpenTelemetry `semconv` (Semantic Conventions) to be imported
  into `xopconst` (or perhaps somewhere else).

- Performance has been neglected for a while to focus on other things. Performance
  will be a focus again.

## Historical context

Observability code and technique is rapidly evolving.  The 
[Open Telemetry](https://opentelemetry.io/)
project is the focus of most of the energy now. Until Open Telemetry 
releases a Go logger, there still isn't a well integrated logs and traces
package. 

That is beginning to change.  There is now a 
[Zap/OTEL integration](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelzap).

Xop is currently the only Go structured logs and tracing system.  Performance-wise,
it's better that Zap, and was on-par with Zerolog, but recently fell behind.

Where Xop shines is in it's API design.  Xop manages to be very flexible, has
lots of features, is easy to use and has high performance.  Meeting all of those
goals simultaneously made Xop somewhat difficult to build.
Making logging type-safe is difficult because most ways to
accomplish it make logging more diffuclt and more complex. Xop tries
to strike a blance between safety and usability.  Metadata on spans are 
fully type-safe and keywords must be pre-registered.  Data elements on log
lines are mostly type-safe but do not need to be pre-registered.

## Base loggers

Xop is a two-level logger. The top-level logger provides the API for 
logging lines and spans, etc. The bottom-level loggers translate the logs
to different formats.

Some of the bottom-level loggers are "full fidelity" which means that
they are bundled with a function that can consume their own output and
re-log it to a different bottom-level logger thus translating from one
format to another. Xop bottom-level loggers must implement the 
[xopbase](https://pkg.go.dev/github.com/xoplog/xop-go/xopbase) Logger interface.

| name | full fidelity | description |
| -- | -- | -- |
| [xopjson](https://pkg.go.dev/github.com/xoplog/xop-go/xopjson) | yes | JSON output |
| [xopotel](https://pkg.go.dev/github.com/xoplog/xopotel-go) | yes | Output though OpenTelemetry spans (Go logger not available) |
| [xopcon](https://pkg.go.dev/github.com/xoplog/xop-go/xopcon) | no | Console/text logger emphasizing human readability |
| [xopconsole](https://pkg.go.dev/github.com/xoplog/xop-go/xopconsole) | yes | Console/text logger with no information loss |
| [xoppb](https://pkg.go.dev/github.com/xoplog/xop-go/xoppb) | yes | Protobuf output |
| [xoprecorder](https://pkg.go.dev/github.com/xoplog/xop-go/xoprecorder) | yes | Output into a structured in-memory buffer |
| [xoptest](https://pkg.go.dev/github.com/xoplog/xop-go/xoptest) | no | Output to testing.T logger |

## Using xop

To log, you must have a `*Log` object.  To create one you must start with a
`Seed`.  `Seed`s are created with `NewSeed(mods ...SeedModifier)`.  The
`SeedModifier`s are where you specify where the logs actually go by supplying
a bottom level, log exporter: a `xopbase.Logger`.  There are various 
bottom level loggers: `xoptest` for logging to a `*testing.T`, `xopjson` for
generating JSON logs, and `xopotel` for exporting traces (and logs) via
OpenTelemtry.

```go
seed := xop.NewSeed(xop.WithBase(xopjson.New(xopbytes.WriteToIOWriter(io.Stdout))))
```

When you've got a contrete task, for example responding to an HTTP
request or running a cronjob, you convert the `Seed` into a `*Log` with the
`Request()` method.  This can be hooked into your HTTP router so that
a `*Log` is injected into the request's `Context`.

```go
log := seed.Request("GET /users")
r = r.WithContext(log.IntoContext(r.Context()))
```

Once you have a `*Log`, you can log individual "lines", text with 
optional attached data elements.

The creation of a log line is done with chained methods.  It starts
with selecting the log level.

```go
log.Info().String("username", "john").Msg("created new user")
```

Logs are more useful when they have context.  Xop supports adding context
by making it easy to create sub-spans.  There are two flavors of sub-spans:
one for when doing things in parallel and one for when doing a sequence of
actions.

```go
forkA := log.Sub().Fork("do something in a go-routine")
step1 := log.Sub().Step("do the first step of a sequence")
```

Later, when looking at the various span and requests, it is helpful to have
metadata attached.  The metadata keys must be pre-registered.

```go
var BillingAccountKey = xopat.Make{
	Key: "billing.account",
	Namespace: "myApp",
	Indexed: true,
	Prominence: 10,
	Description: "A billing account number",
}.Int64Attribute()

step1.Span().Int64(BillingAccountKey, 299232)
```

There are many other features including:

- creating sub-loggers (span, etc) that prefill line attributes
- fetch logger out of `Context`
- adjust the logging level based on environment variables so that
  different Go packages can log at different levels
- change the set of base loggers on the fly
- mark with spans are done
- adjust `Seed` values as `*Log` is created
- redact sensitive values as they're being logged
- create a seed from a `*Log` or `*Span`

Although xop supports a global logger, it's use is discouraged because 
it doesn't provide enough context for the resulting logs to be useful.

### Performance

The performance of Xop is good enough.  See the benchmark results
at [logbench](https://github.com/xoplog/logbench).

In general: faster than 
[zap](https://github.com/uber-go/zap);
about the same as
[zerolog](https://github.com/rs/zerolog);
but not as quick as
[onelog](https://github.com/francoispqt/onelog) or
[phuslog](https://github.com/phuslu/log).

Xop has a much richer feature set than onelog or phuslog and a nicer
API than zap.

### Propagation

Tracing is inter-process.  Xop supports both B3 and WC3 trace headers in the
propagators that have been written.

#### Incoming

Incoming propagation is when we learn our parent trace id from a request
made to our server.  Xop currently only supports HTTP.  It is done with the 
xopmiddle package.  

xopmiddle generates middleware in various flavors so that you can incorporate
into various http router frameworks.

#### Outgoing

Outgoing propagation is sharing the current trace id as the parent request
to another server when making a request.  Xop currently only supports HTTP
and that only with [resty](https://github.com/go-resty/resty) in the 
[xopresty](https://github.com/xoplog/xopresty-go) package.  Adding additional
outgoing propagators is an outstanding priority.

### Version compatibility

xop is currently tested with go1.18 through go1.20. It is probably 
compatible with go1.17 and perhaps earlier.

## Terminology

A "trace" is the the entire set of spans relating to one starting request or action.  It can
span multiple servers.

A "request" is a single request or action being handled by one program.  It does not span multiple
servers.  There can be multiple requests in a trace.

A "span" is a linear portion of the processing required to handle a request.  A single span should
not include multiple threads of execution.  Span should represent a logical component to of the
work being done.  Breaking the work into spans is an exercise for the programmer.

A "logger" is something that is used throughout code to generate log lines and spans.

A "base logger" is the layer below that the "logger" uses to send output to different systems.

A "bytes logger" is an optional layer below "base logger" that works with logs that have already
become []bytes.

## Naming

### Name registry

Arbitrary names are supported for tagging log lines. For attributes to be displayed
specially in front-ends, they need to follow standards. Standard attribute groups are
pre-registered as structs. These can be shared between organizations by contributing
them to the [Xop repository](https://github.com/xoplog/xop-go/xopconst).

The following names are reserved.  What happens if they're used is undefined and up
to the individual base loggers.

- `xop`.  Used to indicate the kind of item begin emitted in a stream of objects. Empty for lines, `span` for spans.  `enum` to establish enum -> string mappings.  `chunk` for things broken up because they're too big.  `template` for lines that need template expansion.
- `msg`.  Used for the text of a log line.
- `ts`.  Used for the timestamp of the log event, if included.
- `stack`.  Used for stacktraces when errors or alerts are logged.
- `span`.  Used for the span-id of log lines for some base loggers.
- `caller`.  Used to indicate the immediate caller (file & line) when that's desired.
- `level`.  The log level (debug, trace, info, warn, error, alert)

The data associated with spans, traces, and requests must come from pre-registered
keys.

### Attribute/Key naming

#### Open Telementry

OpenTelemetry has invested heavily in naming.  They call it `semconv` (Semantic Conventions).
Although not yet complete, an open TODO for xop is to import the entirty of the 
OpenTelemetry semantic conventions into attributes.  We'll do this for two resons: 

1. Compatibility
2. The effenciency of not re-inventing the wheel.

[They](https://opentelemetry.io/docs/reference/specification/common/attribute-naming/) say to use
dots (`.`) to separate namespaces in attribute names and underscores (`_`) to separate words within a name.
Do not use a namespace as an attribute.

They have lots of examples for:

- [Resources](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/)
- [Traces](https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/)
- [Metrics](https://opentelemetry.io/docs/reference/specification/metrics/semantic_conventions/)

#### Open Tracing

The Open Tracing project has been "archived" in favor of Open Telementry.  That said, they have a
much shorter set of [semantic conventions](https://opentracing.io/specification/conventions/).

#### Zipkin

While lacking a full set of semantic conventions, Zipkin has some sage advice around
[how to instrument spans](https://zipkin.io/pages/instrumenting.html)

#### OpenCensus

OpenCensus lacks a full set of semantic conventions, but it does having suggestions for
how to [name spans](https://opencensus.io/tracing/span/name/).  In OpenCensus, tags names
need to be [registered](https://opencensus.io/tag/key/).

## Philosophy

Xop is opinionated.  It gently nudges in certain directions.  Perhaps
the biggest nudge is that there is no support for generating
logs outside of a span.  

### Log less

Do not log details that don't materialy add to the value of the log

### Log more

Use logs as a narritive of what's going on in the program so that when
you look at the logs, you can follow along with what's going on.

### Always log in context

Logs are best viewed in context: without without needing to search
and correlate, you should know how you go to the point of the log line
you're looking at.  This means the line itself needs less detail
and it contributes to the context of the lines around it.

### No log.Fatal

Panic should be caught and logged.  If panic is caught, `log.Fatal()` 
is not needed and is even redundant as it would problaby panic itself
causing multiple `log.Alert()` for the same event.

### Defer work

Most logs won't be looked at.  Ever.  When possilbe defer the work of assembling the log
to when it viewed.

## Other systems

This logger is primarily inspired by a proprietary logger at [BlueOwl](https://blueowl.xyz).
Other structured loggers also provided inspiration:
[onelog](https://github.com/francoispqt/onelog);
[phuslog](https://github.com/phuslu/log);
[zap](https://github.com/uber-go/zap);
[zerolog](https://github.com/rs/zerolog);
[Open Telementry](https://opentelemetry.io);
and
[Jaeger](https://www.jaegertracing.io/).

Special thanks to [phuslog](https://github.com/phuslu/log) as some of its
code was used.

