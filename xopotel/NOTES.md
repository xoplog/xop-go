
TODO/XXX: test changing OTEL Span name

## Doc notes

In this doc, we'll use the names of contants, like `xopVersion` 
rather than the values. The values can be found in `models.go`.

## Where OTEL and XOP differ

As of the time of writing, OTEL doesn't have a Go logger so logging
is done as span Events.

Data types.  The set of data types is different with OTEL having 
slices that XOP doesn't have and XOP having Any/Model, unsigned int,
time.Duration, and time.Time which OTEL doesn't have.

Both OTEL and XOP require some information at the creation of a span,
but there are differences:

- XOP span creation only (where OTEL can change it later)
  - Name/Description
  - "SourceInfo"
- OTEL span creation only (where XOP can change it later)
  - Links
  - "Instrumentation.Scope"

In XOP, links can be log data. Links only have attributes (other
than a text description) only when they're log data and not when
they're span-level attributes (metadata).

In OTEL, links are only span-level attributes and each link can
have arbitary attributes.

## XOP -> OTEL

Data that originaged in XOP is identified by having a set of 
XOP-specific span-level attributes, at least on the request.

- xopVersion a version string 
- xopOTELVersiona version string 
- xopSource encoding the SourceInfo Source & SourceVersion
- xopNamespace encoding the SourceInfo Namespace & Namespace Version

These four attributes will always be present but the presense of 
xopVersion (`"xop.version"`) is enough to judge span as originating
with XOP.

Translating from XOP to OTEL means encoding log lines as span Events
since (at the time of this writing) OTEL doesn't have a Go logger.

Since the types don't match up and XOP has more types, almost all XOP
types are encoded as OTEL `StringSlice` with the value as the the first
element in the slice and the type as the second element. Some XOP types,
like, Any, use additional slice elements. Xop bool gets encoded as a
`Bool` since there is only one kind of bool.

The encoding of XOP links varies a bit depending on which XOP->OTEL
encoder is used.

- For `BufferedReplayLogger` and `ReadOnlySpan`s that have been 
  post-processed with `NewUnhacker`, span Metadata links
  become OTEL span-level links directly.
- For other encoders links in span Metadata becomes sub-spans that are 
  marked as existing just to encode links in their parent span. These
  extra `Span`s are marked with a spanIsLinkAttributeKey attribute.
- For log lines that are links, if they're encoded both as an `Event`
  and also with a sub-Span that is used just to encode the links. These
  extra `Span`s are marked with a spanIsLinkEventKey attribute.
  WIth `BufferedReplayLogger`, line links are also added to the span
  attributes.

- Baggage
  While OTEL includes functions for manipulating Baggage, the Baggage is
  not retained in the span directly. XOP baggage is saved in OTEL as
  a span-level string attribute keyed by the `xopBaggage` constant.

## XOP -> OTEL -> XOP

When translating from OTEL to XOP, data that originated in XOP is detected
so that it can be reconstructed at full fidelity.

The detection is based upon the presense of specific attributes in the 
`Span`s.

## OTEL -> XOP 

Data imported from OTEL is marked with a MetadataAny: otelReplayStuff.
That marking allows XOP -> OTEL translation to detect where the data came
from origianlly so that it can be reconstructed. This object contains
all of the attributes that aren't easily mapped to XOP. The type of this
object is `otelStuff`.

Some of the OTEL types translate directly to XOP and they're sent with their
natural representation.

The remaining OTEL types are mostly encoded using `xopbase.ModelArg` and sent
at the line live with `Model()` and at the span level with `MetadataAny`.  

OTEL links are all span-level, but they all have attribute slices rather than
just name/description so they do not fit well as MetadataLinks. OTEL links are
sent twice, once as MetadataLink with `otelLink` and also as line Links where
their attributes (if any) are included, always described as `xopOTELLinkDetail`.
