package mdb_mongodb_database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	mongodbv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
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

func createDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, spec *mongodb.DatabaseSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MongoDB().Database().Create(ctx, &mongodb.CreateDatabaseRequest{
			ClusterId:    cid,
			DatabaseSpec: spec,
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

func updateDatabase(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, req *mongodb.UpdateDatabaseRequest) {
	op, err := retry.ConflictingOperationV2(ctx, providerConfig.SDKv2, func() (*mongodbv1sdk.DatabaseUpdateOperation, error) {
		return mongodbv1sdk.NewDatabaseClient(providerConfig.SDKv2).Update(ctx, req)
	})
	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update MongoDB database:"+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update MongoDB database:"+err.Error(),
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
