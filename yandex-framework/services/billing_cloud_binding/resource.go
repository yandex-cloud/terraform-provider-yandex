package billing_cloud_binding

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/encoding/protojson"
)

type bindingResource struct {
	providerConfig             *provider_config.Config
	serviceInstanceType        string
	serviceInstanceIdFieldName string
}

func NewResource(bindingServiceInstanceType, bindingServiceInstanceIdFieldName string) resource.Resource {
	return &bindingResource{
		serviceInstanceType:        bindingServiceInstanceType,
		serviceInstanceIdFieldName: bindingServiceInstanceIdFieldName,
	}
}

func (r *bindingResource) Schema(ctx context.Context,
	req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bind cloud to billing account. Creating the bind, which connect the cloud to the billing account.\n For more information, see [the official documentation](https://yandex.cloud/docs/billing/operations/pin-cloud).\n\n**Note**: Currently resource deletion do not unbind cloud from billing account. Instead it does no-operations.\n\n",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["id"],
			},
			"billing_account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of billing account to bind cloud to.",
			},
			r.serviceInstanceIdFieldName: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service Instance ID.",
			},
		},
	}
}

func (r *bindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_cloud_binding"
}

func (r *bindingResource) bindAccountToServiceInstance(ctx context.Context, billingAccountId,
	serviceInstanceId string, diagnostics *diag.Diagnostics) (resourceID string) {
	billableObject := billing.BillableObject{
		Type: r.serviceInstanceType,
		Id:   serviceInstanceId,
	}
	bindRequest := billing.BindBillableObjectRequest{
		BillingAccountId: billingAccountId,
		BillableObject:   &billableObject,
	}

	op, err := r.providerConfig.SDK.Billing().BillingAccount().BindBillableObject(
		ctx,
		&bindRequest,
	)

	if opErr := op.GetError(); opErr != nil {
		log.Printf("[WARN] Operation ended with error: %s", protojson.Format(opErr))
		diagnostics.AddError("Failed to bind billing object", fmt.Sprintf("%v [%v]", opErr.Message, opErr.Code))
		return
	}

	if err != nil {
		diagnostics.AddError(fmt.Sprintf("Error while requesting API binding %s to billing account", r.serviceInstanceType), err.Error())
		return
	}

	id := InstanceID{
		BillingAccountId:    billingAccountId,
		ServiceInstanceType: r.serviceInstanceType,
		ServiceInstanceId:   serviceInstanceId,
	}

	return id.compute()
}

func (r *bindingResource) Create(ctx context.Context,
	req resource.CreateRequest, resp *resource.CreateResponse) {
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/create
	var state yandexBillingBindingState
	getAllRequestAttributes(ctx, &state, r.serviceInstanceIdFieldName, req.Plan, &resp.Diagnostics)

	resourceID := r.bindAccountToServiceInstance(ctx, state.billingAccountID.ValueString(), state.serviceInstanceID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.id = types.StringValue(resourceID)
	setAllResponseAttributes(ctx, state, r.serviceInstanceIdFieldName, &resp.State, &resp.Diagnostics)
}

func (r *bindingResource) Read(ctx context.Context,
	req resource.ReadRequest, resp *resource.ReadResponse) {
	var state yandexBillingBindingState

	getAllRequestAttributes(ctx, &state, r.serviceInstanceIdFieldName, req.State, &resp.Diagnostics)

	if !isObjectExist(ctx, r.providerConfig.SDK, r.serviceInstanceType, state.billingAccountID.ValueString(), state.serviceInstanceID.ValueString()) {
		resp.Diagnostics.AddError("Failed to read resource",
			fmt.Sprintf("Bound %s to billing account not found", r.serviceInstanceType))
		resp.State.RemoveResource(ctx)
		return
	}

	setAllResponseAttributes(ctx, state, r.serviceInstanceIdFieldName, &resp.State, &resp.Diagnostics)
}

func (r *bindingResource) Update(ctx context.Context,
	req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state yandexBillingBindingState

	getAllRequestAttributes(ctx, &plan, r.serviceInstanceIdFieldName, req.Plan, &resp.Diagnostics)
	getAllRequestAttributes(ctx, &state, r.serviceInstanceIdFieldName, req.State, &resp.Diagnostics)

	plan.id = types.StringValue(r.bindAccountToServiceInstance(ctx, plan.billingAccountID.ValueString(),
		plan.serviceInstanceID.ValueString(), &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}

	// Update should set only changed attributes
	if plan.id != state.id {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(idFieldName), plan.id)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if plan.billingAccountID != state.billingAccountID {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(accountIDFieldName), plan.billingAccountID)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if plan.serviceInstanceID != state.serviceInstanceID {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.serviceInstanceIdFieldName), plan.serviceInstanceID)...)
	}
}

func (r *bindingResource) Delete(ctx context.Context,
	req resource.DeleteRequest, resp *resource.DeleteResponse) {
	log.Printf("[INFO] The resource of binding to billign account is deleted " +
		"however the binding itself still exists. " +
		"This is an excepted behaviour. See documentation for details.")
	// The Delete method will automatically call resp.State.RemoveResource() if there are no errors
}

func (r *bindingResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parsedID, err := ParseBindServiceInstanceId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: "+
				"'BillingAccountId/ServiceInstanceType/ObjectID'. Got: %q", req.ID),
		)
		return
	}
	state := yandexBillingBindingState{
		id:                types.StringValue(req.ID),
		billingAccountID:  types.StringValue(parsedID.BillingAccountId),
		serviceInstanceID: types.StringValue(parsedID.ServiceInstanceId),
	}

	setAllResponseAttributes(ctx, state, r.serviceInstanceIdFieldName, &resp.State, &resp.Diagnostics)
}

func (r *bindingResource) Configure(ctx context.Context,
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
