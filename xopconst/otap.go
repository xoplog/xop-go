package xopconst

// The descriptions are lifted from https://opentelemetry.io/ and are thus Copyright(c)
// the Open Telementry authors.

var HTTPMethod = Make{Key: "http.method", Namespace: "OTAP", Indexed: true, Description: "HTTP request method"}.StrAttribute()

var URL = Make{Key: "http.url", Namespace: "OTAP", Indexed: true,
	Description: "Full HTTP request URL in the form scheme://host[:port]/path?query[#fragment]." +
		" Usually the fragment is not transmitted over HTTP, but if it is known," +
		" it should be included nevertheless"}.StrAttribute()

var HTTPTarget = Make{Key: "http.target", Namespace: "OTAP", Indexed: true,
	Description: "The full request target as passed in a HTTP request line or equivalent"}.StrAttribute()

var HTTPHost = Make{Key: "http.host", Namespace: "OTAP", Indexed: true,
	Description: "The value of the HTTP host header. An empty Host header should also be reported"}.StrAttribute()

var HTTPStatusCode = Make{Key: "http.status_code", Namespace: "OTAP", Indexed: true, Description: "HTTP response status code"}.IntAttribute()
