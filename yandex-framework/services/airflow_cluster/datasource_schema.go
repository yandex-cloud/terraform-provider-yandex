package airflow_cluster

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ClusterDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Password that is used to log in to Apache Airflow web UI under `admin` user.",
			},
			"airflow_config": schema.MapAttribute{
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
				Computed:            true,
				MarkdownDescription: "Configuration of the Apache Airflow application itself. The value of this attribute is a two-level map. Keys of top-level map are the names of [configuration sections](https://airflow.apache.org/docs/apache-airflow/stable/configurations-ref.html#airflow-configuration-options). Keys of inner maps are the names of configuration options within corresponding section.",
				Validators: []validator.Map{
					airflowConfigValidator(),
				},
			},
			"code_sync": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"s3": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"bucket": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The name of the Object Storage bucket that stores DAG files used in the cluster.",
							},
						},
						CustomType: S3Type{
							ObjectType: types.ObjectType{
								AttrTypes: S3Value{}.AttributeTypes(ctx),
							},
						},
						Computed:            true,
						MarkdownDescription: "Currently only Object Storage (S3) is supported as the source of DAG files.",
					},
				},
				CustomType: CodeSyncType{
					ObjectType: types.ObjectType{
						AttrTypes: CodeSyncValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Parameters of the location and access to the code that will be executed in the cluster.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the resource.",
			},
			"deb_packages": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "System packages that are installed in the cluster.",
			},
			"deletion_protection": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "The `true` value means that resource is protected from accidental deletion.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource description.",
			},
			"folder_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.",
			},
			"health": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The resource identifier. Exactly one of the attributes `id` or `name` should be specified.",
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.Expressions{path.MatchRoot("name")}...),
				},
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "A set of key/value label pairs which assigned to resource.",
			},
			"lockbox_secrets_backend": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Enables usage of Lockbox Secrets Backend.",
					},
				},
				CustomType: LockboxSecretsBackendType{
					ObjectType: types.ObjectType{
						AttrTypes: LockboxSecretsBackendValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Configuration of Lockbox Secrets Backend. [See documentation](https://yandex.cloud/docs/managed-airflow/tutorials/lockbox-secrets-in-maf-cluster) for details.",
			},
			"logging": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Enables delivery of logs generated by the Airflow components to [Cloud Logging](https://yandex.cloud/docs/logging/).",
					},
					"folder_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Logs are written to **default log group** of specified folder.",
					},
					"log_group_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Logs are written to the **specified log group**.",
					},
					"min_level": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Minimum level of messages that are sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`.",
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
				MarkdownDescription: "Cloud Logging configuration.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The resource name. Exactly one of the attributes `id` or `name` should be specified.",
			},
			"pip_packages": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Python packages that are installed in the cluster.",
			},
			"scheduler": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The number of scheduler instances in the cluster.",
					},
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
					},
				},
				CustomType: SchedulerType{
					ObjectType: types.ObjectType{
						AttrTypes: SchedulerValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Configuration of scheduler instances.",
			},
			"security_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "The list of security groups applied to resource or their components.",
			},
			"service_account_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "[Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) which linked to the resource. For more information, see [documentation](https://yandex.cloud/docs/managed-airflow/concepts/impersonation).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).",
			},
			"subnet_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "The list of VPC subnets identifiers which resource is attached.",
			},
			"triggerer": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The number of triggerer instances in the cluster.",
					},
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
					},
				},
				CustomType: TriggererType{
					ObjectType: types.ObjectType{
						AttrTypes: TriggererValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Configuration of `triggerer` instances.",
			},
			"webserver": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The number of webserver instances in the cluster.",
					},
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
					},
				},
				CustomType: WebserverType{
					ObjectType: types.ObjectType{
						AttrTypes: WebserverValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Configuration of `webserver` instances.",
			},
			"worker": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"max_count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The maximum number of worker instances in the cluster.",
					},
					"min_count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The minimum number of worker instances in the cluster.",
					},
					"resource_preset_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The identifier of the preset for computational resources available to an instance (CPU, memory etc.).",
					},
				},
				CustomType: WorkerType{
					ObjectType: types.ObjectType{
						AttrTypes: WorkerValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				MarkdownDescription: "Configuration of worker instances.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read: true,
			}),
		},
	}
}
