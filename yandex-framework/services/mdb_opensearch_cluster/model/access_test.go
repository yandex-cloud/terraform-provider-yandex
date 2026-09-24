package model

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
)

func TestConfigToStateAccess(t *testing.T) {
	accessValue := func(dataTransfer, serverless types.Bool) types.Object {
		return types.ObjectValueMust(accessAttrTypes, map[string]attr.Value{
			"data_transfer": dataTransfer,
			"serverless":    serverless,
		})
	}
	nullAccess := types.ObjectNull(accessAttrTypes)
	disabled := accessValue(types.BoolValue(false), types.BoolValue(false))
	partial := accessValue(types.BoolValue(false), types.BoolNull())
	empty := accessValue(types.BoolNull(), types.BoolNull())
	enabled := accessValue(types.BoolValue(true), types.BoolValue(false))

	for _, tc := range []struct {
		name  string
		prior types.Object
		api   *opensearch.Access
		want  types.Object
	}{
		{name: "explicit false omitted by API", prior: disabled, want: disabled},
		{name: "explicit false returned by API", prior: disabled, api: &opensearch.Access{}, want: disabled},
		{name: "absent block omitted by API", prior: nullAccess, want: nullAccess},
		{name: "absent block returned empty by API", prior: nullAccess, api: &opensearch.Access{}, want: nullAccess},
		{name: "partially configured block", prior: partial, want: partial},
		{name: "empty configured block", prior: empty, want: empty},
		{name: "optional field omitted", prior: partial, api: &opensearch.Access{}, want: partial},
		{name: "enabled flag reset outside Terraform", prior: enabled, want: disabled},
		{name: "flag enabled outside Terraform", prior: disabled, api: &opensearch.Access{DataTransfer: true}, want: enabled},
		{name: "absent flag enabled outside Terraform", prior: partial, api: &opensearch.Access{Serverless: true}, want: accessValue(types.BoolValue(false), types.BoolValue(true))},
		{name: "unknown flags become known", prior: accessValue(types.BoolUnknown(), types.BoolUnknown()), want: disabled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			attrs := resourceConfigValue(t, types.Int64Null()).Attributes()
			attrs["access"] = tc.prior
			state := &OpenSearch{Config: types.ObjectValueMust(ConfigAttrTypes, attrs)}
			apiConfig := &opensearch.ClusterConfig{
				Version:    "2",
				Opensearch: &opensearch.OpenSearch{},
				Access:     tc.api,
			}

			got, diags := configToState(ctx, apiConfig, state)
			if diags.HasError() {
				t.Fatalf("configToState() diagnostics: %#v", diags)
			}
			if actual := got.Attributes()["access"]; !actual.Equal(tc.want) {
				t.Fatalf("access = %s, want %s", actual, tc.want)
			}
		})
	}
}
