package security_group

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	ycsdk "github.com/yandex-cloud/go-sdk"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	sg_api "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc/security_group/api"
	"time"
)

const YandexVPCSecurityGroupDefaultTimeout = 3 * time.Minute

var (
	groupResourceAttributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"network_id": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"created_at": schema.StringAttribute{Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name":        schema.StringAttribute{Optional: true},
		"description": schema.StringAttribute{Optional: true},
		"labels":      schema.MapAttribute{Optional: true, ElementType: types.StringType},
		"folder_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	ruleResourceAttributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
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
	}

	_ resource.Resource                = &securityGroupResource{}
	_ resource.ResourceWithConfigure   = &securityGroupResource{}
	_ resource.ResourceWithImportState = &securityGroupResource{}
)

type securityGroupResource struct {
	providerConfig *provider_config.Config
}

func (g *securityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_security_group"
}
func NewResource() resource.Resource {
	return &securityGroupResource{}
}

func (g *securityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing VPC SecurityGroup schema")
	resp.Schema = schema.Schema{
		Attributes: groupResourceAttributes,
		Blocks: map[string]schema.Block{
			"ingress": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: ruleResourceAttributes,
				},
			},
			"egress": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: ruleResourceAttributes,
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (g *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (g *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (g *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateState(ctx, g.providerConfig.SDK, &state, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (g *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, YandexVPCSecurityGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	sg_api.DeleteSecurityGroup(ctx, g.providerConfig.SDK, &resp.Diagnostics, state.ID.ValueString())
}

func (g *securityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	g.providerConfig = providerConfig
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *securityGroupModel, diag *diag.Diagnostics, createIfMissing bool) {
	sgID := state.ID.ValueString()
	tflog.Debug(ctx, "Reading VPC SecurityGroup", map[string]interface{}{"id": sgID})
	sg := sg_api.ReadSecurityGroup(ctx, sdk, diag, sgID)
	if diag.HasError() {
		return
	}

	if sg == nil {
		if createIfMissing {
			// To create a new security group if missing
			state.ID = types.StringUnknown()
			return
		}

		diag.AddError(
			"Failed to get SecurityGroup",
			fmt.Sprintf("SecurityGroup with id %s not found", sgID))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("updateState: VPC SecurityGroup state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("updateState: Received VPC SecurityGroup data: %+v", sg))

	diags := securityGroupToState(ctx, sg, state)
	diag.Append(diags...)
}
