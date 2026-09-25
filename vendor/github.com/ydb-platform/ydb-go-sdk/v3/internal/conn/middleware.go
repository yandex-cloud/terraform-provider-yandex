package conn

import (
	"context"

	"google.golang.org/grpc"
)

var _ grpc.ClientConnInterface = (*middleware)(nil)

type (
	invoker  func(context.Context, string, interface{}, interface{}, ...grpc.CallOption) error
	streamer func(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error)
)

type middleware struct {
	invoke    invoker
	newStream streamer
}

func (m *middleware) Invoke(
	ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption,
) error {
	return m.invoke(ctx, method, args, reply, opts...)
}

func (m *middleware) NewStream(
	ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return m.newStream(ctx, desc, method, opts...)
}

func WithContextModifier(
	cc grpc.ClientConnInterface,
	modifyCtx func(ctx context.Context) context.Context,
) grpc.ClientConnInterface {
	return &middleware{
		invoke: func(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
			ctx = modifyCtx(ctx)

			return cc.Invoke(ctx, method, args, reply, opts...)
		},
		newStream: func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (
			grpc.ClientStream, error,
		) {
			ctx = modifyCtx(ctx)

			return cc.NewStream(ctx, desc, method, opts...)
		},
	}
}

func WithAppendOptions(cc grpc.ClientConnInterface, appendOpts ...grpc.CallOption) grpc.ClientConnInterface {
	return &middleware{
		invoke: func(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
			opts = append(opts, appendOpts...)

			return cc.Invoke(ctx, method, args, reply, opts...)
		},
		newStream: func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (
			grpc.ClientStream, error,
		) {
			opts = append(opts, appendOpts...)

			return cc.NewStream(ctx, desc, method, opts...)
		},
	}
}

func WithBeforeFunc(
	cc grpc.ClientConnInterface,
	before func(),
) grpc.ClientConnInterface {
	return &middleware{
		invoke: func(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
			before()

			return cc.Invoke(ctx, method, args, reply, opts...)
		},
		newStream: func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (
			grpc.ClientStream, error,
		) {
			before()

			return cc.NewStream(ctx, desc, method, opts...)
		},
	}
}
