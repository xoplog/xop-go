package xopat

var reservedKeys = map[string]struct{}{
	"dur":           {},
	"impl":          {},
	"name":          {},
	"request":       {},
	"request.id":    {},
	"span":          {},
	"span.ver":      {},
	"trace":         {},
	"trace.baggage": {},
	"trace.header":  {},
	"trace.id":      {},
	"trace.parent":  {},
	"trace.state":   {},
	"ts":            {},
	"type":          {},
}
