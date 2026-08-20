package mdb_mongodb_user

import (
	"context"
	"github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *mongodb.User {
	user, err := mongodbsdk.NewUserClient(sdk).Get(ctx, &mongodb.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get MongoDB user: "+err.Error(),
		)
		return nil
	}
	return user
}

func createUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *mongodb.UserSpec) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*mongodbsdk.UserCreateOperation, error) {
		return mongodbsdk.NewUserClient(sdk).Create(ctx, &mongodb.CreateUserRequest{
			ClusterId: cid,
			UserSpec:  user,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create MongoDB user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MongoDB user: "+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *mongodb.UserSpec, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*mongodbsdk.UserUpdateOperation, error) {
		return mongodbsdk.NewUserClient(sdk).Update(ctx, &mongodb.UpdateUserRequest{
			ClusterId:          cid,
			UserName:           user.Name,
			Password:           user.Password,
			Permissions:        user.Permissions,
			DeletionProtection: user.DeletionProtection,
			UpdateMask:         &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update MongoDB user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update MongoDB user: "+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*mongodbsdk.UserDeleteOperation, error) {
		return mongodbsdk.NewUserClient(sdk).Delete(ctx, &mongodb.DeleteUserRequest{
			ClusterId: cid,
			UserName:  userName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete MongoDB user: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MongoDB user: "+err.Error(),
		)
	}
}
