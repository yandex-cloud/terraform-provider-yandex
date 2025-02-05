package datasphere_project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

type projectDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasphere_project"
}

func (d *projectDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Info(ctx, "Initializing project resource schema")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Required: true},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			"community_id": schema.StringAttribute{Computed: true},
			"name":         schema.StringAttribute{Computed: true},
			"description":  schema.StringAttribute{Computed: true},
			"labels": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"settings": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"service_account_id":   schema.StringAttribute{Computed: true},
					"subnet_id":            schema.StringAttribute{Computed: true},
					"data_proc_cluster_id": schema.StringAttribute{Computed: true},
					"security_group_ids": schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"default_folder_id":       schema.StringAttribute{Computed: true},
					"stale_exec_timeout_mode": schema.StringAttribute{Computed: true},
				},
				Computed: true,
			},
			"limits": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"max_units_per_hour":      schema.Int64Attribute{Computed: true},
					"max_units_per_execution": schema.Int64Attribute{Computed: true},
					"balance":                 schema.Int64Attribute{Computed: true},
				},
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading project data source")
	var projectModel projectDataModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &projectModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx,
		fmt.Sprintf("Making API call to fetch project data for project %s", projectModel.Id.ValueString()),
	)

	existingProject, err := d.providerConfig.SDK.Datasphere().Project().Get(
		ctx,
		&datasphere.GetProjectRequest{ProjectId: projectModel.Id.ValueString()},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch DataSource",
			fmt.Sprintf("An unexpected error occurred while fetching the datasource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	existingUnitBalance, err := d.providerConfig.SDK.Datasphere().Project().GetUnitBalance(
		ctx,
		&datasphere.GetUnitBalanceRequest{ProjectId: projectModel.Id.ValueString()},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch DataSource",
			fmt.Sprintf("An unexpected error occurred while fetching the datasource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}

	convertToTerraformModel(ctx, &projectModel, existingProject, &resp.Diagnostics, existingUnitBalance.UnitBalance)
	resp.Diagnostics.Append(resp.State.Set(ctx, &projectModel)...)
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
