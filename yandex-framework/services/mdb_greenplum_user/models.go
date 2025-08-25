package mdb_greenplum_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

type User struct {
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
}

func userFromState(_ context.Context, state *User) *greenplum.User {
	u := &greenplum.User{
		Name:          state.Name.ValueString(),
		Password:      "",
		ResourceGroup: state.ResourceGroup.ValueString(),
	}
	if state.Password != nil {
		u.Password = *state.Password
	}
	return u
}
