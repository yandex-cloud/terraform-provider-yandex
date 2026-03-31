package mdb_mysql_database_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func ReadDatabase(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, cid string, dbName string) *mysql.Database {
	db, err := mysqlv1sdk.NewDatabaseClient(providerConfig.SDKv2).Get(ctx, &mysql.GetDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			fmt.Sprintf("Error while requesting API to read MySQL database %q in cluster %q: %s", dbName, cid, err.Error()),
		)
		return nil
	}
	return db
}

func ListDatabases(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, cid string) []*mysql.Database {
	var databases []*mysql.Database
	pageToken := ""

	for {
		resp, err := mysqlv1sdk.NewDatabaseClient(providerConfig.SDKv2).List(ctx, &mysql.ListDatabasesRequest{
			ClusterId: cid,
			PageSize:  1000,
			PageToken: pageToken,
		})

		if err != nil {
			diag.AddError(
				"Failed to List databases",
				fmt.Sprintf("Error while requesting API to list MySQL databases in cluster %q: %s", cid, err.Error()),
			)
			return nil
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases
}

func CreateDatabase(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, req *mysql.CreateDatabaseRequest) {
	op, err := mysqlv1sdk.NewDatabaseClient(providerConfig.SDKv2).Create(ctx, req)

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while requesting API to create MySQL database %q in cluster %q: %s", req.DatabaseSpec.Name, req.ClusterId, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while waiting for operation to create MySQL database %q in cluster %q: %s", req.DatabaseSpec.Name, req.ClusterId, err.Error()),
		)
	}
}

func UpdateDatabase(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, req *mysql.UpdateDatabaseRequest) {
	op, err := mysqlv1sdk.NewDatabaseClient(providerConfig.SDKv2).Update(ctx, req)

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			fmt.Sprintf("Error while requesting API to update MySQL database %q in cluster %q: %s", req.DatabaseName, req.ClusterId, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			fmt.Sprintf("Error while waiting for operation to update MySQL database %q in cluster %q: %s", req.DatabaseName, req.ClusterId, err.Error()),
		)
	}
}

func DeleteDatabase(ctx context.Context, providerConfig *provider_config.Config, diag *diag.Diagnostics, cid string, dbName string) {
	op, err := mysqlv1sdk.NewDatabaseClient(providerConfig.SDKv2).Delete(ctx, &mysql.DeleteDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while requesting API to delete MySQL database %q in cluster %q: %s", dbName, cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while waiting for operation to delete MySQL database %q in cluster %q: %s", dbName, cid, err.Error()),
		)
	}
}
