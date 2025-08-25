package mdb_redis_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type User struct {
	Id          types.String   `tfsdk:"id"`
	ClusterID   types.String   `tfsdk:"cluster_id"`
	Name        types.String   `tfsdk:"name"`
	Permissions types.Object   `tfsdk:"permissions"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	Passwords   types.Set      `tfsdk:"passwords"`
	ACLOptions  types.String   `tfsdk:"acl_options"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

type Permissions struct {
	Commands        types.String `tfsdk:"commands"`
	Categories      types.String `tfsdk:"categories"`
	Patterns        types.String `tfsdk:"patterns"`
	PubSubChannels  types.String `tfsdk:"pub_sub_channels"`
	SanitizePayload types.String `tfsdk:"sanitize_payload"`
}

var permissionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"commands":         types.StringType,
		"categories":       types.StringType,
		"patterns":         types.StringType,
		"pub_sub_channels": types.StringType,
		"sanitize_payload": types.StringType,
	},
}

func userToState(ctx context.Context, user *redis.User, state *User) diag.Diagnostics {
	state.Name = types.StringValue(user.Name)
	state.ClusterID = types.StringValue(user.ClusterId)
	state.Enabled = types.BoolValue(user.Enabled)
	state.ACLOptions = types.StringValue(user.AclOptions)

	if state.Passwords.IsNull() {
		state.Passwords = types.SetNull(types.StringType)
	}

	return permissionsToState(ctx, user.Permissions, state)
}

func permissionsToState(ctx context.Context, perms *redis.Permissions, state *User) diag.Diagnostics {
	var permissions Permissions

	permissions.Commands = types.StringValue(perms.Commands.GetValue())
	permissions.Categories = types.StringValue(perms.Categories.GetValue())
	permissions.Patterns = types.StringValue(perms.Patterns.GetValue())
	permissions.PubSubChannels = types.StringValue(perms.PubSubChannels.GetValue())
	permissions.SanitizePayload = types.StringValue(perms.SanitizePayload.GetValue())

	permissionsObject, diags := types.ObjectValueFrom(ctx, permissionType.AttrTypes, permissions)

	state.Permissions = permissionsObject

	return diags
}

func userFromState(ctx context.Context, state *User) (*redis.UserSpec, diag.Diagnostics) {
	permissions, diags := permissionsFromState(ctx, state.Permissions)

	var passwords = make([]string, 0, len(state.Passwords.Elements()))
	diags.Append(state.Passwords.ElementsAs(ctx, &passwords, false)...)

	return &redis.UserSpec{
		Name:        state.Name.ValueString(),
		Passwords:   passwords,
		Enabled:     &wrapperspb.BoolValue{Value: state.Enabled.ValueBool()},
		Permissions: permissions,
	}, diags
}

func permissionsFromState(ctx context.Context, u types.Object) (*redis.Permissions, diag.Diagnostics) {
	if !utils.IsPresent(u) {
		return nil, nil
	}

	permissions := &Permissions{}
	diags := u.As(ctx, permissions, basetypes.ObjectAsOptions{})

	if diags.HasError() {
		return nil, nil
	}

	res := &redis.Permissions{}

	if utils.IsPresent(permissions.Commands) {
		res.Commands = &wrapperspb.StringValue{Value: permissions.Commands.ValueString()}
	}
	if utils.IsPresent(permissions.Categories) {
		res.Categories = &wrapperspb.StringValue{Value: permissions.Categories.ValueString()}
	}
	if utils.IsPresent(permissions.Patterns) {
		res.Patterns = &wrapperspb.StringValue{Value: permissions.Patterns.ValueString()}
	}
	if utils.IsPresent(permissions.PubSubChannels) {
		res.PubSubChannels = &wrapperspb.StringValue{Value: permissions.PubSubChannels.ValueString()}
	}
	if utils.IsPresent(permissions.SanitizePayload) {
		res.SanitizePayload = &wrapperspb.StringValue{Value: permissions.SanitizePayload.ValueString()}
	}

	return res, diags
}
