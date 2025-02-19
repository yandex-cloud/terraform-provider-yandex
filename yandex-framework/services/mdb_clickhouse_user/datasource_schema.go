package mdb_clickhouse_user

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func DataSourceUserSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a ClickHouse user within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-clickhouse/concepts).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the ClickHouse cluster. Provided by the client when the user is created.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the ClickHouse user. Provided by the client when the user is created.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": DataSourcePermissionSchema(),
			"quota":      DataSourceQuotasSchema(),
			"settings":   DataSourceSettingsSchema(),
		},
	}
}

func DataSourcePermissionSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database_name": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}

func DataSourceQuotasSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval_duration": schema.Int64Attribute{
					Computed: true,
				},
				"queries": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"errors": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"result_rows": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"read_rows": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"execution_time": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func DataSourceSettingsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"readonly": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"allow_ddl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"insert_quorum": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"connect_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"receive_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"send_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"insert_quorum_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"select_sequential_consistency": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_replica_delay_for_distributed_queries": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"fallback_to_stale_replicas_for_distributed_queries": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"replication_alter_partitions_sync": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"distributed_product_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"distributed_aggregation_memory_efficient": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"distributed_ddl_task_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"skip_unavailable_shards": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"compile_expressions": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"min_count_to_compile_expression": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_block_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_insert_block_size_rows": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_insert_block_size_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_insert_block_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_bytes_to_use_direct_io": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"use_uncompressed_cache": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"merge_tree_max_rows_to_use_cache": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"merge_tree_max_bytes_to_use_cache": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"merge_tree_min_rows_for_concurrent_read": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"merge_tree_min_bytes_for_concurrent_read": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_before_external_group_by": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_before_external_sort": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"group_by_two_level_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"group_by_two_level_threshold_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_threads": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_memory_usage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_memory_usage_for_user": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_network_bandwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_network_bandwidth_for_user": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"force_index_by_date": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"force_primary_key": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_to_read": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_to_read": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"read_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_to_group_by": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"group_by_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_to_sort": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_to_sort": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"sort_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_result_rows": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_result_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"result_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_in_distinct": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_in_distinct": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"distinct_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_to_transfer": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_to_transfer": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"transfer_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_execution_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"timeout_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_in_set": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_in_set": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"set_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_rows_in_join": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_bytes_in_join": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"join_overflow_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_columns_to_read": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_temporary_columns": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_temporary_non_const_columns": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_query_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_ast_depth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_ast_elements": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_expanded_ast_elements": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_execution_speed": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_execution_speed_bytes": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"count_distinct_implementation": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"input_format_values_interpret_expressions": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"input_format_defaults_for_omitted_fields": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"output_format_json_quote_64bit_integers": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"output_format_json_quote_denormals": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"low_cardinality_allow_in_native_format": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"empty_result_for_aggregation_by_empty_set": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"joined_subquery_requires_alias": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"join_use_nulls": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"transform_null_in": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"http_connection_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"http_receive_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"http_send_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"enable_http_compression": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"send_progress_in_http_headers": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"http_headers_progress_interval": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"add_http_cors_header": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"quota_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"max_concurrent_queries_for_user": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"memory_profiler_step": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"memory_profiler_sample_probability": schema.Float64Attribute{
				Optional: true,
				Computed: true,
			},
			"insert_null_as_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"allow_suspicious_low_cardinality_types": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"connect_timeout_with_failover": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"allow_introspection_functions": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"async_insert": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"async_insert_threads": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"wait_for_async_insert": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"wait_for_async_insert_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"async_insert_max_data_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"async_insert_busy_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"async_insert_stale_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"timeout_before_checking_execution_speed": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"cancel_http_readonly_queries_on_client_close": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"flatten_nested": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_http_get_redirects": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"input_format_import_nested_json": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"input_format_parallel_parsing": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_final_threads": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_read_buffer_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"local_filesystem_read_method": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"remote_filesystem_read_method": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"insert_keeper_max_retries": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_temporary_data_on_disk_size_for_user": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_temporary_data_on_disk_size_for_query": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"max_parser_depth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"memory_overcommit_ratio_denominator": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"memory_overcommit_ratio_denominator_for_user": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"memory_usage_overcommit_max_wait_microseconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"log_query_threads": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_insert_threads": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"use_hedged_requests": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"idle_connection_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"hedged_connection_timeout_ms": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"load_balancing": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"prefer_localhost_replica": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"date_time_input_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"date_time_output_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"format_regexp": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"format_regexp_skip_unmatched": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"input_format_with_names_use_header": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"input_format_null_as_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"insert_quorum_parallel": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"max_partitions_per_insert_block": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"deduplicate_blocks_in_dependent_materialized_views": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"any_join_distinct_right_table_keys": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"join_algorithm": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}

}
