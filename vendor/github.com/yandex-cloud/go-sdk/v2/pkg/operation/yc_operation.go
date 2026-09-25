package operation

import (
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type YCOperation interface {
	GetId() string
	GetDescription() string
	GetMetadata() *anypb.Any
	GetCreatedAt() *timestamppb.Timestamp
	GetError() *status.Status
	GetCreatedBy() string
	GetResponse() *anypb.Any
	GetDone() bool
	ProtoReflect() protoreflect.Message
}
