package xslices

func Map[Key comparable, T any](x []T, key func(t T) Key) map[Key]T {
	m := make(map[Key]T, len(x))

	for i := range x {
		m[key(x[i])] = x[i]
	}

	return m
}
