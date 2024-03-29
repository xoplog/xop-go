
# Xop internals

## Done() & Flush()

We want to call Done() on base spans as soon as we know they're
done, or just before a Flush() if there is activity since the last
Done().

Each log keeps track of it's non-detached children that aren't Done.

A log can only be Done() if it's dependent children are Done(). 

activeChildren are those that are not explicitly Done() (both dependent and detached)

Non-detached children are Done when their parent is Done or if Done() is called
on them explicitly.

span.knownActive: true if listed in parents activeChildren

## Ordering

Change atomics and then take action based upon the value of the 
atomic.  For example set knownActive to 1 and then if it had been
zero, add to depndentents and such.

On the reverse, set knownActive to 0 and then call Done(), etc.

Try not to hold locks while calling things: make todo lists while
holding the lock but then do the work w/o the lock.  

# Philosophy

## The problem with the existing model

The industry model of tracing as documented in the W3C spec requires that spans
have full identifiers.  If you give each part of dealing with a request inside
a single server, lots of different spans, then how can you quickly reference the
request-level span from one the sub-spans or one of the other requests that 
is a child of the main request?  There is no standard way to distinguish a span
that is simply a separate thread of execution or one that is a related
request on a different server.  (Well, okay, Open Telemetry could be considered
a standard, and it provides a way)

The format of logs isn't easy to extend because there are is no meta-level or
standard for what log fields mean.  The closest for this is the naming semantics
that are included in the Open Telementry project.

Once the logs are generated, good logging systems tag each line with trance and
span identifiers, but the logs are still stored, searched, and displayed as
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

# Misc

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

Note: this is not enforced (yet)

## Span-level

Spans have a text description

Spans have arbitrary key/value pairs that are added as users create
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

The key of "type" should be displayed prominently by the front-end and
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


