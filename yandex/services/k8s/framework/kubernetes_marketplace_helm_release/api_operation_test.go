package kubernetes_marketplace_helm_release

import (
	"context"
	"strings"
	"testing"
	"time"

	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdkop "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestWaitOperationWithoutResponseIgnoresResponseType(t *testing.T) {
	response, err := anypb.New(&emptypb.Empty{})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	poll := func(context.Context, string, ...grpc.CallOption) (sdkop.YCOperation, error) {
		return &operationpb.Operation{
			Id:     "operation-id",
			Done:   true,
			Result: &operationpb.Operation_Response{Response: response},
		}, nil
	}

	err = waitOperationWithoutResponse(context.Background(), "operation-id", poll, func(int) time.Duration { return 0 })
	if err != nil {
		t.Fatalf("wait operation: %v", err)
	}
}

func TestWaitOperationWithoutResponseReturnsOperationError(t *testing.T) {
	poll := func(context.Context, string, ...grpc.CallOption) (sdkop.YCOperation, error) {
		return &operationpb.Operation{
			Id:   "operation-id",
			Done: true,
			Result: &operationpb.Operation_Error{Error: &statuspb.Status{
				Code:    int32(codes.Internal),
				Message: "install failed",
			}},
		}, nil
	}

	err := waitOperationWithoutResponse(context.Background(), "operation-id", poll, func(int) time.Duration { return 0 })
	if err == nil || !strings.Contains(err.Error(), "install failed") {
		t.Fatalf("expected operation error, got %v", err)
	}
}
