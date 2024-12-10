package security_group_rule

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/globallock"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc/security_group"
	sg_api "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc/security_group/api"
	spm "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/stringplanmodifier"
	"google.golang.org/genproto/protobuf/field_mask"
)

var (
	_ resource.Resource                   = &securityGroupRuleResource{}
	_ resource.ResourceWithConfigure      = &securityGroupRuleResource{}
	_ resource.ResourceWithImportState    = &securityGroupRuleResource{}
	_ resource.ResourceWithValidateConfig = &securityGroupRuleResource{}
)

type securityGroupRuleResource struct {
	providerConfig *provider_config.Config
}

func (r *securityGroupRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_security_group_rule"
}
func NewResource() resource.Resource {
	return &securityGroupRuleResource{}
}

func (r *securityGroupRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing VPC SecurityGroupRule schema")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					spm.RequiresRefreshIf(func(ctx context.Context, req planmodifier.StringRequest, resp *spm.RequiresRefreshIfFuncResponse) {
						var plan securityGroupRuleModel
						var state securityGroupRuleModel
						resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
						resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
						if resp.Diagnostics.HasError() {
							return
						}
						resp.RequiresRefresh = !state.BodyEqual(plan)
					},
						"Refresh if rule body modified",
						"Refresh if rule body modified",
					),
				},
			},
			"security_group_binding": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("ingress", "egress"),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
					int64validator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("from_port"),
						path.MatchRelative().AtParent().AtName("to_port"),
					),
					int64validator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("protocol"),
					),
				},
				Default: int64default.StaticInt64(-1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"from_port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
					int64validator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("port"),
					),
					int64validator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("protocol"),
					),
				},
				Default: int64default.StaticInt64(-1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"to_port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
					int64validator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("port"),
					),
					int64validator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("protocol"),
					),
				},
				Default: int64default.StaticInt64(-1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"v4_cidr_blocks": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("security_group_id"),
						path.MatchRelative().AtParent().AtName("predefined_target"),
					}...),
				},
			},
			"v6_cidr_blocks": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("security_group_id"),
						path.MatchRelative().AtParent().AtName("predefined_target"),
					}...),
				},
			},
			"security_group_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("v4_cidr_blocks"),
						path.MatchRelative().AtParent().AtName("v6_cidr_blocks"),
						path.MatchRelative().AtParent().AtName("predefined_target"),
					}...),
				},
			},
			"predefined_target": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("v4_cidr_blocks"),
						path.MatchRelative().AtParent().AtName("v6_cidr_blocks"),
						path.MatchRelative().AtParent().AtName("security_group_id"),
					}...),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *securityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sgID, ruleID, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	var state securityGroupRuleModel
	state.SecurityGroupBinding = types.StringValue(sgID)
	state.ID = types.StringValue(ruleID)
	state.Labels = types.MapNull(types.StringType)
	// all complex object fields must be explicitly set to corresponding "null" values...
	state.V4CidrBlocks = types.ListNull(types.StringType)
	state.V6CidrBlocks = types.ListNull(types.StringType)
	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"update": types.StringType,
		}),
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *securityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := plan.SecurityGroupBinding.ValueString()
	mutexKV := globallock.GetMutexKV()
	mutexKV.Lock(sgID)
	defer mutexKV.Unlock(sgID)

	createTimeout, diags := plan.Timeouts.Create(ctx, security_group.YandexVPCSecurityGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	ruleSpec, diags := stateToSecurityGroupRuleSpec(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	meta := sg_api.UpdateSecurityGroupRules(ctx, r.providerConfig.SDK, &resp.Diagnostics, sgID, ruleSpec, "")
	if resp.Diagnostics.HasError() {
		return
	}
	if meta.GetAddedRuleIds() == nil || len(meta.GetAddedRuleIds()) != 1 {
		resp.Diagnostics.AddError(
			"Error adding rule",
			"Rule was not added or not a singleton",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(meta.AddedRuleIds[0])

	updateRuleState(ctx, r.providerConfig.SDK, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRuleState(ctx, r.providerConfig.SDK, &state, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityGroupRuleModel
	var state securityGroupRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := state.SecurityGroupBinding.ValueString()
	mutexKV := globallock.GetMutexKV()
	mutexKV.Lock(sgID)
	defer mutexKV.Unlock(sgID)

	updateTimeout, diags := state.Timeouts.Update(ctx, security_group.YandexVPCSecurityGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var needFullUpdate = false

	var updateRuleReq = &vpc.UpdateSecurityGroupRuleRequest{
		SecurityGroupId: state.SecurityGroupBinding.ValueString(),
		RuleId:          state.ID.ValueString(),
		UpdateMask:      &field_mask.FieldMask{},
	}

	if !state.Description.Equal(plan.Description) {
		updateRuleReq.Description = plan.Description.ValueString()
		updateRuleReq.UpdateMask.Paths = append(updateRuleReq.UpdateMask.Paths, "description")
	}

	if !state.Labels.Equal(plan.Labels) {
		labels := make(map[string]string, len(state.Labels.Elements()))
		resp.Diagnostics.Append(state.Labels.ElementsAs(ctx, &labels, false)...)
		updateRuleReq.SetLabels(labels)
		updateRuleReq.UpdateMask.Paths = append(updateRuleReq.UpdateMask.Paths, "labels")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.BodyEqual(plan) {
		needFullUpdate = true
	}

	if needFullUpdate {
		ruleSpec, diags := stateToSecurityGroupRuleSpec(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		meta := sg_api.UpdateSecurityGroupRules(ctx, r.providerConfig.SDK, &resp.Diagnostics, sgID, ruleSpec, state.ID.ValueString())
		if meta.GetAddedRuleIds() == nil || len(meta.GetAddedRuleIds()) != 1 {
			resp.Diagnostics.AddError(
				"Error replacing rule",
				"Rule was not added or not a singleton",
			)
		}
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ID = types.StringValue(meta.AddedRuleIds[0])
	} else if len(updateRuleReq.UpdateMask.Paths) > 0 {
		sg_api.UpdateSecurityGroupRuleMetadata(ctx, r.providerConfig.SDK, &resp.Diagnostics, updateRuleReq)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateRuleState(ctx, r.providerConfig.SDK, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, security_group.YandexVPCSecurityGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := state.SecurityGroupBinding.ValueString()
	mutexKV := globallock.GetMutexKV()
	mutexKV.Lock(sgID)
	defer mutexKV.Unlock(sgID)

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	sg_api.UpdateSecurityGroupRules(ctx, r.providerConfig.SDK, &resp.Diagnostics, sgID, nil, state.ID.ValueString())
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (r *securityGroupRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var state securityGroupRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.FromPort.IsUnknown() || state.ToPort.IsUnknown() {
		return
	}
	if state.FromPort.IsNull() || state.ToPort.IsNull() {
		return
	}
	if state.FromPort.ValueInt64() == -1 && state.ToPort.ValueInt64() == -1 {
		return
	}
	if state.FromPort.ValueInt64() == state.ToPort.ValueInt64() {
		resp.Diagnostics.AddError(
			"Invalid SecurityGroupRule",
			"Use port attribute to specify single port value",
		)
	} else if state.FromPort.ValueInt64() > state.ToPort.ValueInt64() {
		resp.Diagnostics.AddError(
			"Invalid SecurityGroupRule",
			fmt.Sprintf("from_port (%d) must be less than to_port (%d)", state.FromPort.ValueInt64(), state.ToPort.ValueInt64()),
		)
	}
}

func (r *securityGroupRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func updateRuleState(ctx context.Context, sdk *ycsdk.SDK, state *securityGroupRuleModel, diag *diag.Diagnostics) {
	sgID := state.SecurityGroupBinding.ValueString()
	ruleID := state.ID.ValueString()
	rule := sg_api.FindSecurityGroupRule(ctx, sdk, diag, sgID, ruleID)
	if diag.HasError() {
		return
	}

	if rule == nil {
		diag.AddError(
			"Failed to get SecurityGroupRule data",
			fmt.Sprintf("SecurityGroupRule with id %s not found", ruleID))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("updateRuleState: VPC SecurityGroupRule state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("updateRuleState: Received VPC SecurityGroupRule data: %+v", rule))

	diags := securityGroupRuleToState(ctx, rule, state)
	diag.Append(diags...)
}
