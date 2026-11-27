package mdb_postgresql_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestPostgreSQLHostBackupPrioritySchema(t *testing.T) {
	ctx := context.Background()
	var resp frameworkresource.SchemaResponse
	NewPostgreSQLClusterResourceV2().Schema(ctx, frameworkresource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema diagnostics: %v", resp.Diagnostics)

	hosts := resp.Schema.Attributes["hosts"].(schema.MapNestedAttribute)
	backupPriority, ok := hosts.NestedObject.Attributes["backup_priority"].(schema.Int64Attribute)
	require.True(t, ok, "hosts.backup_priority must be an int64 attribute")
	assert.True(t, backupPriority.IsOptional())
	assert.True(t, backupPriority.IsComputed())

	for value, wantError := range map[int64]bool{-1: true, 0: false, 100: false, 101: true} {
		var validation validator.Int64Response
		for _, v := range backupPriority.Int64Validators() {
			v.ValidateInt64(ctx, validator.Int64Request{
				Path:        path.Root("hosts"),
				ConfigValue: types.Int64Value(value),
			}, &validation)
		}
		assert.Equal(t, wantError, validation.Diagnostics.HasError(), "backup_priority = %d", value)
	}
}

func TestGetChangesBackupPriority(t *testing.T) {
	base := stateHost("zone-a", "subnet-a", "host-a"+testHostDomain)
	base.BackupPriority = types.Int64Value(5)

	t.Run("unknown backup priority -> no change", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, spec)
	})

	t.Run("same backup priority -> no change", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		plan.BackupPriority = types.Int64Value(5)
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, spec)
	})

	t.Run("backup priority reset to zero -> update", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		plan.BackupPriority = types.Int64Value(0)
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, spec)
		assert.Equal(t, []string{"backup_priority"}, spec.GetUpdateMask().GetPaths())
		require.NotNil(t, spec.GetBackupPriority())
		assert.Equal(t, int64(0), spec.GetBackupPriority().GetValue())
	})

	t.Run("differing backup priority -> update", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		plan.BackupPriority = types.Int64Value(70)
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, spec)
		assert.Equal(t, []string{"backup_priority"}, spec.GetUpdateMask().GetPaths())
		assert.Equal(t, int64(70), spec.GetBackupPriority().GetValue())
	})
}

func TestGetChangesHostAttributes(t *testing.T) {
	base := stateHost("zone-a", "subnet-a", "host-a"+testHostDomain)

	t.Run("public ip and replication source -> update", func(t *testing.T) {
		plan := planHost("zone-a", "subnet-a")
		plan.AssignPublicIp = types.BoolValue(true)
		plan.ReplicationSource = types.StringValue("host-b" + testHostDomain)
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, spec)
		assert.Equal(t, "host-a"+testHostDomain, spec.GetHostName())
		assert.Equal(t, []string{"assign_public_ip", "replication_source"}, spec.GetUpdateMask().GetPaths())
		assert.True(t, spec.GetAssignPublicIp())
		assert.Equal(t, "host-b"+testHostDomain, spec.GetReplicationSource())
	})

	t.Run("zone change -> error", func(t *testing.T) {
		plan := planHost("zone-b", "subnet-a")
		spec, diags := postgresqlHostService.GetChanges(plan, base)
		assert.Nil(t, spec)
		require.True(t, diags.HasError())
		assert.Equal(t, "Wrong changes for host", diags.Errors()[0].Summary())
	})
}

func TestConvertToProtoSendsBackupPriority(t *testing.T) {
	host := planHost("zone-a", "subnet-a")
	host.BackupPriority = types.Int64Value(70)

	spec := postgresqlHostService.ConvertToProto(host)

	assert.Equal(t, int64(70), spec.GetBackupPriority().GetValue())
}

func TestReadHostsReturnsBackupPriority(t *testing.T) {
	ctx := context.Background()
	planValue, diags := types.MapValueFrom(ctx, hostType, map[string]Host{
		"with-priority":    planHost("sas", "subnet-sas"),
		"without-priority": planHost("vla", "subnet-vla"),
	})
	require.False(t, diags.HasError(), "building hosts plan: %v", diags)

	api := testPostgresqlHostAPI{hosts: []*postgresql.Host{
		{Name: "host-sas" + testHostDomain, ZoneId: "sas", SubnetId: "subnet-sas", BackupPriority: wrapperspb.Int64(70)},
		{Name: "host-vla" + testHostDomain, ZoneId: "vla", SubnetId: "subnet-vla"},
	}}
	var readDiags diag.Diagnostics

	got := mdbcommon.ReadHosts(ctx, nil, &readDiags, postgresqlHostService, api, planValue, "cluster-id")

	require.False(t, readDiags.HasError(), "reading hosts: %v", readDiags)
	require.Len(t, got, 2)
	assert.Equal(t, types.Int64Value(70), got["with-priority"].BackupPriority)
	assert.Equal(t, types.Int64Value(0), got["without-priority"].BackupPriority)
}
