package xopconst

// ParentLink is added automatically by xop in all situations where the information is present
var ParentLink = Make{Key: "parent", Namespace: "xop", Indexed: true, Description: "Parent span"}.LinkAttribute()

var EndpointRoute = Make{Key: "http.route", Namespace: "xop", Indexed: true,
	Description: "HTTP handler route used to handle the request." +
		" If there are path parameters in the route their generic names should be used," +
		" eg '/invoice/{number}' or '/invoice/:number' depending on the router used"}.StrAttribute()

var Boring = Make{Key: "boring", Namespace: "xop", Indexed: false,
	Description: "spans are boring if they're an internal span (created by log.Fork() or" +
		" log.Step()) or they're a request and log.Boring() has been called, and if" +
		" there have has been nothing logged at the Error or Alert level"}.BoolAttribute()
