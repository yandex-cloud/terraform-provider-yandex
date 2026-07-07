package mdb_mongodb_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
)

const (
	authTypePassword = "PASSWORD"
	authTypeIam      = "IAM"
)

type User struct {
	Id                 types.String   `tfsdk:"id"`
	ClusterID          types.String   `tfsdk:"cluster_id"`
	Name               types.String   `tfsdk:"name"`
	Password           types.String   `tfsdk:"password"`
	AuthType           types.String   `tfsdk:"auth_type"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	Permission         types.Set      `tfsdk:"permission"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

// authTypeToState maps a MongoDB API auth type into a Terraform-friendly value.
func authTypeToState(authType mongodb.AuthType) types.String {
	switch authType {
	case mongodb.AuthType_AUTH_TYPE_IAM:
		return types.StringValue(authTypeIam)
	default:
		return types.StringValue(authTypePassword)
	}
}

// authTypeFromState maps a Terraform value into a MongoDB API auth type.
func authTypeFromState(authType types.String) mongodb.AuthType {
	switch authType.ValueString() {
	case authTypeIam:
		return mongodb.AuthType_AUTH_TYPE_IAM
	default:
		return mongodb.AuthType_AUTH_TYPE_PASSWORD
	}
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
	state.AuthType = authTypeToState(user.GetAuthType())
	state.DeletionProtection = wrappers.BoolToTF(user.GetDeletionProtection())

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
		Name:               state.Name.ValueString(),
		Password:           state.Password.ValueString(),
		AuthType:           authTypeFromState(state.AuthType),
		DeletionProtection: wrappers.BoolFromTF(state.DeletionProtection),
		Permissions:        permissions,
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
