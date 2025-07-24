package mdb_greenplum_user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *greenplum.User {
	users, err := sdk.MDB().Greenplum().User().List(ctx, &greenplum.ListUsersRequest{
		ClusterId: cid,
	})
	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get Greenplum user: "+err.Error(),
		)
		return nil
	}

	for _, u := range users.GetUsers() {
		if u.GetName() == userName {
			return u
		}
	}
	diag.AddError(
		"Failed to Read resource",
		fmt.Sprintf("Greenplum user %q not found", userName),
	)
	return nil
}

func createUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *greenplum.User) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().User().Create(ctx, &greenplum.CreateUserRequest{
			ClusterId: cid,
			User:      user,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create Greenplum user: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Greenplum user: "+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *greenplum.User, updatePaths []string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().User().Update(ctx, &greenplum.UpdateUserRequest{
			ClusterId:  cid,
			User:       user,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update Greenplum user: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update Greenplum user: "+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().User().Delete(ctx, &greenplum.DeleteUserRequest{
			ClusterId: cid,
			UserName:  userName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Greenplum user: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Greenplum user: "+err.Error(),
		)
	}
}
