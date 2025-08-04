package mdb_sharded_postgresql_database

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

var shardedPostgreSQLAPI = ShardedPostgreSQLAPI{}

type ShardedPostgreSQLAPI struct{}

func (r *ShardedPostgreSQLAPI) ReadDatabase(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, dbname string) *spqr.Database {
	dbs, err := sdk.MDB().SPQR().Database().List(ctx, &spqr.ListDatabasesRequest{
		ClusterId: cid,
	})
	if err != nil {
		diags.AddError(
			"Failed to Read resources",
			fmt.Sprintf("Error while requesting API to get Sharded PostgreSQL database: %s", err.Error()),
		)
		return nil
	}

	for _, u := range dbs.GetDatabases() {
		if u.GetName() == dbname {
			return u
		}
	}

	diags.AddError(
		"Failed to Read resource",
		fmt.Sprintf("Sharded PostgreSQL database %q not found", dbname),
	)
	return nil
}

func (r *ShardedPostgreSQLAPI) CreateDatabase(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, db *spqr.DatabaseSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().SPQR().Database().Create(ctx, &spqr.CreateDatabaseRequest{
			ClusterId:    cid,
			DatabaseSpec: db,
		})
	})
	if err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while requesting API to create Sharded PostgreSQL database: %s", err.Error()),
		)
		return
	}
	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while waiting for operation to create Sharded PostgreSQL database: %s", err.Error()),
		)
	}
}

func (r *ShardedPostgreSQLAPI) UpdateDatabase(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, db *spqr.DatabaseSpec, updatePaths []string) {
	diags.AddError("update database is not implemented yet", "")
}

func (r *ShardedPostgreSQLAPI) DeleteDatabase(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, dbname string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().SPQR().Database().Delete(ctx, &spqr.DeleteDatabaseRequest{
			ClusterId:    cid,
			DatabaseName: dbname,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while requesting API to delete Sharded PostgreSQL database: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while waiting for operation to delete Sharded PostgreSQL database: %s", err.Error()),
		)
	}
}
