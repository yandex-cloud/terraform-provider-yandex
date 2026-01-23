package mdb_clickhouse_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func DataSourceClusterSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Get information about a Yandex Managed ClickHouse cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).\n\n~> Either `cluster_id` or `name` should be specified.",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: common.ResourceDescriptions["id"],
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the ClickHouse cluster. This ID is assigned by MDB at creation time.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the ClickHouse cluster. Provided by the client when the cluster is created.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["description"],
				Computed:            true,
			},
			"folder_id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Computed:            true,
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["network_id"],
				Computed:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Deployment environment of the ClickHouse cluster.",
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: common.ResourceDescriptions["labels"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: common.ResourceDescriptions["deletion_protection"],
				Computed:            true,
			},
			"disk_encryption_key_id": schema.StringAttribute{
				MarkdownDescription: "ID of the KMS key for cluster disk encryption.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["created_at"],
				Computed:            true,
			},
			"security_group_ids": schema.SetAttribute{
				MarkdownDescription: common.ResourceDescriptions["security_group_ids"],
				ElementType:         types.StringType,
				Computed:            true,
			},
			"service_account_id": &schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["service_account_id"],
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the ClickHouse server software.",
				Computed:            true,
			},
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "A password used to authorize as user `admin` when `sql_user_management` enabled.",
				Computed:            true,
				Sensitive:           true,
			},
			"sql_user_management": schema.BoolAttribute{
				MarkdownDescription: "Enables `admin` user with user management permission.",
				Computed:            true,
			},
			"sql_database_management": schema.BoolAttribute{
				MarkdownDescription: "Grants `admin` user database management permission.",
				Computed:            true,
			},
			"embedded_keeper": schema.BoolAttribute{
				MarkdownDescription: "Whether to use ClickHouse Keeper as a coordination system.",
				Computed:            true,
			},
			"backup_retain_period_days": schema.Int64Attribute{
				MarkdownDescription: "The period in days during which backups are stored.",
				Computed:            true,
			},
			"copy_schema_on_new_hosts": schema.BoolAttribute{
				MarkdownDescription: "Whether to copy schema on new ClickHouse hosts.",
				Computed:            true,
			},
			"clickhouse":          DataSourceClickHouseSchema(),
			"zookeeper":           DataSourceZooKeeperSchema(),
			"cloud_storage":       DataSourceCloudStorageSchema(),
			"backup_window_start": DataSourceBackupWindowStart(),
			"access":              DataSourceAccessSchema(),
			"hosts":               DataSourceHostsSchema(),
			"shards":              DataSourceShardsSchema(),
		},
		Blocks: map[string]schema.Block{
			"shard_group":        DataSourceShardGroupSchema(),
			"format_schema":      DataSourceFormatSchemaSchema(),
			"ml_model":           DataSourceMlModelSchema(),
			"maintenance_window": DataSourceMaintenanceWindowSchema(),
		},
	}
}

func DataSourceHostsSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "A host configuration of the ClickHouse cluster.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"zone": schema.StringAttribute{
					MarkdownDescription: common.ResourceDescriptions["zone"],
					Computed:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "The type of the host to be deployed. Can be either `CLICKHOUSE` or `ZOOKEEPER`.",
					Computed:            true,
				},
				"subnet_id": schema.StringAttribute{
					MarkdownDescription: "ID of the subnet where the host is located.",
					Computed:            true,
				},
				"assign_public_ip": schema.BoolAttribute{
					MarkdownDescription: "Whether the host should get a public IP address.",
					Computed:            true,
				},
				"shard_name": schema.StringAttribute{
					MarkdownDescription: "The name of the shard to which the host belongs.",
					Computed:            true,
				},
				"fqdn": schema.StringAttribute{
					MarkdownDescription: "The fully qualified domain name of the host.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceZooKeeperSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Configuration of the ZooKeeper subcluster.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"resources": DataSourceResourcesSchema(),
		},
	}
}

func DataSourceCloudStorageSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Cloud Storage settings.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.",
				Computed:            true,
			},
			"move_factor": schema.NumberAttribute{
				MarkdownDescription: "Sets the minimum free space ratio in the cluster storage. If the free space is lower than this value, the data is transferred to Yandex Object Storage. Acceptable values are 0 to 1, inclusive.",
				Computed:            true,
			},
			"data_cache_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables temporary storage in the cluster repository of data requested from the object repository.",
				Computed:            true,
			},
			"data_cache_max_size": schema.Int64Attribute{
				MarkdownDescription: "Defines the maximum amount of memory (in bytes) allocated in the cluster storage for temporary storage of data requested from the object storage.",
				Computed:            true,
			},
			"prefer_not_to_merge": schema.BoolAttribute{
				MarkdownDescription: "Disables merging of data parts in `Yandex Object Storage`.",
				Computed:            true,
			},
		},
	}
}

func DataSourceBackupWindowStart() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Time to start the daily backup, in the UTC timezone.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"hours": schema.Int64Attribute{
				MarkdownDescription: "The hour at which backup will be started (UTC).",
				Computed:            true,
			},
			"minutes": schema.Int64Attribute{
				MarkdownDescription: "The minute at which backup will be started (UTC).",
				Computed:            true,
			},
		},
	}
}

func DataSourceAccessSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Access policy to the ClickHouse cluster.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"web_sql": schema.BoolAttribute{
				MarkdownDescription: "Allow access for Web SQL.",
				Computed:            true,
			},
			"data_lens": schema.BoolAttribute{
				MarkdownDescription: "Allow access for DataLens.",
				Computed:            true,
			},
			"metrika": schema.BoolAttribute{
				MarkdownDescription: "Allow access for Yandex.Metrika.",
				Computed:            true,
			},
			"serverless": schema.BoolAttribute{
				MarkdownDescription: "Allow access for Serverless.",
				Computed:            true,
			},
			"data_transfer": schema.BoolAttribute{
				MarkdownDescription: "Allow access for DataTransfer.",
				Computed:            true,
			},
			"yandex_query": schema.BoolAttribute{
				MarkdownDescription: "Allow access for YandexQuery.",
				Computed:            true,
			},
		},
	}
}

func DataSourceClickHouseSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Configuration of the ClickHouse subcluster.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"resources": DataSourceResourcesSchema(),
			"config":    DataSourceClickHouseConfigSchema(),
		},
	}
}

func DataSourceClickHouseConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Configuration of the ClickHouse subcluster.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"log_level": schema.StringAttribute{
				MarkdownDescription: "Logging level.",
				Computed:            true,
			},
			"max_connections": schema.Int64Attribute{
				MarkdownDescription: "Max server connections.",
				Computed:            true,
			},
			"max_concurrent_queries": schema.Int64Attribute{
				MarkdownDescription: "Limit on total number of concurrently executed queries.",
				Computed:            true,
			},
			"keep_alive_timeout": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds that ClickHouse waits for incoming requests for HTTP protocol before closing the connection.",
				Computed:            true,
			},
			"uncompressed_cache_size": schema.Int64Attribute{
				MarkdownDescription: "Cache size (in bytes) for uncompressed data used by table engines from the MergeTree family. Zero means disabled.",
				Computed:            true,
			},
			"max_table_size_to_drop": schema.Int64Attribute{
				MarkdownDescription: "Restriction on deleting tables.",
				Computed:            true,
			},
			"max_partition_size_to_drop": schema.Int64Attribute{
				MarkdownDescription: "Restriction on dropping partitions.",
				Computed:            true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "The server's time zone.",
				Computed:            true,
			},
			"geobase_uri": schema.StringAttribute{
				MarkdownDescription: "Address of the archive with the user geobase in Object Storage.",
				Computed:            true,
			},
			"geobase_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable geobase.",
				Computed:            true,
			},
			"query_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that query_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"query_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that query_log records will be retained before removal.",
				Computed:            true,
			},
			"query_thread_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable query_thread_log system table.",
				Computed:            true,
			},
			"query_thread_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that query_thread_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"query_thread_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that query_thread_log records will be retained before removal.",
				Computed:            true,
			},
			"part_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that part_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"part_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that part_log records will be retained before removal.",
				Computed:            true,
			},
			"metric_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable metric_log system table.",
				Computed:            true,
			},
			"metric_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that metric_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"metric_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that metric_log records will be retained before removal.",
				Computed:            true,
			},
			"trace_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable trace_log system table.",
				Computed:            true,
			},
			"trace_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that trace_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"trace_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that trace_log records will be retained before removal.",
				Computed:            true,
			},
			"text_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable text_log system table.",
				Computed:            true,
			},
			"text_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that text_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"text_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that text_log records will be retained before removal.",
				Computed:            true,
			},
			"opentelemetry_span_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable opentelemetry_span_log system table.",
				Computed:            true,
			},
			"opentelemetry_span_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that opentelemetry_span_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"opentelemetry_span_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that opentelemetry_span_log records will be retained before removal.",
				Computed:            true,
			},
			"query_views_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable query_views_log system table.",
				Computed:            true,
			},
			"query_views_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that query_views_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"query_views_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that query_views_log records will be retained before removal.",
				Computed:            true,
			},
			"asynchronous_metric_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable asynchronous_metric_log system table.",
				Computed:            true,
			},
			"asynchronous_metric_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that asynchronous_metric_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"asynchronous_metric_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that asynchronous_metric_log records will be retained before removal.",
				Computed:            true,
			},
			"session_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable session_log system table.",
				Computed:            true,
			},
			"session_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that session_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"session_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that session_log records will be retained before removal.",
				Computed:            true,
			},
			"zookeeper_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable zookeeper_log system table.",
				Computed:            true,
			},
			"zookeeper_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that zookeeper_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"zookeeper_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that zookeeper_log records will be retained before removal.",
				Computed:            true,
			},
			"asynchronous_insert_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable asynchronous_insert_log system table.",
				Computed:            true,
			},
			"asynchronous_insert_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that asynchronous_insert_log can grow to before old data will be removed.",
				Computed:            true,
			},
			"asynchronous_insert_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that asynchronous_insert_log records will be retained before removal.",
				Computed:            true,
			},
			"processors_profile_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables processors_profile_log system table.",
				Computed:            true,
			},
			"processors_profile_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that processors_profile_log records will be retained before removal. If set to **0**, automatic removal of processors_profile_log data based on time is disabled.",
				Computed:            true,
			},
			"processors_profile_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "Enables or disables error_log system table.",
				Computed:            true,
			},
			"error_log_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables error_log system table.",
				Computed:            true,
			},
			"error_log_retention_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum size that error_log can grow to before old data will be removed. If set to **0**, automatic removal of error_log data based on size is disabled.",
				Computed:            true,
			},
			"error_log_retention_time": schema.Int64Attribute{
				MarkdownDescription: "The maximum time that error_log records will be retained before removal. If set to **0**, automatic removal of error_log data based on time is disabled.",
				Computed:            true,
			},
			"access_control_improvements": DataSourceAccessControlImprovementsSchema(),
			"text_log_level": schema.StringAttribute{
				MarkdownDescription: "Logging level for text_log system table.",
				Computed:            true,
			},
			"background_pool_size": schema.Int64Attribute{
				MarkdownDescription: "Sets the number of threads performing background merges and mutations for MergeTree-engine tables.",
				Computed:            true,
			},
			"background_schedule_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for constantly executing some lightweight periodic operations for replicated tables, Kafka streaming, and DNS cache updates.",
				Computed:            true,
			},
			"background_fetches_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for fetching data parts from another replica for MergeTree-engine tables in a background.",
				Computed:            true,
			},
			"background_move_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for moving data parts to another disk or volume for MergeTree-engine tables in a background.",
				Computed:            true,
			},
			"background_distributed_schedule_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for executing distributed sends.",
				Computed:            true,
			},
			"background_buffer_flush_schedule_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for performing flush operations for Buffer-engine tables in the background.",
				Computed:            true,
			},
			"background_message_broker_schedule_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for executing background operations for message streaming.",
				Computed:            true,
			},
			"background_common_pool_size": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads that will be used for performing a variety of operations (mostly garbage collection) for MergeTree-engine tables in a background.",
				Computed:            true,
			},
			"background_merges_mutations_concurrency_ratio": schema.Int64Attribute{
				MarkdownDescription: "Sets a ratio between the number of threads and the number of background merges and mutations that can be executed concurrently.",
				Computed:            true,
			},
			"default_database": schema.StringAttribute{
				MarkdownDescription: "Default database name.",
				Computed:            true,
			},
			"total_memory_profiler_step": schema.Int64Attribute{
				MarkdownDescription: "Whenever server memory usage becomes larger than every next step in number of bytes the memory profiler will collect the allocating stack trace.",
				Computed:            true,
			},
			"total_memory_tracker_sample_probability": schema.Float64Attribute{
				MarkdownDescription: "Allows to collect random allocations and de-allocations and writes them in the system.trace_log system table with trace_type equal to a MemorySample with the specified probability.",
				Computed:            true,
			},
			"async_insert_threads": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of threads to parse and insert data in background.",
				Computed:            true,
			},
			"backup_threads": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads to execute **BACKUP** requests.",
				Computed:            true,
			},
			"restore_threads": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of threads to execute **RESTORE** requests.",
				Computed:            true,
			},
			"dictionaries_lazy_load": schema.BoolAttribute{
				MarkdownDescription: "Lazy loading of dictionaries. If true, then each dictionary is loaded on the first use.",
				Computed:            true,
			},
			"mysql_protocol": schema.BoolAttribute{
				MarkdownDescription: "Enables or disables MySQL interface on ClickHouse server.",
				Computed:            true,
			},
			"jdbc_bridge":         DataSourceJdbcBridgeSchema(),
			"rabbitmq":            DataSourceRabbitMQSchema(),
			"kafka":               DataSourceKafkaSchema(),
			"merge_tree":          DataSourceMergeTreeSchema(),
			"query_cache":         DataSourceQueryCacheSchema(),
			"compression":         DataSourceCompressionSchema(),
			"graphite_rollup":     DataSourceGraphiteRollupSchema(),
			"query_masking_rules": DataSourceQueryMaskingRulesSchema(),
			"custom_macros":       DataSourceCustomMacrosSchema(),
		},
	}
}

func DataSourceResourcesSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Resources allocated to hosts.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"resource_preset_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",
				Computed:            true,
			},
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: "Volume of the storage available to a host, in gigabytes.",
				Computed:            true,
			},
			"disk_type_id": schema.StringAttribute{
				MarkdownDescription: "Type of the storage of hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
				Computed:            true,
			},
		},
	}
}

func DataSourceShardsSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "A shards of the ClickHouse cluster.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"weight": schema.Int64Attribute{
					MarkdownDescription: "The weight of shard.",
					Computed:            true,
				},
				"resources": DataSourceResourcesSchema(),
			},
		},
	}
}

func DataSourceShardGroupSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "A group of clickhouse shards.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "The name of the shard group, used as cluster name in Distributed tables.",
					Computed:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "MarkdownDescription of the shard group.",
					Computed:            true,
				},
				"shard_names": schema.ListAttribute{
					MarkdownDescription: "List of shards names that belong to the shard group.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceFormatSchemaSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "A set of `protobuf` or `capnproto` format schemas.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "The name of the format schema.",
					Computed:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "Type of the format schema.",
					Computed:            true,
				},
				"uri": schema.StringAttribute{
					MarkdownDescription: "Format schema file URL. You can only use format schemas stored in Yandex Object Storage.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceMlModelSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "A group of machine learning models.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "The name of the ml model.",
					Computed:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "Type of the model.",
					Computed:            true,
				},
				"uri": schema.StringAttribute{
					MarkdownDescription: "Model file URL. You can only use models stored in Yandex Object Storage.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceMaintenanceWindowSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Maintenance window settings.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
				Computed:            true,
			},
			"day": schema.StringAttribute{
				MarkdownDescription: "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
				Computed:            true,
			},
			"hour": schema.Int64Attribute{
				MarkdownDescription: "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
				Computed:            true,
			},
		},
	}
}

func DataSourceAccessControlImprovementsSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Access control settings.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"select_from_system_db_requires_grant": schema.BoolAttribute{
				MarkdownDescription: "Sets whether `SELECT * FROM system.<table>` requires any grants and can be executed by any user. If set to true then this query requires `GRANT SELECT ON system.<table>` just as for non-system tables.",
				Computed:            true,
			},
			"select_from_information_schema_requires_grant": schema.BoolAttribute{
				MarkdownDescription: "Sets whether `SELECT * FROM information_schema.<table>` requires any grants and can be executed by any user. If set to true, then this query requires `GRANT SELECT ON information_schema.<table>`, just as for ordinary tables.",
				Computed:            true,
			},
		},
	}
}

func DataSourceCustomMacrosSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Custom ClickHouse macros.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the macro.",
					Computed:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value of the macro.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceQueryMaskingRulesSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Query masking rules configuration.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name for the rule.",
					Computed:            true,
				},
				"regexp": schema.StringAttribute{
					MarkdownDescription: "RE2 compatible regular expression.",
					Computed:            true,
				},
				"replace": schema.StringAttribute{
					MarkdownDescription: "Substitution string for sensitive data. Default value: six asterisks.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceGraphiteRollupSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Graphite rollup configuration.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Graphite rollup configuration name.",
					Computed:            true,
				},
				"path_column_name": schema.StringAttribute{
					MarkdownDescription: "The name of the column storing the metric name (Graphite sensor). Default value: Path.",
					Computed:            true,
				},
				"time_column_name": schema.StringAttribute{
					MarkdownDescription: "The name of the column storing the time of measuring the metric. Default value: Time.",
					Computed:            true,
				},
				"value_column_name": schema.StringAttribute{
					MarkdownDescription: "The name of the column storing the value of the metric at the time set in `time_column_name`. Default value: Value.",
					Computed:            true,
				},
				"version_column_name": schema.StringAttribute{
					MarkdownDescription: "The name of the column storing the version of the metric. Default value: Timestamp.",
					Computed:            true,
				},
				"patterns": DataSourcePatternsSchema(),
			},
		},
	}
}

func DataSourcePatternsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Set of thinning rules.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"regexp": schema.StringAttribute{
					MarkdownDescription: "Regular expression that the metric name must match.",
					Computed:            true,
				},
				"function": schema.StringAttribute{
					MarkdownDescription: "Aggregation function name.",
					Computed:            true,
				},
				"retention": DataSourceRetentionSchema(),
			},
		},
	}
}

func DataSourceRetentionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Retain parameters.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"age": schema.Int64Attribute{
					MarkdownDescription: "Minimum data age in seconds.",
					Computed:            true,
				},
				"precision": schema.Int64Attribute{
					MarkdownDescription: "Accuracy of determining the age of the data in seconds.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceCompressionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Data compression configuration.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"method": schema.StringAttribute{
					MarkdownDescription: "Compression method. Two methods are available: `LZ4` and `zstd`.",
					Computed:            true,
				},
				"min_part_size": schema.Int64Attribute{
					MarkdownDescription: "Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value.",
					Computed:            true,
				},
				"min_part_size_ratio": schema.NumberAttribute{
					MarkdownDescription: "Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value.",
					Computed:            true,
				},
				"level": schema.Int64Attribute{
					MarkdownDescription: "Compression level for `ZSTD` method.",
					Computed:            true,
				},
			},
		},
	}
}

func DataSourceQueryCacheSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Query cache configuration.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"max_size_in_bytes": schema.Int64Attribute{
				MarkdownDescription: "The maximum cache size in bytes. 0 means the query cache is disabled. Default value: 1073741824 (1 GiB).",
				Computed:            true,
			},
			"max_entries": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of SELECT query results stored in the cache. Default value: 1024.",
				Computed:            true,
			},
			"max_entry_size_in_bytes": schema.Int64Attribute{
				MarkdownDescription: "The maximum size in bytes SELECT query results may have to be saved in the cache. Default value: 1048576 (1 MiB).",
				Computed:            true,
			},
			"max_entry_size_in_rows": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of rows SELECT query results may have to be saved in the cache. Default value: 30000000 (30 mil).",
				Computed:            true,
			},
		},
	}
}

func DataSourceMergeTreeSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "MergeTree engine configuration.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"replicated_deduplication_window": schema.Int64Attribute{
				MarkdownDescription: "Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted).",
				Computed:            true,
			},
			"replicated_deduplication_window_seconds": schema.Int64Attribute{
				MarkdownDescription: "Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones will be deleted).",
				Computed:            true,
			},
			"fsync_after_insert": schema.BoolAttribute{
				MarkdownDescription: "Do fsync for every inserted part. Significantly decreases performance of inserts, not recommended to use with wide parts.",
				Computed:            true,
			},
			"fsync_part_directory": schema.BoolAttribute{
				MarkdownDescription: "Do fsync for part directory after all part operations (writes, renames, etc.).",
				Computed:            true,
			},
			"min_compressed_bytes_to_fsync_after_fetch": schema.Int64Attribute{
				MarkdownDescription: "Minimal number of rows to do fsync for part after merge. **0** means disabled.",
				Computed:            true,
			},
			"min_compressed_bytes_to_fsync_after_merge": schema.Int64Attribute{
				MarkdownDescription: "Minimal number of compressed bytes to do fsync for part after merge. **0** means disabled.",
				Computed:            true,
			},
			"min_rows_to_fsync_after_merge": schema.Int64Attribute{
				MarkdownDescription: "Minimal number of rows to do fsync for part after merge. **0** means disabled.",
				Computed:            true,
			},
			"parts_to_delay_insert": schema.Int64Attribute{
				MarkdownDescription: "Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table",
				Computed:            true,
			},
			"parts_to_throw_insert": schema.Int64Attribute{
				MarkdownDescription: "Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception.",
				Computed:            true,
			},
			"inactive_parts_to_delay_insert": schema.Int64Attribute{
				MarkdownDescription: "If the number of inactive parts in a single partition in the table at least that many the inactive_parts_to_delay_insert value, an INSERT artificially slows down. It is useful when a server fails to clean up parts quickly enough.",
				Computed:            true,
			},
			"inactive_parts_to_throw_insert": schema.Int64Attribute{
				MarkdownDescription: "If the number of inactive parts in a single partition more than the inactive_parts_to_throw_insert value, INSERT is interrupted with the `Too many inactive parts (N). Parts cleaning are processing significantly slower than inserts` exception.",
				Computed:            true,
			},
			"max_replicated_merges_in_queue": schema.Int64Attribute{
				MarkdownDescription: "Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time.",
				Computed:            true,
			},
			"number_of_free_entries_in_pool_to_lower_max_size_of_merge": schema.Int64Attribute{
				MarkdownDescription: "Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges.",
				Computed:            true,
			},
			"max_bytes_to_merge_at_min_space_in_pool": schema.Int64Attribute{
				MarkdownDescription: "Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum.",
				Computed:            true,
			},
			"max_bytes_to_merge_at_max_space_in_pool": schema.Int64Attribute{
				MarkdownDescription: "The maximum total parts size (in bytes) to be merged into one part, if there are enough resources available. Roughly corresponds to the maximum possible part size created by an automatic background merge.",
				Computed:            true,
			},
			"min_bytes_for_wide_part": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of bytes in a data part that can be stored in Wide format. You can set one, both or none of these settings.",
				Computed:            true,
			},
			"min_rows_for_wide_part": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of rows in a data part that can be stored in Wide format. You can set one, both or none of these settings.",
				Computed:            true,
			},
			"ttl_only_drop_parts": schema.BoolAttribute{
				MarkdownDescription: "Enables zero-copy replication when a replica is located on a remote filesystem.",
				Computed:            true,
			},
			"merge_with_ttl_timeout": schema.Int64Attribute{
				MarkdownDescription: "Minimum delay in seconds before repeating a merge with delete TTL. Default value: 14400 seconds (4 hours).",
				Computed:            true,
			},
			"merge_with_recompression_ttl_timeout": schema.Int64Attribute{
				MarkdownDescription: "Minimum delay in seconds before repeating a merge with recompression TTL. Default value: 14400 seconds (4 hours).",
				Computed:            true,
			},
			"max_parts_in_total": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of parts in all partitions.",
				Computed:            true,
			},
			"max_number_of_merges_with_ttl_in_pool": schema.Int64Attribute{
				MarkdownDescription: "When there is more than specified number of merges with TTL entries in pool, do not assign new merge with TTL.",
				Computed:            true,
			},
			"materialize_ttl_recalculate_only": schema.BoolAttribute{
				MarkdownDescription: "Only recalculate ttl info when **MATERIALIZE TTL**.",
				Computed:            true,
			},
			"cleanup_delay_period": schema.Int64Attribute{
				MarkdownDescription: "Minimum period to clean old queue logs, blocks hashes and parts.",
				Computed:            true,
			},
			"number_of_free_entries_in_pool_to_execute_mutation": schema.Int64Attribute{
				MarkdownDescription: "When there is less than specified number of free entries in pool, do not execute part mutations. This is to leave free threads for regular merges and avoid `Too many parts`. Default value: 20.",
				Computed:            true,
			},
			"max_avg_part_size_for_too_many_parts": schema.Int64Attribute{
				MarkdownDescription: "The `too many parts` check will be active only if the average part size is not larger than the specified threshold. This allows large tables if parts are successfully merged.",
				Computed:            true,
			},
			"min_age_to_force_merge_seconds": schema.Int64Attribute{
				MarkdownDescription: "Merge parts if every part in the range is older than the value of `min_age_to_force_merge_seconds`.",
				Computed:            true,
			},
			"min_age_to_force_merge_on_partition_only": schema.BoolAttribute{
				MarkdownDescription: "Whether min_age_to_force_merge_seconds should be applied only on the entire partition and not on subset.",
				Computed:            true,
			},
			"merge_selecting_sleep_ms": schema.Int64Attribute{
				MarkdownDescription: "Sleep time for merge selecting when no part is selected. Lower values increase ZooKeeper requests in large clusters.",
				Computed:            true,
			},
			"merge_max_block_size": schema.Int64Attribute{
				MarkdownDescription: "The number of rows that are read from the merged parts into memory. Default value: 8192.",
				Computed:            true,
			},
			"deduplicate_merge_projection_mode": schema.StringAttribute{
				MarkdownDescription: "Determines the behavior of background merges for MergeTree tables with projections.",
				Computed:            true,
			},
			"lightweight_mutation_projection_mode": schema.StringAttribute{
				MarkdownDescription: "Determines the behavior of lightweight deletes for MergeTree tables with projections.",
				Computed:            true,
			},
			"check_sample_column_is_correct": schema.BoolAttribute{
				MarkdownDescription: "Enables the check at table creation that the sampling column type is correct. Default value: true.",
				Computed:            true,
			},
			"max_merge_selecting_sleep_ms": schema.Int64Attribute{
				MarkdownDescription: "Maximum sleep time for merge selecting. Default value: 60000 milliseconds (60 seconds).",
				Computed:            true,
			},
			"max_cleanup_delay_period": schema.Int64Attribute{
				MarkdownDescription: "Maximum period to clean old queue logs, blocks hashes and parts. Default value: 300 seconds.",
				Computed:            true,
			},
		},
	}
}

func DataSourceKafkaSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Kafka connection configuration.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"security_protocol": schema.StringAttribute{
				MarkdownDescription: "Security protocol used to connect to kafka server.",
				Computed:            true,
			},
			"sasl_mechanism": schema.StringAttribute{
				MarkdownDescription: "SASL mechanism used in kafka authentication.",
				Computed:            true,
			},
			"sasl_username": schema.StringAttribute{
				MarkdownDescription: "Username on kafka server.",
				Computed:            true,
			},
			"sasl_password": schema.StringAttribute{
				MarkdownDescription: "User password on kafka server.",
				Computed:            true,
				Sensitive:           true,
			},
			"enable_ssl_certificate_verification": schema.BoolAttribute{
				MarkdownDescription: "Enable verification of SSL certificates.",
				Computed:            true,
			},
			"max_poll_interval_ms": schema.Int64Attribute{
				MarkdownDescription: "Maximum allowed time between calls to consume messages. If exceeded, consumer is considered failed.",
				Computed:            true,
			},
			"session_timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "Client group session and failure detection timeout.",
				Computed:            true,
			},
			"debug": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of debug contexts to enable.",
				Computed:            true,
			},
			"auto_offset_reset": schema.StringAttribute{
				MarkdownDescription: "Action when no initial offset: 'smallest','earliest','largest','latest','error'.",
				Computed:            true,
			},
		},
	}
}

func DataSourceRabbitMQSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "RabbitMQ connection configuration.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "RabbitMQ username.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "RabbitMQ user password.",
				Computed:            true,
				Sensitive:           true,
			},
			"vhost": schema.StringAttribute{
				MarkdownDescription: "RabbitMQ vhost. Default: `\\`.",
				Computed:            true,
			},
		},
	}
}

func DataSourceJdbcBridgeSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "JDBC bridge configuration.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Host of jdbc bridge.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port of jdbc bridge. Default value: 9019.",
				Computed:            true,
			},
		},
	}
}
