package api

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

func ClusterResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages an Apache Airflow cluster within Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-airflow/concepts/).",
		Attributes: map[string]schema.Attribute{
			"id":        defaultschema.Id(),
			"folder_id": defaultschema.FolderId(),
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"],
				Required:            true,
			},
			"description":         defaultschema.Description(),
			"labels":              defaultschema.Labels(),
			"created_at":          defaultschema.CreatedAt(),
			"deletion_protection": defaultschema.DeletionProtection(),
			"service_account_id":  defaultschema.ServiceAccountId(),
			"subnet_ids":          defaultschema.SubnetIds(),
			"security_group_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			/*
				"service_account_id": schema.StringAttribute{
					Required: true,
				},
			*/
			// Add to Markdown description:
			// For more information, see [documentation](https://yandex.cloud/docs/managed-airflow/concepts/impersonation).
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "Password that is used to log in to Apache Airflow web UI under `admin` user.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					adminPasswordModifier{},
				},
			},
			"airflow_config": schema.MapAttribute{
				MarkdownDescription: "Configuration of the Apache Airflow application itself. The value of this attribute is a two-level map. Keys of top-level map are the names of [configuration sections](https://airflow.apache.org/docs/apache-airflow/stable/configurations-ref.html#airflow-configuration-options). Keys of inner maps are the names of configuration options within corresponding section.",
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				Validators: []validator.Map{
					airflowConfigValidator(),
				},
			},
			"code_sync": schema.SingleNestedAttribute{
				MarkdownDescription: "Parameters of the location and access to the code that will be executed in the cluster.",
				Attributes: map[string]schema.Attribute{
					"s3": schema.SingleNestedAttribute{
						MarkdownDescription: "Currently only Object Storage (S3) is supported as the source of DAG files.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"bucket": schema.StringAttribute{
								MarkdownDescription: "The name of the Object Storage bucket that stores DAG files used in the cluster.",
								Required:            true,
							},
						},
						CustomType: S3Type{
							ObjectType: types.ObjectType{
								AttrTypes: S3Value{}.AttributeTypes(ctx),
							},
						},
					},
				},
				CustomType: CodeSyncType{
					ObjectType: types.ObjectType{
						AttrTypes: CodeSyncValue{}.AttributeTypes(ctx),
					},
				},
				Required: true,
			},
			"deb_packages": schema.SetAttribute{
				MarkdownDescription: "System packages that are installed in the cluster.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"health": schema.StringAttribute{
				MarkdownDescription: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lockbox_secrets_backend": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of Lockbox Secrets Backend. [See documentation](https://yandex.cloud/docs/managed-airflow/tutorials/lockbox-secrets-in-maf-cluster) for details.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enables usage of Lockbox Secrets Backend.",
						Required:            true,
					},
				},
				CustomType: LockboxSecretsBackendType{
					ObjectType: types.ObjectType{
						AttrTypes: LockboxSecretsBackendValue{}.AttributeTypes(ctx),
					},
				},
				Optional: true,
			},
			"logging": schema.SingleNestedAttribute{
				MarkdownDescription: "Cloud Logging configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enables delivery of logs generated by the Airflow components to [Cloud Logging](https://yandex.cloud/docs/logging/).",
						Required:            true,
					},
					"folder_id": schema.StringAttribute{
						MarkdownDescription: "Logs will be written to **default log group** of specified folder. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
						Optional:            true,
					},
					"log_group_id": schema.StringAttribute{
						MarkdownDescription: "Logs will be written to the **specified log group**. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.",
						Optional:            true,
					},
					"min_level": schema.StringAttribute{
						MarkdownDescription: "Minimum level of messages that will be sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. If not set then server default is applied (currently `INFO`).",
						Optional:            true,
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
				Optional: true,
			},
			"pip_packages": schema.SetAttribute{
				MarkdownDescription: "Python packages that are installed in the cluster.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"scheduler": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of scheduler instances.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						MarkdownDescription: "The number of scheduler instances in the cluster.",
						Required:            true,
					},
					"resource_preset_id": schema.StringAttribute{
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
						Required:            true,
					},
				},
				CustomType: SchedulerType{
					ObjectType: types.ObjectType{
						AttrTypes: SchedulerValue{}.AttributeTypes(ctx),
					},
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"triggerer": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of `triggerer` instances.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						MarkdownDescription: "The number of triggerer instances in the cluster.",
						Required:            true,
					},
					"resource_preset_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the preset for computational resources available to an instance (CPU, memory etc.).",
						Required:            true,
					},
				},
				CustomType: TriggererType{
					ObjectType: types.ObjectType{
						AttrTypes: TriggererValue{}.AttributeTypes(ctx),
					},
				},
			},
			"webserver": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of `webserver` instances.",
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						MarkdownDescription: "The number of webserver instances in the cluster.",
						Required:            true,
					},
					"resource_preset_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the preset for computational resources available to an instance (CPU, memory etc.).",
						Required:            true,
					},
				},
				CustomType: WebserverType{
					ObjectType: types.ObjectType{
						AttrTypes: WebserverValue{}.AttributeTypes(ctx),
					},
				},
				Required: true,
			},
			"worker": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration of worker instances.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"max_count": schema.Int64Attribute{
						MarkdownDescription: "The maximum number of worker instances in the cluster.",
						Required:            true,
					},
					"min_count": schema.Int64Attribute{
						MarkdownDescription: "The minimum number of worker instances in the cluster.",
						Required:            true,
					},
					"resource_preset_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the preset for computational resources available to an instance (CPU, memory etc.).",
						Required:            true,
					},
				},
				CustomType: WorkerType{
					ObjectType: types.ObjectType{
						AttrTypes: WorkerValue{}.AttributeTypes(ctx),
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
	}
}
