package gitlab_instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var _ datasource.DataSourceWithConfigure = (*gitlabInstanceDatasource)(nil)

type gitlabInstanceDatasource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &gitlabInstanceDatasource{}
}

func (d *gitlabInstanceDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_instance"
}

func (d *gitlabInstanceDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gitlabInstanceDatasource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_email": schema.StringAttribute{
				Computed:            true,
				Description:         "An email of admin user in Gitlab.",
				MarkdownDescription: "An email of admin user in Gitlab.",
			},
			"admin_login": schema.StringAttribute{
				Computed:            true,
				Description:         "A login of admin user in Gitlab.",
				MarkdownDescription: "A login of admin user in Gitlab.",
			},
			"approval_rules_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Approval rules configuration. One of: NONE, BASIC, STANDARD, ADVANCED.",
				MarkdownDescription: "Approval rules configuration. One of: NONE, BASIC, STANDARD, ADVANCED.",
			},
			"approval_rules_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "Approval rules token.",
				MarkdownDescription: "Approval rules token.",
			},
			"backup_retain_period_days": schema.Int64Attribute{
				Computed:            true,
				Description:         "Auto backups retain period in days.",
				MarkdownDescription: "Auto backups retain period in days.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         common.ResourceDescriptions["created_at"],
				MarkdownDescription: common.ResourceDescriptions["created_at"],
			},
			"deletion_protection": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         common.ResourceDescriptions["deletion_protection"],
				MarkdownDescription: common.ResourceDescriptions["deletion_protection"],
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         common.ResourceDescriptions["description"],
				MarkdownDescription: common.ResourceDescriptions["description"],
			},
			"disk_size": schema.Int64Attribute{
				Computed:            true,
				Description:         "Amount of disk storage available to a instance in GB.",
				MarkdownDescription: "Amount of disk storage available to a instance in GB.",
			},
			"domain": schema.StringAttribute{
				Computed:            true,
				Description:         "Domain of the Gitlab instance.",
				MarkdownDescription: "Domain of the Gitlab instance.",
			},
			"folder_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         common.ResourceDescriptions["folder_id"],
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
			},
			"gitlab_version": schema.StringAttribute{
				Computed:            true,
				Description:         "Version of Gitlab on instance.",
				MarkdownDescription: "Version of Gitlab on instance.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         common.ResourceDescriptions["id"],
				MarkdownDescription: common.ResourceDescriptions["id"],
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         common.ResourceDescriptions["labels"],
				MarkdownDescription: common.ResourceDescriptions["labels"],
			},
			"maintenance_delete_untagged": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The `true` value means that untagged images will be deleted during maintenance.",
				MarkdownDescription: "The `true` value means that untagged images will be deleted during maintenance.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         common.ResourceDescriptions["name"],
				MarkdownDescription: common.ResourceDescriptions["name"],
			},
			"resource_preset_id": schema.StringAttribute{
				Computed:            true,
				Description:         "ID of the preset for computational resources available to the instance (CPU, memory etc.). One of: s2.micro, s2.small, s2.medium, s2.large.",
				MarkdownDescription: "ID of the preset for computational resources available to the instance (CPU, memory etc.). One of: s2.micro, s2.small, s2.medium, s2.large.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Status of the instance.",
				MarkdownDescription: "Status of the instance.",
			},
			"subnet_id": schema.StringAttribute{
				Computed:            true,
				Description:         "ID of the subnet where the GitLab instance is located.",
				MarkdownDescription: "ID of the subnet where the GitLab instance is located.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the instance was updated.",
				MarkdownDescription: "The timestamp when the instance was updated.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
		Description: "Managed Gitlab instance.",
	}
}

func (d *gitlabInstanceDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := d.providerConfig.SDK.Gitlab().Instance().Get(
		ctx,
		&gitlab.GetInstanceRequest{
			InstanceId: state.Id.ValueString(),
		},
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

	state.Id = types.StringValue(instance.Id)
	updateState(ctx, d.providerConfig.SDK, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
