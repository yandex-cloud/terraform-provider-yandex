package yandex

import (
	"reflect"
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

func TestPGChangedDatabases(t *testing.T) {
	database := func(name, owner string) map[string]interface{} {
		return map[string]interface{}{"name": name, "owner": owner}
	}

	tests := []struct {
		name     string
		oldSpecs []interface{}
		newSpecs []interface{}
		expected []pgDatabaseChange
	}{
		{
			name:     "owner changed",
			oldSpecs: []interface{}{database("alice", "u1")},
			newSpecs: []interface{}{database("alice", "u2")},
			expected: []pgDatabaseChange{
				{Spec: &postgresql.DatabaseSpec{Name: "alice", Owner: "u2"}, UpdatePath: []string{"owner"}},
			},
		},
		{
			name:     "nothing changed",
			oldSpecs: []interface{}{database("alice", "u1")},
			newSpecs: []interface{}{database("alice", "u1")},
			expected: []pgDatabaseChange{},
		},
		{
			name:     "owner changed for the database shifted by a deleted one",
			oldSpecs: []interface{}{database("alice", "u1"), database("bob", "u2")},
			newSpecs: []interface{}{database("bob", "u1")},
			expected: []pgDatabaseChange{
				{Spec: &postgresql.DatabaseSpec{Name: "bob", Owner: "u1"}, UpdatePath: []string{"owner"}},
			},
		},
		{
			name:     "databases reordered without changes",
			oldSpecs: []interface{}{database("alice", "u1"), database("bob", "u2")},
			newSpecs: []interface{}{database("bob", "u2"), database("alice", "u1")},
			expected: []pgDatabaseChange{},
		},
		{
			name:     "added database is not an update",
			oldSpecs: []interface{}{database("alice", "u1")},
			newSpecs: []interface{}{database("alice", "u1"), database("bob", "u2")},
			expected: []pgDatabaseChange{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pgChangedDatabases(tt.oldSpecs, tt.newSpecs)
			if err != nil {
				t.Fatalf("pgChangedDatabases() returned error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("pgChangedDatabases() returned %d changes, want %d", len(result), len(tt.expected))
			}
			for i, change := range result {
				if change.Spec.Name != tt.expected[i].Spec.Name || change.Spec.Owner != tt.expected[i].Spec.Owner {
					t.Errorf("pgChangedDatabases()[%d].Spec = %v, want %v", i, change.Spec, tt.expected[i].Spec)
				}
				if !reflect.DeepEqual(change.UpdatePath, tt.expected[i].UpdatePath) {
					t.Errorf("pgChangedDatabases()[%d].UpdatePath = %v, want %v", i, change.UpdatePath, tt.expected[i].UpdatePath)
				}
			}
		})
	}
}
