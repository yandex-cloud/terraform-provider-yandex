package mdb_postgresql_cluster_beta

import (
	"context"
	"math/rand"
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

func builTestOneOfIfConfiguredConfigSchema(attrName string) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			attrName: schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"attr1": schema.StringAttribute{
						Optional: true,
					},
					"attr2": schema.Int64Attribute{
						Optional: true,
					},
					"attr3": schema.BoolAttribute{
						Computed: true,
						Optional: true,
					},
					"attr4": schema.Int64Attribute{
						Optional: true,
					},
					"attr5": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func builTestOneOfIfConfiguredBlockObjectsRequest(attrName string) validator.ObjectRequest {
	return validator.ObjectRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				attrName: tftypes.NewValue(
					tftypes.Object{}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "string"),
						"attr2": tftypes.NewValue(tftypes.Number, 1),
						"attr3": tftypes.NewValue(tftypes.Bool, nil),
						"attr4": tftypes.NewValue(tftypes.Number, nil),
						"attr5": tftypes.NewValue(tftypes.String, nil),
					},
				),
			}),
			Schema: builTestOneOfIfConfiguredConfigSchema(attrName),
		},
		Path: path.Root(attrName),
	}
}

func TestYandexProvider_MDBPostgresClusterOneOfIfConfiguredValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	attrName := "one_of_if_configured_block"

	req := builTestOneOfIfConfiguredBlockObjectsRequest(attrName)
	req.ConfigValue = basetypes.NewObjectValueMust(
		map[string]attr.Type{
			"attr1": types.StringType,
			"attr2": types.Int64Type,
			"attr3": types.BoolType,
			"attr4": types.Int64Type,
			"attr5": types.StringType,
		},
		map[string]attr.Value{
			"attr1": types.StringValue("string"),
			"attr2": types.Int64Value(1),
			"attr3": types.BoolUnknown(),
			"attr4": types.Int64Null(),
			"attr5": types.StringNull(),
		},
	)

	cases := []struct {
		validator     *atLeastIfConfiguredValidator
		testname      string
		expectedError bool
	}{
		{
			testname: "AllConfiguredAttributes",
			validator: NewAtLeastIfConfiguredValidator(
				path.MatchRoot(attrName).AtName("attr1"),
				path.MatchRoot(attrName).AtName("attr2"),
			),
		},
		{
			testname: "PartlyConfiguredAttributes",
			validator: NewAtLeastIfConfiguredValidator(
				path.MatchRoot(attrName).AtName("attr1"),
				path.MatchRoot(attrName).AtName("attr3"),
				path.MatchRoot(attrName).AtName("attr4"),
			),
		},
		{
			testname: "UnknownConfiguredAttributes",
			validator: NewAtLeastIfConfiguredValidator(
				path.MatchRoot(attrName).AtName("attr3"),
			),
			expectedError: true,
		},
		{
			testname: "NullConfiguredAttributes",
			validator: NewAtLeastIfConfiguredValidator(
				path.MatchRoot(attrName).AtName("attr5"),
			),
			expectedError: true,
		},
		{
			testname: "NullUnknownConfiguredAttributes",
			validator: NewAtLeastIfConfiguredValidator(
				path.MatchRoot(attrName).AtName("attr3"),
				path.MatchRoot(attrName).AtName("attr4"),
			),
			expectedError: true,
		},
		{
			testname:      "EmptyAttributes",
			validator:     NewAtLeastIfConfiguredValidator(),
			expectedError: true,
		},
	}

	for _, c := range cases {
		var resp validator.ObjectResponse
		c.validator.ValidateObject(ctx, req, &resp)
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

func TestYandexProvider_MDBPostgresClusterOneOfIfConfiguredEmptyValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	attrName := "one_of_if_configured_block_empty"

	reqEmptyVal := builTestOneOfIfConfiguredBlockObjectsRequest(attrName)
	reqEmptyVal.ConfigValue = basetypes.NewObjectNull(
		map[string]attr.Type{
			"attr1": types.StringType,
			"attr2": types.Int64Type,
			"attr3": types.BoolType,
			"attr4": types.Int64Type,
			"attr5": types.StringType,
		},
	)
	resp := validator.ObjectResponse{}
	NewAtLeastIfConfiguredValidator(
		path.MatchRoot(attrName).AtName("attr1"),
		path.MatchRoot(attrName).AtName("attr2"),
	).ValidateObject(ctx, reqEmptyVal, &validator.ObjectResponse{})

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected validation status: expected %t, actual %t with errors: %v", false, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
	}
}

func builTestGreaterValidatorRequest(value int64) validator.Int64Request {
	reqConf := tfsdk.Config{
		Raw: tftypes.NewValue(
			tftypes.Object{}, map[string]tftypes.Value{
				"attr1": tftypes.NewValue(tftypes.Number, 5),
				"block1": tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
					"attr3": tftypes.NewValue(tftypes.Number, 10),
					"attr4": tftypes.NewValue(tftypes.Number, 15),
				}),
			},
		),
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"attr1": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						Int64GreaterValidator(),
					},
				},
				"attr2": schema.Int64Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"block1": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"attr3": schema.Int64Attribute{
							Required: true,
						},
						"attr4": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
		},
	}

	return validator.Int64Request{
		Config:      reqConf,
		ConfigValue: types.Int64Value(value),
		Path:        path.Root("attr1"),
	}
}

func TestYandexProvider_MDBPostgresClusterInt64GreaterValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		validator     *int64GreaterValidator
		req           validator.Int64Request
		expectedError bool
	}{
		{
			testname:      "CheckWithNullComparing",
			validator:     Int64GreaterValidator(path.MatchRoot("attr2")),
			req:           builTestGreaterValidatorRequest(rand.Int63()),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64ComparingSuccess",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3")),
			req:           builTestGreaterValidatorRequest(11),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64ComparingFailed",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3")),
			req:           builTestGreaterValidatorRequest(5),
			expectedError: true,
		},
		{
			testname:      "CheckWithInt64SeveralComparingSuccess",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3"), path.MatchRoot("block1").AtName("attr4")),
			req:           builTestGreaterValidatorRequest(20),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64SeveralComparingFailed",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3"), path.MatchRoot("block1").AtName("attr4")),
			req:           builTestGreaterValidatorRequest(12),
			expectedError: true,
		},
	}

	for _, c := range cases {
		var resp validator.Int64Response
		c.validator.ValidateInt64(ctx, c.req, &resp)
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
