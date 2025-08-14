package metastore_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"

	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ datasource.DataSource              = &metastoreClusterDatasource{}
	_ datasource.DataSourceWithConfigure = &metastoreClusterDatasource{}
)

func NewDatasource() datasource.DataSource {
	return &metastoreClusterDatasource{}
}

type metastoreClusterDatasource struct {
	providerConfig *provider_config.Config
}

func (d *metastoreClusterDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metastore_cluster"
}

func (d *metastoreClusterDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	if id == "" {
		folderID, diags := validate.FolderID(state.FolderId, &d.providerConfig.ProviderState)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		id, diags = objectid.ResolveByNameAndFolderID(ctx, d.providerConfig.SDK, folderID, state.Name.ValueString(), sdkresolvers.MetastoreClusterResolver)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Id = types.StringValue(id)
	}

	refreshState(ctx, d.providerConfig.SDK, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *metastoreClusterDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.providerConfig = providerConfig
}

func (d *metastoreClusterDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_config": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
					},
				},
				CustomType: ClusterConfigType{
					ObjectType: types.ObjectType{
						AttrTypes: ClusterConfigValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Hive Metastore cluster configuration.",
				MarkdownDescription: "Hive Metastore cluster configuration.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The creation timestamp of the resource.",
				MarkdownDescription: "The creation timestamp of the resource.",
			},
			"deletion_protection": schema.BoolAttribute{
				Computed:            true,
				Description:         "The `true` value means that resource is protected from accidental deletion. By default is set to `false`.",
				MarkdownDescription: "The `true` value means that resource is protected from accidental deletion. By default is set to `false`.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The resource description.",
				MarkdownDescription: "The resource description.",
			},
			"endpoint_ip": schema.StringAttribute{
				Computed:            true,
				Description:         "IP address of Metastore server balancer endpoint.",
				MarkdownDescription: "IP address of Metastore server balancer endpoint.",
			},
			"folder_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.",
				MarkdownDescription: "The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.",
			},
			"health": schema.StringAttribute{
				Computed:            true,
				Description:         "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.",
				MarkdownDescription: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The resource identifier.",
				MarkdownDescription: "The resource identifier.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A set of key/value label pairs which assigned to resource.",
				MarkdownDescription: "A set of key/value label pairs which assigned to resource.",
			},
			"logging": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						Description:         "Enables delivery of logs generated by Metastore to [Cloud Logging](https://yandex.cloud/docs/logging/).",
						MarkdownDescription: "Enables delivery of logs generated by Metastore to [Cloud Logging](https://yandex.cloud/docs/logging/).",
					},
					"folder_id": schema.StringAttribute{
						Computed:            true,
						Description:         "Logs will be written to **default log group** of specified folder. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
						MarkdownDescription: "Logs will be written to **default log group** of specified folder. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
					},
					"log_group_id": schema.StringAttribute{
						Computed:            true,
						Description:         "Logs will be written to the **specified log group**. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
						MarkdownDescription: "Logs will be written to the **specified log group**. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
					},
					"min_level": schema.StringAttribute{
						Computed:            true,
						Description:         "Minimum level of messages that will be sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. If not set then server default is applied (currently `INFO`).",
						MarkdownDescription: "Minimum level of messages that will be sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. If not set then server default is applied (currently `INFO`).",
						Validators: []validator.String{
							logLevelValidator(),
						},
					},
				},
				CustomType: LoggingType{
					ObjectType: types.ObjectType{
						AttrTypes: LoggingValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Cloud Logging configuration.",
				MarkdownDescription: "Cloud Logging configuration.",
			},
			"maintenance_window": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"day": schema.StringAttribute{
						Computed:            true,
						Description:         "Day of week for maintenance window. One of `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
						MarkdownDescription: "Day of week for maintenance window. One of `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
						Validators: []validator.String{
							mwDayValidator(),
						},
					},
					"hour": schema.Int64Attribute{
						Computed:            true,
						Description:         "Hour of day in UTC time zone (1-24) for maintenance window.",
						MarkdownDescription: "Hour of day in UTC time zone (1-24) for maintenance window.",
						Validators: []validator.Int64{
							mwHourValidator(),
						},
					},
					"type": schema.StringAttribute{
						Computed:            true,
						Description:         "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. If `WEEKLY`, day and hour must be specified.",
						MarkdownDescription: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. If `WEEKLY`, day and hour must be specified.",
						Validators: []validator.String{
							mwTypeValidator(),
						},
					},
				},
				CustomType: MaintenanceWindowType{
					ObjectType: types.ObjectType{
						AttrTypes: MaintenanceWindowValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Configuration of window for maintenance operations.",
				MarkdownDescription: "Configuration of window for maintenance operations.",
				Validators: []validator.Object{
					mwValidator(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The resource name.",
				MarkdownDescription: "The resource name.",
			},
			"network_id": schema.StringAttribute{
				Computed:            true,
				Description:         "VPC network identifier which resource is attached.",
				MarkdownDescription: "VPC network identifier which resource is attached.",
			},
			"security_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The list of security groups applied to resource or their components.",
				MarkdownDescription: "The list of security groups applied to resource or their components.",
			},
			"service_account_id": schema.StringAttribute{
				Computed:            true,
				Description:         "[Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) with role `managed-metastore.integrationProvider`. For more information, see [documentation](https://yandex.cloud/docs/metadata-hub/concepts/metastore-impersonation).",
				MarkdownDescription: "[Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) with role `managed-metastore.integrationProvider`. For more information, see [documentation](https://yandex.cloud/docs/metadata-hub/concepts/metastore-impersonation).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.",
				MarkdownDescription: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.",
			},
			"subnet_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The list of VPC subnets identifiers which resource is attached.",
				MarkdownDescription: "The list of VPC subnets identifiers which resource is attached.",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Metastore server version.",
				MarkdownDescription: "Metastore server version.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read: true,
			}),
		},
		Description: "Managed Metastore cluster.",
	}
}
