package mdb_mysql_user_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func ReadUser(
	ctx context.Context,
	providerConfig *provider_config.Config,
	diags *diag.Diagnostics,
	cid string,
	userName string,
) *mysql.User {
	user, err := mysqlv1sdk.NewUserClient(providerConfig.SDKv2).Get(ctx, &mysql.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})
	if err != nil {
		diags.AddError(
			"Failed to Read resource",
			fmt.Sprintf(
				"Error while requesting API to read MySQL user %q in cluster %q: %s",
				userName, cid, err.Error(),
			),
		)
		return nil
	}
	return user
}

func ListUsers(
	ctx context.Context,
	providerConfig *provider_config.Config,
	diags *diag.Diagnostics,
	cid string,
) []*mysql.User {
	var users []*mysql.User
	pageToken := ""

	for {
		resp, err := mysqlv1sdk.NewUserClient(providerConfig.SDKv2).List(ctx, &mysql.ListUsersRequest{
			ClusterId: cid,
			PageSize:  1000,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to List users",
				fmt.Sprintf(
					"Error while requesting API to list MySQL users in cluster %q: %s",
					cid, err.Error(),
				),
			)
			return nil
		}

		users = append(users, resp.Users...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return users
}

func CreateUser(
	ctx context.Context,
	providerConfig *provider_config.Config,
	diags *diag.Diagnostics,
	req *mysql.CreateUserRequest,
) {
	op, err := mysqlv1sdk.NewUserClient(providerConfig.SDKv2).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf(
				"Error while requesting API to create MySQL user %q in cluster %q: %s",
				req.UserSpec.Name, req.ClusterId, err.Error(),
			),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf(
				"Error while waiting for operation to create MySQL user %q in cluster %q: %s",
				req.UserSpec.Name, req.ClusterId, err.Error(),
			),
		)
	}
}

func UpdateUser(
	ctx context.Context,
	providerConfig *provider_config.Config,
	diags *diag.Diagnostics,
	req *mysql.UpdateUserRequest,
) {
	op, err := mysqlv1sdk.NewUserClient(providerConfig.SDKv2).Update(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to Update resource",
			fmt.Sprintf(
				"Error while requesting API to update MySQL user %q in cluster %q: %s",
				req.UserName, req.ClusterId, err.Error(),
			),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Update resource",
			fmt.Sprintf(
				"Error while waiting for operation to update MySQL user %q in cluster %q: %s",
				req.UserName, req.ClusterId, err.Error(),
			),
		)
	}
}

func DeleteUser(
	ctx context.Context,
	providerConfig *provider_config.Config,
	diags *diag.Diagnostics,
	cid string,
	userName string,
) {
	op, err := mysqlv1sdk.NewUserClient(providerConfig.SDKv2).Delete(ctx, &mysql.DeleteUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})
	if err != nil {
		diags.AddError(
			"Failed to Delete resource",
			fmt.Sprintf(
				"Error while requesting API to delete MySQL user %q in cluster %q: %s",
				userName, cid, err.Error(),
			),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Delete resource",
			fmt.Sprintf(
				"Error while waiting for operation to delete MySQL user %q in cluster %q: %s",
				userName, cid, err.Error(),
			),
		)
	}
}
