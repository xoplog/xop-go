package generic

func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	n := make(map[K]V)
	for k, v := range m {
		n[k] = v
	}
	return n
}

