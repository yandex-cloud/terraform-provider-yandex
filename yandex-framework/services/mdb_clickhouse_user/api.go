package mdb_clickhouse_user

import (
	"context"
	"github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, userName string) *clickhouse.User {
	user, err := clickhousesdk.NewUserClient(sdk).Get(ctx, &clickhouse.GetUserRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.UserCreateOperation, error) {
		return clickhousesdk.NewUserClient(sdk).Create(ctx, &clickhouse.CreateUserRequest{
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create ClickHouse user:"+err.Error(),
		)
	}
}

func updateUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, user *clickhouse.UserSpec, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.UserUpdateOperation, error) {
		return clickhousesdk.NewUserClient(sdk).Update(ctx, &clickhouse.UpdateUserRequest{
			ClusterId:        cid,
			UserName:         user.Name,
			Password:         user.Password,
			Permissions:      user.Permissions,
			Settings:         user.Settings,
			Quotas:           user.Quotas,
			GeneratePassword: user.GeneratePassword,
			AuthMethod:       user.AuthMethod,
			UpdateMask:       &fieldmaskpb.FieldMask{Paths: updatePaths},
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update ClickHouse user:"+err.Error(),
		)
	}
}

func deleteUser(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, userName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.UserDeleteOperation, error) {
		return clickhousesdk.NewUserClient(sdk).Delete(ctx, &clickhouse.DeleteUserRequest{
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete ClickHouse user:"+err.Error(),
		)
	}
}
