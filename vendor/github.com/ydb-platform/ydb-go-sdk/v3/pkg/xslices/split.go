package xslices

func Split[T any](x []T, isOk func(t T) bool) (good, bad []T) {
	good = make([]T, 0, len(x))
	bad = make([]T, 0, len(x))

	for i := 0; i < len(x); i++ {
		if isOk(x[i]) {
			good = append(good, x[i])
		} else {
			bad = append(bad, x[i])
		}
	}

	return good, bad
}
