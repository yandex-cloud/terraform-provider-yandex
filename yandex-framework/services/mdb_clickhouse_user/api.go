package mdb_clickhouse_user

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *clickhouse.User {
	user, err := sdk.MDB().Clickhouse().User().Get(ctx, &clickhouse.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})

	if err != nil {

		if validate.IsStatusWithCode(err, codes.NotFound) {
			diag.AddWarning(
				"Failed to Read resource",
				"User "+userName+" not found in cluster "+cid,
			)
		} else {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to get ClickHouse user:"+err.Error(),
			)

		}
		return nil
	}
	return user
}

func createUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *clickhouse.UserSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Clickhouse().User().Create(ctx, &clickhouse.CreateUserRequest{
			ClusterId: cid,
			UserSpec:  user,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create ClickHouse user:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create ClickHouse user:"+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *clickhouse.UserSpec, updatePaths []string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Clickhouse().User().Update(ctx, &clickhouse.UpdateUserRequest{
			ClusterId:   cid,
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
			Settings:    user.Settings,
			Quotas:      user.Quotas,
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		if !strings.EqualFold(errorMessage(err), "no changes detected") {
			diag.AddError(
				"Failed to Update resource",
				"Error while requesting API to update ClickHouse user:"+err.Error(),
			)
		}
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update ClickHouse user:"+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Clickhouse().User().Delete(ctx, &clickhouse.DeleteUserRequest{
			ClusterId: cid,
			UserName:  userName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete ClickHouse user:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete ClickHouse user:"+err.Error(),
		)
	}
}
