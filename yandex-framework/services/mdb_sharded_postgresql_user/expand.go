package mdb_sharded_postgresql_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func expandSettings(ctx context.Context, settings mdbcommon.SettingsMapValue) (*spqr.UserSettings, diag.Diagnostics) {
	a := protobuf_adapter.NewProtobufMapDataAdapter()
	userSettings := &spqr.UserSettings{}
	var diags diag.Diagnostics
	a.Fill(ctx, userSettings, settings.PrimitiveElements(ctx, &diags), &diags)
	return userSettings, diags
}

func expandGrants(ctx context.Context, grants types.Set) ([]string, diag.Diagnostics) {
	grantsType := make([]string, 0, len(grants.Elements()))
	diag := grants.ElementsAs(ctx, &grantsType, false)
	return grantsType, diag
}

func expandPermissions(ctx context.Context, perms types.Set) ([]*spqr.Permission, diag.Diagnostics) {
	permType := make([]Permission, 0, len(perms.Elements()))
	d := perms.ElementsAs(ctx, &permType, false)
	permissions := make([]*spqr.Permission, 0, len(permType))
	for _, p := range permType {
		permissions = append(permissions, &spqr.Permission{
			DatabaseName: p.DatabaseName.ValueString(),
		})
	}
	return permissions, d
}
