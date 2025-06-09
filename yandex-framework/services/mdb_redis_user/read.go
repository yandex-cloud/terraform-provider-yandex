package mdb_redis_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	redis "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func userRead(ctx context.Context, sdk *ycsdk.SDK, diagnostics *diag.Diagnostics, state *User) {
	cid := state.ClusterID.ValueString()
	name := state.Name.ValueString()
	user := readUser(ctx, sdk, diagnostics, cid, name)
	if diagnostics.HasError() {
		return
	}

	state.ACLOptions = types.StringValue(user.AclOptions)
	state.Enabled = types.BoolValue(user.Enabled)
	permissions, diags := flattenPermissions(ctx, user.Permissions)
	state.Permissions = permissions
	diagnostics.Append(diags...)
}

func flattenPermissions(ctx context.Context, permissions *redis.Permissions) (types.Object, diag.Diagnostics) {
	res := Permissions{
		Commands:        types.StringValue(permissions.Commands.GetValue()),
		Categories:      types.StringValue(permissions.Categories.GetValue()),
		Patterns:        types.StringValue(permissions.Patterns.GetValue()),
		PubSubChannels:  types.StringValue(permissions.PubSubChannels.GetValue()),
		SanitizePayload: types.StringValue(permissions.SanitizePayload.GetValue()),
	}

	return types.ObjectValueFrom(ctx, permissionType.AttributeTypes(), res)
}
