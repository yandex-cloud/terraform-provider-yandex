package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"testing"
)

func TestExpandGreenplumConfigSpecGreenplumConfig_Positive(t *testing.T) {
	for _, tt := range []struct {
		name               string
		rawConfig          map[string]interface{}
		expectedConfigMask *field_mask.FieldMask
		expectedConfig     greenplum.ConfigSpec_GreenplumConfig
	}{
		{
			name: "6.25 single field",
			rawConfig: map[string]interface{}{
				"version": "6.25",
				"greenplum_config": map[string]interface{}{
					"max_connections": 100,
				},
			},
			expectedConfigMask: &field_mask.FieldMask{
				Paths: []string{
					"config_spec.greenplum_config_6.max_connections",
				},
			},
			expectedConfig: &greenplum.ConfigSpec_GreenplumConfig_6{
				GreenplumConfig_6: &greenplum.GreenplumConfig6{
					MaxConnections: wrapperspb.Int64(100),
				},
			},
		},
		{
			name: "6.25 all supported fields",
			rawConfig: map[string]interface{}{
				"version": "6.25",
				"greenplum_config": map[string]interface{}{
					"max_connections":                      100,
					"max_slot_wal_keep_size":               101,
					"gp_workfile_limit_per_segment":        102,
					"gp_workfile_limit_per_query":          103,
					"gp_workfile_limit_files_per_query":    104,
					"max_prepared_transactions":            106,
					"gp_workfile_compression":              true,
					"max_statement_mem":                    107,
					"log_statement":                        2,
					"gp_add_column_inherits_table_setting": true,
					"gp_enable_global_deadlock_detector":   true,
					"gp_global_deadlock_detector_period":   108,
				},
			},
			expectedConfigMask: &field_mask.FieldMask{
				Paths: []string{
					"config_spec.greenplum_config_6.max_slot_wal_keep_size",
					"config_spec.greenplum_config_6.max_connections",
					"config_spec.greenplum_config_6.gp_workfile_limit_per_segment",
					"config_spec.greenplum_config_6.gp_workfile_limit_per_query",
					"config_spec.greenplum_config_6.gp_workfile_limit_files_per_query",
					"config_spec.greenplum_config_6.max_prepared_transactions",
					"config_spec.greenplum_config_6.gp_workfile_compression",
					"config_spec.greenplum_config_6.max_statement_mem",
					"config_spec.greenplum_config_6.log_statement",
					"config_spec.greenplum_config_6.gp_add_column_inherits_table_setting",
					"config_spec.greenplum_config_6.gp_enable_global_deadlock_detector",
					"config_spec.greenplum_config_6.gp_global_deadlock_detector_period",
				},
			},
			expectedConfig: &greenplum.ConfigSpec_GreenplumConfig_6{
				GreenplumConfig_6: &greenplum.GreenplumConfig6{
					MaxConnections:                  wrapperspb.Int64(100),
					MaxSlotWalKeepSize:              wrapperspb.Int64(101),
					GpWorkfileLimitPerSegment:       wrapperspb.Int64(102),
					GpWorkfileLimitPerQuery:         wrapperspb.Int64(103),
					GpWorkfileLimitFilesPerQuery:    wrapperspb.Int64(104),
					MaxPreparedTransactions:         wrapperspb.Int64(106),
					GpWorkfileCompression:           wrapperspb.Bool(true),
					MaxStatementMem:                 wrapperspb.Int64(107),
					LogStatement:                    greenplum.LogStatement_DDL,
					GpAddColumnInheritsTableSetting: wrapperspb.Bool(true),
					GpEnableGlobalDeadlockDetector:  wrapperspb.Bool(true),
					GpGlobalDeadlockDetectorPeriod:  wrapperspb.Int64(108),
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rd := schema.TestResourceDataRaw(t, resourceYandexMDBGreenplumCluster().Schema, tt.rawConfig)

			config, configMask, err := expandGreenplumConfigSpecGreenplumConfig(rd)

			require.NoError(t, err)
			assert.Equal(t, config, tt.expectedConfig)
			assert.ElementsMatch(t, tt.expectedConfigMask.GetPaths(), configMask.GetPaths())
		})
	}
}

func TestExpandGreenplumConfigSpecGreenplumConfig_Negative(t *testing.T) {
	for _, tt := range []struct {
		name                 string
		rawConfig            map[string]interface{}
		expectedErrorMessage string
	}{
		{
			name:                 "unsupported version 6.17",
			rawConfig:            map[string]interface{}{"version": "6.17"},
			expectedErrorMessage: "unsupported Greenplum version '6.17'",
		},
		{
			name:                 "unsupported version 6.19",
			rawConfig:            map[string]interface{}{"version": "6.19"},
			expectedErrorMessage: "unsupported Greenplum version '6.19'",
		},
		{
			name:                 "unsupported version 6.22",
			rawConfig:            map[string]interface{}{"version": "6.22"},
			expectedErrorMessage: "unsupported Greenplum version '6.22'",
		},
		{
			name:                 "unsupported version unknown",
			rawConfig:            map[string]interface{}{"version": "unknown"},
			expectedErrorMessage: "unknown Greenplum version 'unknown'",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rd := schema.TestResourceDataRaw(t, resourceYandexMDBGreenplumCluster().Schema, tt.rawConfig)

			config, configMask, err := expandGreenplumConfigSpecGreenplumConfig(rd)

			assert.EqualError(t, err, tt.expectedErrorMessage)
			assert.Nil(t, config)
			assert.Nil(t, configMask)
		})
	}
}
