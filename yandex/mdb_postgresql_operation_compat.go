package yandex

import (
	"context"

	postgresql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	postgresqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/postgresql/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
)

// requestPostgreSQLDatabaseUpdate starts the operation through genproto so the
// provider does not apply SDK v2's Database response assertion. The service
// returns google.protobuf.Empty here, as it did while the provider used SDK v1.
func requestPostgreSQLDatabaseUpdate(
	ctx context.Context,
	sdk *ycsdk.SDK,
	req *postgresql.UpdateDatabaseRequest,
) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, postgresqlsdk.DatabaseUpdate)
	if err != nil {
		return nil, err
	}
	return postgresql.NewDatabaseServiceClient(conn).Update(ctx, req)
}
