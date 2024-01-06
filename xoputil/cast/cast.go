package cast

func To[T any](v any) T {
	c, ok := v.(T)
	if ok {
		return c
	}
	var zero T
	return zero
}
