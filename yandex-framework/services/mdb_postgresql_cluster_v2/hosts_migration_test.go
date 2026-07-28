package mdb_postgresql_cluster_v2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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

// TestMigrateHostLabelsNoRecreate reproduces the scenario from the plan diff:
// state hosts are keyed by fqdn while the plan keys them by custom labels
// (psql1/psql2/psql3). We assert that ModifyStateDependsPlan collapses the labels
// onto the existing hosts and that the resulting HostsDiff neither creates nor
// deletes any node.
func TestMigrateHostLabelsNoRecreate(t *testing.T) {
	fqdnA := "host-a" + testHostDomain
	fqdnB := "host-b" + testHostDomain
	fqdnC := "host-c" + testHostDomain

	plan := map[string]Host{
		"psql1": planHost("zone-a", "subnet-a"),
		"psql2": planHost("zone-b", "subnet-b"),
		"psql3": planHost("zone-c", "subnet-c"),
	}

	state := map[string]Host{
		fqdnA: stateHost("zone-a", "subnet-a", fqdnA),
		fqdnB: stateHost("zone-b", "subnet-b", fqdnB),
		fqdnC: stateHost("zone-c", "subnet-c", fqdnC),
	}

	// 1) Labels are collapsed: fqdn-keys become psql1/2/3, mapped to the SAME hosts.
	fixedState := mdbcommon.ModifyStateDependsPlan(postgresqlHostService, plan, state)

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
	toCreate, toUpdate, toDelete, diags := mdbcommon.HostsDiff(postgresqlHostService, plan, fixedState)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Empty(t, toCreate, "no host must be created")
	assert.Empty(t, toDelete, "no host must be deleted")
	assert.Empty(t, toUpdate, "no host must be updated (priority is Unknown in plan)")
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
