package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestPGPerformanceDiagnosticsAdvancedMode(t *testing.T) {
	t.Run("expand configured advanced mode", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBPostgreSQLCluster().Schema, map[string]any{
			"config": []any{
				map[string]any{
					"performance_diagnostics": []any{
						map[string]any{
							"enabled":                      true,
							"advanced_mode":                true,
							"sessions_sampling_interval":   60,
							"statements_sampling_interval": 600,
						},
					},
				},
			},
		})

		got := expandPGPerformanceDiagnostics(resourceData)
		if got == nil {
			t.Fatal("expandPGPerformanceDiagnostics() returned nil")
		}
		if !got.AdvancedMode {
			t.Error("expandPGPerformanceDiagnostics() did not preserve advanced_mode=true")
		}
	})

	t.Run("flatten API advanced mode", func(t *testing.T) {
		got := flattenPGPerformanceDiagnostics(&postgresql.PerformanceDiagnostics{AdvancedMode: true})
		if len(got) != 1 {
			t.Fatalf("flattenPGPerformanceDiagnostics() returned %d items, want 1", len(got))
		}

		values, ok := got[0].(map[string]any)
		if !ok {
			t.Fatalf("flattenPGPerformanceDiagnostics() item has type %T, want map[string]any", got[0])
		}
		if advancedMode, ok := values["advanced_mode"].(bool); !ok || !advancedMode {
			t.Errorf("flattenPGPerformanceDiagnostics() advanced_mode = %v, want true", values["advanced_mode"])
		}
	})
}

func TestComparePGNoNamedHostInfo(t *testing.T) {
	tests := []struct {
		name        string
		existedHost *pgHostInfo
		newHost     *pgHostInfo
		expected    bool
	}{
		{
			name: "equal zone and subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11",
			},
			expected: true,
		},
		{
			name: "not equal zone and equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z12", subnetID: "sn11",
			},
			expected: false,
		},
		{
			name: "equal zone and not equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn12",
			},
			expected: false,
		},
		{
			name: "equal zone and empty new subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := matchesPGNoNamedHostInfo(tt.existedHost, tt.newHost); result != tt.expected {
				t.Errorf("matchesPGNoNamedHostInfo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComparePGNamedHostInfo(t *testing.T) {
	tests := []struct {
		name        string
		existedHost *pgHostInfo
		newHost     *pgHostInfo
		expected    int
	}{
		{
			name: "equal zone and subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", name: "n1",
			},
			expected: 2,
		},
		{
			name: "not equal zone and equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z12", subnetID: "sn11", name: "n1",
			},
			expected: 0,
		},
		{
			name: "equal zone and not equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn12", name: "n1",
			},
			expected: 0,
		},
		{
			name: "equal zone and empty new subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "", name: "n1",
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := comparePGNamedHostInfo(tt.existedHost, tt.newHost, map[string]string{}); result != tt.expected {
				t.Errorf("comparePGNamedHostInfo() = %v, want %v", result, tt.expected)
			}
		})
	}
}
