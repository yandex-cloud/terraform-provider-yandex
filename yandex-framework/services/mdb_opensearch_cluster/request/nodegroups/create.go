package nodegroups

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func PrepareOpenSearchCreate(ctx context.Context, cfg model.OpenSearchSubConfig) ([]*opensearch.OpenSearchCreateSpec_NodeGroup, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	nodeGroups := make([]model.OpenSearchNode, 0, len(cfg.NodeGroups.Elements()))
	diags.Append(cfg.NodeGroups.ElementsAs(ctx, &nodeGroups, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*opensearch.OpenSearchCreateSpec_NodeGroup, 0, len(nodeGroups))
	for _, ng := range nodeGroups {

		resources, d := prepareResources(ctx, ng)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		zoneIDs := make([]string, 0, len(ng.ZoneIDs.Elements()))
		diags.Append(ng.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		var subnetIDs []string
		if !(ng.SubnetIDs.IsUnknown() || ng.SubnetIDs.IsNull()) {
			diags.Append(ng.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
			if diags.HasError() {
				return nil, diags
			}
		}

		roles := make([]opensearch.OpenSearch_GroupRole, 0, len(ng.Roles.Elements()))
		stringRoles := make([]string, 0, len(ng.Roles.Elements()))
		diags.Append(ng.Roles.ElementsAs(ctx, &stringRoles, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for _, role := range stringRoles {
			roleId := opensearch.OpenSearch_GroupRole_value[strings.ToUpper(role)]
			roles = append(roles, opensearch.OpenSearch_GroupRole(roleId))
		}

		result = append(result, &opensearch.OpenSearchCreateSpec_NodeGroup{
			Name:           ng.Name.ValueString(),
			Resources:      resources,
			HostsCount:     ng.HostsCount.ValueInt64(),
			ZoneIds:        zoneIDs,
			SubnetIds:      subnetIDs,
			AssignPublicIp: ng.AssignPublicIP.ValueBool(),
			Roles:          roles,
		})
	}

	return result, diags
}

func PrepareDashboardsCreate(ctx context.Context, cfg *model.DashboardsSubConfig) ([]*opensearch.DashboardsCreateSpec_NodeGroup, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if cfg == nil {
		return nil, diags
	}

	nodeGroups := make([]model.DashboardNode, 0, len(cfg.NodeGroups.Elements()))
	diags.Append(cfg.NodeGroups.ElementsAs(ctx, &nodeGroups, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*opensearch.DashboardsCreateSpec_NodeGroup, 0, len(nodeGroups))
	for _, ng := range nodeGroups {

		resources, diags := prepareResources(ctx, ng)
		if diags.HasError() {
			return nil, diags
		}

		zoneIDs := make([]string, 0, len(ng.ZoneIDs.Elements()))
		diags.Append(ng.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		var subnetIDs []string
		if !ng.SubnetIDs.IsUnknown() && !ng.SubnetIDs.IsNull() {
			diags.Append(ng.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
			if diags.HasError() {
				return nil, diags
			}
		}

		result = append(result, &opensearch.DashboardsCreateSpec_NodeGroup{
			Name:           ng.Name.ValueString(),
			Resources:      resources,
			HostsCount:     ng.HostsCount.ValueInt64(),
			ZoneIds:        zoneIDs,
			SubnetIds:      subnetIDs,
			AssignPublicIp: ng.AssignPublicIP.ValueBool(),
		})
	}

	return result, diags
}

func prepareResources(ctx context.Context, ng model.WithResources) (*opensearch.Resources, diag.Diagnostics) {
	r := ng.GetResources()
	if r.IsUnknown() || r.IsNull() {
		return &opensearch.Resources{}, diag.Diagnostics{}
	}

	resource, diags := model.ParseNodeResource(ctx, ng)
	if diags.HasError() {
		return nil, diags
	}

	return &opensearch.Resources{
		ResourcePresetId: resource.ResourcePresetID.ValueString(),
		DiskSize:         resource.DiskSize.ValueInt64(),
		DiskTypeId:       resource.DiskTypeID.ValueString(),
	}, diag.Diagnostics{}
}
