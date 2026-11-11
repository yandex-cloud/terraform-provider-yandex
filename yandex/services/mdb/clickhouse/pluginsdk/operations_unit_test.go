package pluginsdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdkoperationv2 "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	svc_clickhouse "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/pluginsdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestWaitOperationV2PreservesSDKWaitSemantics(t *testing.T) {
	response, err := anypb.New(&emptypb.Empty{})
	require.NoError(t, err)

	polls := 0
	op, err := sdkoperationv2.NewOperation(
		&operationpb.Operation{Id: "test-operation"},
		&sdkoperationv2.Concretization{
			Poll: func(context.Context, string, ...grpc.CallOption) (sdkoperationv2.YCOperation, error) {
				polls++
				if polls <= 3 {
					return nil, status.Error(codes.NotFound, "operation is not visible yet")
				}
				return &operationpb.Operation{
					Id:   "test-operation",
					Done: true,
					Result: &operationpb.Operation_Response{
						Response: response,
					},
				}, nil
			},
			ResponseType: &emptypb.Empty{},
		},
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, svc_clickhouse.WaitOperationV2(ctx, op, 0))
	require.Equal(t, 4, polls)
}
