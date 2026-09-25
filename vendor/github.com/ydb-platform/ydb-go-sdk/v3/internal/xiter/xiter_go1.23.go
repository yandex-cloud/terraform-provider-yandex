//go:build go1.23

package xiter

import (
	"iter"
)

type (
	Seq[T any]     iter.Seq[T]
	Seq2[K, V any] iter.Seq2[K, V]
)
