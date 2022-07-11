# Things not built, but considered

These are things that could be built if there is sufficient interest.

## Per-line prefill

Add per-line data for all lines in a span.  Allow the set of data to
change.

In the base loggers:

```go
type Span interface {
	... other stuff

	// On a per-Span basis, there will only ever be one outstanding Prefill
	// in progress at a time and it will not overlap Flush()
	Prefill() Prefilling

	type Prefilling interface {
		ObjectParts
		Replace() // replaces per-line prefill
		AddTo() // adds to per-line prefill
	}
}
```

