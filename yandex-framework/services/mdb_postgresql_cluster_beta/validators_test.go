package mdb_postgresql_cluster_beta

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func builTestMaintenanceWindowConfigSchema(blockName string) schema.Schema {
	return schema.Schema{
		Blocks: map[string]schema.Block{
			blockName: schema.SingleNestedBlock{
				Validators: []validator.Object{
					NewMaintenanceWindowStructValidator(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional: true,
					},
					"day": schema.StringAttribute{
						Optional: true,
					},
					"hour": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func builTestMaintenanceWindowExplicitBlockObjectsRequest(mwType, mwDay *string, mwHour *int64) validator.ObjectRequest {
	const testBlockName = "maintenance_window_test_block_explicit"

	reqConf := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
			testBlockName: tftypes.NewValue(
				tftypes.Object{}, map[string]tftypes.Value{
					"type": tftypes.NewValue(tftypes.String, mwType),
					"day":  tftypes.NewValue(tftypes.String, mwDay),
					"hour": tftypes.NewValue(tftypes.Number, mwHour),
				},
			),
		}),
		Schema: builTestMaintenanceWindowConfigSchema(testBlockName),
	}

	return validator.ObjectRequest{
		Config: reqConf,
		ConfigValue: basetypes.NewObjectValueMust(
			map[string]attr.Type{
				"type": types.StringType,
				"day":  types.StringType,
				"hour": types.Int64Type,
			},
			map[string]attr.Value{
				"type": types.StringPointerValue(mwType),
				"day":  types.StringPointerValue(mwDay),
				"hour": types.Int64PointerValue(mwHour),
			},
		),
		Path: path.Root(testBlockName),
	}
}

func builTestMaintenanceWindowEmptyBlockObjectsRequest() validator.ObjectRequest {
	const testBlockName = "maintenance_window_test_block_empty_block"

	return validator.ObjectRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
			Schema: builTestMaintenanceWindowConfigSchema(testBlockName),
		},
		ConfigValue: basetypes.NewObjectNull(
			map[string]attr.Type{
				"type": types.StringType,
				"day":  types.StringType,
				"hour": types.Int64Type,
			},
		),
		Path: path.Root(testBlockName),
	}
}

func TestYandexProvider_MDBPostgresClusterMaintenanceWindowStructValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := NewMaintenanceWindowStructValidator()

	anytimeType, weeklyType := "ANYTIME", "WEEKLY"
	weekday := "SAT"
	var hour int64 = 1

	cases := []struct {
		testname      string
		req           validator.ObjectRequest
		expectedError bool
	}{
		// Check ANYTIME and WEEKLY structures
		{
			testname:      "AnytimeWithWeekdayAndHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&anytimeType, &weekday, &hour),
			expectedError: true,
		},
		{
			testname:      "WeeklyWithWeekdayAndHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&weeklyType, &weekday, &hour),
			expectedError: false,
		},
		{
			testname:      "AnytimeWithoutWeekdayAndHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&anytimeType, nil, nil),
			expectedError: false,
		},
		{
			testname:      "WeeklyWithoutWeekdayAndHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&weeklyType, nil, nil),
			expectedError: true,
		},
		{
			testname:      "WeeklyWithoutWeekday",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&weeklyType, nil, &hour),
			expectedError: true,
		},
		{
			testname:      "WeeklyWithoutHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&weeklyType, &weekday, nil),
			expectedError: true,
		},
		{
			testname:      "AnytimeWithWeekday",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&anytimeType, &weekday, nil),
			expectedError: true,
		},
		{
			testname:      "AnytimeWithHour",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(&anytimeType, nil, &hour),
			expectedError: true,
		},
		{
			testname:      "EmptyRequest",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(nil, nil, nil),
			expectedError: true,
		},
		{
			testname:      "WithoutMWType",
			req:           builTestMaintenanceWindowExplicitBlockObjectsRequest(nil, &weekday, &hour),
			expectedError: true,
		},
		{
			testname:      "WithoutMWBlock",
			req:           builTestMaintenanceWindowEmptyBlockObjectsRequest(),
			expectedError: false,
		},
	}

	for _, c := range cases {
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, c.req, &resp)
		if resp.Diagnostics.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected validation status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				resp.Diagnostics.HasError(),
				resp.Diagnostics.Errors(),
			)
		}
	}
}
