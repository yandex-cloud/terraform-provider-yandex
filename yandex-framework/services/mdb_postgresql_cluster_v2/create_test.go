package mdb_postgresql_cluster_v2

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	pconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	expectedConfigAttrs = map[string]attr.Type{
		"version":                   types.StringType,
		"resources":                 types.ObjectType{AttrTypes: mdbcommon.ResourceType.AttrTypes},
		"autofailover":              types.BoolType,
		"access":                    types.ObjectType{AttrTypes: mdbcommon.AccessAttrTypes},
		"performance_diagnostics":   types.ObjectType{AttrTypes: expectedPDAttrs},
		"backup_window_start":       types.ObjectType{AttrTypes: mdbcommon.BackupWindowType.AttrTypes},
		"backup_retain_period_days": types.Int64Type,
		"postgresql_config":         mdbcommon.NewSettingsMapType(pgAttrProvider),
		"pooler_config":             types.ObjectType{AttrTypes: expectedPCAttrTypes},
		"disk_size_autoscaling":     types.ObjectType{AttrTypes: expectedDiskSizeAutoscalingAttrs},
	}
	expectedBWSAttrs = map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	}
	expectedPDAttrs = map[string]attr.Type{
		"enabled":                      types.BoolType,
		"sessions_sampling_interval":   types.Int64Type,
		"statements_sampling_interval": types.Int64Type,
	}
	expectedClusterAttrs = map[string]attr.Type{
		"name":                   types.StringType,
		"description":            types.StringType,
		"labels":                 types.MapType{ElemType: types.StringType},
		"environment":            types.StringType,
		"network_id":             types.StringType,
		"maintenance_window":     types.ObjectType{AttrTypes: mdbcommon.MaintenanceWindowType.AttrTypes},
		"security_group_ids":     types.SetType{ElemType: types.StringType},
		"restore":                types.ObjectType{AttrTypes: expectedRestoreAttrTypes},
		"config":                 types.ObjectType{AttrTypes: expectedConfigAttrs},
		"deletion_protection":    types.BoolType,
		"folder_id":              types.StringType,
		"hosts":                  types.MapType{ElemType: types.StringType},
		"id":                     types.StringType,
		"disk_encryption_key_id": types.StringType,
		"timeouts":               timeouts.Type{},
	}
	expectedPCAttrTypes = map[string]attr.Type{
		"pool_discard": types.BoolType,
		"pooling_mode": types.StringType,
	}
	expectedDiskSizeAutoscalingAttrs = map[string]attr.Type{
		"disk_size_limit":           types.Int64Type,
		"planned_usage_threshold":   types.Int64Type,
		"emergency_usage_threshold": types.Int64Type,
	}
	expectedRestoreAttrTypes = map[string]attr.Type{
		"backup_id":      types.StringType,
		"time_inclusive": types.BoolType,
		"time":           types.StringType,
	}
	baseConfig = types.ObjectValueMust(
		expectedConfigAttrs,
		map[string]attr.Value{
			"version": types.StringValue("15"),
			"resources": types.ObjectValueMust(
				mdbcommon.ResourceType.AttrTypes,
				map[string]attr.Value{
					"resource_preset_id": types.StringValue("s1.micro"),
					"disk_type_id":       types.StringValue("network-ssd"),
					"disk_size":          types.Int64Value(10),
				},
			),
			"autofailover": types.BoolNull(),
			"backup_window_start": types.ObjectNull(
				expectedBWSAttrs,
			),
			"backup_retain_period_days": types.Int64Null(),
			"performance_diagnostics": types.ObjectNull(
				expectedPDAttrs,
			),
			"access": types.ObjectNull(mdbcommon.AccessAttrTypes),
			"postgresql_config": NewPgSettingsMapValueMust(map[string]attr.Value{
				"max_connections": types.Int64Value(100),
			}),
			"pooler_config": types.ObjectValueMust(expectedPCAttrTypes, map[string]attr.Value{
				"pool_discard": types.BoolValue(true),
				"pooling_mode": types.StringValue(postgresql.ConnectionPoolerConfig_SESSION.String()),
			}),
			"disk_size_autoscaling": types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{
				"disk_size_limit":           types.Int64Value(5),
				"planned_usage_threshold":   types.Int64Value(20),
				"emergency_usage_threshold": types.Int64Value(20),
			}),
		},
	)
)

func TestYandexProvider_MDBPostgresClusterPrepareCreateRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.CreateClusterRequest
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: types.ObjectValueMust(
				expectedClusterAttrs,
				map[string]attr.Value{
					"id": types.StringUnknown(),
					"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
						"host1": types.StringValue("host1"),
						"host2": types.StringValue("host2"),
					}),
					"folder_id":   types.StringValue("test-folder"),
					"name":        types.StringValue("test-cluster"),
					"description": types.StringValue("test-description"),
					"labels": types.MapValueMust(types.StringType, map[string]attr.Value{
						"key": types.StringValue("value"),
					}),
					"environment": types.StringValue("PRESTABLE"),
					"network_id":  types.StringValue("test-network"),
					"maintenance_window": types.ObjectValueMust(
						mdbcommon.MaintenanceWindowType.AttrTypes,
						map[string]attr.Value{
							"type": types.StringValue("ANYTIME"),
							"day":  types.StringValue("MON"),
							"hour": types.Int64Value(1),
						},
					),
					"config":              baseConfig,
					"deletion_protection": types.BoolValue(true),
					"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("test-sg"),
					}),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id":      types.StringNull(),
						"time_inclusive": types.BoolNull(),
						"time":           types.StringNull(),
					}),
					"disk_encryption_key_id": types.StringValue("test-key"),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &postgresql.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "test-description",
				Labels: map[string]string{
					"key": "value",
				},
				Environment: postgresql.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				ConfigSpec: &postgresql.ConfigSpec{
					Version: "15",
					Resources: &postgresql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &postgresql.Access{},
					PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
						PostgresqlConfig_15: &pconfig.PostgresqlConfig15{
							MaxConnections: wrapperspb.Int64(100),
						},
					},
					PoolerConfig: &postgresql.ConnectionPoolerConfig{
						PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
						PoolDiscard: wrapperspb.Bool(true),
					},
					DiskSizeAutoscaling: &postgresql.DiskSizeAutoscaling{
						DiskSizeLimit:           datasize.ToBytes(5),
						PlannedUsageThreshold:   20,
						EmergencyUsageThreshold: 20,
					},
				},
				SecurityGroupIds:   []string{"test-sg"},
				DeletionProtection: true,
				FolderId:           "test-folder",
				MaintenanceWindow: &postgresql.MaintenanceWindow{
					Policy: &postgresql.MaintenanceWindow_Anytime{
						Anytime: &postgresql.AnytimeMaintenanceWindow{},
					},
				},
				DiskEncryptionKeyId: wrapperspb.String("test-key"),
			},
		},
		{
			testname: "CheckPartlyAttributes",
			reqVal: types.ObjectValueMust(
				expectedClusterAttrs,
				map[string]attr.Value{
					"id": types.StringUnknown(),
					"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
						"host1": types.StringValue("host1"),
						"host2": types.StringValue("host2"),
					}),
					"folder_id":           types.StringValue("test-folder"),
					"name":                types.StringValue("test-cluster"),
					"description":         types.StringNull(),
					"labels":              types.MapNull(types.StringType),
					"environment":         types.StringValue("PRODUCTION"),
					"network_id":          types.StringValue("test-network"),
					"config":              baseConfig,
					"maintenance_window":  types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes),
					"deletion_protection": types.BoolNull(),
					"security_group_ids":  types.SetNull(types.StringType),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id":      types.StringNull(),
						"time_inclusive": types.BoolNull(),
						"time":           types.StringNull(),
					}),
					"disk_encryption_key_id": types.StringNull(),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &postgresql.CreateClusterRequest{
				FolderId:    "test-folder",
				Name:        "test-cluster",
				Environment: postgresql.Cluster_PRODUCTION,
				NetworkId:   "test-network",
				ConfigSpec: &postgresql.ConfigSpec{
					Version: "15",
					Resources: &postgresql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
						PostgresqlConfig_15: &pconfig.PostgresqlConfig15{
							MaxConnections: wrapperspb.Int64(100),
						},
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &postgresql.Access{},
					PoolerConfig: &postgresql.ConnectionPoolerConfig{
						PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
						PoolDiscard: wrapperspb.Bool(true),
					},
					DiskSizeAutoscaling: &postgresql.DiskSizeAutoscaling{
						DiskSizeLimit:           datasize.ToBytes(5),
						PlannedUsageThreshold:   20,
						EmergencyUsageThreshold: 20,
					},
				},
				DiskEncryptionKeyId: nil,
			},
		},
	}

	for _, c := range cases {
		cluster := &Cluster{}
		diags := c.reqVal.As(ctx, cluster, datasize.DefaultOpts)
		if diags.HasError() {
			t.Errorf(
				"Unexpected prepare create status diagnostics status %s test errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		req, diags := prepareCreateRequest(ctx, cluster, &config.State{}, nil)
		if diags.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected expand diagnostics status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				diags.HasError(),
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(req, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\nexpected %s\nactual %s",
				c.testname,
				c.expectedVal,
				req,
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterPrepareRestoreRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *postgresql.RestoreClusterRequest
		expectedError bool
	}{
		{
			testname: "CheckFullAttributes",
			reqVal: types.ObjectValueMust(
				expectedClusterAttrs,
				map[string]attr.Value{
					"id": types.StringUnknown(),
					"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
						"host1": types.StringValue("host1"),
						"host2": types.StringValue("host2"),
					}),
					"folder_id":   types.StringValue("test-folder"),
					"name":        types.StringValue("test-cluster"),
					"description": types.StringValue("test-description"),
					"labels": types.MapValueMust(types.StringType, map[string]attr.Value{
						"key": types.StringValue("value"),
					}),
					"environment": types.StringValue("PRESTABLE"),
					"network_id":  types.StringValue("test-network"),
					"maintenance_window": types.ObjectValueMust(
						mdbcommon.MaintenanceWindowType.AttrTypes,
						map[string]attr.Value{
							"type": types.StringValue("ANYTIME"),
							"day":  types.StringValue("MON"),
							"hour": types.Int64Value(1),
						},
					),
					"config":              baseConfig,
					"deletion_protection": types.BoolValue(true),
					"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("test-sg"),
					}),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id":      types.StringValue("backup_id"),
						"time_inclusive": types.BoolValue(true),
						"time":           types.StringValue("2006-01-02T15:04:05"),
					}),
					"disk_encryption_key_id": types.StringValue("test-key"),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &postgresql.RestoreClusterRequest{
				BackupId:      "backup_id",
				TimeInclusive: true,
				Time:          timestamppb.New(parceTime("2006-01-02T15:04:05")),
				Name:          "test-cluster",
				Description:   "test-description",
				Labels: map[string]string{
					"key": "value",
				},
				Environment: postgresql.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				ConfigSpec: &postgresql.ConfigSpec{
					Version: "15",
					Resources: &postgresql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &postgresql.Access{},
					PostgresqlConfig: &postgresql.ConfigSpec_PostgresqlConfig_15{
						PostgresqlConfig_15: &pconfig.PostgresqlConfig15{
							MaxConnections: wrapperspb.Int64(100),
						},
					},
					PoolerConfig: &postgresql.ConnectionPoolerConfig{
						PoolingMode: postgresql.ConnectionPoolerConfig_SESSION,
						PoolDiscard: wrapperspb.Bool(true),
					},
					DiskSizeAutoscaling: &postgresql.DiskSizeAutoscaling{
						DiskSizeLimit:           datasize.ToBytes(5),
						PlannedUsageThreshold:   20,
						EmergencyUsageThreshold: 20,
					},
				},
				SecurityGroupIds:   []string{"test-sg"},
				DeletionProtection: true,
				FolderId:           "test-folder",
				MaintenanceWindow: &postgresql.MaintenanceWindow{
					Policy: &postgresql.MaintenanceWindow_Anytime{
						Anytime: &postgresql.AnytimeMaintenanceWindow{},
					},
				},
				DiskEncryptionKeyId: wrapperspb.String("test-key"),
			},
		},
	}

	for _, c := range cases {
		cluster := &Cluster{}
		diags := c.reqVal.As(ctx, cluster, datasize.DefaultOpts)
		if diags.HasError() {
			t.Errorf(
				"Unexpected prepare create status diagnostics status %s test errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		req, diags := prepareRestoreRequest(ctx, cluster, &config.State{}, nil)
		if diags.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected expand diagnostics status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				diags.HasError(),
				diags.Errors(),
			)
			continue
		}

		if !reflect.DeepEqual(req, c.expectedVal) {
			t.Errorf(
				"Unexpected expand result value %s test:\nexpected %s\nactual %s",
				c.testname,
				c.expectedVal,
				req,
			)
		}
	}
}

func parceTime(time string) time.Time {
	v, _ := mdbcommon.ParseStringToTime(time)
	return v
}
