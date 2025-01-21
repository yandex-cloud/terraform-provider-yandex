package vpc_security_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

type securityGroupModel struct {
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	ID          types.String   `tfsdk:"id"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Labels      types.Map      `tfsdk:"labels"`
	FolderID    types.String   `tfsdk:"folder_id"`
	NetworkID   types.String   `tfsdk:"network_id"`
	Status      types.String   `tfsdk:"status"`
	Ingress     types.Set      `tfsdk:"ingress"`
	Egress      types.Set      `tfsdk:"egress"`
}

type securityGroupDataSourceModel struct {
	securityGroupModel

	SecurityGroupID types.String `tfsdk:"security_group_id"`
}

var ruleType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":                types.StringType,
		"description":       types.StringType,
		"labels":            types.MapType{ElemType: types.StringType},
		"protocol":          types.StringType,
		"port":              types.Int64Type,
		"from_port":         types.Int64Type,
		"to_port":           types.Int64Type,
		"v4_cidr_blocks":    types.ListType{ElemType: types.StringType},
		"v6_cidr_blocks":    types.ListType{ElemType: types.StringType},
		"security_group_id": types.StringType,
		"predefined_target": types.StringType,
	},
}

func securityGroupToState(ctx context.Context, sg *vpc.SecurityGroup, state *securityGroupModel) diag.Diagnostics {
	state.FolderID = types.StringValue(sg.GetFolderId())
	state.NetworkID = types.StringValue(sg.GetNetworkId())
	state.Status = types.StringValue(sg.GetStatus().String())
	state.CreatedAt = types.StringValue(timestamp.Get(sg.GetCreatedAt()))
	state.Name = types.StringValue(sg.GetName())
	if state.Description.IsUnknown() || sg.GetDescription() != "" {
		state.Description = types.StringValue(sg.GetDescription())
	}

	if state.Labels.IsUnknown() || sg.Labels != nil {
		labels, diags := types.MapValueFrom(ctx, types.StringType, sg.Labels)
		if diags.HasError() {
			return diags
		}
		state.Labels = labels
	}

	var ingress, egress, diags = flattenRules(ctx, sg.GetRules())
	if diags.HasError() {
		return diags
	}
	state.Ingress = ingress
	state.Egress = egress

	return diags
}

func flattenRules(ctx context.Context, rules []*vpc.SecurityGroupRule) (types.Set, types.Set, diag.Diagnostics) {
	var ingressValues []attr.Value
	var egressValues []attr.Value

	var diags diag.Diagnostics

	for _, rule := range rules {
		labels, diagnostics := types.MapValueFrom(ctx, types.StringType, rule.Labels)
		diags.Append(diagnostics...)
		if diags.HasError() {
			continue
		}

		port, fromPort, toPort := FlattenRulePorts(rule)

		v4Cidrs, v6Cidrs := SplitCidrs(rule.GetCidrBlocks())

		v4CidrsValue, diagnostics := NullableStringSliceToList(ctx, v4Cidrs)
		diags.Append(diagnostics...)
		if diags.HasError() {
			continue
		}
		v6CidrsValue, diagnostics := NullableStringSliceToList(ctx, v6Cidrs)
		diags.Append(diagnostics...)
		if diags.HasError() {
			continue
		}

		ruleValue, diagnostics := types.ObjectValue(ruleType.AttrTypes, map[string]attr.Value{
			"id":                types.StringValue(rule.Id),
			"description":       types.StringValue(rule.Description),
			"labels":            labels,
			"protocol":          types.StringValue(rule.ProtocolName),
			"port":              types.Int64Value(port),
			"from_port":         types.Int64Value(fromPort),
			"to_port":           types.Int64Value(toPort),
			"v4_cidr_blocks":    v4CidrsValue,
			"v6_cidr_blocks":    v6CidrsValue,
			"security_group_id": types.StringValue(rule.GetSecurityGroupId()),
			"predefined_target": types.StringValue(rule.GetPredefinedTarget()),
		})
		diags.Append(diagnostics...)
		if diags.HasError() {
			continue
		}

		switch rule.GetDirection() {
		case vpc.SecurityGroupRule_INGRESS:
			ingressValues = append(ingressValues, ruleValue)
		case vpc.SecurityGroupRule_EGRESS:
			egressValues = append(egressValues, ruleValue)
		}
	}

	ingress, diagnostics := types.SetValue(ruleType, ingressValues)
	diags.Append(diagnostics...)
	egress, diagnostics := types.SetValue(ruleType, egressValues)
	diags.Append(diagnostics...)

	return ingress, egress, diags
}

func SplitCidrs(cidrs *vpc.CidrBlocks) ([]string, []string) {
	return cidrs.GetV4CidrBlocks(), cidrs.GetV6CidrBlocks()
}

func FlattenRulePorts(g *vpc.SecurityGroupRule) (port, fromPort, toPort int64) {
	port = -1
	fromPort = -1
	toPort = -1

	if ports := g.GetPorts(); ports != nil {
		if ports.FromPort == ports.ToPort {
			port = ports.FromPort
		} else {
			fromPort = ports.FromPort
			toPort = ports.ToPort
		}
	}

	return
}

func NullableStringSliceToList(ctx context.Context, s []string) (types.List, diag.Diagnostics) {
	if s == nil {
		return types.ListNull(types.StringType), diag.Diagnostics{}
	}

	return types.ListValueFrom(ctx, types.StringType, s)
}
