package billing_cloud_binding

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type bindingDataSource struct {
	providerConfig             *provider_config.Config
	serviceInstanceType        string
	serviceInstanceIdFieldName string
}

func NewDataSource(bindingServiceInstanceType, bindingServiceInstanceIdFieldName string) datasource.DataSource {
	return &bindingDataSource{
		serviceInstanceType:        bindingServiceInstanceType,
		serviceInstanceIdFieldName: bindingServiceInstanceIdFieldName,
	}
}

func (d *bindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier.",
			},
			"billing_account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of billing account to bind cloud to.",
			},
			d.serviceInstanceIdFieldName: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service Instance ID.",
			},
		},
		/*
			Attributes: map[string]schema.Attribute{
				idFieldName: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Service generated identifier for the thing.",
				},
				accountIDFieldName: schema.StringAttribute{
					Required: true,
				},
				d.serviceInstanceIdFieldName: schema.StringAttribute{
					Required: true,
				},
			},
		*/
	}
}

func (d *bindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_cloud_binding"
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state yandexBillingBindingState
	getAllRequestAttributes(ctx, &state, d.serviceInstanceIdFieldName, req.Config, &resp.Diagnostics)

	if !isObjectExist(ctx, d.providerConfig.SDK, d.serviceInstanceType, state.billingAccountID.ValueString(), state.serviceInstanceID.ValueString()) {
		resp.Diagnostics.AddError("Failed to read datasource",
			fmt.Sprintf("Bound %s to billing account not found", d.serviceInstanceType))
		return
	}

	setAllResponseAttributes(ctx, state, d.serviceInstanceIdFieldName, &resp.State, &resp.Diagnostics)
}

func (d *bindingDataSource) Configure(ctx context.Context,
	req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}
