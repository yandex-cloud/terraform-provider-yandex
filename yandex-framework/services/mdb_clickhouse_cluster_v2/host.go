package mdb_clickhouse_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var clickhouseHostService = &ClickHouseHostService{}

type ClickHouseHostService struct{}

func (r ClickHouseHostService) FullyMatch(planHost models.Host, stateHost models.Host) bool {
	return planHost.Zone.ValueString() == stateHost.Zone.ValueString() &&
		(planHost.ShardName.IsUnknown() || planHost.ShardName.ValueString() == stateHost.ShardName.ValueString()) &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.ValueString() == stateHost.SubnetId.ValueString()) &&
		(planHost.Type.IsUnknown() || planHost.Type.ValueString() == stateHost.Type.ValueString()) &&
		planHost.AssignPublicIp.ValueBool() == stateHost.AssignPublicIp.ValueBool()
}

func (r ClickHouseHostService) PartialMatch(planHost models.Host, stateHost models.Host) bool {
	return planHost.Zone.Equal(stateHost.Zone) &&
		(planHost.FQDN.IsUnknown() || planHost.FQDN.Equal(stateHost.FQDN)) &&
		(planHost.ShardName.IsUnknown() || planHost.ShardName.Equal(stateHost.ShardName)) &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.Equal(stateHost.SubnetId))
}

func (r ClickHouseHostService) GetChanges(plan models.Host, state models.Host) (*clickhouse.UpdateHostSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !r.PartialMatch(plan, state) {
		diags.AddError(
			"Wrong changes for host",
			"Attributes zone, subnet_id, shard_name can't be changed. Try to replace this host to new one",
		)
		return nil, diags
	}

	if plan.AssignPublicIp.Equal(state.AssignPublicIp) {
		return nil, nil
	}

	return &clickhouse.UpdateHostSpec{
		HostName: state.FQDN.ValueString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"assign_public_ip"},
		},
		AssignPublicIp: &wrapperspb.BoolValue{Value: plan.AssignPublicIp.ValueBool()},
	}, diags
}

func (r ClickHouseHostService) ConvertToProto(h models.Host) *clickhouse.HostSpec {
	a := &clickhouse.HostSpec{
		ZoneId:         h.Zone.ValueString(),
		Type:           HostTypeToProto(h.Type.ValueString()),
		ShardName:      h.ShardName.ValueString(),
		SubnetId:       h.SubnetId.ValueString(),
		AssignPublicIp: h.AssignPublicIp.ValueBool(),
	}
	return a
}

func (r ClickHouseHostService) ConvertFromProto(apiHost *clickhouse.Host) models.Host {
	return models.Host{
		Zone:           types.StringValue(apiHost.ZoneId),
		Type:           types.StringValue(apiHost.GetType().String()),
		SubnetId:       types.StringValue(apiHost.SubnetId),
		AssignPublicIp: types.BoolValue(apiHost.AssignPublicIp),
		ShardName:      types.StringValue(apiHost.ShardName),
		FQDN:           types.StringValue(apiHost.Name),
	}
}

func HostTypeToProto(hostType string) clickhouse.Host_Type {
	switch hostType {
	case "CLICKHOUSE":
		return clickhouse.Host_CLICKHOUSE
	case "ZOOKEEPER":
		return clickhouse.Host_ZOOKEEPER
	case "KEEPER":
		return clickhouse.Host_KEEPER
	default:
		return clickhouse.Host_TYPE_UNSPECIFIED
	}
}

func splitHostSpecsByType(ctx context.Context, hosts types.Map, diags *diag.Diagnostics) (types.Map, types.Map) {
	elemType := hosts.ElementType(ctx)
	empty := types.MapNull(elemType)

	if hosts.IsNull() || hosts.IsUnknown() {
		return empty, empty
	}

	var hostMap map[string]models.Host
	diags.Append(hosts.ElementsAs(ctx, &hostMap, false)...)
	if diags.HasError() {
		return empty, empty
	}

	chHostMap := make(map[string]models.Host)
	keeperHostMap := make(map[string]models.Host)

	for label, host := range hostMap {
		switch host.Type.ValueString() {
		case "CLICKHOUSE":
			chHostMap[label] = host
		case "ZOOKEEPER", "KEEPER":
			keeperHostMap[label] = host
		}
	}

	chHostResult, d := types.MapValueFrom(ctx, elemType, chHostMap)
	if d.HasError() {
		return empty, empty
	}
	keeperHostResult, d := types.MapValueFrom(ctx, elemType, keeperHostMap)
	if d.HasError() {
		return empty, empty
	}

	return chHostResult, keeperHostResult
}
