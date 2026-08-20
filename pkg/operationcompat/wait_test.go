package operationcompat

import (
	"context"
	"strings"
	"testing"

	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	vpc "github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	sdkop "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestWaitIgnoresSuccessfulPayloadTypes(t *testing.T) {
	response, err := anypb.New(&emptypb.Empty{})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	metadata, err := anypb.New(&vpc.UpdateSecurityGroupMetadata{SecurityGroupId: "security-group-id"})
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}

	poll := func(context.Context, string, ...grpc.CallOption) (sdkop.YCOperation, error) {
		return &operationpb.Operation{
			Id:       "operation-id",
			Done:     true,
			Metadata: metadata,
			Result:   &operationpb.Operation_Response{Response: response},
		}, nil
	}

	_, err = wait(context.Background(), "operation-id", poll)
	if err != nil {
		t.Fatalf("wait operation: %v", err)
	}
}

func TestWaitReturnsOperationError(t *testing.T) {
	poll := func(context.Context, string, ...grpc.CallOption) (sdkop.YCOperation, error) {
		return &operationpb.Operation{
			Id:   "operation-id",
			Done: true,
			Result: &operationpb.Operation_Error{Error: &statuspb.Status{
				Code:    int32(codes.Internal),
				Message: "operation failed",
			}},
		}, nil
	}

	_, err := wait(context.Background(), "operation-id", poll)
	if err == nil || !strings.Contains(err.Error(), "operation failed") {
		t.Fatalf("expected operation error, got %v", err)
	}
}

func TestWaitResponseIgnoresMetadataTypeAndDecodesResponse(t *testing.T) {
	response, err := anypb.New(&vpc.SecurityGroup{Id: "security-group-id"})
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

	actual := &vpc.SecurityGroup{}
	err = waitResponse(context.Background(), "operation-id", poll, actual)
	if err != nil {
		t.Fatalf("wait operation response: %v", err)
	}
	if actual.GetId() != "security-group-id" {
		t.Fatalf("unexpected response id: %q", actual.GetId())
	}
}
