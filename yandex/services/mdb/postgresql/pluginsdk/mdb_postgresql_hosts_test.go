package pluginsdk

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func pgClusterDataWithHosts(t *testing.T, hosts ...map[string]any) *schema.ResourceData {
	t.Helper()
	return pgClusterDataOnApply(t, nil, hosts...)
}

func pgClusterDataOnApply(t *testing.T, state *terraform.InstanceState, hosts ...map[string]any) *schema.ResourceData {
	t.Helper()
	rawHosts := make([]any, 0, len(hosts))
	hostConfigs := make([]cty.Value, 0, len(hosts))
	for _, host := range hosts {
		rawHosts = append(rawHosts, host)
		hostConfigs = append(hostConfigs, pgHostConfigFromMap(t, host))
	}
	clusterSchema := schema.InternalMap(ResourceYandexMDBPostgreSQLCluster().Schema)

	diff, err := clusterSchema.Diff(
		context.Background(), state, terraform.NewResourceConfigRaw(map[string]any{"host": rawHosts}), nil, nil, true)
	require.NoError(t, err)
	require.NotNil(t, diff)
	diff.RawConfig = pgClusterRawConfig(hostConfigs...)

	d, err := clusterSchema.Data(state, diff)
	require.NoError(t, err)
	return d
}

func pgHostConfigFromMap(t *testing.T, host map[string]any) cty.Value {
	t.Helper()
	attrs := map[string]cty.Value{}
	for name, value := range host {
		attr, err := gocty.ToCtyValue(value, pgHostConfigType().AttributeType(name))
		require.NoError(t, err)
		attrs[name] = attr
	}
	return pgHostConfig(attrs)
}

func pgClusterStateAfterRefresh(t *testing.T, apiHosts []*postgresql.Host, hosts ...map[string]any) *terraform.InstanceState {
	t.Helper()
	d := pgClusterDataWithHosts(t, hosts...)
	d.SetId("cluster-id")
	d, _ = refreshedPGClusterData(t, d, apiHosts)
	return d.State()
}

func TestExpandPGHostPriorities(t *testing.T) {
	t.Run("values are sent as configured", func(t *testing.T) {
		host := expandPGHost(map[string]any{"zone": "z1", "name": "na", "priority": 5, "backup_priority": 70})
		assert.Equal(t, int64(5), host.HostSpec.GetPriority().GetValue())
		assert.Equal(t, int64(70), host.HostSpec.GetBackupPriority().GetValue())
	})

	t.Run("zero values are sent explicitly", func(t *testing.T) {
		host := expandPGHost(map[string]any{"zone": "z1", "priority": 0, "backup_priority": 0})
		require.NotNil(t, host.HostSpec.GetPriority())
		require.NotNil(t, host.HostSpec.GetBackupPriority())
		assert.Equal(t, int64(0), host.HostSpec.GetPriority().GetValue())
		assert.Equal(t, int64(0), host.HostSpec.GetBackupPriority().GetValue())
	})

	t.Run("missing values are not sent", func(t *testing.T) {
		host := expandPGHost(map[string]any{"zone": "z1"})
		assert.Nil(t, host.HostSpec.GetPriority())
		assert.Nil(t, host.HostSpec.GetBackupPriority())
	})
}

func TestExpandPGHostFields(t *testing.T) {
	host := expandPGHost(map[string]any{
		"zone":                    "z1",
		"subnet_id":               "sn1",
		"assign_public_ip":        true,
		"fqdn":                    "fqdn-a",
		"name":                    "na",
		"replication_source_name": "nb",
	})

	assert.Equal(t, "fqdn-a", host.Fqdn)
	assert.Equal(t, "z1", host.HostSpec.GetZoneId())
	assert.Equal(t, "sn1", host.HostSpec.GetSubnetId())
	assert.True(t, host.HostSpec.GetAssignPublicIp())
	assert.Equal(t, "nb", host.HostSpec.GetReplicationSource())
}

type pgRawConfig struct {
	value cty.Value
}

func (c pgRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func pgHostConfigType() cty.Type {
	return ResourceYandexMDBPostgreSQLCluster().CoreConfigSchema().ImpliedType().AttributeType("host").ElementType()
}

func pgHostConfig(attrs map[string]cty.Value) cty.Value {
	host := map[string]cty.Value{}
	for name, attrType := range pgHostConfigType().AttributeTypes() {
		host[name] = cty.NullVal(attrType)
	}
	maps.Copy(host, attrs)
	return cty.ObjectVal(host)
}

func pgClusterRawConfig(hosts ...cty.Value) cty.Value {
	config := ResourceYandexMDBPostgreSQLCluster().CoreConfigSchema().EmptyValue().AsValueMap()
	if len(hosts) > 0 {
		config["host"] = cty.ListVal(hosts)
	}
	return cty.ObjectVal(config)
}

func TestPGHostPriorityRequiresNameInConfig(t *testing.T) {
	const priorityWithoutName = "priority can be set only when name is specified"
	zone := cty.StringVal("z1")
	named := pgHostConfig(map[string]cty.Value{"zone": zone, "name": cty.StringVal("na"), "priority": cty.NumberIntVal(10)})

	tests := []struct {
		name      string
		config    cty.Value
		wantError string
	}{
		{
			name:   "no configuration on read",
			config: cty.NullVal(cty.EmptyObject),
		},
		{
			name:   "no hosts",
			config: pgClusterRawConfig(),
		},
		{
			name:   "unnamed host without priority",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{"zone": zone})),
		},
		{
			name:   "unnamed host with zero priority",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{"zone": zone, "priority": cty.NumberIntVal(0)})),
		},
		{
			name: "unnamed host with backup priority only",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{
				"zone": zone, "backup_priority": cty.NumberIntVal(70),
			})),
		},
		{
			name:   "named host with priority",
			config: pgClusterRawConfig(named),
		},
		{
			name: "priority is known only after apply",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{
				"zone": zone, "priority": cty.UnknownVal(cty.Number),
			})),
		},
		{
			name: "name is known only after apply",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{
				"zone": zone, "name": cty.UnknownVal(cty.String), "priority": cty.NumberIntVal(10),
			})),
		},
		{
			name:   "host is known only after apply",
			config: pgClusterRawConfig(cty.UnknownVal(pgHostConfigType())),
		},
		{
			name:   "host is null",
			config: pgClusterRawConfig(cty.NullVal(pgHostConfigType())),
		},
		{
			name:      "unnamed host with the lowest non-zero priority",
			config:    pgClusterRawConfig(pgHostConfig(map[string]cty.Value{"zone": zone, "priority": cty.NumberIntVal(1)})),
			wantError: priorityWithoutName,
		},
		{
			name: "host with empty name and priority",
			config: pgClusterRawConfig(pgHostConfig(map[string]cty.Value{
				"zone": zone, "name": cty.StringVal(""), "priority": cty.NumberIntVal(10),
			})),
			wantError: priorityWithoutName,
		},
		{
			name: "only the last host is unnamed",
			config: pgClusterRawConfig(named, pgHostConfig(map[string]cty.Value{
				"zone": zone, "priority": cty.NumberIntVal(10),
			})),
			wantError: priorityWithoutName,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePGHostPriorityRequiresName(pgRawConfig{value: test.config})
			if test.wantError == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, test.wantError)
		})
	}
}

func TestPGClusterPlanRejectsHostPriorityWithoutName(t *testing.T) {
	rawConfig := pgClusterRawConfig(pgHostConfig(map[string]cty.Value{
		"zone": cty.StringVal("z1"), "priority": cty.NumberIntVal(10),
	}))
	config := terraform.NewResourceConfigRaw(map[string]any{
		"host": []any{map[string]any{"zone": "z1", "priority": 10}},
	})

	_, err := ResourceYandexMDBPostgreSQLCluster().SimpleDiff(
		context.Background(), &terraform.InstanceState{RawConfig: rawConfig}, config, nil)

	assert.ErrorContains(t, err, "priority can be set only when name is specified")
}

func TestPGHostBackupPrioritySchemaRange(t *testing.T) {
	backupPriority := resourceYandexMDBPostgreSQLClusterHost().Schema["backup_priority"]
	require.NotNil(t, backupPriority)
	assert.True(t, backupPriority.Optional)
	assert.True(t, backupPriority.Computed)

	for value, wantError := range map[int]bool{-1: true, 0: false, 100: false, 101: true} {
		_, errs := backupPriority.ValidateFunc(value, "backup_priority")
		assert.Equal(t, wantError, len(errs) > 0, "backup_priority = %d", value)
	}
}

func TestPGHostBackupPriorityUpdate(t *testing.T) {
	apiHosts := []*postgresql.Host{
		{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1", BackupPriority: wrapperspb.Int64(5)},
		{Name: "fqdn-b", ZoneId: "z2", SubnetId: "sn2", BackupPriority: wrapperspb.Int64(5)},
		{Name: "fqdn-c", ZoneId: "z3", SubnetId: "sn3"},
	}
	d := pgClusterDataWithHosts(t,
		map[string]any{"zone": "z1", "subnet_id": "sn1", "name": "na", "backup_priority": 70},
		map[string]any{"zone": "z2", "subnet_id": "sn2", "name": "nb", "backup_priority": 5},
		map[string]any{"zone": "z3", "subnet_id": "sn3", "name": "nc"},
	)

	compared, err := comparePGHostsInfo(d, apiHosts, true)
	require.NoError(t, err)
	assert.Empty(t, compared.createHostsInfo)
	for fqdn, hostInfo := range compared.hostsInfo {
		assert.True(t, hostInfo.inTargetSet, "host %s is declared in config and must be kept", fqdn)
	}

	changed := pgHostChanges(compared.hostsInfo["fqdn-a"])
	require.NotNil(t, changed)
	assert.Equal(t, "fqdn-a", changed.GetHostName())
	assert.Equal(t, []string{"backup_priority"}, changed.GetUpdateMask().GetPaths())
	assert.Equal(t, int64(70), changed.GetBackupPriority().GetValue())

	assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-b"]))
	assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-c"]))
}

func TestPGHostChangesKeepsExistingMaskPaths(t *testing.T) {
	spec := pgHostChanges(&PgHostInfo{
		fqdn:                 "fqdn-a",
		oldReplicationSource: "fqdn-b",
		newReplicationSource: "fqdn-c",
		oldAssignPublicIP:    false,
		newAssignPublicIP:    true,
		oldPriority:          1,
		newPriority:          2,
	})

	require.NotNil(t, spec)
	assert.Equal(t, []string{"replication_source", "assign_public_ip", "priority"}, spec.GetUpdateMask().GetPaths())
	assert.Equal(t, "fqdn-c", spec.GetReplicationSource())
	assert.True(t, spec.GetAssignPublicIp())
	assert.Equal(t, int64(2), spec.GetPriority().GetValue())
}

func TestPGHostBackupPriorityCreate(t *testing.T) {
	d := pgClusterDataWithHosts(t,
		map[string]any{"zone": "z1", "subnet_id": "sn1", "name": "na"},
		map[string]any{"zone": "z2", "subnet_id": "sn2", "name": "nb", "backup_priority": 70, "priority": 3, "assign_public_ip": true},
	)
	apiHosts := []*postgresql.Host{{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1"}}

	compared, err := comparePGHostsInfo(d, apiHosts, true)
	require.NoError(t, err)
	require.Len(t, compared.createHostsInfo, 1)

	spec := pgCreateHostSpec(compared.createHostsInfo[0], compared.haveHostWithName)

	assert.Equal(t, "z2", spec.GetZoneId())
	assert.Equal(t, "sn2", spec.GetSubnetId())
	assert.True(t, spec.GetAssignPublicIp())
	assert.Equal(t, int64(3), spec.GetPriority().GetValue())
	assert.Equal(t, int64(70), spec.GetBackupPriority().GetValue())
}

func TestPGUnnamedHostBackupPriorityUpdate(t *testing.T) {
	d := pgClusterDataWithHosts(t, map[string]any{"zone": "z1", "subnet_id": "sn1", "backup_priority": 70})
	apiHosts := []*postgresql.Host{{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1", BackupPriority: wrapperspb.Int64(5)}}

	compared, err := comparePGHostsInfo(d, apiHosts, true)
	require.NoError(t, err)

	spec := pgHostChanges(compared.hostsInfo["fqdn-a"])
	require.NotNil(t, spec)
	assert.Equal(t, []string{"backup_priority"}, spec.GetUpdateMask().GetPaths())
	assert.Equal(t, int64(70), spec.GetBackupPriority().GetValue())
}

func TestPGHostPrioritiesMissingInConfigStayOnTheirHosts(t *testing.T) {
	apiHosts := []*postgresql.Host{
		{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1", Priority: wrapperspb.Int64(7), BackupPriority: wrapperspb.Int64(100)},
		{Name: "fqdn-b", ZoneId: "z2", SubnetId: "sn2"},
		{Name: "fqdn-c", ZoneId: "z3", SubnetId: "sn3"},
	}
	na := map[string]any{"zone": "z1", "subnet_id": "sn1", "name": "na"}
	nb := map[string]any{"zone": "z2", "subnet_id": "sn2", "name": "nb"}
	nc := map[string]any{"zone": "z3", "subnet_id": "sn3", "name": "nc"}
	state := pgClusterStateAfterRefresh(t, apiHosts, na, nb, nc)
	requireFirstRowCarriesOverPriorities := func(t *testing.T, d *schema.ResourceData) {
		t.Helper()
		require.Equal(t, 7, d.Get("host.0.priority"))
		require.Equal(t, 100, d.Get("host.0.backup_priority"))
	}

	t.Run("host is removed from the head of the list", func(t *testing.T) {
		d := pgClusterDataOnApply(t, state, nb, nc)
		requireFirstRowCarriesOverPriorities(t, d)

		compared, err := comparePGHostsInfo(d, apiHosts, true)
		require.NoError(t, err)

		assert.False(t, compared.hostsInfo["fqdn-a"].inTargetSet)
		assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-b"]))
		assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-c"]))
	})

	t.Run("host is inserted at the head of the list", func(t *testing.T) {
		nd := map[string]any{"zone": "z4", "subnet_id": "sn4", "name": "nd"}
		d := pgClusterDataOnApply(t, state, nd, na, nb, nc)
		requireFirstRowCarriesOverPriorities(t, d)

		compared, err := comparePGHostsInfo(d, apiHosts, true)
		require.NoError(t, err)

		require.Len(t, compared.createHostsInfo, 1)
		created := pgCreateHostSpec(compared.createHostsInfo[0], compared.haveHostWithName)
		require.NotNil(t, created.GetPriority(), "zero is sent explicitly: without it the API picks its own default")
		assert.Equal(t, int64(0), created.GetPriority().GetValue())
		assert.Equal(t, int64(0), created.GetBackupPriority().GetValue())
		for fqdn, hostInfo := range compared.hostsInfo {
			assert.Nil(t, pgHostChanges(hostInfo), "host %s", fqdn)
		}
	})

	t.Run("values from config are applied next to carried over ones", func(t *testing.T) {
		ncWithValues := map[string]any{"zone": "z3", "subnet_id": "sn3", "name": "nc", "priority": 4, "backup_priority": 30}
		d := pgClusterDataOnApply(t, state, nb, ncWithValues)

		compared, err := comparePGHostsInfo(d, apiHosts, true)
		require.NoError(t, err)

		assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-b"]))
		changed := pgHostChanges(compared.hostsInfo["fqdn-c"])
		require.NotNil(t, changed)
		assert.Equal(t, []string{"priority", "backup_priority"}, changed.GetUpdateMask().GetPaths())
		assert.Equal(t, int64(4), changed.GetPriority().GetValue())
		assert.Equal(t, int64(30), changed.GetBackupPriority().GetValue())
	})

	t.Run("zeros from config reset the values", func(t *testing.T) {
		naWithZeros := map[string]any{"zone": "z1", "subnet_id": "sn1", "name": "na", "priority": 0, "backup_priority": 0}
		d := pgClusterDataOnApply(t, state, naWithZeros, nb, nc)

		compared, err := comparePGHostsInfo(d, apiHosts, true)
		require.NoError(t, err)

		changed := pgHostChanges(compared.hostsInfo["fqdn-a"])
		require.NotNil(t, changed)
		assert.Equal(t, []string{"priority", "backup_priority"}, changed.GetUpdateMask().GetPaths())
		require.NotNil(t, changed.GetPriority())
		require.NotNil(t, changed.GetBackupPriority())
		assert.Equal(t, int64(0), changed.GetPriority().GetValue())
		assert.Equal(t, int64(0), changed.GetBackupPriority().GetValue())
	})
}

func refreshedPGClusterData(t *testing.T, d *schema.ResourceData, apiHosts []*postgresql.Host) (*schema.ResourceData, []map[string]any) {
	t.Helper()
	ordered, err := flattenPGHostsInfo(d, apiHosts)
	require.NoError(t, err)
	hosts := flattenPGHostsFromHostInfos(d, ordered, false)
	require.NoError(t, d.Set("host", hosts))
	return ResourceYandexMDBPostgreSQLCluster().Data(d.State()), hosts
}

func TestPGUnnamedHostPrioritiesSetOutsideTerraformSurviveRefresh(t *testing.T) {
	apiHosts := []*postgresql.Host{{
		Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1",
		Priority:       wrapperspb.Int64(5),
		BackupPriority: wrapperspb.Int64(70),
	}}
	d := pgClusterDataWithHosts(t, map[string]any{"zone": "z1", "subnet_id": "sn1"})
	d.SetId("cluster-id")

	for refresh := 1; refresh <= 2; refresh++ {
		var hosts []map[string]any
		d, hosts = refreshedPGClusterData(t, d, apiHosts)

		require.Len(t, hosts, 1, "refresh %d", refresh)
		assert.NotContains(t, hosts[0], "name")
		assert.Equal(t, int64(5), hosts[0]["priority"])
		assert.Equal(t, int64(70), hosts[0]["backup_priority"])
	}
}

func TestPGUnnamedHostPriorityFromStateDoesNotBlockUpdate(t *testing.T) {
	apiHosts := []*postgresql.Host{{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1", Priority: wrapperspb.Int64(5)}}
	d := pgClusterDataWithHosts(t, map[string]any{"zone": "z1", "subnet_id": "sn1"})
	d.SetId("cluster-id")
	d, _ = refreshedPGClusterData(t, d, apiHosts)

	compared, err := comparePGHostsInfo(d, apiHosts, true)

	require.NoError(t, err)
	assert.Empty(t, compared.createHostsInfo)
	assert.Nil(t, pgHostChanges(compared.hostsInfo["fqdn-a"]))
}

func TestFlattenPGHostsBackupPriority(t *testing.T) {
	apiHosts := []*postgresql.Host{
		{Name: "fqdn-a", ZoneId: "z1", SubnetId: "sn1", BackupPriority: wrapperspb.Int64(70)},
		{Name: "fqdn-b", ZoneId: "z2", SubnetId: "sn2"},
	}

	t.Run("resource", func(t *testing.T) {
		d := pgClusterDataWithHosts(t,
			map[string]any{"zone": "z1", "subnet_id": "sn1", "name": "na"},
			map[string]any{"zone": "z2", "subnet_id": "sn2", "name": "nb"},
		)
		ordered, err := flattenPGHostsInfo(d, apiHosts)
		require.NoError(t, err)

		hosts := flattenPGHostsFromHostInfos(d, ordered, false)

		require.Len(t, hosts, 2)
		assert.Equal(t, "na", hosts[0]["name"])
		assert.Equal(t, int64(70), hosts[0]["backup_priority"])
		assert.Equal(t, "nb", hosts[1]["name"])
		assert.Equal(t, int64(0), hosts[1]["backup_priority"])
		require.NoError(t, d.Set("host", hosts))
	})

	t.Run("data source", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, DataSourceYandexMDBPostgreSQLCluster().Schema, map[string]any{})
		ordered, err := flattenPGHostsInfo(d, apiHosts)
		require.NoError(t, err)

		hosts := flattenPGHostsFromHostInfos(d, ordered, true)

		require.Len(t, hosts, 2)
		assert.NotContains(t, hosts[0], "name")
		require.NoError(t, d.Set("host", hosts))
		assert.ElementsMatch(t, []any{70, 0}, []any{d.Get("host.0.backup_priority"), d.Get("host.1.backup_priority")})
	})
}
