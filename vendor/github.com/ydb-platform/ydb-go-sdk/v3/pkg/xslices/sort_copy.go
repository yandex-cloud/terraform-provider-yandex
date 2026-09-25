package xslices

import "slices"

func SortCopy[T any](in []T, cmp func(lhs, rhs T) int) (out []T) {
	out = slices.Clone(in)

	slices.SortFunc(out, cmp)

	return out
}
