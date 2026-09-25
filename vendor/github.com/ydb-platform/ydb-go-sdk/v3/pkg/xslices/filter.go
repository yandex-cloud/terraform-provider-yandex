package xslices

func Filter[T any](in []T, filter func(t T) bool) (out []T) {
	out = make([]T, 0, len(in))

	for i := 0; i < len(in); i++ {
		if filter(in[i]) {
			out = append(out, in[i])
		}
	}

	return out
}
