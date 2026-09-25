//go:build !go1.23

package xiter

type (
	Seq[T any]     func(yield func(T) bool)
	Seq2[K, V any] func(yield func(K, V) bool)
)
