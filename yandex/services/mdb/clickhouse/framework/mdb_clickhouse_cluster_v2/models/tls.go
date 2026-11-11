package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type ClickhouseTls struct {
	TrustedCertificates types.List `tfsdk:"trusted_certificates"`
}

var ClickhouseTlsAttrTypes = map[string]attr.Type{
	"trusted_certificates": types.ListType{ElemType: types.StringType},
}

func flattenClickhouseTls(ctx context.Context, tls *clickhouseConfig.ClickhouseConfig_Tls, diags *diag.Diagnostics) types.Object {
	if tls == nil {
		return types.ObjectNull(ClickhouseTlsAttrTypes)
	}

	certs, d := types.ListValueFrom(ctx, types.StringType, tls.TrustedCertificates)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, ClickhouseTlsAttrTypes, ClickhouseTls{
		TrustedCertificates: certs,
	})
	diags.Append(d...)

	return obj
}

func expandClickhouseTls(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_Tls {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var tls ClickhouseTls
	diags.Append(obj.As(ctx, &tls, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	var certs []string
	diags.Append(tls.TrustedCertificates.ElementsAs(ctx, &certs, false)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_Tls{
		TrustedCertificates: certs,
	}
}
