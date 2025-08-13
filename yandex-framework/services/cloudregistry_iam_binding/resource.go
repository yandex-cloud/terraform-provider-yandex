package cloudregistry_iam_binding

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	yandexCloudRegistryIAMBindingDefaultTimeout = 1 * time.Minute
	defaultListSize                             = 1000
)

type CloudRegistryIAMBindingResource struct {
	config *provider_config.Config
}

func NewResource() resource.Resource {
	return &CloudRegistryIAMBindingResource{}
}

func (r *CloudRegistryIAMBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "yandex_cloud_registry_iam_binding"
}

func (r *CloudRegistryIAMBindingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.config = config
}

func (r *CloudRegistryIAMBindingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IAM binding for a Yandex Cloud Registry.",

		Attributes: map[string]schema.Attribute{
			"registry_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the registry to apply the binding to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role to assign to the members.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Identities (users, service accounts, groups) to bind the role to.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Computed identifier of the IAM binding.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type cloudRegistryIAMBindingModel struct {
	RegistryID types.String `tfsdk:"registry_id"`
	Role       types.String `tfsdk:"role"`
	Members    types.Set    `tfsdk:"members"`
	ID         types.String `tfsdk:"id"`
}

func (r *CloudRegistryIAMBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudRegistryIAMBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bindings, err := getCloudRegistryAccessBindings(ctx, r.config, plan.RegistryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IAM bindings",
			fmt.Sprintf("Could not read access bindings for registry %s: %s",
				plan.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	var members []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, member := range members {
		parts := strings.Split(member, ":")
		if len(parts) < 2 {
			resp.Diagnostics.AddError(
				"Invalid member format",
				fmt.Sprintf("Member %s must be in format 'subjectType:subjectId'", member),
			)
			return
		}

		bindings = append(bindings, &access.AccessBinding{
			RoleId: plan.Role.ValueString(),
			Subject: &access.Subject{
				Type: parts[0],
				Id:   strings.Join(parts[1:], ":"),
			},
		})
	}

	if err := setCloudRegistryAccessBindings(ctx, r.config, plan.RegistryID.ValueString(), bindings); err != nil {
		resp.Diagnostics.AddError(
			"Error setting IAM bindings",
			fmt.Sprintf("Could not set access bindings for registry %s: %s",
				plan.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.RegistryID.ValueString(), plan.Role.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CloudRegistryIAMBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudRegistryIAMBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bindings, err := getCloudRegistryAccessBindings(ctx, r.config, state.RegistryID.ValueString())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading IAM bindings",
			fmt.Sprintf("Could not read access bindings for registry %s: %s",
				state.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	var members []string
	role := state.Role.ValueString()
	for _, binding := range bindings {
		if binding.RoleId == role {
			members = append(members, fmt.Sprintf("%s:%s", binding.Subject.Type, binding.Subject.Id))
		}
	}

	if len(members) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	membersSet, diags := types.SetValueFrom(ctx, types.StringType, members)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Members = membersSet
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CloudRegistryIAMBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cloudRegistryIAMBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bindings, err := getCloudRegistryAccessBindings(ctx, r.config, plan.RegistryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IAM bindings",
			fmt.Sprintf("Could not read access bindings for registry %s: %s",
				plan.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	role := state.Role.ValueString()
	var newBindings []*access.AccessBinding
	for _, binding := range bindings {
		if binding.RoleId != role {
			newBindings = append(newBindings, binding)
		}
	}

	var members []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, member := range members {
		parts := strings.Split(member, ":")
		if len(parts) < 2 {
			resp.Diagnostics.AddError(
				"Invalid member format",
				fmt.Sprintf("Member %s must be in format 'subjectType:subjectId'", member),
			)
			return
		}

		newBindings = append(newBindings, &access.AccessBinding{
			RoleId: plan.Role.ValueString(),
			Subject: &access.Subject{
				Type: parts[0],
				Id:   strings.Join(parts[1:], ":"),
			},
		})
	}

	if err := setCloudRegistryAccessBindings(ctx, r.config, plan.RegistryID.ValueString(), newBindings); err != nil {
		resp.Diagnostics.AddError(
			"Error setting IAM bindings",
			fmt.Sprintf("Could not set access bindings for registry %s: %s",
				plan.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.RegistryID.ValueString(), plan.Role.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CloudRegistryIAMBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudRegistryIAMBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bindings, err := getCloudRegistryAccessBindings(ctx, r.config, state.RegistryID.ValueString())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return
		}

		resp.Diagnostics.AddError(
			"Error reading IAM bindings",
			fmt.Sprintf("Could not read access bindings for registry %s: %s",
				state.RegistryID.ValueString(), err.Error()),
		)
		return
	}

	role := state.Role.ValueString()
	var newBindings []*access.AccessBinding
	for _, binding := range bindings {
		if binding.RoleId != role {
			newBindings = append(newBindings, binding)
		}
	}

	if len(bindings) == len(newBindings) {
		log.Printf("[DEBUG] Binding for role %s not found in registry %s, assuming already deleted",
			role, state.RegistryID.ValueString())
		return
	}

	if err := setCloudRegistryAccessBindings(ctx, r.config, state.RegistryID.ValueString(), newBindings); err != nil {
		resp.Diagnostics.AddError(
			"Error setting IAM bindings",
			fmt.Sprintf("Could not set access bindings for registry %s: %s",
				state.RegistryID.ValueString(), err.Error()),
		)
	}
}

func (r *CloudRegistryIAMBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, " ")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format <registry_id> <role>",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("registry_id"),
		parts[0],
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("role"),
		parts[1],
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%s:%s", parts[0], parts[1]),
	)...)
}

func getCloudRegistryAccessBindings(ctx context.Context, config *provider_config.Config, registryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.SDK.CloudRegistry().Registry().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: registryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings: %w", err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}

func setCloudRegistryAccessBindings(ctx context.Context, config *provider_config.Config, registryID string, bindings []*access.AccessBinding) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     registryID,
		AccessBindings: bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexCloudRegistryIAMBindingDefaultTimeout)
	defer cancel()

	op, err := config.SDK.WrapOperation(config.SDK.CloudRegistry().Registry().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting access bindings: %w", err)
	}

	return op.Wait(ctx)
}
