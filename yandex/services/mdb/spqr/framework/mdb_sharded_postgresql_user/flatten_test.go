package mdb_sharded_postgresql_user

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBSPQRUserSettingsFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname string
		settings *spqr.UserSettings
		expected mdbcommon.SettingsMapValue
	}{
		{
			"CheckFullAttributes",
			&spqr.UserSettings{
				ConnectionLimit:   wrapperspb.Int64(10),
				ConnectionRetries: wrapperspb.Int64(10),
			},
			mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{
				"connection_limit":   types.StringValue("10"),
				"connection_retries": types.StringValue("10"),
			}, attrProvider),
		},
		{
			"CheckPartialAttributes",
			&spqr.UserSettings{
				ConnectionLimit: wrapperspb.Int64(10),
			},
			mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{
				"connection_limit": types.StringValue("10"),
			}, attrProvider),
		},
		{
			"CheckNullAttributes",
			nil,
			mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{}, attrProvider),
		},
	}

	check := func(t *testing.T, testname string, cc attr.Value, exp attr.Value) {
		if (cc.IsNull() && !exp.IsNull()) || (!cc.IsNull() && exp.IsNull()) || (!exp.IsNull() && !exp.Equal(cc)) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				testname,
				exp,
				cc,
			)
		}
	}

	for _, tc := range cases {
		diags := &diag.Diagnostics{}
		output := flattenSettings(ctx, tc.settings, diags)
		if diags.HasError() {
			t.Errorf("Unexpected flatten diagnostics status %s test: errors: %v", tc.testname, diags.Errors())
			continue
		}
		check(t, tc.testname, output, tc.expected)
	}
}
