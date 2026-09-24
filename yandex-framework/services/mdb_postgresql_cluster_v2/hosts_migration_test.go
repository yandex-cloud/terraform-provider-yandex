package mdb_postgresql_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

const testHostDomain = ".mdb.example.net"

// planHost reproduces how terraform builds a host in the PLAN when the user gives
// it a custom label (psql1/2/3): the computed attributes (fqdn, priority,
// replication_source) are still "known after apply", i.e. Unknown.
func planHost(zone, subnetID string) Host {
	return Host{
		Zone:                  types.StringValue(zone),
		SubnetId:              types.StringValue(subnetID),
		AssignPublicIp:        types.BoolValue(false),
		FQDN:                  types.StringUnknown(),
		ReplicationSource:     types.StringUnknown(),
		ReplicationSourceName: types.StringNull(),
		Priority:              types.Int64Unknown(),
	}
}

func computedSubnetPlanHost(zone string) Host {
	host := planHost(zone, "")
	host.SubnetId = types.StringUnknown()
	return host
}

// stateHost reproduces the host currently in STATE, keyed by its fqdn.
func stateHost(zone, subnetID, fqdn string) Host {
	return Host{
		Zone:                  types.StringValue(zone),
		SubnetId:              types.StringValue(subnetID),
		AssignPublicIp:        types.BoolValue(false),
		FQDN:                  types.StringValue(fqdn),
		ReplicationSource:     types.StringValue(""),
		ReplicationSourceName: types.StringNull(),
		Priority:              types.Int64Value(0),
	}
}

type testPostgresqlHostAPI struct {
	hosts []*postgresql.Host
}

func (a testPostgresqlHostAPI) ListHosts(context.Context, *ycsdk.SDK, *diag.Diagnostics, string) []*postgresql.Host {
	return a.hosts
}

func (testPostgresqlHostAPI) CreateHosts(context.Context, *ycsdk.SDK, *diag.Diagnostics, string, []*postgresql.HostSpec, struct{}) {
}

func (testPostgresqlHostAPI) UpdateHosts(context.Context, *ycsdk.SDK, *diag.Diagnostics, string, []*postgresql.UpdateHostSpec) {
}

func (testPostgresqlHostAPI) DeleteHosts(context.Context, *ycsdk.SDK, *diag.Diagnostics, string, []string) {
}

func TestSubnetMatchesOrComputed(t *testing.T) {
	tests := []struct {
		name  string
		plan  types.String
		state types.String
		want  bool
	}{
		{name: "empty matches empty", plan: types.StringValue(""), state: types.StringValue(""), want: true},
		{name: "empty differs from API value", plan: types.StringValue(""), state: types.StringValue("subnet-api"), want: false},
		{name: "null is computed by API", plan: types.StringNull(), state: types.StringValue("subnet-api"), want: true},
		{name: "unknown is computed by API", plan: types.StringUnknown(), state: types.StringValue("subnet-api"), want: true},
		{name: "configured subnet matches", plan: types.StringValue("subnet-a"), state: types.StringValue("subnet-a"), want: true},
		{name: "configured subnet differs", plan: types.StringValue("subnet-a"), state: types.StringValue("subnet-b"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, subnetMatchesOrComputed(test.plan, test.state))
		})
	}
}

func TestPostgreSQLHostSubnetSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewPostgreSQLClusterResourceV2().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema diagnostics: %v", resp.Diagnostics)

	hosts := resp.Schema.Attributes["hosts"].(schema.MapNestedAttribute)
	subnet := hosts.NestedObject.Attributes["subnet_id"].(schema.StringAttribute)
	assert.True(t, subnet.IsOptional())
	assert.True(t, subnet.IsComputed())
	require.Len(t, subnet.StringValidators(), 1)
}

func TestReadHostsMatchesComputedSubnets(t *testing.T) {
	ctx := context.Background()
	plan := map[string]Host{
		"psql1": computedSubnetPlanHost("sas"),
		"psql2": computedSubnetPlanHost("vla"),
		"psql3": computedSubnetPlanHost("klg"),
	}
	planValue, diags := types.MapValueFrom(ctx, hostType, plan)
	require.False(t, diags.HasError(), "building hosts plan: %v", diags)

	api := testPostgresqlHostAPI{hosts: []*postgresql.Host{
		{Name: "host-sas" + testHostDomain, ZoneId: "sas", SubnetId: "subnet-sas"},
		{Name: "host-vla" + testHostDomain, ZoneId: "vla", SubnetId: "subnet-vla"},
		{Name: "host-klg" + testHostDomain, ZoneId: "klg", SubnetId: "subnet-klg"},
	}}
	var readDiags diag.Diagnostics

	got := mdbcommon.ReadHosts(ctx, nil, &readDiags, postgresqlHostService, api, planValue, "cluster-id")

	require.False(t, readDiags.HasError(), "reading hosts: %v", readDiags)
	require.Len(t, got, 3)
	assert.Equal(t, "host-sas"+testHostDomain, got["psql1"].FQDN.ValueString())
	assert.Equal(t, "host-vla"+testHostDomain, got["psql2"].FQDN.ValueString())
	assert.Equal(t, "host-klg"+testHostDomain, got["psql3"].FQDN.ValueString())
	assert.Equal(t, "subnet-sas", got["psql1"].SubnetId.ValueString())
	assert.Equal(t, "subnet-vla", got["psql2"].SubnetId.ValueString())
	assert.Equal(t, "subnet-klg", got["psql3"].SubnetId.ValueString())
}

// TestMigrateHostLabelsNoRecreate reproduces the scenario from the plan diff:
// state hosts are keyed by fqdn while the plan keys them by custom labels
// (psql1/psql2/psql3). We assert that ModifyStateDependsPlan collapses the labels
// onto the existing hosts and that the resulting HostsDiff neither creates nor
// deletes any node.
func TestMigrateHostLabelsNoRecreate(t *testing.T) {
	fqdnA := "host-a" + testHostDomain
	fqdnB := "host-b" + testHostDomain
	fqdnC := "host-c" + testHostDomain

	tests := []struct {
		name  string
		plan  map[string]Host
		state map[string]Host
	}{
		{
			name: "configured subnets",
			plan: map[string]Host{
				"psql1": planHost("zone-a", "subnet-a"),
				"psql2": planHost("zone-b", "subnet-b"),
				"psql3": planHost("zone-c", "subnet-c"),
			},
			state: map[string]Host{
				fqdnA: stateHost("zone-a", "subnet-a", fqdnA),
				fqdnB: stateHost("zone-b", "subnet-b", fqdnB),
				fqdnC: stateHost("zone-c", "subnet-c", fqdnC),
			},
		},
		{
			name: "automatically selected subnets",
			plan: map[string]Host{
				"psql1": computedSubnetPlanHost("sas"),
				"psql2": computedSubnetPlanHost("vla"),
				"psql3": computedSubnetPlanHost("klg"),
			},
			state: map[string]Host{
				fqdnA: stateHost("sas", "subnet-sas", fqdnA),
				fqdnB: stateHost("vla", "subnet-vla", fqdnB),
				fqdnC: stateHost("klg", "subnet-klg", fqdnC),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 1) Labels are collapsed: fqdn-keys become psql1/2/3, mapped to the SAME hosts.
			fixedState := mdbcommon.ModifyStateDependsPlan(postgresqlHostService, test.plan, test.state)

			assert.Len(t, fixedState, 3, "expected exactly 3 hosts, no phantom entries")
			assert.Equal(t, fqdnA, fixedState["psql1"].FQDN.ValueString())
			assert.Equal(t, fqdnB, fixedState["psql2"].FQDN.ValueString())
			assert.Equal(t, fqdnC, fixedState["psql3"].FQDN.ValueString())
			// none of the old fqdn-keys survive
			for label := range fixedState {
				assert.NotContains(t, label, testHostDomain, "old fqdn key must be gone: %s", label)
			}

			// 2) With the fixed state the diff is a full no-op: nothing to create, update or delete.
			// In particular there is no spurious priority update: the plan's priority is Unknown
			// ("known after apply") and GetChanges guards against Unknown, so no host is touched.
			toCreate, toUpdate, toDelete, diags := mdbcommon.HostsDiff(postgresqlHostService, test.plan, fixedState)
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Empty(t, toCreate, "no host must be created")
			assert.Empty(t, toDelete, "no host must be deleted")
			assert.Empty(t, toUpdate, "no host must be updated (priority is Unknown in plan)")
		})
	}
}

// TestGetChangesPriorityUnknown locks the fix directly on GetChanges: an Unknown priority
// in the plan must NOT produce an update, while an explicitly set, differing priority still must.
func TestGetChangesPriorityUnknown(t *testing.T) {
	base := stateHost("zone-a", "subnet-a", "host-a"+testHostDomain)

	t.Run("unknown priority -> no change", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a") // Priority is Unknown
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, spec, "Unknown priority must not trigger an update")
	})

	t.Run("explicit differing priority -> update", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		plan.Priority = types.Int64Value(10) // user explicitly sets priority
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		if assert.NotNil(t, spec, "explicit priority change must trigger an update") {
			assert.Contains(t, spec.GetUpdateMask().GetPaths(), "priority")
			assert.Equal(t, int64(10), spec.GetPriority().GetValue())
		}
	})
}
