package mdb_clickhouse_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/customplanmodifiers"
)

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ClickHouse cluster within the Yandex Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/). [How to connect to the DB](https://yandex.cloud/en/docs/managed-clickhouse/quickstart#connect). To connect, use port 9440. The port number is not configurable.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Description: common.ResourceDescriptions["id"],
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the ClickHouse cluster. Provided by the client when the cluster is created.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": defaultschema.FolderId(),
			"network_id": schema.StringAttribute{
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
			},
			"environment": schema.StringAttribute{
				Description: "Deployment environment of the ClickHouse cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels":              defaultschema.Labels(),
			"deletion_protection": defaultschema.DeletionProtection(),
			"disk_encryption_key_id": schema.StringAttribute{
				Description: "ID of the KMS key for cluster disk encryption.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at":         defaultschema.CreatedAt(),
			"security_group_ids": defaultschema.SecurityGroupIds(),
			"service_account_id": &schema.StringAttribute{
				Description: common.ResourceDescriptions["service_account_id"],
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the ClickHouse server software.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.NoneOf(""),
				},
			},
			"admin_password": schema.StringAttribute{
				Description: "A password used to authorize as user `admin` when `sql_user_management` enabled.",
				Optional:    true,
				Sensitive:   true,
			},
			"sql_user_management": schema.BoolAttribute{
				Description: "Enables `admin` user with user management permission.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"sql_database_management": schema.BoolAttribute{
				Description: "Grants `admin` user database management permission.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"embedded_keeper": schema.BoolAttribute{
				Description: "Whether to use ClickHouse Keeper as a coordination system.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"backup_retain_period_days": schema.Int64Attribute{
				Description: "The period in days during which backups are stored.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(7),
			},
			"copy_schema_on_new_hosts": schema.BoolAttribute{
				Description: "Whether to copy schema on new ClickHouse hosts.",
				Optional:    true,
			},
			"clickhouse":          ClickHouseSchema(),
			"zookeeper":           ZooKeeperSchema(),
			"cloud_storage":       CloudStorageSchema(),
			"backup_window_start": BackupWindowStart(),
			"access":              AccessSchema(),
			"hosts":               HostsSchema(),
			"shards":              ShardsSchema(),
		},
		Blocks: map[string]schema.Block{
			"shard_group":        ShardGroupSchema(),
			"format_schema":      FormatSchemaSchema(),
			"ml_model":           MlModelSchema(),
			"maintenance_window": MaintenanceWindowSchema(),
		},
	}
}

func HostsSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		Description: "A host configuration of the ClickHouse cluster.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"zone": schema.StringAttribute{
					Description: common.ResourceDescriptions["zone"],
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"type": schema.StringAttribute{
					Description: "The type of the host to be deployed. Can be either `CLICKHOUSE` or `ZOOKEEPER`.",
					Required:    true,
				},
				"subnet_id": schema.StringAttribute{
					Description: "ID of the subnet where the host is located.",
					Optional:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"assign_public_ip": schema.BoolAttribute{
					Description: "Whether the host should get a public IP address.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
				},
				"shard_name": schema.StringAttribute{
					Description: "The name of the shard to which the host belongs.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"fqdn": schema.StringAttribute{
					Description: "The fully qualified domain name of the host.",
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

func ZooKeeperSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Configuration of the ZooKeeper subcluster.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"resources": ResourcesSchema(),
		},
	}
}

func CloudStorageSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Cloud Storage settings.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
			customplanmodifiers.CloudStoragePlanModifier(),
		},
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.",
				Required:    true,
			},
			"move_factor": schema.NumberAttribute{
				Description: "Sets the minimum free space ratio in the cluster storage. If the free space is lower than this value, the data is transferred to Yandex Object Storage. Acceptable values are 0 to 1, inclusive.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Number{
					numberplanmodifier.UseStateForUnknown(),
				},
			},
			"data_cache_enabled": schema.BoolAttribute{
				Description: "Enables temporary storage in the cluster repository of data requested from the object repository.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"data_cache_max_size": schema.Int64Attribute{
				Description: "Defines the maximum amount of memory (in bytes) allocated in the cluster storage for temporary storage of data requested from the object storage.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"prefer_not_to_merge": schema.BoolAttribute{
				Description: "Disables merging of data parts in `Yandex Object Storage`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func BackupWindowStart() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Time to start the daily backup, in the UTC timezone.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"hours": schema.Int64Attribute{
				Description: "The hour at which backup will be started (UTC).",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 23),
				},
			},
			"minutes": schema.Int64Attribute{
				Description: "The minute at which backup will be started (UTC).",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 59),
				},
			},
		},
	}
}

func AccessSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Access policy to the ClickHouse cluster.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"web_sql": schema.BoolAttribute{
				Description: "Allow access for Web SQL.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"data_lens": schema.BoolAttribute{
				Description: "Allow access for DataLens.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"metrika": schema.BoolAttribute{
				Description: "Allow access for Yandex.Metrika.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"serverless": schema.BoolAttribute{
				Description: "Allow access for Serverless.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"data_transfer": schema.BoolAttribute{
				Description: "Allow access for DataTransfer.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"yandex_query": schema.BoolAttribute{
				Description: "Allow access for YandexQuery.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func ClickHouseSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Configuration of the ClickHouse subcluster.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"resources": ResourcesSchema(),
			"config":    ClickHouseConfigSchema(),
		},
	}
}

func ClickHouseConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Configuration of the ClickHouse subcluster.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"log_level": schema.StringAttribute{
				Description: "Logging level.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_connections": schema.Int64Attribute{
				Description: "Max server connections.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_concurrent_queries": schema.Int64Attribute{
				Description: "Limit on total number of concurrently executed queries.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"keep_alive_timeout": schema.Int64Attribute{
				Description: "The number of seconds that ClickHouse waits for incoming requests for HTTP protocol before closing the connection.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uncompressed_cache_size": schema.Int64Attribute{
				Description: "Cache size (in bytes) for uncompressed data used by table engines from the MergeTree family. Zero means disabled.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_table_size_to_drop": schema.Int64Attribute{
				Description: "Restriction on deleting tables.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_partition_size_to_drop": schema.Int64Attribute{
				Description: "Restriction on dropping partitions.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Description: "The server's time zone.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"geobase_uri": schema.StringAttribute{
				Description: "Address of the archive with the user geobase in Object Storage.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"geobase_enabled": schema.BoolAttribute{
				Description: "Enable or disable geobase.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"query_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that query_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that query_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query_thread_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable query_thread_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"query_thread_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that query_thread_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query_thread_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that query_thread_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"part_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that part_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"part_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that part_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"metric_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable metric_log system table.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"metric_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that metric_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"metric_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that metric_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"trace_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable trace_log system table.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"trace_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that trace_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"trace_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that trace_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"text_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable text_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"text_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that text_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"text_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that text_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"opentelemetry_span_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable opentelemetry_span_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"opentelemetry_span_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that opentelemetry_span_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"opentelemetry_span_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that opentelemetry_span_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query_views_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable query_views_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"query_views_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that query_views_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query_views_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that query_views_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_metric_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable asynchronous_metric_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_metric_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that asynchronous_metric_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_metric_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that asynchronous_metric_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"session_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable session_log system table.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"session_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that session_log can grow to before old data will be removed.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"session_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that session_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"zookeeper_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable zookeeper_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"zookeeper_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that zookeeper_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"zookeeper_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that zookeeper_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_insert_log_enabled": schema.BoolAttribute{
				Description: "Enable or disable asynchronous_insert_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_insert_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that asynchronous_insert_log can grow to before old data will be removed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"asynchronous_insert_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that asynchronous_insert_log records will be retained before removal.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"processors_profile_log_enabled": schema.BoolAttribute{
				Description: "Enables or disables processors_profile_log system table.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"processors_profile_log_retention_size": schema.Int64Attribute{
				Description: "The maximum time that processors_profile_log records will be retained before removal. If set to **0**, automatic removal of processors_profile_log data based on time is disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"processors_profile_log_retention_time": schema.Int64Attribute{
				Description: "Enables or disables error_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"error_log_enabled": schema.BoolAttribute{
				Description: "Enables or disables error_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"error_log_retention_size": schema.Int64Attribute{
				Description: "The maximum size that error_log can grow to before old data will be removed. If set to **0**, automatic removal of error_log data based on size is disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"error_log_retention_time": schema.Int64Attribute{
				Description: "The maximum time that error_log records will be retained before removal. If set to **0**, automatic removal of error_log data based on time is disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"access_control_improvements": AccessControlImprovementsSchema(),
			"text_log_level": schema.StringAttribute{
				Description: "Logging level for text_log system table.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"background_pool_size": schema.Int64Attribute{
				Description: "Sets the number of threads performing background merges and mutations for MergeTree-engine tables.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_schedule_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for constantly executing some lightweight periodic operations for replicated tables, Kafka streaming, and DNS cache updates.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_fetches_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for fetching data parts from another replica for MergeTree-engine tables in a background.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_move_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for moving data parts to another disk or volume for MergeTree-engine tables in a background.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_distributed_schedule_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for executing distributed sends.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_buffer_flush_schedule_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for performing flush operations for Buffer-engine tables in the background.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_message_broker_schedule_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for executing background operations for message streaming.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_common_pool_size": schema.Int64Attribute{
				Description: "The maximum number of threads that will be used for performing a variety of operations (mostly garbage collection) for MergeTree-engine tables in a background.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"background_merges_mutations_concurrency_ratio": schema.Int64Attribute{
				Description: "Sets a ratio between the number of threads and the number of background merges and mutations that can be executed concurrently.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"default_database": schema.StringAttribute{
				Description: "Default database name.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"total_memory_profiler_step": schema.Int64Attribute{
				Description: "Whenever server memory usage becomes larger than every next step in number of bytes the memory profiler will collect the allocating stack trace.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"total_memory_tracker_sample_probability": schema.Float64Attribute{
				Description: "Allows to collect random allocations and de-allocations and writes them in the system.trace_log system table with trace_type equal to a MemorySample with the specified probability.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"async_insert_threads": schema.Int64Attribute{
				Description: "Maximum number of threads to parse and insert data in background.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"backup_threads": schema.Int64Attribute{
				Description: "The maximum number of threads to execute **BACKUP** requests.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"restore_threads": schema.Int64Attribute{
				Description: "The maximum number of threads to execute **RESTORE** requests.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dictionaries_lazy_load": schema.BoolAttribute{
				Description: "Lazy loading of dictionaries. If true, then each dictionary is loaded on the first use.",
				Optional:    true,
				// Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"mysql_protocol": schema.BoolAttribute{
				Description: "Enables or disables MySQL interface on ClickHouse server.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"jdbc_bridge":         JdbcBridgeSchema(),
			"rabbitmq":            RabbitMQSchema(),
			"kafka":               KafkaSchema(),
			"merge_tree":          MergeTreeSchema(),
			"query_cache":         QueryCacheSchema(),
			"compression":         CompressionSchema(),
			"graphite_rollup":     GraphiteRollupSchema(),
			"query_masking_rules": QueryMaskingRulesSchema(),
			"custom_macros":       CustomMacrosSchema(),
		},
	}
}

func ResourcesSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Resources allocated to hosts.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"resource_preset_id": schema.StringAttribute{
				Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",
				Required:    true,
			},
			"disk_size": schema.Int64Attribute{
				Description: "Volume of the storage available to a host, in gigabytes.",
				Required:    true,
			},
			"disk_type_id": schema.StringAttribute{
				Description: "Type of the storage of hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
				Required:    true,
			},
		},
	}
}

func ShardResourcesSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Resources allocated to hosts.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"resource_preset_id": schema.StringAttribute{
				Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",
				Optional:    true,
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Description: "Volume of the storage available to a host, in gigabytes.",
				Optional:    true,
				Computed:    true,
			},
			"disk_type_id": schema.StringAttribute{
				Description: "Type of the storage of hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func ShardsSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		Description: "A shards of the ClickHouse cluster.",
		Computed:    true,
		Optional:    true,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"weight": schema.Int64Attribute{
					Description: "The weight of shard.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"resources": ShardResourcesSchema(),
			},
		},
	}
}

func ShardGroupSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "A group of clickhouse shards.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "The name of the shard group, used as cluster name in Distributed tables.",
					Required:    true,
				},
				"description": schema.StringAttribute{
					Description: "Description of the shard group.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"shard_names": schema.ListAttribute{
					Description: "List of shards names that belong to the shard group.",
					Required:    true,
					ElementType: types.StringType,
				},
			},
		},
	}
}

func FormatSchemaSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Description: "A set of `protobuf` or `capnproto` format schemas.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "The name of the format schema.",
					Required:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of the format schema.",
					Required:    true,
				},
				"uri": schema.StringAttribute{
					Description: "Format schema file URL. You can only use format schemas stored in Yandex Object Storage.",
					Required:    true,
				},
			},
		},
	}
}

func MlModelSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Description: "A group of machine learning models.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "The name of the ml model.",
					Required:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of the model.",
					Required:    true,
				},
				"uri": schema.StringAttribute{
					Description: "Model file URL. You can only use models stored in Yandex Object Storage.",
					Required:    true,
				},
			},
		},
	}
}

func MaintenanceWindowSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Maintenance window settings.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ANYTIME", "WEEKLY"),
				},
			},
			"day": schema.StringAttribute{
				Description: "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"MON", "TUE",
						"WED", "THU",
						"FRI", "SAT",
						"SUN",
					),
				},
			},
			"hour": schema.Int64Attribute{
				Description: "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 24),
				},
			},
		},
	}
}

func AccessControlImprovementsSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Access control settings.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"select_from_system_db_requires_grant": schema.BoolAttribute{
				Description: "Sets whether **SELECT * FROM system.<table>** requires any grants and can be executed by any user. If set to true then this query requires **GRANT SELECT ON system.<table>** just as for non-system tables.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"select_from_information_schema_requires_grant": schema.BoolAttribute{
				Description: "Sets whether **SELECT * FROM information_schema.<table>** requires any grants and can be executed by any user. If set to true, then this query requires **GRANT SELECT ON information_schema.<table>**, just as for ordinary tables.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func CustomMacrosSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Custom ClickHouse macros.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Name of the macro.",
					Required:    true,
				},
				"value": schema.StringAttribute{
					Description: "Value of the macro.",
					Required:    true,
				},
			},
		},
	}
}

func QueryMaskingRulesSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Query masking rules configuration.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Name for the rule.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"regexp": schema.StringAttribute{
					Description: "RE2 compatible regular expression.",
					Required:    true,
				},
				"replace": schema.StringAttribute{
					Description: "Substitution string for sensitive data. Default value: six asterisks.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

func GraphiteRollupSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Graphite rollup configuration.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Graphite rollup configuration name.",
					Required:    true,
				},
				"path_column_name": schema.StringAttribute{
					Description: "The name of the column storing the metric name (Graphite sensor). Default value: Path.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"time_column_name": schema.StringAttribute{
					Description: "The name of the column storing the time of measuring the metric. Default value: Time.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"value_column_name": schema.StringAttribute{
					Description: "The name of the column storing the value of the metric at the time set in `time_column_name`. Default value: Value.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"version_column_name": schema.StringAttribute{
					Description: "The name of the column storing the version of the metric. Default value: Timestamp.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"patterns": PatternsSchema(),
			},
		},
	}
}

func PatternsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Set of thinning rules.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"regexp": schema.StringAttribute{
					Description: "Regular expression that the metric name must match.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"function": schema.StringAttribute{
					Description: "Aggregation function name.",
					Required:    true,
				},
				"retention": RetentionSchema(),
			},
		},
	}
}

func RetentionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Retain parameters.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"age": schema.Int64Attribute{
					Description: "Minimum data age in seconds.",
					Required:    true,
				},
				"precision": schema.Int64Attribute{
					Description: "Accuracy of determining the age of the data in seconds.",
					Required:    true,
				},
			},
		},
	}
}

func CompressionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Data compression configuration.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"method": schema.StringAttribute{
					Description: "Compression method. Two methods are available: `LZ4` and `zstd`.",
					Required:    true,
				},
				"min_part_size": schema.Int64Attribute{
					Description: "Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.",
					Required:    true,
				},
				"min_part_size_ratio": schema.NumberAttribute{
					Description: "Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.",
					Required:    true,
				},
				"level": schema.Int64Attribute{
					Description: "Compression level for `ZSTD` method.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

func QueryCacheSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Query cache configuration.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"max_size_in_bytes": schema.Int64Attribute{
				Description: "The maximum cache size in bytes. 0 means the query cache is disabled. Default value: 1073741824 (1 GiB).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_entries": schema.Int64Attribute{
				Description: "The maximum number of SELECT query results stored in the cache. Default value: 1024.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_entry_size_in_bytes": schema.Int64Attribute{
				Description: "The maximum size in bytes SELECT query results may have to be saved in the cache. Default value: 1048576 (1 MiB).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_entry_size_in_rows": schema.Int64Attribute{
				Description: "The maximum number of rows SELECT query results may have to be saved in the cache. Default value: 30000000 (30 mil).",
				Optional:    true,
				Computed:    true, PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func MergeTreeSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "MergeTree engine configuration.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"replicated_deduplication_window": schema.Int64Attribute{
				Description: "Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"replicated_deduplication_window_seconds": schema.Int64Attribute{
				Description: "Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones will be deleted).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fsync_after_insert": schema.BoolAttribute{
				Description: "Do fsync for every inserted part. Significantly decreases performance of inserts, not recommended to use with wide parts.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"fsync_part_directory": schema.BoolAttribute{
				Description: "Do fsync for part directory after all part operations (writes, renames, etc.).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"min_compressed_bytes_to_fsync_after_fetch": schema.Int64Attribute{
				Description: "Minimal number of rows to do fsync for part after merge. **0** means disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_compressed_bytes_to_fsync_after_merge": schema.Int64Attribute{
				Description: "Minimal number of compressed bytes to do fsync for part after merge. **0** means disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_rows_to_fsync_after_merge": schema.Int64Attribute{
				Description: "Minimal number of rows to do fsync for part after merge. **0** means disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"parts_to_delay_insert": schema.Int64Attribute{
				Description: "Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"parts_to_throw_insert": schema.Int64Attribute{
				Description: "Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"inactive_parts_to_delay_insert": schema.Int64Attribute{
				Description: "If the number of inactive parts in a single partition in the table at least that many the inactive_parts_to_delay_insert value, an INSERT artificially slows down. It is useful when a server fails to clean up parts quickly enough.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"inactive_parts_to_throw_insert": schema.Int64Attribute{
				Description: "If the number of inactive parts in a single partition more than the inactive_parts_to_throw_insert value, INSERT is interrupted with the `Too many inactive parts (N). Parts cleaning are processing significantly slower than inserts` exception.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_replicated_merges_in_queue": schema.Int64Attribute{
				Description: "Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"number_of_free_entries_in_pool_to_lower_max_size_of_merge": schema.Int64Attribute{
				Description: "Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_bytes_to_merge_at_min_space_in_pool": schema.Int64Attribute{
				Description: "Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_bytes_to_merge_at_max_space_in_pool": schema.Int64Attribute{
				Description: "The maximum total parts size (in bytes) to be merged into one part, if there are enough resources available. Roughly corresponds to the maximum possible part size created by an automatic background merge.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_bytes_for_wide_part": schema.Int64Attribute{
				Description: "Minimum number of bytes in a data part that can be stored in Wide format. You can set one, both or none of these settings.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_rows_for_wide_part": schema.Int64Attribute{
				Description: "Minimum number of rows in a data part that can be stored in Wide format. You can set one, both or none of these settings.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ttl_only_drop_parts": schema.BoolAttribute{
				Description: "Enables zero-copy replication when a replica is located on a remote filesystem.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"merge_with_ttl_timeout": schema.Int64Attribute{
				Description: "Minimum delay in seconds before repeating a merge with delete TTL. Default value: 14400 seconds (4 hours).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"merge_with_recompression_ttl_timeout": schema.Int64Attribute{
				Description: "Minimum delay in seconds before repeating a merge with recompression TTL. Default value: 14400 seconds (4 hours).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_parts_in_total": schema.Int64Attribute{
				Description: "Maximum number of parts in all partitions.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_number_of_merges_with_ttl_in_pool": schema.Int64Attribute{
				Description: "When there is more than specified number of merges with TTL entries in pool, do not assign new merge with TTL.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"materialize_ttl_recalculate_only": schema.BoolAttribute{
				Description: "Only recalculate ttl info when **MATERIALIZE TTL**.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cleanup_delay_period": schema.Int64Attribute{
				Description: "Minimum period to clean old queue logs, blocks hashes and parts.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"number_of_free_entries_in_pool_to_execute_mutation": schema.Int64Attribute{
				Description: "When there is less than specified number of free entries in pool, do not execute part mutations. This is to leave free threads for regular merges and avoid `Too many parts`. Default value: 20.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_avg_part_size_for_too_many_parts": schema.Int64Attribute{
				Description: "The `too many parts` check will be active only if the average part size is not larger than the specified threshold. This allows large tables if parts are successfully merged.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_age_to_force_merge_seconds": schema.Int64Attribute{
				Description: "Merge parts if every part in the range is older than the value of `min_age_to_force_merge_seconds`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_age_to_force_merge_on_partition_only": schema.BoolAttribute{
				Description: "Whether min_age_to_force_merge_seconds should be applied only on the entire partition and not on subset.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"merge_selecting_sleep_ms": schema.Int64Attribute{
				Description: "Sleep time for merge selecting when no part is selected. Lower values increase ZooKeeper requests in large clusters.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"merge_max_block_size": schema.Int64Attribute{
				Description: "The number of rows that are read from the merged parts into memory. Default value: 8192.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"deduplicate_merge_projection_mode": schema.StringAttribute{
				Description: "Determines the behavior of background merges for MergeTree tables with projections.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lightweight_mutation_projection_mode": schema.StringAttribute{
				Description: "Determines the behavior of lightweight deletes for MergeTree tables with projections.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"check_sample_column_is_correct": schema.BoolAttribute{
				Description: "Enables the check at table creation that the sampling column type is correct. Default value: true.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"max_merge_selecting_sleep_ms": schema.Int64Attribute{
				Description: "Maximum sleep time for merge selecting. Default value: 60000 milliseconds (60 seconds).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_cleanup_delay_period": schema.Int64Attribute{
				Description: "Maximum period to clean old queue logs, blocks hashes and parts. Default value: 300 seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func KafkaSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Kafka connection configuration.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"security_protocol": schema.StringAttribute{
				Description: "Security protocol used to connect to kafka server.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sasl_mechanism": schema.StringAttribute{
				Description: "SASL mechanism used in kafka authentication.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sasl_username": schema.StringAttribute{
				Description: "Username on kafka server.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sasl_password": schema.StringAttribute{
				Description: "User password on kafka server.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Sensitive: true,
			},
			"enable_ssl_certificate_verification": schema.BoolAttribute{
				Description: "Enable verification of SSL certificates.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"max_poll_interval_ms": schema.Int64Attribute{
				Description: "Maximum allowed time between calls to consume messages. If exceeded, consumer is considered failed.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"session_timeout_ms": schema.Int64Attribute{
				Description: "Client group session and failure detection timeout.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"debug": schema.StringAttribute{
				Description: "A comma-separated list of debug contexts to enable.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_offset_reset": schema.StringAttribute{
				Description: "Action when no initial offset: 'smallest','earliest','largest','latest','error'.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func RabbitMQSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "RabbitMQ connection configuration.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "RabbitMQ username.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Description: "RabbitMQ user password.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vhost": schema.StringAttribute{
				Description: "RabbitMQ vhost. Default: `\\`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func JdbcBridgeSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "JDBC bridge configuration.",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Host of jdbc bridge.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port of jdbc bridge. Default value: 9019.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
