package xslices

import (
	"cmp"
	"sort"
)

func Uniq[T cmp.Ordered](in []T) (out []T) {
	sort.Slice(in, func(i, j int) bool {
		return in[i] < in[j]
	})

	out = make([]T, 0, len(in))
	out = append(out, in[0])

	for i := 1; i < len(in); i++ {
		if in[i] != out[len(out)-1] {
			out = append(out, in[i])
		}
	}

	return out
}
