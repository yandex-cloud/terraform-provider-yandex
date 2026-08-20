package mdb_sharded_postgresql_user

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-sdk/services/mdb/spqr/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var shardedPostgreSQLAPI = ShardedPostgreSQLAPI{}

type ShardedPostgreSQLAPI struct{}

func (r *ShardedPostgreSQLAPI) ReadUser(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, userName string) *spqr.User {
	users, err := spqrsdk.NewUserClient(sdk).List(ctx, &spqr.ListUsersRequest{
		ClusterId: cid,
	})
	if err != nil {
		diags.AddError(
			"Failed to Read resources",
			fmt.Sprintf("Error while requesting API to get Sharded PostgreSQL user: %s", err.Error()),
		)
		return nil
	}

	for _, u := range users.GetUsers() {
		if u.GetName() == userName {
			return u
		}
	}

	diags.AddError(
		"Failed to Read resource",
		fmt.Sprintf("Sharded PostgreSQL user %q not found", userName),
	)
	return nil
}

func (r *ShardedPostgreSQLAPI) CreateUser(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, user *spqr.UserSpec) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*spqrsdk.UserCreateOperation, error) {
		return spqrsdk.NewUserClient(sdk).Create(ctx, &spqr.CreateUserRequest{
			ClusterId: cid,
			UserSpec:  user,
		})
	})
	if err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while requesting API to create Sharded PostgreSQL user: %s", err.Error()),
		)
		return
	}
	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Create resource",
			fmt.Sprintf("Error while waiting for operation to create Sharded PostgreSQL user: %s", err.Error()),
		)
	}
}

func (r *ShardedPostgreSQLAPI) UpdateUser(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, user *spqr.UserSpec, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*spqrsdk.UserUpdateOperation, error) {
		return spqrsdk.NewUserClient(sdk).Update(ctx, &spqr.UpdateUserRequest{
			ClusterId:   cid,
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
			Settings:    user.Settings,
			Grants:      user.Grants,
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		diags.AddError(
			"Failed to Update resource",
			fmt.Sprintf("Error while requesting API to update Sharded PostgreSQL user: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to Update resource",
			fmt.Sprintf("Error while waiting for operation to update Sharded PostgreSQL user: %s", err.Error()),
		)
	}
}

func (r *ShardedPostgreSQLAPI) DeleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*spqrsdk.UserDeleteOperation, error) {
		return spqrsdk.NewUserClient(sdk).Delete(ctx, &spqr.DeleteUserRequest{
			ClusterId: cid,
			UserName:  userName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while requesting API to delete Sharded PostgreSQL user: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			fmt.Sprintf("Error while waiting for operation to delete Sharded PostgreSQL user: %s", err.Error()),
		)
	}
}
