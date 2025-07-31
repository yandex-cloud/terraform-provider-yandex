package mdb_sharded_postgresql_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type User struct {
	Id          types.String               `tfsdk:"id"`
	ClusterID   types.String               `tfsdk:"cluster_id"`
	Name        types.String               `tfsdk:"name"`
	Password    types.String               `tfsdk:"password"`
	Grants      types.Set                  `tfsdk:"grants"`
	Permissions types.Set                  `tfsdk:"permissions"`
	Settings    mdbcommon.SettingsMapValue `tfsdk:"settings"`
}

type Permission struct {
	DatabaseName types.String `tfsdk:"database"`
}

var permissionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database": types.StringType,
	},
}

func userToState(ctx context.Context, user *spqr.User, state *User) diag.Diagnostics {
	state.ClusterID = types.StringValue(user.ClusterId)
	state.Name = types.StringValue(user.Name)
	state.Permissions = flattenPermissions(user.Permissions)
	state.Grants = flattenGrants(user.Grants)
	var diags diag.Diagnostics
	state.Settings = flattenSettings(ctx, user.Settings, &diags)
	return diags
}

func userFromState(ctx context.Context, state *User) (*spqr.UserSpec, diag.Diagnostics) {
	settings, diags := expandSettings(ctx, state.Settings)
	grants, grantDiags := expandGrants(ctx, state.Grants)
	diags.Append(grantDiags...)
	perms, permDiags := expandPermissions(ctx, state.Permissions)
	diags.Append(permDiags...)
	u := &spqr.UserSpec{
		Name:        state.Name.ValueString(),
		Settings:    settings,
		Grants:      grants,
		Permissions: perms,
		Password:    state.Password.ValueString(),
	}

	return u, diags
}
