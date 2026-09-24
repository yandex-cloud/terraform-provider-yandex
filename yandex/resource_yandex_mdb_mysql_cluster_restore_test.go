package yandex

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testBackupID       = "c9qnlqr37bgp53r9pbek:mdbel77v1so5qiu199ua"
	testSourceCluster  = "c9qnlqr37bgp53r9pbek"
	testRestoreTimeStr = "2025-08-26T14:04:05"
)

func buildBaseMysqlClusterConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":        "test-cluster",
		"environment": "PRODUCTION",
		"network_id":  "network-id",
		"version":     "8.0",
		"resources": []interface{}{
			map[string]interface{}{
				"resource_preset_id": "s2.micro",
				"disk_type_id":       "network-ssd",
				"disk_size":          10,
			},
		},
		"host": []interface{}{
			map[string]interface{}{
				"zone": "ru-central1-a",
			},
		},
	}
}

func buildRestoreResourceData(t *testing.T, restore map[string]interface{}) *schema.ResourceData {
	raw := buildBaseMysqlClusterConfig()
	if restore != nil {
		raw["restore"] = []interface{}{restore}
	}
	return schema.TestResourceDataRaw(t, resourceYandexMDBMySQLCluster().Schema, raw)
}

func TestMysqlRestoreSchemaValidation(t *testing.T) {
	resourceSchema := schema.InternalMap(resourceYandexMDBMySQLCluster().Schema)

	cases := []struct {
		name      string
		restore   map[string]interface{}
		wantError bool
	}{
		{
			name: "backup_id only",
			restore: map[string]interface{}{
				"backup_id": testBackupID,
			},
			wantError: false,
		},
		{
			name: "backup_id and time",
			restore: map[string]interface{}{
				"backup_id": testBackupID,
				"time":      testRestoreTimeStr,
			},
			wantError: false,
		},
		{
			name: "time and source_cluster_id",
			restore: map[string]interface{}{
				"time":              testRestoreTimeStr,
				"source_cluster_id": testSourceCluster,
			},
			wantError: false,
		},
		{
			name:      "none set - empty restore block",
			restore:   map[string]interface{}{},
			wantError: true,
		},
		{
			name: "time and empty source_cluster_id",
			restore: map[string]interface{}{
				"time":              testRestoreTimeStr,
				"source_cluster_id": "",
			},
			wantError: true,
		},
		{
			name: "invalid time and source_cluster_id",
			restore: map[string]interface{}{
				"time":              "aaa",
				"source_cluster_id": testSourceCluster,
			},
			wantError: true,
		},
		{
			name: "time only",
			restore: map[string]interface{}{
				"time": testRestoreTimeStr,
			},
			wantError: true,
		},
		{
			name: "source_cluster_id only",
			restore: map[string]interface{}{
				"source_cluster_id": testSourceCluster,
			},
			wantError: true,
		},
		{
			name: "backup_id and source_cluster_id - mutually exclusive",
			restore: map[string]interface{}{
				"backup_id":         testBackupID,
				"source_cluster_id": testSourceCluster,
			},
			wantError: true,
		},
		{
			name: "backup_id, time, and source_cluster_id - mutually exclusive",
			restore: map[string]interface{}{
				"backup_id":         testBackupID,
				"time":              testRestoreTimeStr,
				"source_cluster_id": testSourceCluster,
			},
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := buildBaseMysqlClusterConfig()
			raw["restore"] = []interface{}{tc.restore}

			diags := resourceSchema.Validate(sdkTerraform.NewResourceConfigRaw(raw))
			for _, d := range diags {
				t.Logf("diagnostic: severity=%v summary=%q detail=%q", d.Severity, d.Summary, d.Detail)
			}
			if gotError := diags.HasError(); gotError != tc.wantError {
				t.Fatalf("schema validation error = %t, want %t; diagnostics: %#v", gotError, tc.wantError, diags)
			}
		})
	}
}

type testMysqlRestoreRawConfig struct {
	value cty.Value
}

func (c testMysqlRestoreRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func TestCheckBackupIdIsNotEmptyIfSpecified(t *testing.T) {
	cases := []struct {
		name      string
		restore   cty.Value
		wantError bool
	}{
		{
			name:      "no restore block",
			restore:   cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"backup_id": cty.String}))),
			wantError: false,
		},
		{
			name: "empty restore block list",
			restore: cty.ListValEmpty(
				cty.Object(map[string]cty.Type{"backup_id": cty.String}),
			),
			wantError: false,
		},
		{
			name: "no backup_id field (null)",
			restore: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"backup_id": cty.NullVal(cty.String),
				}),
			}),
			wantError: false,
		},
		{
			name: "backup_id unknown",
			restore: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"backup_id": cty.UnknownVal(cty.String),
				}),
			}),
			wantError: false,
		},
		{
			name: "backup_id set to non-empty value",
			restore: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"backup_id": cty.StringVal(testBackupID),
				}),
			}),
			wantError: false,
		},
		{
			name: "backup_id explicitly set to empty string",
			restore: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"backup_id": cty.StringVal(""),
				}),
			}),
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := testMysqlRestoreRawConfig{
				value: cty.ObjectVal(map[string]cty.Value{
					"restore": tc.restore,
				}),
			}
			err := checkBackupIdIsNotEmptyIfSpecified(config)
			if tc.wantError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestRestoreBlockExists(t *testing.T) {
	cases := []struct {
		name     string
		restore  map[string]interface{}
		expected bool
	}{
		{
			name:     "no restore block",
			restore:  nil,
			expected: false,
		},
		{
			name:     "empty restore block",
			restore:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "restore block",
			restore: map[string]interface{}{
				"backup_id": testBackupID,
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := buildRestoreResourceData(t, tc.restore)
			got := restoreBlockExists(d)
			if got != tc.expected {
				t.Fatalf("expected %t, got %t", tc.expected, got)
			}
		})
	}
}
