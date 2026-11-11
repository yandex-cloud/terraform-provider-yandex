package datasphere_community

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	dataspheresdk "github.com/yandex-cloud/go-sdk/services/datasphere/v2"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
)

func requestCreateCommunity(ctx context.Context, sdk *ycsdk.SDK, req *datasphere.CreateCommunityRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, dataspheresdk.CommunityCreate)
	if err != nil {
		return nil, err
	}
	return datasphere.NewCommunityServiceClient(conn).Create(ctx, req)
}
