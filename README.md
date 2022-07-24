# xop - Golang structured log generation combined with tracing (Cross Obserability Platform)

[![GoDoc](https://godoc.org/github.com/muir/xoplog?status.png)](https://pkg.go.dev/github.com/muir/xoplog)

# Development status

In development, not ready for use.

# Context

## The problem with the existing model

The industry model of tracing as documented in the W3C spec requires that spans
have full identifiers.  If you give each part of dealing with a request inside
a single server, lots of different spans, then how can you quickly reference the
request-level span from one the sub-spans or one of the other requests that 
is a child of the main request.  There is no standard way to distinguish a span
that is simply a separate thread of execution or one that is a related
request on a different server.

The format of logs isn't easy to extend because there are is no meta-level or
standard for what log fields mean.  The closest for this is the naming semantics
that are included in the Open Telementry project.

Once the logs are generated, good logging systems tag each line with trance and
span identifiers, but the logs are still stored, searched, and dispalyed as
lines.  Most of the value comes from the context of the log so recording them
as lines removes misses the point.

Another issue with most structured loggers is that they over-collect details that
don't matter and un-invest in how the logs are presented.  The extra details are
sometimes useful, but they increase the cost to process and store the logs.  More
importantly, they can clutter the display so that it has lots of data but does
not present much information.

The standard model does not lend itself to experimentation with what kinds of things
are logged and how they're presented.  At [BlueOwl](https://www.blueowl.xyz), we
discovered that logging tables was very valuable.  We had support for displaying
tables in our log viewer.  For some very complicated bugs, displaying tables in the
logs was instrumental in finding the problem.

## Alternatives

- [Open Telementry](https://opentelemetry.io/) (which gobbled up 
[Open Tracing](https://github.com/opentracing/opentracing-go), 
[OpneCensus](https://opencensus.io/), and 
[Jaeger](https://www.jaegertracing.io/))
- [Zipkin](https://github.com/openzipkin/zipkin-go)
- AppDynamics
- Datadog

### other takes

[Software Tracing With Go](https://www.doxsey.net/blog/software-tracing-with-go/)

# Architecture

A complete logging solution has:

## Generating Logs

Logging API for writing code.  There are a couple of flavors that are commonly
found: Printf-style; Key/Value style; Structure with functional args style; and
Structured with methods style.

Tracing APIs have less variation.  Opentracing is a fine example.

## Writing logs

Logs can be sent directly to a server; they can be written to STDOUT. Most systems
use JSON for structured logging, but Open Telementry uses Protobufs.  

## Gathering logs

Most frameworks include an intermediary program that scrapes up logs, especially
STDOUT logs, and sends them on to servers relibably.  This has an advantage of lowering
the likelihood of backpressure.

Should backpressure exist at all?  Maybe.  Some logs may be required for auditing
purposes and such logs may need to be sent reliably.

## Indexing logs

ElasticSearch vs OpenSearch vs Solr?  Elastic loses on the licensing front.

It makes sense to do full-text indexing and also trace-level indexing separately
so that the trace-level indexing can be retained for a longer duration.

## Storing logs

S3 and other bulk storage

# Generating tracelogs API

## Request-level

At the request-level (HTTP request) or program being run, etc, tag
the request with: fields and indexes.  The fields are descriptive
and the indexes are searchable.  Both have the same format.

The request model and response model are good examples of fields.

The URL is a good example of an index.

One required indexed key is 'type'.  In an organization, each type
can have different required fields and associated value types.

For example, the 'http' type can have: 'endpoint', 'path', 'method',
'response-code' indexes and 'request-data', 'response-data' fields.

Should this be enforced?  XXX how?

## Span-level

Spans have a text description

Spans have arbitray key/value pairs that are added as users create
the span.

Spans have a W3C-compatible id

Spans have a .1 .A prefix to show how they relate to the other spans
in the same request

Spans are either included in the request indexes or not included.  If
included, they have type=span.

Spans have a full W3C trace context 

## Log-line level

Log lines do not have any inherited data except their span reference

Log lines can have key/value pairs

If the key is "span" then the value should be full span reference

The key of "type" should be displayed prominately by the front-end and
the following types are suggested:

- "reference": requires a "span" and referes to that span

- "request": a request to another system, if that system
  replied with a "tracereference" then record that here.

# Generating Tracelogs - Data format

## Log lines

[Open Telemetry Log Specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md)

Every log line gets a span id and trace id. 

In the key/values attached to log lines, a key of "reference" 
to another trace/span

## Spans

Span records are only emitted after the span is complete or N minutes have passed.

Must reference their parent span, if possilbe

May have a "thread.context" like "B.3." to indicate it's the third thing in the 2nd parallel execution.
If there is no "thread.context", then attributes are searchable.

May have an http.status_code even if it's not an http request

An "error.count" attribute is automatically added

Spans must have names.  For parallel forks, this is the same as the thread context.

Requests to other services get their own span.

Open Telementy's "SpanKind" isn't rich enough.  It's missing CLI_COMMAND, TASK

Span data can continue to be updated even after the span is first sent.

# Forwarding tracelogs

No changes are needed for log forwarding

[LogStash introduction](https://www.elastic.co/guide/en/logstash/current/introduction.html)
[Open Telementry data collection](https://opentelemetry.io/docs/concepts/data-collection/)
[Open Telementry protocol](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)

# Reforming tracelogs

Accepts both logs and traces.

Step 1: save the raw data.  Data is written to persistent disk in the "inbox".  Logs that
don't have trace ids and span ids are dropped.

Step 2: If there are multiple servers, forward raw data to the appropriate server based on
sharding of the trace id.

Step 3: Write data to appropriate trace-collection buckets (files, one per trace).  Mark
buckets as not-indexed.

Step 4: Once written to buckets, inbox files can be removed.

Step 5: Add/update Elastic indexes.  

NOTES



# Terminology

A "trace" is the the entire set of spans relating to one starting request or action.  It can
span multiple servers.

A "request" is a single request or action being handled by one program.  It does not span multiple
servers.  There can be multiple requests in a trace.

A "span" is a linear portion of the processing required to handle a request.  A single span should
not include multiple threads of execution.  Span should represent a logical component to of the
work being done.  Breaking the work into spans is an exercise for the programmer.

# Naming

## Name registry

Arbitrary names are supported for tagging log lines. For attributes to be displayed
specially in front-ends, they need to follow standards. Standard attribute groups are
pre-registered as structs. These can be shared between organizations by contributing
them to the [Xop repository](https://github.com/muir/xoplog/xopconst).

The following names are reserved.  What happens if they're used is undefined and up
to the individual base loggers.

- `xop`.  Used to indicate the kind of item begin emitted in a stream of objects. Empty for lines, `span` for spans.  `enum` to establish enum -> string mappings.  `chunk` for things broken up because they're too big.  `template` for lines that need template expansion.
- `msg`.  Used for the text of a log line.
- `time`.  Used for the timestamp of the log event, if included.
- `stack`.  Used for stacktraces when errors or alerts are logged.
- `span`.  Used for the span-id of log lines for some base loggers.
- `caller`.  Used to indicate the immediate caller (file & line) when that's desired.
- `level`.  The log level (debug, trace, info, warn, error, alert)

The data associated with spans, traces, and requests must come from pre-registered
keys.

# Philosphy

## Log less

Do not log details that don't materialy add to the value of the log

## Log more

Use logs as a narritive of what's going on in the program so that when
you look at the logs, you can follow along with what's going on.

## Defer work

Most logs won't be looked at.  Ever.  When possilbe defer the work of assembling the log
to when it viewed.

## Other systems

This logger is inspired by a proprietary logger at [BlueOwl](https://blueowl.xyz);
[onelog](https://github.com/francoispqt/onelog);
[phuslog](https://github.com/phuslu/log);
[zap](https://github.com/uber-go/zap);
[zerolog](https://github.com/rs/zerolog);
[Open Telementry](https://opentelemetry.io);
and
[Jaeger](https://www.jaegertracing.io/).


### Open Telementry

[They](https://opentelemetry.io/docs/reference/specification/common/attribute-naming/) say to use
dots (`.`) to separate namespaces in attribute names and underscores (`_`) to separate words within a name.
Do not use a namespace as an attribute.

They have lots of examples for:

- [Resources](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/)
- [Traces](https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/)
- [Metrics](https://opentelemetry.io/docs/reference/specification/metrics/semantic_conventions/)

### Open Tracing

The Open Tracing project has been "archived" in favor of Open Telementry.  That said, they have a
much shorter set of [semantic conventions](https://opentracing.io/specification/conventions/).

### Zipkin

While lacking a full set of semantic conventions, Zipkin has some sage advice around
[how to instrument spans](https://zipkin.io/pages/instrumenting.html)

### OpenCensus

OpenCensus lacks a full set of semantic conventions, but it does having suggestions for
how to [name spans](https://opencensus.io/tracing/span/name/).  In OpenCensus, tags names
need to be [registered](https://opencensus.io/tag/key/).

