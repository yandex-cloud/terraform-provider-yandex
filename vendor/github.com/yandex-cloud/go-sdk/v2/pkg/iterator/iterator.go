package iterator

import (
	"context"

	"google.golang.org/grpc"
)

const defaultPageSize int64 = 1000

type PageRequest interface {
	GetPageSize() int64
	SetPageSize(int64)
	GetPageToken() string
	SetPageToken(string)
}

type PageResponse[T any] interface {
	Items() []T
	GetNextPageToken() string
}

type PageFetcher[Req PageRequest, T any] func(
	ctx context.Context,
	req Req,
	opts ...grpc.CallOption,
) (PageResponse[T], error)

type Iterator[Req PageRequest, T any] struct {
	ctx  context.Context
	opts []grpc.CallOption

	err           error
	started       bool
	requestedSize int64
	pageSize      int64

	req   Req
	fetch PageFetcher[Req, T]
	items []T
}

func NewIterator[Req PageRequest, T any](
	ctx context.Context,
	req Req,
	fetch PageFetcher[Req, T],
	opts ...grpc.CallOption,
) *Iterator[Req, T] {

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	return &Iterator[Req, T]{
		ctx:      ctx,
		opts:     opts,
		req:      req,
		fetch:    fetch,
		pageSize: pageSize,
	}
}

func (it *Iterator[Req, T]) Next() bool {
	if it.err != nil {
		return false
	}

	if len(it.items) > 1 {
		it.items = it.items[1:]
		return true
	}
	it.items = nil

	if it.started && it.req.GetPageToken() == "" {
		return false
	}
	it.started = true

	if it.requestedSize == 0 || it.requestedSize > it.pageSize {
		it.req.SetPageSize(it.pageSize)
	} else {
		it.req.SetPageSize(it.requestedSize)
	}

	resp, err := it.fetch(it.ctx, it.req, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = resp.Items()
	it.req.SetPageToken(resp.GetNextPageToken())
	return len(it.items) > 0
}

func (it *Iterator[Req, T]) Value() T {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *Iterator[Req, T]) Take(size int64) ([]T, error) {
	if size == 0 {
		size = 1 << 32
	}
	it.requestedSize = size
	defer func() { it.requestedSize = 0 }()

	var out []T
	for it.requestedSize > 0 && it.Next() {
		it.requestedSize--
		out = append(out, it.Value())
	}
	return out, it.err
}

func (it *Iterator[Req, T]) TakeAll() ([]T, error) {
	return it.Take(0)
}

func (it *Iterator[Req, T]) Error() error {
	return it.err
}

func (it *Iterator[Req, T]) NextPageToken() string {
	return it.req.GetPageToken()
}
