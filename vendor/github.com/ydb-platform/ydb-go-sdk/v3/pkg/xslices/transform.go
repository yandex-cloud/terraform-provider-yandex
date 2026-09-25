package xslices

func Transform[T1, T2 any](in []T1, f func(t T1) T2) (out []T2) {
	out = make([]T2, len(in))

	for i, t := range in {
		out[i] = f(t)
	}

	return out
}
