package vpc_security_group_rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	sg "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc_security_group"
	sg_api "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc_security_group/api"
)

var (
	_ datasource.DataSource              = &securityGroupRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &securityGroupRuleDataSource{}
)

type securityGroupRuleDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &securityGroupRuleDataSource{}
}

func (r securityGroupRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_security_group_rule"
}

func (r securityGroupRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing VPC SecurityGroupRule datasource schema")
	var attributes = sg.RuleDataSourceAttributes
	attributes["rule_id"] = schema.StringAttribute{Required: true}
	attributes["security_group_binding"] = schema.StringAttribute{Required: true}
	attributes["direction"] = schema.StringAttribute{Computed: true}
	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *securityGroupRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r securityGroupRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading VPC SecurityGroupRule datasource")
	tflog.Debug(ctx, fmt.Sprintf("Read: VPC SecurityGroupRule raw state: %+v", req.Config.Raw))
	var state securityGroupRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = state.RuleID
	updateRuleDatasourceState(ctx, r.providerConfig.SDK, &state, &resp.Diagnostics, false)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func updateRuleDatasourceState(ctx context.Context, sdk *ycsdk.SDK, state *securityGroupRuleDataSourceModel, diag *diag.Diagnostics, createIfMissing bool) {
	sgID := state.SecurityGroupBinding.ValueString()
	ruleID := state.ID.ValueString()
	rule := sg_api.FindSecurityGroupRule(ctx, sdk, diag, sgID, ruleID)
	if diag.HasError() {
		return
	}

	if rule == nil {
		if createIfMissing {
			// To create a new security group rule if missing
			state.ID = types.StringUnknown()
			return
		}

		diag.AddError(
			"Failed to get SecurityGroupRule data",
			fmt.Sprintf("SecurityGroupRule with id %s not found", ruleID))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("updateRuleState: VPC SecurityGroupRule state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("updateRuleState: Received VPC SecurityGroupRule data: %+v", rule))

	diags := securityGroupRuleToDatasourceState(ctx, rule, state)
	diag.Append(diags...)
}

// copypasta suxx, but one does not simply inherit a struct...
func securityGroupRuleToDatasourceState(ctx context.Context, rule *vpc.SecurityGroupRule, state *securityGroupRuleDataSourceModel) diag.Diagnostics {
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

	state.V4CidrBlocks, diags = sg.NullableStringSliceToList(ctx, v4Cidrs)
	if diags.HasError() {
		return diags
	}
	state.V6CidrBlocks, diags = sg.NullableStringSliceToList(ctx, v6Cidrs)
	if diags.HasError() {
		return diags
	}

	return diags
}
