package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

// Set access to default if null
func expandAccess(ctx context.Context, cfgAccess types.Object, diags *diag.Diagnostics) *postgresql.Access {
	var access Access
	diags.Append(cfgAccess.As(ctx, &access, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}
	return &postgresql.Access{
		WebSql:       access.WebSql.ValueBool(),
		DataLens:     access.DataLens.ValueBool(),
		DataTransfer: access.DataTransfer.ValueBool(),
		Serverless:   access.Serverless.ValueBool(),
	}
}
