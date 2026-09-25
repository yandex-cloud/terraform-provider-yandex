package iterator

import (
	"context"

	"google.golang.org/grpc"
)

// SimpleResponse is the interface for responses that contain a list of items
// but don't use pagination (return all items in a single response).
type SimpleResponse[T any] interface {
	Items() []T
}

// SimpleFetcher is the function type that fetches all items in a single request.
type SimpleFetcher[Req any, T any] func(
	ctx context.Context,
	req Req,
	opts ...grpc.CallOption,
) (SimpleResponse[T], error)

// SimpleIterator iterates over items from a non-paginated List method.
// Unlike Iterator, it fetches all items in a single request.
type SimpleIterator[Req any, T any] struct {
	ctx  context.Context
	opts []grpc.CallOption

	err     error
	fetched bool

	req   Req
	fetch SimpleFetcher[Req, T]
	items []T
}

// NewSimpleIterator creates a new iterator for non-paginated responses.
func NewSimpleIterator[Req any, T any](
	ctx context.Context,
	req Req,
	fetch SimpleFetcher[Req, T],
	opts ...grpc.CallOption,
) *SimpleIterator[Req, T] {
	return &SimpleIterator[Req, T]{
		ctx:   ctx,
		opts:  opts,
		req:   req,
		fetch: fetch,
	}
}

// Next advances the iterator to the next item.
// Returns true if there is an item available, false otherwise.
func (it *SimpleIterator[Req, T]) Next() bool {
	if it.err != nil {
		return false
	}

	if len(it.items) > 1 {
		it.items = it.items[1:]
		return true
	}
	it.items = nil

	if it.fetched {
		return false
	}
	it.fetched = true

	resp, err := it.fetch(it.ctx, it.req, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = resp.Items()
	return len(it.items) > 0
}

// Value returns the current item.
// Panics if called on an empty iterator.
func (it *SimpleIterator[Req, T]) Value() T {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

// Take returns up to size items from the iterator.
// If size is 0, returns all items.
func (it *SimpleIterator[Req, T]) Take(size int64) ([]T, error) {
	if size == 0 {
		size = 1 << 32
	}

	var out []T
	for size > 0 && it.Next() {
		size--
		out = append(out, it.Value())
	}
	return out, it.err
}

// TakeAll returns all items from the iterator.
func (it *SimpleIterator[Req, T]) TakeAll() ([]T, error) {
	return it.Take(0)
}

// Error returns the error that occurred during iteration, if any.
func (it *SimpleIterator[Req, T]) Error() error {
	return it.err
}
