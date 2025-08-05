package spark_cluster

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func ClusterDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"dependencies": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"deb_packages": schema.SetAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								Description:         "Deb-packages that need to be installed using system package manager.",
								MarkdownDescription: "Deb-packages that need to be installed using system package manager.",
							},
							"pip_packages": schema.SetAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								Description:         "Python packages that need to be installed using pip (in pip requirement format).",
								MarkdownDescription: "Python packages that need to be installed using pip (in pip requirement format).",
							},
						},
						CustomType: DependenciesType{
							ObjectType: types.ObjectType{
								AttrTypes: DependenciesValue{}.AttributeTypes(ctx),
							},
						},
						Computed:            true,
						Description:         "Environment dependencies.",
						MarkdownDescription: "Environment dependencies.",
					},
					"history_server": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Computed:            true,
								Description:         "Enable Spark History Server.",
								MarkdownDescription: "Enable Spark History Server.",
							},
						},
						CustomType: HistoryServerType{
							ObjectType: types.ObjectType{
								AttrTypes: HistoryServerValue{}.AttributeTypes(ctx),
							},
						},
						Computed:            true,
						Description:         "History Server configuration.",
						MarkdownDescription: "History Server configuration.",
					},
					"metastore": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"cluster_id": schema.StringAttribute{
								Computed:            true,
								Description:         "Metastore cluster ID for default spark configuration.",
								MarkdownDescription: "Metastore cluster ID for default spark configuration.",
							},
						},
						CustomType: MetastoreType{
							ObjectType: types.ObjectType{
								AttrTypes: MetastoreValue{}.AttributeTypes(ctx),
							},
						},
						Computed:            true,
						Description:         "Metastore configuration.",
						MarkdownDescription: "Metastore configuration.",
					},
					"resource_pools": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"driver": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"max_size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Maximum node count for the driver pool with autoscaling.",
										MarkdownDescription: "Maximum node count for the driver pool with autoscaling.",
									},
									"min_size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Minimum node count for the driver pool with autoscaling.",
										MarkdownDescription: "Minimum node count for the driver pool with autoscaling.",
									},
									"resource_preset_id": schema.StringAttribute{
										Computed:            true,
										Description:         "Resource preset ID for the driver pool.",
										MarkdownDescription: "Resource preset ID for the driver pool.",
									},
									"size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Node count for the driver pool with fixed size.",
										MarkdownDescription: "Node count for the driver pool with fixed size.",
									},
								},
								CustomType: DriverType{
									ObjectType: types.ObjectType{
										AttrTypes: DriverValue{}.AttributeTypes(ctx),
									},
								},
								Computed:            true,
								Description:         "Computational resources for the driver pool.",
								MarkdownDescription: "Computational resources for the driver pool.",
							},
							"executor": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"max_size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Maximum node count for the executor pool with autoscaling.",
										MarkdownDescription: "Maximum node count for the executor pool with autoscaling.",
									},
									"min_size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Minimum node count for the executor pool with autoscaling.",
										MarkdownDescription: "Minimum node count for the executor pool with autoscaling.",
									},
									"resource_preset_id": schema.StringAttribute{
										Computed:            true,
										Description:         "Resource preset ID for the executor pool.",
										MarkdownDescription: "Resource preset ID for the executor pool.",
									},
									"size": schema.Int64Attribute{
										Computed:            true,
										Description:         "Node count for the executor pool with fixed size.",
										MarkdownDescription: "Node count for the executor pool with fixed size.",
									},
								},
								CustomType: ExecutorType{
									ObjectType: types.ObjectType{
										AttrTypes: ExecutorValue{}.AttributeTypes(ctx),
									},
								},
								Computed:            true,
								Description:         "Computational resources for the executor pool.",
								MarkdownDescription: "Computational resources for the executor pool.",
							},
						},
						CustomType: ResourcePoolsType{
							ObjectType: types.ObjectType{
								AttrTypes: ResourcePoolsValue{}.AttributeTypes(ctx),
							},
						},
						Computed:            true,
						Description:         "Computational resources.",
						MarkdownDescription: "Computational resources.",
					},
				},
				CustomType: ConfigType{
					ObjectType: types.ObjectType{
						AttrTypes: ConfigValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Configuration of the Spark cluster.",
				MarkdownDescription: "Configuration of the Spark cluster.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the cluster was created.",
				MarkdownDescription: "The timestamp when the cluster was created.",
			},
			"deletion_protection": schema.BoolAttribute{
				Computed:            true,
				Description:         "The `true` value means that resource is protected from accidental deletion.",
				MarkdownDescription: "The `true` value means that resource is protected from accidental deletion.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "Description of the cluster. 0-256 characters long.",
				MarkdownDescription: "Description of the cluster. 0-256 characters long.",
			},
			"folder_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "ID of the cloud folder that the cluster belongs to.",
				MarkdownDescription: "ID of the cloud folder that the cluster belongs to.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Unique ID of the cluster.",
				MarkdownDescription: "Unique ID of the cluster.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "Cluster labels as key/value pairs.",
				MarkdownDescription: "Cluster labels as key/value pairs.",
			},
			"logging": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						Description:         "Enable log delivery to Cloud Logging.",
						MarkdownDescription: "Enable log delivery to Cloud Logging.",
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
				Description:         "Configuration of the window for maintenance operations.",
				MarkdownDescription: "Configuration of the window for maintenance operations.",
				Validators: []validator.Object{
					mwValidator(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Name of the cluster. The name is unique within the folder.",
				MarkdownDescription: "Name of the cluster. The name is unique within the folder.",
			},
			"network": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"security_group_ids": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "Network security groups.",
						MarkdownDescription: "Network security groups.",
					},
					"subnet_ids": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "Network subnets.",
						MarkdownDescription: "Network subnets.",
					},
				},
				CustomType: NetworkType{
					ObjectType: types.ObjectType{
						AttrTypes: NetworkValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Network configuration.",
				MarkdownDescription: "Network configuration.",
			},
			"service_account_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The service account used by the cluster to access cloud resources.",
				MarkdownDescription: "The service account used by the cluster to access cloud resources.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Status of the cluster.",
				MarkdownDescription: "Status of the cluster.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
		Description: "Managed Spark cluster.",
	}
}
