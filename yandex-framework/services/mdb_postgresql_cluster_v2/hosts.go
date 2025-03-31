package mdb_postgresql_cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var postgresqlHostService = &PostgresqlHostService{}

type PostgresqlHostService struct {
}

func (r PostgresqlHostService) FullyMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.ValueString() == stateHost.Zone.ValueString() &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.ValueString() == stateHost.SubnetId.ValueString()) &&
		planHost.AssignPublicIp.ValueBool() == stateHost.AssignPublicIp.ValueBool() &&
		(planHost.ReplicationSource.IsUnknown() || planHost.ReplicationSource.Equal(stateHost.ReplicationSource))
}

func (r PostgresqlHostService) PartialMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.Equal(stateHost.Zone) &&
		(planHost.FQDN.IsUnknown() || planHost.FQDN.Equal(stateHost.FQDN)) &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.Equal(stateHost.SubnetId))
}

func (r PostgresqlHostService) GetChanges(plan Host, state Host) (*postgresql.UpdateHostSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !r.PartialMatch(plan, state) {
		diags.AddError(
			"Wrong changes for host",
			"Attributes zone, subnet_id can't be changed. Try to replace this host to new one",
		)
		return nil, diags
	}
	if plan.AssignPublicIp.Equal(state.AssignPublicIp) && plan.ReplicationSource.Equal(state.ReplicationSource) {
		return nil, nil
	}
	return &postgresql.UpdateHostSpec{
		HostName: state.FQDN.ValueString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"assign_public_ip", "replica_priority"},
		},
		AssignPublicIp:    plan.AssignPublicIp.ValueBool(),
		ReplicationSource: plan.ReplicationSource.ValueString(),
	}, diags
}

func (r PostgresqlHostService) ConvertToProto(h Host) *postgresql.HostSpec {
	return &postgresql.HostSpec{
		ZoneId:            h.Zone.ValueString(),
		SubnetId:          h.SubnetId.ValueString(),
		AssignPublicIp:    h.AssignPublicIp.ValueBool(),
		ReplicationSource: h.ReplicationSource.ValueString(),
	}
}

func (r PostgresqlHostService) ConvertFromProto(apiHost *postgresql.Host) Host {
	return Host{
		Zone: types.StringValue(apiHost.ZoneId),

		SubnetId:          types.StringValue(apiHost.SubnetId),
		AssignPublicIp:    types.BoolValue(apiHost.AssignPublicIp),
		ReplicationSource: types.StringValue(apiHost.ReplicationSource),
		FQDN:              types.StringValue(apiHost.Name),
	}
}

func (h Host) GetFQDN() types.String {
	return h.FQDN
}
