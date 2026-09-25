package xslices

import (
	"cmp"
	"slices"
)

func Keys[Key cmp.Ordered, T any](m map[Key]T) []Key {
	keys := make([]Key, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
