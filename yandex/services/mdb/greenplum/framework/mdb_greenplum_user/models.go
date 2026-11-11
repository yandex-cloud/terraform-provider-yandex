package mdb_greenplum_user

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

type User struct {
	Id                types.String   `tfsdk:"id"`
	ClusterID         types.String   `tfsdk:"cluster_id"`
	Name              types.String   `tfsdk:"name"`
	Password          *string        `tfsdk:"password"`
	PasswordWo        types.String   `tfsdk:"password_wo"`
	PasswordWoVersion types.Int64    `tfsdk:"password_wo_version"`
	ResourceGroup     types.String   `tfsdk:"resource_group"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

type dataSourceUser struct {
	Id            types.String   `tfsdk:"id"`
	ClusterID     types.String   `tfsdk:"cluster_id"`
	Name          types.String   `tfsdk:"name"`
	Password      *string        `tfsdk:"password"`
	ResourceGroup types.String   `tfsdk:"resource_group"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func userToState(user *greenplum.User, state *User) {
	state.Name = types.StringValue(user.Name)
	state.ResourceGroup = types.StringValue(user.ResourceGroup)
	state.PasswordWo = types.StringNull()
}

func dataSourceUserToState(user *greenplum.User, state *dataSourceUser) {
	state.Name = types.StringValue(user.Name)
	state.ResourceGroup = types.StringValue(user.ResourceGroup)
}

func userFromState(state *User, password string) *greenplum.User {
	return &greenplum.User{
		Name:          state.Name.ValueString(),
		Password:      password,
		ResourceGroup: state.ResourceGroup.ValueString(),
	}
}
