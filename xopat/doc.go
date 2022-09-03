/*
Package xopat provides a mechanism to pre-register span attributes.

Pre-registered attributes can be found in the xopconst package.

It is anticipated that spans/traces/requests will be searchable with key/value
pairs.  To increase the quality of the key/value pairs, the keys must be pre-registered
and the values are statically typed.  Pre-registration provides flexibility in
describing the keys: should a particular key be indexed (searchable)?  Should it have
a single value or list?  If a list, should it only hold distinct values (a set)?
If a single value, should the first value be kept or the last?

Attributes on log lines are not pre-registered for several reasons, the main one
being that pre-registration takes effort and there are likely lots more line attributes
that span attributes.

Most span attributes are typed simply: int32, string, etc.  Any is also a valid
type.

For example:

	var URL = xopat.Make{Key: "http.url", Namespace: "OTEL", Indexed: true, Prominence: 12,
		Description: "Full HTTP request URL in the form scheme://host[:port]/path?query[#fragment]." +
		" Usually the fragment is not transmitted over HTTP, but if it is known," +
		" it should be included nevertheless"}.
		StringAttribute()

	log.Span().String(URL, url.String())

Enums

Enums are another valid type.  Enums values require both a string representation and
an integer.  It is up to a base logger to decide which value to keep (or both).  Enums
come in two forms: ones where the value embeds the key and ones where the key and value
are provided separately.  When the value embeds the key, there is full type safety: the
key and value will match each other.  The disadvantage of embedded enums is that there
is only one key for the value and sometimes that doesn't provide enough flexiblity.

There are multiple ways to pre-register enum keys and values but none of the methods
are completely frictionless:

	Make{}.EmbeddedEnumAttribute() // assign int and string values explicitly
	Make{}.IotaEnumAttribute()     // assign string values, ints are automatic
	Make{}.EnumAttribute() 	       // for values already implement .String() & .Int64()

With Enum:

	log.Span().Enum(xopconst.SpanKind, xopconst.SpanKindClient)

With EmbeddedEnum:

	log.Span().EmbeddedEnum(xopconst.SpanTypeHTTPClientRequest)

There is currently no good solution for using enums from third
parties.  The issue is that an enum needs to provide both an integer
value and a string and Go generics do not provide a solution since
you cannot say that something must be ~int|~int64 and implement
fmt.Stringer.  Since that isn't possible, enums must implement
String() and Int64() methods and you cannot define methods on
third-party types.

*/
package xopat
