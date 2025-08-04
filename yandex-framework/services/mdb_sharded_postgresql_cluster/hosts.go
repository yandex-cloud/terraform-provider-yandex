package mdb_sharded_postgresql_cluster

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var spqrHostService = &SPQRHostService{}

type SPQRHostService struct {
}

func (r SPQRHostService) FullyMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.ValueString() == stateHost.Zone.ValueString() &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.ValueString() == stateHost.SubnetId.ValueString()) &&
		planHost.AssignPublicIp.ValueBool() == stateHost.AssignPublicIp.ValueBool()
}

func (r SPQRHostService) PartialMatch(planHost Host, stateHost Host) bool {
	return planHost.Zone.Equal(stateHost.Zone) &&
		(planHost.FQDN.IsUnknown() || planHost.FQDN.Equal(stateHost.FQDN)) &&
		(planHost.SubnetId.IsUnknown() || planHost.SubnetId.Equal(stateHost.SubnetId))
}

func (r SPQRHostService) GetChanges(plan Host, state Host) (*spqr.UpdateHostSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !r.PartialMatch(plan, state) {
		diags.AddError(
			"Wrong changes for host",
			"Attributes shard_name, zone, subnet_id can't be changed. Try to replace this host to new one",
		)
		return nil, diags
	}
	if plan.AssignPublicIp.Equal(state.AssignPublicIp) {
		return nil, nil
	}
	return &spqr.UpdateHostSpec{
		HostName: state.FQDN.ValueString(),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"assign_public_ip"},
		},
		AssignPublicIp: plan.AssignPublicIp.ValueBool(),
	}, diags
}

func (r SPQRHostService) ConvertToProto(h Host) *spqr.HostSpec {
	return &spqr.HostSpec{
		ZoneId:         h.Zone.ValueString(),
		SubnetId:       h.SubnetId.ValueString(),
		AssignPublicIp: h.AssignPublicIp.ValueBool(),
		Type:           mapHostTypeToProto[strings.ToUpper(h.Type.ValueString())],
	}
}

func (r SPQRHostService) ConvertFromProto(apiHost *spqr.Host) Host {
	return Host{
		Zone: types.StringValue(apiHost.ZoneId),

		SubnetId:       types.StringValue(apiHost.SubnetId),
		AssignPublicIp: types.BoolValue(apiHost.AssignPublicIp),
		FQDN:           types.StringValue(apiHost.Name),
		Type:           types.StringValue(mapHostTypeFromProto[apiHost.Type]),
	}
}

func (h Host) GetFQDN() types.String {
	return h.FQDN
}

var (
	mapHostTypeToProto = map[string]spqr.Host_Type{
		"ROUTER":      spqr.Host_ROUTER,
		"COORDINATOR": spqr.Host_COORDINATOR,
		"INFRA":       spqr.Host_INFRA,
		"POSTGRESQL":  spqr.Host_POSTGRESQL,
	}
	mapHostTypeFromProto = map[spqr.Host_Type]string{
		spqr.Host_ROUTER:      "ROUTER",
		spqr.Host_COORDINATOR: "COORDINATOR",
		spqr.Host_INFRA:       "INFRA",
		spqr.Host_POSTGRESQL:  "POSTGRESQL",
	}
)
