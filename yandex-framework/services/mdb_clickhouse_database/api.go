package mdb_clickhouse_database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
)

func readDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) *clickhouse.Database {
	db, err := sdk.MDB().Clickhouse().Database().Get(ctx, &clickhouse.GetDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	})

	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			diag.AddWarning(
				"Failed to Read resource",
				"Database "+dbName+" not found in cluster "+cid,
			)
		} else {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get ClickHouse database:"+err.Error(),
			)

		}
		return nil
	}

	return db
}

func createDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbSpec *clickhouse.DatabaseSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Clickhouse().Database().Create(ctx, &clickhouse.CreateDatabaseRequest{
			ClusterId:    cid,
			DatabaseSpec: dbSpec,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create ClickHouse database:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create ClickHouse database:"+err.Error(),
		)
	}
}

func deleteDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Clickhouse().Database().Delete(ctx, &clickhouse.DeleteDatabaseRequest{
			ClusterId:    cid,
			DatabaseName: dbName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete ClickHouse database: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete ClickHouse database: "+err.Error(),
		)
	}
}
