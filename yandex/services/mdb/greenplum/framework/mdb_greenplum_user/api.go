package mdb_greenplum_user

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/operationcompat"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func requestCreateUser(ctx context.Context, sdk *ycsdk.SDK, req *greenplum.CreateUserRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, greenplumsdk.UserCreate)
	if err != nil {
		return nil, err
	}
	return greenplum.NewUserServiceClient(conn).Create(ctx, req)
}

func requestUpdateUser(ctx context.Context, sdk *ycsdk.SDK, req *greenplum.UpdateUserRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, greenplumsdk.UserUpdate)
	if err != nil {
		return nil, err
	}
	return greenplum.NewUserServiceClient(conn).Update(ctx, req)
}

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *greenplum.User {
	users, err := greenplumsdk.NewUserClient(sdk).List(ctx, &greenplum.ListUsersRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*operationpb.Operation, error) {
		return requestCreateUser(ctx, sdk, &greenplum.CreateUserRequest{
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

	if err = operationcompat.Wait(ctx, sdk, op.GetId()); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Greenplum user: "+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *greenplum.User, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*operationpb.Operation, error) {
		return requestUpdateUser(ctx, sdk, &greenplum.UpdateUserRequest{
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

	if err = operationcompat.Wait(ctx, sdk, op.GetId()); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update Greenplum user: "+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*greenplumsdk.UserDeleteOperation, error) {
		return greenplumsdk.NewUserClient(sdk).Delete(ctx, &greenplum.DeleteUserRequest{
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Greenplum user: "+err.Error(),
		)
	}
}
