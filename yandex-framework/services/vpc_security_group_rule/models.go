package vpc_security_group_rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	sg "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc_security_group"
)

type securityGroupRuleModel struct {
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	ID                   types.String   `tfsdk:"id"`
	SecurityGroupBinding types.String   `tfsdk:"security_group_binding"`
	Direction            types.String   `tfsdk:"direction"`
	Description          types.String   `tfsdk:"description"`
	Labels               types.Map      `tfsdk:"labels"`
	Protocol             types.String   `tfsdk:"protocol"`
	Port                 types.Int64    `tfsdk:"port"`
	FromPort             types.Int64    `tfsdk:"from_port"`
	ToPort               types.Int64    `tfsdk:"to_port"`
	V4CidrBlocks         types.List     `tfsdk:"v4_cidr_blocks"`
	V6CidrBlocks         types.List     `tfsdk:"v6_cidr_blocks"`
	SecurityGroupID      types.String   `tfsdk:"security_group_id"`
	PredefinedTarget     types.String   `tfsdk:"predefined_target"`
}

type securityGroupRuleDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	RuleID               types.String `tfsdk:"rule_id"`
	SecurityGroupBinding types.String `tfsdk:"security_group_binding"`
	Direction            types.String `tfsdk:"direction"`
	Description          types.String `tfsdk:"description"`
	Labels               types.Map    `tfsdk:"labels"`
	Protocol             types.String `tfsdk:"protocol"`
	Port                 types.Int64  `tfsdk:"port"`
	FromPort             types.Int64  `tfsdk:"from_port"`
	ToPort               types.Int64  `tfsdk:"to_port"`
	V4CidrBlocks         types.List   `tfsdk:"v4_cidr_blocks"`
	V6CidrBlocks         types.List   `tfsdk:"v6_cidr_blocks"`
	SecurityGroupID      types.String `tfsdk:"security_group_id"`
	PredefinedTarget     types.String `tfsdk:"predefined_target"`
}

func (r securityGroupRuleModel) BodyEqual(o securityGroupRuleModel) bool {
	return r.Direction.Equal(o.Direction) &&
		r.Protocol.Equal(o.Protocol) &&
		r.Port.Equal(o.Port) &&
		r.FromPort.Equal(o.FromPort) &&
		r.ToPort.Equal(o.ToPort) &&
		r.PredefinedTarget.Equal(o.PredefinedTarget) &&
		r.SecurityGroupID.Equal(o.SecurityGroupID) &&
		r.V4CidrBlocks.Equal(o.V4CidrBlocks) &&
		r.V6CidrBlocks.Equal(o.V6CidrBlocks)
}

func expandRulePorts(port, fromPort, toPort int64) (*vpc.PortRange, error) {
	if port == -1 && fromPort == -1 && toPort == -1 {
		return nil, nil
	}

	if port != -1 {
		if fromPort != -1 || toPort != -1 {
			return nil, fmt.Errorf("cannot set from_port/to_port with port")
		}
		fromPort = port
		toPort = port
	} else if fromPort == -1 || toPort == -1 {
		return nil, fmt.Errorf("port or from_port + to_port must be defined")
	}

	return &vpc.PortRange{FromPort: fromPort, ToPort: toPort}, nil
}

func securityGroupRuleToState(ctx context.Context, rule *vpc.SecurityGroupRule, state *securityGroupRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Direction = types.StringValue(strings.ToLower(rule.GetDirection().String()))
	state.Protocol = types.StringValue(rule.ProtocolName)
	if state.Description.IsUnknown() || rule.GetDescription() != "" {
		state.Description = types.StringValue(rule.GetDescription())
	}

	if state.Labels.IsUnknown() || rule.Labels != nil {
		labels, diags := types.MapValueFrom(ctx, types.StringType, rule.Labels)
		if diags.HasError() {
			return diags
		}
		state.Labels = labels
	}

	port, fromPort, toPort := sg.FlattenRulePorts(rule)

	state.Port = types.Int64Value(port)
	state.FromPort = types.Int64Value(fromPort)
	state.ToPort = types.Int64Value(toPort)

	state.SecurityGroupID = types.StringValue(rule.GetSecurityGroupId())
	state.PredefinedTarget = types.StringValue(rule.GetPredefinedTarget())

	v4Cidrs, v6Cidrs := sg.SplitCidrs(rule.GetCidrBlocks())

	v4CidrBlocks, diags := sg.NullableStringSliceToList(ctx, v4Cidrs)
	if diags.HasError() {
		return diags
	}
	if state.V4CidrBlocks.IsUnknown() || !v4CidrBlocks.IsNull() {
		state.V4CidrBlocks = v4CidrBlocks
	}
	v6CidrBlocks, diags := sg.NullableStringSliceToList(ctx, v6Cidrs)
	if diags.HasError() {
		return diags
	}
	if state.V6CidrBlocks.IsUnknown() || !v6CidrBlocks.IsNull() {
		state.V6CidrBlocks = v6CidrBlocks
	}

	return diags
}

func stateToSecurityGroupRuleSpec(ctx context.Context, state *securityGroupRuleModel) (*vpc.SecurityGroupRuleSpec, diag.Diagnostics) {
	var diags = diag.Diagnostics{}

	portRange, err := expandRulePorts(state.Port.ValueInt64(), state.FromPort.ValueInt64(), state.ToPort.ValueInt64())
	if err != nil {
		diags.AddError(
			"Failed to construct PortRange",
			fmt.Sprintf("Error while constructing PortRange: %s", err.Error()),
		)
		return nil, diags
	}

	directionId, ok := vpc.SecurityGroupRule_Direction_value[strings.ToUpper(state.Direction.ValueString())]
	if !ok {
		diags.AddError(
			"Invalid direction value",
			fmt.Sprintf("Invalid direction value: %s", state.Direction.ValueString()),
		)
		return nil, diags
	}
	var spec = &vpc.SecurityGroupRuleSpec{
		Description: state.Description.ValueString(),
		Direction:   vpc.SecurityGroupRule_Direction(directionId),
		Ports:       portRange,
		Protocol: &vpc.SecurityGroupRuleSpec_ProtocolName{
			ProtocolName: state.Protocol.ValueString(),
		},
	}
	if !state.Labels.IsNull() && !state.Labels.IsUnknown() {
		labels := make(map[string]string, len(state.Labels.Elements()))
		diags = state.Labels.ElementsAs(ctx, &labels, false)
		spec.SetLabels(labels)
	}
	if diags.HasError() {
		return nil, diags
	}
	if !state.SecurityGroupID.IsUnknown() && state.SecurityGroupID.ValueString() != "" {
		spec.SetSecurityGroupId(state.SecurityGroupID.ValueString())
	}
	if !state.PredefinedTarget.IsUnknown() && state.PredefinedTarget.ValueString() != "" {
		spec.SetPredefinedTarget(state.PredefinedTarget.ValueString())
	}
	cidrs, diags := collectCidrs(ctx, state)
	if diags.HasError() {
		return nil, diags
	}
	if cidrs != nil {
		spec.SetCidrBlocks(cidrs)
	}

	return spec, diags
}

func collectCidrs(ctx context.Context, state *securityGroupRuleModel) (*vpc.CidrBlocks, diag.Diagnostics) {
	var diags = diag.Diagnostics{}

	v4Blocks := make([]string, 0, len(state.V4CidrBlocks.Elements()))
	if !state.V4CidrBlocks.IsUnknown() {
		diags.Append(state.V4CidrBlocks.ElementsAs(ctx, &v4Blocks, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("collectCidrs: V6CidrBlocks: %+v", state.V6CidrBlocks))
	v6Blocks := make([]string, 0, len(state.V6CidrBlocks.Elements()))
	if !state.V6CidrBlocks.IsUnknown() {
		diags.Append(state.V6CidrBlocks.ElementsAs(ctx, &v6Blocks, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	if len(v4Blocks) > 0 || len(v6Blocks) > 0 {
		var resp = &vpc.CidrBlocks{
			V4CidrBlocks: v4Blocks,
			V6CidrBlocks: v6Blocks,
		}
		if len(v4Blocks) > 0 {
			resp.V4CidrBlocks = v4Blocks
		}
		if len(v6Blocks) > 0 {
			resp.V6CidrBlocks = v6Blocks
		}
		return resp, diags
	}

	return nil, diags
}
