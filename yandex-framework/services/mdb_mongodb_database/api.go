package mdb_mongodb_database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

func readDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) *mongodb.Database {
	db, err := sdk.MDB().MongoDB().Database().Get(ctx, &mongodb.GetDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get MongoDB database:"+err.Error(),
		)
		return nil
	}
	return db
}

func createDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, dbName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MongoDB().Database().Create(ctx, &mongodb.CreateDatabaseRequest{
			ClusterId: cid,
			DatabaseSpec: &mongodb.DatabaseSpec{
				Name: dbName,
			},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create MongoDB database:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MongoDB database:"+err.Error(),
		)
	}
}

func deleteDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, dbName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MongoDB().Database().Delete(ctx, &mongodb.DeleteDatabaseRequest{
			ClusterId:    cid,
			DatabaseName: dbName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete MongoDB database: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MongoDB database: "+err.Error(),
		)
	}
}
