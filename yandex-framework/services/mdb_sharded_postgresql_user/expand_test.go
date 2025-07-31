package mdb_sharded_postgresql_user

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBSPQRUserGrantsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname string
		grants   types.Set
		expected []string
	}{
		{
			"ManyGrants",
			types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("writer"),
				types.StringValue("reader"),
				types.StringValue("admin"),
			}),
			[]string{
				"writer",
				"reader",
				"admin",
			},
		},
		{
			"EmptyGrants",
			types.SetValueMust(types.StringType, []attr.Value{}),
			[]string{},
		},
	}

	for _, tc := range cases {
		output, diags := expandGrants(ctx, tc.grants)
		if diags.HasError() {
			t.Errorf("Unexpected expand diagnostics status %s test: errors: %v", tc.testname, diags.Errors())
			continue
		}

		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				tc.testname,
				tc.expected,
				output,
			)
		}
	}
}

func TestYandexProvider_MDBSPQRUserPermissionsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname    string
		permissions types.Set
		expected    []*spqr.Permission
	}{
		{
			"ManyPermissions",
			types.SetValueMust(permissionType, []attr.Value{
				types.ObjectValueMust(permissionType.AttrTypes, map[string]attr.Value{
					"database": types.StringValue("testdb"),
				}),
				types.ObjectValueMust(permissionType.AttrTypes, map[string]attr.Value{
					"database": types.StringValue("anotherdb"),
				}),
			}),
			[]*spqr.Permission{
				{
					DatabaseName: "testdb",
				},
				{
					DatabaseName: "anotherdb",
				},
			},
		},
		{
			"EmptyPermissions",
			types.SetValueMust(types.StringType, []attr.Value{}),
			[]*spqr.Permission{},
		},
	}

	for _, tc := range cases {
		output, diags := expandPermissions(ctx, tc.permissions)
		if diags.HasError() {
			t.Errorf("Unexpected expand diagnostics status %s test: errors: %v", tc.testname, diags.Errors())
			continue
		}

		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				tc.testname,
				tc.expected,
				output,
			)
		}
	}
}

func TestYandexProvider_MDBSPQRUserSettingsExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname string
		settings mdbcommon.SettingsMapValue
		expected *spqr.UserSettings
	}{
		{
			"CheckFullAttributes",
			mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{
				"connection_limit":   types.StringValue("10"),
				"connection_retries": types.StringValue("10"),
			}, attrProvider),
			&spqr.UserSettings{
				ConnectionLimit:   wrapperspb.Int64(10),
				ConnectionRetries: wrapperspb.Int64(10),
			},
		},
		{
			"CheckEmptyAttributes",
			mdbcommon.NewSettingsMapValueMust(map[string]attr.Value{}, attrProvider),
			&spqr.UserSettings{},
		},
	}

	for _, tc := range cases {
		output, diags := expandSettings(ctx, tc.settings)
		if diags.HasError() {
			t.Errorf("Unexpected expand diagnostics status %s test: errors: %v", tc.testname, diags.Errors())
			continue
		}

		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				tc.testname,
				tc.expected,
				output,
			)
		}
	}
}
