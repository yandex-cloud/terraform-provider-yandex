package mdb_clickhouse_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

func TestDetectKeeperMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateTypes    map[string]string
		planTypes     map[string]string
		wantMigration bool
		wantError     bool
	}{
		{
			name:          "ZooKeeper to Keeper",
			stateTypes:    map[string]string{"a": "ZOOKEEPER", "b": "ZOOKEEPER", "d": "ZOOKEEPER"},
			planTypes:     map[string]string{"a": "KEEPER", "b": "KEEPER", "d": "KEEPER"},
			wantMigration: true,
		},
		{
			name:       "dedicated Keeper added to cluster without coordinator hosts",
			stateTypes: map[string]string{"clickhouse": "CLICKHOUSE"},
			planTypes:  map[string]string{"clickhouse": "CLICKHOUSE", "keeper": "KEEPER"},
		},
		{
			name:       "Keeper topology update",
			stateTypes: map[string]string{"a": "KEEPER", "b": "KEEPER"},
			planTypes:  map[string]string{"a": "KEEPER", "b": "KEEPER", "d": "KEEPER"},
		},
		{
			name:       "mixed target coordinator types",
			stateTypes: map[string]string{"a": "ZOOKEEPER", "b": "ZOOKEEPER"},
			planTypes:  map[string]string{"a": "ZOOKEEPER", "b": "KEEPER"},
			wantError:  true,
		},
		{
			name:       "Keeper to ZooKeeper",
			stateTypes: map[string]string{"a": "KEEPER"},
			planTypes:  map[string]string{"a": "ZOOKEEPER"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			state := makeHostMap(t, ctx, tt.stateTypes)
			plan := makeHostMap(t, ctx, tt.planTypes)
			var diags diag.Diagnostics

			got := detectKeeperMigration(ctx, state, plan, &diags)
			if got != tt.wantMigration {
				t.Fatalf("detectKeeperMigration() = %v, want %v", got, tt.wantMigration)
			}
			if diags.HasError() != tt.wantError {
				t.Fatalf("detectKeeperMigration() diagnostics error = %v, want %v: %v", diags.HasError(), tt.wantError, diags)
			}
		})
	}
}

func TestMarkKeeperFQDNsUnknown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hosts := makeHostMap(t, ctx, map[string]string{
		"clickhouse": "CLICKHOUSE",
		"keeper":     "KEEPER",
	})
	var diags diag.Diagnostics

	result := markKeeperFQDNsUnknown(ctx, hosts, &diags)
	if diags.HasError() {
		t.Fatalf("markKeeperFQDNsUnknown() returned diagnostics: %v", diags)
	}

	var resultHosts map[string]models.Host
	diags.Append(result.ElementsAs(ctx, &resultHosts, false)...)
	if diags.HasError() {
		t.Fatalf("failed to decode result hosts: %v", diags)
	}

	if !resultHosts["keeper"].FQDN.IsUnknown() {
		t.Fatal("Keeper FQDN must be unknown after planning migration")
	}
	if got := resultHosts["clickhouse"].FQDN.ValueString(); got != "clickhouse.example.net" {
		t.Fatalf("ClickHouse FQDN = %q, want %q", got, "clickhouse.example.net")
	}
}

func TestPrepareMigrateToKeeperRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hosts := makeHostMap(t, ctx, map[string]string{
		"keeper-a": "KEEPER",
		"keeper-b": "KEEPER",
		"keeper-d": "KEEPER",
	})
	resources := &clickhouse.Resources{ResourcePresetId: "s2.micro"}
	var diags diag.Diagnostics

	request := prepareMigrateToKeeperRequest(ctx, "cluster-id", hosts, resources, true, &diags)
	if diags.HasError() {
		t.Fatalf("prepareMigrateToKeeperRequest() returned diagnostics: %v", diags)
	}
	if request.GetClusterId() != "cluster-id" {
		t.Fatalf("request cluster ID = %q, want %q", request.GetClusterId(), "cluster-id")
	}
	if request.GetResources() != resources {
		t.Fatal("request resources differ from the planned coordinator resources")
	}
	if len(request.GetHostSpecs()) != 3 {
		t.Fatalf("request host specs count = %d, want 3", len(request.GetHostSpecs()))
	}
	if !request.GetAllowDegradationToReadOnly() {
		t.Fatal("request must allow degradation to read-only")
	}
}

func TestGetAPICoordinatorHostTypes(t *testing.T) {
	t.Parallel()

	result := getAPICoordinatorHostTypes([]*clickhouse.Host{
		{Type: clickhouse.Host_CLICKHOUSE},
		{Type: clickhouse.Host_ZOOKEEPER},
		{Type: clickhouse.Host_KEEPER},
	})
	if !result.hasZooKeeper || !result.hasKeeper {
		t.Fatalf("getAPICoordinatorHostTypes() = %+v, want both ZooKeeper and Keeper", result)
	}
}

func makeHostMap(t *testing.T, ctx context.Context, hostTypes map[string]string) types.Map {
	t.Helper()

	hosts := make(map[string]models.Host, len(hostTypes))
	for label, hostType := range hostTypes {
		hosts[label] = models.Host{
			FQDN:           types.StringValue(label + ".example.net"),
			Zone:           types.StringValue("ru-central1-a"),
			Type:           types.StringValue(hostType),
			ShardName:      types.StringValue(""),
			SubnetId:       types.StringValue("subnet-id"),
			AssignPublicIp: types.BoolValue(false),
		}
	}

	result, diags := types.MapValueFrom(ctx, models.HostType, hosts)
	if diags.HasError() {
		t.Fatalf("failed to create hosts map: %v", diags)
	}

	return result
}
