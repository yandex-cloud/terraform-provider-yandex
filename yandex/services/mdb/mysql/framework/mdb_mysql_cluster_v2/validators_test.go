package mdb_mysql_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func builTestMaintenanceWindowConfigSchema(blockName string) schema.Schema {
	return schema.Schema{
		Description: "Mock MW",
		Blocks: map[string]schema.Block{
			blockName: schema.SingleNestedBlock{
				Validators: []validator.Object{
					NewMaintenanceWindowStructValidator(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Mock type",
						Optional:    true,
					},
					"day": schema.StringAttribute{
						Description: "Mock day",
						Optional:    true,
					},
					"hour": schema.Int64Attribute{
						Description: "Mock hour",
						Optional:    true,
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

// restoreAttributeSchema returns the "restore" attribute schema as defined
// by the actual resource, so tests stay in sync with production code
// automatically instead of duplicating the schema definition.
func restoreAttributeSchema() schema.SingleNestedAttribute {
	r := NewMySQLClusterResourceV2()

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	return schemaResp.Schema.Attributes["restore"].(schema.SingleNestedAttribute)
}

func buildTestRestoreConfigSchema() schema.Schema {
	return schema.Schema{
		Description: "Mock Restore",
		Attributes: map[string]schema.Attribute{
			"restore": restoreAttributeSchema(),
		},
	}
}

func runRestoreAttributeValidators(ctx context.Context, backupId, sourceClusterId, restoreTime *string) diag.Diagnostics {
	sch := buildTestRestoreConfigSchema()

	restoreObj := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"backup_id":         tftypes.String,
				"time":              tftypes.String,
				"source_cluster_id": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"backup_id":         tftypes.NewValue(tftypes.String, backupId),
			"time":              tftypes.NewValue(tftypes.String, restoreTime),
			"source_cluster_id": tftypes.NewValue(tftypes.String, sourceClusterId),
		},
	)

	cfg := tfsdk.Config{
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"restore": restoreObj.Type(),
				},
			},
			map[string]tftypes.Value{
				"restore": restoreObj,
			},
		),
		Schema: sch,
	}

	var diags diag.Diagnostics

	attrPaths := map[string]*string{
		"backup_id":         backupId,
		"time":              restoreTime,
		"source_cluster_id": sourceClusterId,
	}

	restoreAttrs := restoreAttributeSchema().Attributes
	for name, val := range attrPaths {
		attrSchema := restoreAttrs[name].(schema.StringAttribute)
		p := path.Root("restore").AtName(name)

		req := validator.StringRequest{
			Config:         cfg,
			ConfigValue:    types.StringPointerValue(val),
			Path:           p,
			PathExpression: p.Expression(),
		}
		for _, v := range attrSchema.Validators {
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			diags.Append(resp.Diagnostics...)
		}
	}

	return diags
}

func TestYandexProvider_MDBMySQLClusterRestoreAttributeValidators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	backupId := "backup-id"
	sourceClusterId := "source-cluster-id"
	restoreTime := "2006-01-02T15:04:05"
	invalidTime := "not-a-time"
	empty := ""

	cases := []struct {
		testname        string
		backupId        *string
		sourceClusterId *string
		restoreTime     *string
		expectedError   bool
	}{
		{
			testname:      "BackupIdOnly",
			backupId:      &backupId,
			expectedError: false,
		},
		{
			testname:      "BackupIdAndTime",
			backupId:      &backupId,
			restoreTime:   &restoreTime,
			expectedError: false,
		},
		{
			testname:        "TimeAndSourceClusterId",
			sourceClusterId: &sourceClusterId,
			restoreTime:     &restoreTime,
			expectedError:   false,
		},
		{
			testname:      "NoneSet",
			expectedError: true,
		},
		{
			testname:      "EmptyBackupIdOnly",
			backupId:      &empty,
			expectedError: true,
		},
		{
			testname:      "TimeOnly",
			restoreTime:   &restoreTime,
			expectedError: true,
		},
		{
			testname:        "SourceClusterIdOnly",
			sourceClusterId: &sourceClusterId,
			expectedError:   true,
		},
		{
			testname:        "BackupIdAndSourceClusterId",
			backupId:        &backupId,
			sourceClusterId: &sourceClusterId,
			expectedError:   true,
		},
		{
			testname:        "BackupIdAndSourceClusterIdBothEmpty",
			backupId:        &empty,
			sourceClusterId: &empty,
			expectedError:   true,
		},
		{
			testname:        "AllThreeSet",
			backupId:        &backupId,
			sourceClusterId: &sourceClusterId,
			restoreTime:     &restoreTime,
			expectedError:   true,
		},
		{
			testname:        "SourceClusterIdWithInvalidTime",
			sourceClusterId: &sourceClusterId,
			restoreTime:     &invalidTime,
			expectedError:   true,
		},
		{
			testname:        "TimeWithEmptySourceClusterId",
			sourceClusterId: &empty,
			restoreTime:     &restoreTime,
			expectedError:   true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.testname, func(t *testing.T) {
			diags := runRestoreAttributeValidators(ctx, c.backupId, c.sourceClusterId, c.restoreTime)
			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected validation status %s test: expected %t, actual %t with errors: %v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
			}
		})
	}
}
