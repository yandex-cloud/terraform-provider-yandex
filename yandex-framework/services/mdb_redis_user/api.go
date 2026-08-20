package mdb_redis_user

import (
	"context"
	"github.com/yandex-cloud/go-sdk/services/mdb/redis/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *redis.User {
	user, err := redissdk.NewUserClient(sdk).Get(ctx, &redis.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get Redis user: "+err.Error(),
		)
		return nil
	}
	return user
}

func createUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *redis.UserSpec) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*redissdk.UserCreateOperation, error) {
		return redissdk.NewUserClient(sdk).Create(ctx, &redis.CreateUserRequest{
			ClusterId: cid,
			UserSpec:  user,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create Redis user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Redis user: "+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *redis.UserSpec, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*redissdk.UserUpdateOperation, error) {
		return redissdk.NewUserClient(sdk).Update(ctx, &redis.UpdateUserRequest{
			ClusterId:   cid,
			UserName:    user.Name,
			Passwords:   user.Passwords,
			Permissions: user.Permissions,
			Enabled:     user.Enabled.Value,
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update Redis user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update Redis user: "+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*redissdk.UserDeleteOperation, error) {
		return redissdk.NewUserClient(sdk).Delete(ctx, &redis.DeleteUserRequest{
			ClusterId: cid,
			UserName:  userName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Redis user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Redis user: "+err.Error(),
		)
	}
}
