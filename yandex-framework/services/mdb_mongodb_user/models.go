package mdb_mongodb_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

type User struct {
	Id         types.String `tfsdk:"id"`
	ClusterID  types.String `tfsdk:"cluster_id"`
	Name       types.String `tfsdk:"name"`
	Password   types.String `tfsdk:"password"`
	Permission types.Set    `tfsdk:"permission"`
}

type Permission struct {
	DatabaseName types.String `tfsdk:"database_name"`
	Roles        types.Set    `tfsdk:"roles"`
}

var permissionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database_name": types.StringType,
		"roles":         types.SetType{ElemType: types.StringType},
	},
}

func userToState(user *mongodb.User, state *User) diag.Diagnostics {
	state.Name = types.StringValue(user.Name)
	state.ClusterID = types.StringValue(user.ClusterId)

	return permissionsToState(user.Permissions, state)
}

func permissionsToState(permissions []*mongodb.Permission, state *User) diag.Diagnostics {
	var permissionValues []attr.Value

	var diags diag.Diagnostics
	for _, permission := range permissions {
		var stateRoles []attr.Value
		for _, role := range permission.Roles {
			stateRoles = append(stateRoles, types.StringValue(role))
		}

		value, diagnostics := types.SetValue(types.StringType, stateRoles)
		diags.Append(diagnostics...)
		permissionValue, diagnostics := types.ObjectValue(permissionType.AttrTypes, map[string]attr.Value{
			"database_name": types.StringValue(permission.DatabaseName),
			"roles":         value,
		})

		permissionValues = append(permissionValues, permissionValue)
		diags.Append(diagnostics...)

	}

	value, diagnostics := types.SetValue(permissionType, permissionValues)
	diags.Append(diagnostics...)

	state.Permission = value
	return diags
}

func userFromState(ctx context.Context, state *User) (*mongodb.UserSpec, diag.Diagnostics) {
	permissions, diags := permissionsFromState(ctx, state)
	return &mongodb.UserSpec{
		Name:        state.Name.ValueString(),
		Password:    state.Password.ValueString(),
		Permissions: permissions,
	}, diags
}

func permissionsFromState(ctx context.Context, state *User) ([]*mongodb.Permission, diag.Diagnostics) {
	permissions := make([]*mongodb.Permission, 0, len(state.Permission.Elements()))
	permissionsType := make([]Permission, 0, len(state.Permission.Elements()))
	diags := state.Permission.ElementsAs(ctx, &permissionsType, false)

	for _, permission := range permissionsType {
		roles := make([]string, 0, len(permission.Roles.Elements()))
		diags.Append(permission.Roles.ElementsAs(ctx, &roles, false)...)

		permissions = append(permissions, &mongodb.Permission{
			DatabaseName: permission.DatabaseName.ValueString(),
			Roles:        roles,
		})
	}
	return permissions, diags
}
