package model

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
)

type Host struct {
	FQDN           types.String `tfsdk:"fqdn"`
	Type           types.String `tfsdk:"type"`
	Roles          types.Set    `tfsdk:"roles"`
	AssignPublicIP types.Bool   `tfsdk:"assign_public_ip"`
	Zone           types.String `tfsdk:"zone"`
	SubnetID       types.String `tfsdk:"subnet_id"`
	NodeGroup      types.String `tfsdk:"node_group"`
}

var HostType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"fqdn":             types.StringType,
		"type":             types.StringType,
		"roles":            types.SetType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
		"zone":             types.StringType,
		"subnet_id":        types.StringType,
		"node_group":       types.StringType,
	},
}

func HostsToState(ctx context.Context, hosts []*opensearch.Host) (types.List, diag.Diagnostics) {
	res := make([]Host, 0, len(hosts))

	for _, h := range hosts {
		roles, diags := rolesToSet(h.GetRoles())
		if diags.HasError() {
			diags.AddError("Failed to parse hosts.roles", fmt.Sprintf("Error while parsing roles for host: %s", h.GetName()))
			return types.ListUnknown(HostType), diags
		}

		res = append(res, Host{
			FQDN:           types.StringValue(h.GetName()),
			Type:           types.StringValue(h.GetType().String()),
			Roles:          roles,
			AssignPublicIP: types.BoolValue(h.GetAssignPublicIp()),
			Zone:           types.StringValue(h.GetZoneId()),
			SubnetID:       types.StringValue(h.GetSubnetId()),
			NodeGroup:      types.StringValue(h.GetNodeGroup()),
		})
	}

	return types.ListValueFrom(ctx, HostType, res)
}
