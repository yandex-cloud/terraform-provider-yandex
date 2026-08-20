package mdb_clickhouse_database

import (
	"context"
	"github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
)

func readDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) *clickhouse.Database {
	db, err := clickhousesdk.NewDatabaseClient(sdk).Get(ctx, &clickhouse.GetDatabaseRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.DatabaseCreateOperation, error) {
		return clickhousesdk.NewDatabaseClient(sdk).Create(ctx, &clickhouse.CreateDatabaseRequest{
			ClusterId:    cid,
			DatabaseSpec: dbSpec,
		})
	})

	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			// Try to automatically read existing resource rather than force user to import it manually
			diag.AddWarning(
				"Resource already exists.",
				"Database "+dbSpec.Name+" already exits in cluster "+cid,
			)
		} else {
			diag.AddError(
				"Failed to Create resource",
				"Error while requesting API to create ClickHouse database:"+err.Error(),
			)
		}
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create ClickHouse database:"+err.Error(),
		)
	}
}

func deleteDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.DatabaseDeleteOperation, error) {
		return clickhousesdk.NewDatabaseClient(sdk).Delete(ctx, &clickhouse.DeleteDatabaseRequest{
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete ClickHouse database: "+err.Error(),
		)
	}
}
