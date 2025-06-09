package mdb_redis_user

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
)

const defaultName = "default"

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_redis_user"
}

func (r *bindingResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bindingResource) Schema(_ context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Redis user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster to which user belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"passwords": schema.SetAttribute{
				ElementType:         basetypes.StringType{},
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Set of user passwords",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 1),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Is redis user enabled.",
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"acl_options": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Raw ACL string which has been inserted into the Redis",
			},
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "Set of permissions granted to the user.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"commands": schema.StringAttribute{
						MarkdownDescription: "Commands user can execute.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"categories": schema.StringAttribute{
						MarkdownDescription: "Command categories user has permissions to.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"patterns": schema.StringAttribute{
						MarkdownDescription: "Keys patterns user has permission to.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"pub_sub_channels": schema.StringAttribute{
						MarkdownDescription: "Channel patterns user has permissions to.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sanitize_payload": schema.StringAttribute{
						MarkdownDescription: "SanitizePayload parameter.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive(
								"sanitize-payload",
								"skip-sanitize-payload",
							),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user, err := r.providerConfig.SDK.MDB().Redis().User().Get(ctx, &redis.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})

	if err != nil {
		f := resp.Diagnostics.AddError
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.State.RemoveResource(ctx)
			f = resp.Diagnostics.AddWarning
		}

		f(
			"Failed to Read resource",
			"Error while requesting API to get Redis user:"+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(userToState(ctx, user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, userName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan User
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := plan.ClusterID.ValueString()
	userPlan, diags := userFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, userPlan)
	if resp.Diagnostics.HasError() {
		return
	}

	id := types.StringValue(resourceid.Construct(cid, userPlan.Name))
	userRead(ctx, r.providerConfig.SDK, &resp.Diagnostics, &plan)
	plan.Id = id
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(ctx context.Context, diags *diag.Diagnostics, plan, state User) []string {
	var updatePaths []string

	if !plan.Permissions.Equal(state.Permissions) {
		var planPermissions, statePermissions Permissions
		const permissionsPrefix = "permissions."

		diags.Append(plan.Permissions.As(ctx, &planPermissions, basetypes.ObjectAsOptions{})...)
		diags.Append(state.Permissions.As(ctx, &statePermissions, basetypes.ObjectAsOptions{})...)

		if !planPermissions.Patterns.Equal(statePermissions.Patterns) {
			updatePaths = append(updatePaths, permissionsPrefix+"patterns")
		}
		if !planPermissions.Commands.Equal(statePermissions.Commands) {
			updatePaths = append(updatePaths, permissionsPrefix+"commands")
		}
		if !planPermissions.Categories.Equal(statePermissions.Categories) {
			updatePaths = append(updatePaths, permissionsPrefix+"categories")
		}
		if !planPermissions.PubSubChannels.Equal(statePermissions.PubSubChannels) {
			updatePaths = append(updatePaths, permissionsPrefix+"pub_sub_channels")
		}
		if !planPermissions.SanitizePayload.Equal(statePermissions.SanitizePayload) {
			updatePaths = append(updatePaths, permissionsPrefix+"sanitize_payload")
		}
	}

	if !plan.Enabled.Equal(state.Enabled) {
		updatePaths = append(updatePaths, "enabled")
	}

	if !plan.Passwords.Equal(state.Passwords) {
		updatePaths = append(updatePaths, "passwords")
	}

	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state User
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatePaths := getUpdatePaths(ctx, &resp.Diagnostics, plan, state)
	if resp.Diagnostics.HasError() {
		return
	}

	userPlan, diags := userFromState(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, plan.ClusterID.ValueString(), userPlan, updatePaths)
	if resp.Diagnostics.HasError() {
		return
	}

	userRead(ctx, r.providerConfig.SDK, &resp.Diagnostics, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	name := state.Name.ValueString()
	if name == defaultName {
		return
	}
	deleteUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, name)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, userName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}

	user := readUser(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	var state User
	resp.Diagnostics.Append(userToState(ctx, user, &state)...)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
