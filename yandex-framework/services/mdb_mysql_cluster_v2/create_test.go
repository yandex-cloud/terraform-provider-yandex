package mdb_mysql_cluster_v2

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	msconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	expectedResourcesAttrs = map[string]attr.Type{
		"resource_preset_id": types.StringType,
		"disk_type_id":       types.StringType,
		"disk_size":          types.Int64Type,
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
	expectedDSAAttrs = map[string]attr.Type{
		"disk_size_limit":           types.Int64Type,
		"emergency_usage_threshold": types.Int64Type,
		"planned_usage_threshold":   types.Int64Type,
	}
	expectedMWAttrs = map[string]attr.Type{
		"type": types.StringType,
		"day":  types.StringType,
		"hour": types.Int64Type,
	}
	expectedRestoreAttrTypes = map[string]attr.Type{
		"backup_id": types.StringType,
		"time":      types.StringType,
	}
	expectedClusterAttrs = map[string]attr.Type{
		"name":                      types.StringType,
		"description":               types.StringType,
		"labels":                    types.MapType{ElemType: types.StringType},
		"environment":               types.StringType,
		"network_id":                types.StringType,
		"maintenance_window":        types.ObjectType{AttrTypes: expectedMWAttrs},
		"security_group_ids":        types.SetType{ElemType: types.StringType},
		"deletion_protection":       types.BoolType,
		"folder_id":                 types.StringType,
		"hosts":                     types.MapType{ElemType: types.StringType},
		"id":                        types.StringType,
		"version":                   types.StringType,
		"resources":                 types.ObjectType{AttrTypes: expectedResourcesAttrs},
		"access":                    types.ObjectType{AttrTypes: expectedAccessAttrTypes},
		"performance_diagnostics":   types.ObjectType{AttrTypes: expectedPDAttrs},
		"disk_size_autoscaling":     types.ObjectType{AttrTypes: expectedDSAAttrs},
		"backup_window_start":       types.ObjectType{AttrTypes: expectedBwsAttrTypes},
		"backup_retain_period_days": types.Int64Type,
		"restore":                   types.ObjectType{AttrTypes: expectedRestoreAttrTypes},
		"mysql_config":              mdbcommon.NewSettingsMapType(msAttrProvider),
		"disk_encryption_key_id":    types.StringType,
		"timeouts":                  timeouts.Type{},
	}
	baseCluster = Cluster{
		Id:          types.StringValue("test-id"),
		FolderId:    types.StringValue("test-folder"),
		NetworkId:   types.StringValue("test-network"),
		Name:        types.StringValue("test-cluster"),
		Description: types.StringValue("test-description"),
		Environment: types.StringValue("PRODUCTION"),
		Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
			"key": types.StringValue("value"),
		}),
		MaintenanceWindow: types.ObjectValueMust(
			expectedMWAttrs,
			map[string]attr.Value{
				"type": types.StringValue("WEEKLY"),
				"day":  types.StringValue("MON"),
				"hour": types.Int64Value(1),
			},
		),
		Version: types.StringValue("8.0"),
		Resources: types.ObjectValueMust(
			expectedResourcesAttrs,
			map[string]attr.Value{
				"resource_preset_id": types.StringValue("s1.micro"),
				"disk_type_id":       types.StringValue("network-ssd"),
				"disk_size":          types.Int64Value(10),
			},
		),
		BackupWindowStart: types.ObjectNull(
			expectedBWSAttrs,
		),
		BackupRetainPeriodDays: types.Int64Null(),
		PerformanceDiagnostics: types.ObjectNull(
			expectedPDAttrs,
		),
		DiskSizeAutoscaling: types.ObjectNull(
			expectedDSAAttrs,
		),
		Access:             types.ObjectNull(AccessAttrTypes),
		DeletionProtection: types.BoolValue(true),
		SecurityGroupIds: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("test-sg"),
		}),
		MySQLConfig: NewMsSettingsMapValueMust(map[string]attr.Value{
			"max_connections": types.Int64Value(100),
			"default_authentication_plugin": types.Int64Value(
				int64(msconfig.MysqlConfig8_0_MYSQL_NATIVE_PASSWORD),
			),
			"innodb_print_all_deadlocks": types.BoolValue(true),
		}),
	}
)

func TestYandexProvider_MDBMySQLClusterPrepareCreateRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *mysql.CreateClusterRequest
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
					"version":     types.StringValue("5.7"),
					"resources": types.ObjectValueMust(
						expectedResourcesAttrs,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start": types.ObjectNull(
						expectedBWSAttrs,
					),
					"backup_retain_period_days": types.Int64Null(),
					"performance_diagnostics": types.ObjectNull(
						expectedPDAttrs,
					),
					"disk_size_autoscaling": types.ObjectNull(
						expectedDSAAttrs,
					),
					"access": types.ObjectNull(AccessAttrTypes),
					"maintenance_window": types.ObjectValueMust(
						expectedMWAttrs,
						map[string]attr.Value{
							"type": types.StringValue("WEEKLY"),
							"day":  types.StringValue("MON"),
							"hour": types.Int64Value(1),
						},
					),
					"deletion_protection": types.BoolValue(true),
					"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("test-sg"),
					}),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id": types.StringNull(),
						"time":      types.StringNull(),
					}),
					"mysql_config": NewMsSettingsMapValueMust(map[string]attr.Value{
						"max_connections": types.Int64Value(100),
						"default_authentication_plugin": types.Int64Value(
							int64(msconfig.MysqlConfig8_0_MYSQL_NATIVE_PASSWORD),
						),
						"innodb_print_all_deadlocks": types.BoolValue(true),
					}),
					"disk_encryption_key_id": types.StringValue("test-key"),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &mysql.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "test-description",
				Labels: map[string]string{
					"key": "value",
				},
				Environment: mysql.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				ConfigSpec: &mysql.ConfigSpec{
					Version: "5.7",
					Resources: &mysql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &mysql.Access{},
					MysqlConfig: &mysql.ConfigSpec_MysqlConfig_5_7{
						MysqlConfig_5_7: &msconfig.MysqlConfig5_7{
							MaxConnections:              wrapperspb.Int64(100),
							DefaultAuthenticationPlugin: msconfig.MysqlConfig5_7_MYSQL_NATIVE_PASSWORD,
							InnodbPrintAllDeadlocks:     wrapperspb.Bool(true),
						},
					},
				},
				SecurityGroupIds:   []string{"test-sg"},
				DeletionProtection: true,
				FolderId:           "test-folder",
				MaintenanceWindow: &mysql.MaintenanceWindow{
					Policy: &mysql.MaintenanceWindow_WeeklyMaintenanceWindow{
						WeeklyMaintenanceWindow: &mysql.WeeklyMaintenanceWindow{
							Day:  mysql.WeeklyMaintenanceWindow_MON,
							Hour: 1,
						},
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
					"folder_id":   types.StringValue("test-folder"),
					"name":        types.StringValue("test-cluster"),
					"description": types.StringNull(),
					"labels":      types.MapNull(types.StringType),
					"environment": types.StringValue("PRODUCTION"),
					"network_id":  types.StringValue("test-network"),
					"version":     types.StringValue("8.0"),
					"resources": types.ObjectValueMust(
						expectedResourcesAttrs,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start": types.ObjectNull(
						expectedBWSAttrs,
					),
					"backup_retain_period_days": types.Int64Null(),
					"performance_diagnostics": types.ObjectNull(
						expectedPDAttrs,
					),
					"disk_size_autoscaling": types.ObjectNull(
						expectedDSAAttrs,
					),
					"access":              types.ObjectNull(AccessAttrTypes),
					"maintenance_window":  types.ObjectNull(expectedMWAttrs),
					"deletion_protection": types.BoolNull(),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id": types.StringNull(),
						"time":      types.StringNull(),
					}),
					"security_group_ids":     types.SetNull(types.StringType),
					"mysql_config":           NewMsSettingsMapNull(),
					"disk_encryption_key_id": types.StringNull(),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &mysql.CreateClusterRequest{
				FolderId:    "test-folder",
				Name:        "test-cluster",
				Environment: mysql.Cluster_PRODUCTION,
				NetworkId:   "test-network",
				ConfigSpec: &mysql.ConfigSpec{
					Version: "8.0",
					Resources: &mysql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &mysql.Access{},
					MysqlConfig: &mysql.ConfigSpec_MysqlConfig_8_0{
						MysqlConfig_8_0: &msconfig.MysqlConfig8_0{},
					},
				},
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

func TestYandexProvider_MDBMySQLClusterGetConfigSpec(t *testing.T) {
	t.Parallel()

	req := baseCluster
	expected := Config{
		Version: types.StringValue("8.0"),
		Resources: types.ObjectValueMust(
			expectedResourcesAttrs,
			map[string]attr.Value{
				"resource_preset_id": types.StringValue("s1.micro"),
				"disk_type_id":       types.StringValue("network-ssd"),
				"disk_size":          types.Int64Value(10),
			},
		),
		BackupWindowStart: types.ObjectNull(
			expectedBWSAttrs,
		),
		BackupRetainPeriodDays: types.Int64Null(),
		PerformanceDiagnostics: types.ObjectNull(
			expectedPDAttrs,
		),
		DiskSizeAutoscaling: types.ObjectNull(
			expectedDSAAttrs,
		),
		Access: types.ObjectNull(AccessAttrTypes),
		MySQLConfig: NewMsSettingsMapValueMust(map[string]attr.Value{
			"max_connections": types.Int64Value(100),
			"default_authentication_plugin": types.Int64Value(
				int64(msconfig.MysqlConfig8_0_MYSQL_NATIVE_PASSWORD),
			),
			"innodb_print_all_deadlocks": types.BoolValue(true),
		}),
	}

	diags := diag.Diagnostics{}
	config := getConfigSpecFromState(&req)
	if diags.HasError() {
		t.Errorf(
			"Unexpected get config status diagnostics with status test errors: %v",
			diags.Errors(),
		)
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf(
			"Unexpected config result value test:\nexpected %s\nactual %s",
			expected,
			config,
		)
	}
}

func TestYandexProvider_MDBMySQLClusterPrepareRestoreRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        types.Object
		expectedVal   *mysql.RestoreClusterRequest
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
					"version":     types.StringValue("5.7"),
					"resources": types.ObjectValueMust(
						expectedResourcesAttrs,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s1.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
					"backup_window_start": types.ObjectNull(
						expectedBWSAttrs,
					),
					"backup_retain_period_days": types.Int64Null(),
					"performance_diagnostics": types.ObjectNull(
						expectedPDAttrs,
					),
					"disk_size_autoscaling": types.ObjectNull(
						expectedDSAAttrs,
					),
					"access": types.ObjectNull(AccessAttrTypes),
					"maintenance_window": types.ObjectValueMust(
						expectedMWAttrs,
						map[string]attr.Value{
							"type": types.StringValue("WEEKLY"),
							"day":  types.StringValue("MON"),
							"hour": types.Int64Value(1),
						},
					),
					"deletion_protection": types.BoolValue(true),
					"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("test-sg"),
					}),
					"restore": types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{
						"backup_id": types.StringValue("backup_id"),
						"time":      types.StringValue("2006-01-02T15:04:05"),
					}),
					"mysql_config": NewMsSettingsMapValueMust(map[string]attr.Value{
						"max_connections": types.Int64Value(100),
						"default_authentication_plugin": types.Int64Value(
							int64(msconfig.MysqlConfig8_0_MYSQL_NATIVE_PASSWORD),
						),
						"innodb_print_all_deadlocks": types.BoolValue(true),
					}),
					"disk_encryption_key_id": types.StringValue("test-key"),
					"timeouts":               timeouts.Value{},
				},
			),
			expectedVal: &mysql.RestoreClusterRequest{
				BackupId:    "backup_id",
				Time:        timestamppb.New(parceTime("2006-01-02T15:04:05")),
				Name:        "test-cluster",
				Description: "test-description",
				Labels: map[string]string{
					"key": "value",
				},
				Environment: mysql.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				ConfigSpec: &mysql.ConfigSpec{
					Version: "5.7",
					Resources: &mysql.Resources{
						ResourcePresetId: "s1.micro",
						DiskTypeId:       "network-ssd",
						DiskSize:         datasize.ToBytes(10),
					},
					BackupWindowStart: &timeofday.TimeOfDay{},
					Access:            &mysql.Access{},
					MysqlConfig: &mysql.ConfigSpec_MysqlConfig_5_7{
						MysqlConfig_5_7: &msconfig.MysqlConfig5_7{
							MaxConnections:              wrapperspb.Int64(100),
							DefaultAuthenticationPlugin: msconfig.MysqlConfig5_7_MYSQL_NATIVE_PASSWORD,
							InnodbPrintAllDeadlocks:     wrapperspb.Bool(true),
						},
					},
				},
				SecurityGroupIds:   []string{"test-sg"},
				DeletionProtection: true,
				FolderId:           "test-folder",
				MaintenanceWindow: &mysql.MaintenanceWindow{
					Policy: &mysql.MaintenanceWindow_WeeklyMaintenanceWindow{
						WeeklyMaintenanceWindow: &mysql.WeeklyMaintenanceWindow{
							Day:  mysql.WeeklyMaintenanceWindow_MON,
							Hour: 1,
						},
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
