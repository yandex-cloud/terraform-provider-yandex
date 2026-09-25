package operation

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type AbstractOperation interface {
	ID() string
	Description() string
	CreatedBy() string
	CreatedAt() time.Time
	Metadata() proto.Message
	ResourceID() string
	Done() bool
	PollOnce(ctx context.Context, opts ...grpc.CallOption) error
	Wait(ctx context.Context, opts ...grpc.CallOption) (proto.Message, error)
	Error() error
	Response() proto.Message
	Result() OperationResult
}

type OperationResult interface {
	Response() proto.Message
	Metadata() proto.Message
	Error() error
}
