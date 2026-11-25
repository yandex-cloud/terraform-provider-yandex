package mdb_redis_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
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
	state.PasswordWo = types.StringNull()
	diagnostics.Append(permissionsToState(ctx, user.Permissions, state)...)
}
