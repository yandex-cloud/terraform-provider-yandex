package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBClickHouseClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clickhouse": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_level":                       {Type: schema.TypeString, Optional: true},
									"max_connections":                 {Type: schema.TypeInt, Optional: true},
									"max_concurrent_queries":          {Type: schema.TypeInt, Optional: true},
									"keep_alive_timeout":              {Type: schema.TypeInt, Optional: true},
									"uncompressed_cache_size":         {Type: schema.TypeInt, Optional: true},
									"mark_cache_size":                 {Type: schema.TypeInt, Optional: true},
									"max_table_size_to_drop":          {Type: schema.TypeInt, Optional: true},
									"max_partition_size_to_drop":      {Type: schema.TypeInt, Optional: true},
									"timezone":                        {Type: schema.TypeString, Optional: true},
									"geobase_uri":                     {Type: schema.TypeString, Optional: true},
									"query_log_retention_size":        {Type: schema.TypeInt, Optional: true},
									"query_log_retention_time":        {Type: schema.TypeInt, Optional: true},
									"query_thread_log_enabled":        {Type: schema.TypeBool, Optional: true},
									"query_thread_log_retention_size": {Type: schema.TypeInt, Optional: true},
									"query_thread_log_retention_time": {Type: schema.TypeInt, Optional: true},
									"part_log_retention_size":         {Type: schema.TypeInt, Optional: true},
									"part_log_retention_time":         {Type: schema.TypeInt, Optional: true},
									"metric_log_enabled":              {Type: schema.TypeBool, Optional: true},
									"metric_log_retention_size":       {Type: schema.TypeInt, Optional: true},
									"metric_log_retention_time":       {Type: schema.TypeInt, Optional: true},
									"trace_log_enabled":               {Type: schema.TypeBool, Optional: true},
									"trace_log_retention_size":        {Type: schema.TypeInt, Optional: true},
									"trace_log_retention_time":        {Type: schema.TypeInt, Optional: true},
									"text_log_enabled":                {Type: schema.TypeBool, Optional: true},
									"text_log_retention_size":         {Type: schema.TypeInt, Optional: true},
									"text_log_retention_time":         {Type: schema.TypeInt, Optional: true},
									"text_log_level":                  {Type: schema.TypeString, Optional: true},
									"background_pool_size":            {Type: schema.TypeInt, Optional: true},
									"background_schedule_pool_size":   {Type: schema.TypeInt, Optional: true},

									"merge_tree": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replicated_deduplication_window":                           {Type: schema.TypeInt, Optional: true},
												"replicated_deduplication_window_seconds":                   {Type: schema.TypeInt, Optional: true},
												"parts_to_delay_insert":                                     {Type: schema.TypeInt, Optional: true},
												"parts_to_throw_insert":                                     {Type: schema.TypeInt, Optional: true},
												"max_replicated_merges_in_queue":                            {Type: schema.TypeInt, Optional: true},
												"number_of_free_entries_in_pool_to_lower_max_size_of_merge": {Type: schema.TypeInt, Optional: true},
												"max_bytes_to_merge_at_min_space_in_pool":                   {Type: schema.TypeInt, Optional: true},
											},
										},
									},
									"kafka": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"security_protocol": {Type: schema.TypeString, Optional: true},
												"sasl_mechanism":    {Type: schema.TypeString, Optional: true},
												"sasl_username":     {Type: schema.TypeString, Optional: true},
												"sasl_password":     {Type: schema.TypeString, Optional: true, Sensitive: true},
											},
										},
									},
									"kafka_topic": {
										Type:     schema.TypeList,
										MinItems: 0,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {Type: schema.TypeString, Required: true},
												"settings": {Type: schema.TypeList,
													MinItems: 0,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"security_protocol": {Type: schema.TypeString, Optional: true},
															"sasl_mechanism":    {Type: schema.TypeString, Optional: true},
															"sasl_username":     {Type: schema.TypeString, Optional: true},
															"sasl_password":     {Type: schema.TypeString, Optional: true, Sensitive: true},
														},
													},
												},
											},
										},
									},
									"rabbitmq": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {Type: schema.TypeString, Optional: true},
												"password": {Type: schema.TypeString, Optional: true, Sensitive: true},
											},
										},
									},
									"compression": {
										Type:     schema.TypeList,
										MinItems: 0,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"method":              {Type: schema.TypeString, Required: true},
												"min_part_size":       {Type: schema.TypeInt, Required: true},
												"min_part_size_ratio": {Type: schema.TypeFloat, Required: true},
											},
										},
									},
									"graphite_rollup": {
										Type:     schema.TypeList,
										MinItems: 0,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {Type: schema.TypeString, Required: true},
												"pattern": {
													Type:     schema.TypeList,
													MinItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"regexp":   {Type: schema.TypeString, Optional: true},
															"function": {Type: schema.TypeString, Required: true},
															"retention": {
																Type:     schema.TypeList,
																MinItems: 0,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"age":       {Type: schema.TypeInt, Required: true},
																		"precision": {Type: schema.TypeInt, Required: true},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      clickHouseUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      clickHouseUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"settings": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"readonly":                      {Type: schema.TypeInt, Optional: true, Computed: true},
									"allow_ddl":                     {Type: schema.TypeBool, Optional: true, Computed: true},
									"insert_quorum":                 {Type: schema.TypeInt, Optional: true, Computed: true},
									"connect_timeout":               {Type: schema.TypeInt, Optional: true, Computed: true},
									"receive_timeout":               {Type: schema.TypeInt, Optional: true, Computed: true},
									"send_timeout":                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"insert_quorum_timeout":         {Type: schema.TypeInt, Optional: true, Computed: true},
									"select_sequential_consistency": {Type: schema.TypeBool, Optional: true, Computed: true},
									"max_replica_delay_for_distributed_queries":          {Type: schema.TypeInt, Optional: true, Computed: true},
									"fallback_to_stale_replicas_for_distributed_queries": {Type: schema.TypeBool, Optional: true, Computed: true},
									"replication_alter_partitions_sync":                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"distributed_product_mode":                           {Type: schema.TypeString, Optional: true, Computed: true},
									"distributed_aggregation_memory_efficient":           {Type: schema.TypeBool, Optional: true, Computed: true},
									"distributed_ddl_task_timeout":                       {Type: schema.TypeInt, Optional: true, Computed: true},
									"skip_unavailable_shards":                            {Type: schema.TypeBool, Optional: true, Computed: true},
									"compile":                                            {Type: schema.TypeBool, Optional: true, Computed: true},
									"min_count_to_compile":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"compile_expressions":                                {Type: schema.TypeBool, Optional: true, Computed: true},
									"min_count_to_compile_expression":                    {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_block_size":                                     {Type: schema.TypeInt, Optional: true, Computed: true},
									"min_insert_block_size_rows":                         {Type: schema.TypeInt, Optional: true, Computed: true},
									"min_insert_block_size_bytes":                        {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_insert_block_size":                              {Type: schema.TypeInt, Optional: true, Computed: true},
									"min_bytes_to_use_direct_io":                         {Type: schema.TypeInt, Optional: true, Computed: true},
									"use_uncompressed_cache":                             {Type: schema.TypeBool, Optional: true, Computed: true},
									"merge_tree_max_rows_to_use_cache":                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"merge_tree_max_bytes_to_use_cache":                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"merge_tree_min_rows_for_concurrent_read":            {Type: schema.TypeInt, Optional: true, Computed: true},
									"merge_tree_min_bytes_for_concurrent_read":           {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_before_external_group_by":                 {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_before_external_sort":                     {Type: schema.TypeInt, Optional: true, Computed: true},
									"group_by_two_level_threshold":                       {Type: schema.TypeInt, Optional: true, Computed: true},
									"group_by_two_level_threshold_bytes":                 {Type: schema.TypeInt, Optional: true, Computed: true},
									"priority":                                           {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_threads":                                        {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_memory_usage":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_memory_usage_for_user":                          {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_network_bandwidth":                              {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_network_bandwidth_for_user":                     {Type: schema.TypeInt, Optional: true, Computed: true},
									"force_index_by_date":                                {Type: schema.TypeBool, Optional: true, Computed: true},
									"force_primary_key":                                  {Type: schema.TypeBool, Optional: true, Computed: true},
									"max_rows_to_read":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_to_read":                                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"read_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_to_group_by":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"group_by_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_to_sort":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_to_sort":                                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"sort_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true},
									"max_result_rows":                                    {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_result_bytes":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"result_overflow_mode":                               {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_in_distinct":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_in_distinct":                              {Type: schema.TypeInt, Optional: true, Computed: true},
									"distinct_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_to_transfer":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_to_transfer":                              {Type: schema.TypeInt, Optional: true, Computed: true},
									"transfer_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true},
									"max_execution_time":                                 {Type: schema.TypeInt, Optional: true, Computed: true},
									"timeout_overflow_mode":                              {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_in_set":                                    {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_in_set":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"set_overflow_mode":                                  {Type: schema.TypeString, Optional: true, Computed: true},
									"max_rows_in_join":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_bytes_in_join":                                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"join_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true},
									"max_columns_to_read":                                {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_temporary_columns":                              {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_temporary_non_const_columns":                    {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_query_size":                                     {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_ast_depth":                                      {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_ast_elements":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
									"max_expanded_ast_elements":                          {Type: schema.TypeInt, Optional: true, Computed: true},
									"min_execution_speed":                                {Type: schema.TypeInt, Optional: true, Computed: true},
									"min_execution_speed_bytes":                          {Type: schema.TypeInt, Optional: true, Computed: true},
									"count_distinct_implementation":                      {Type: schema.TypeString, Optional: true, Computed: true},
									"input_format_values_interpret_expressions":          {Type: schema.TypeBool, Optional: true, Computed: true},
									"input_format_defaults_for_omitted_fields":           {Type: schema.TypeBool, Optional: true, Computed: true},
									"output_format_json_quote_64bit_integers":            {Type: schema.TypeBool, Optional: true, Computed: true},
									"output_format_json_quote_denormals":                 {Type: schema.TypeBool, Optional: true, Computed: true},
									"low_cardinality_allow_in_native_format":             {Type: schema.TypeBool, Optional: true, Computed: true},
									"empty_result_for_aggregation_by_empty_set":          {Type: schema.TypeBool, Optional: true, Computed: true},
									"joined_subquery_requires_alias":                     {Type: schema.TypeBool, Optional: true, Computed: true},
									"join_use_nulls":                                     {Type: schema.TypeBool, Optional: true, Computed: true},
									"transform_null_in":                                  {Type: schema.TypeBool, Optional: true, Computed: true},
									"http_connection_timeout":                            {Type: schema.TypeInt, Optional: true, Computed: true},
									"http_receive_timeout":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"http_send_timeout":                                  {Type: schema.TypeInt, Optional: true, Computed: true},
									"enable_http_compression":                            {Type: schema.TypeBool, Optional: true, Computed: true},
									"send_progress_in_http_headers":                      {Type: schema.TypeBool, Optional: true, Computed: true},
									"http_headers_progress_interval":                     {Type: schema.TypeInt, Optional: true, Computed: true},
									"add_http_cors_header":                               {Type: schema.TypeBool, Optional: true, Computed: true},
									"quota_mode":                                         {Type: schema.TypeString, Optional: true, Computed: true},
								},
							},
						},
						"quota": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Set:      clickHouseUserQuotaHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval_duration": {Type: schema.TypeInt, Required: true},
									"queries":           {Type: schema.TypeInt, Optional: true, Computed: true},
									"errors":            {Type: schema.TypeInt, Optional: true, Computed: true},
									"result_rows":       {Type: schema.TypeInt, Optional: true, Computed: true},
									"read_rows":         {Type: schema.TypeInt, Optional: true, Computed: true},
									"execution_time":    {Type: schema.TypeInt, Optional: true, Computed: true},
								},
							},
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      clickHouseDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"shard_group": {
				Type:     schema.TypeList,
				MinItems: 0,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"shard_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"format_schema": {
				Type:     schema.TypeList,
				MinItems: 0,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ml_model": {
				Type:     schema.TypeList,
				MinItems: 0,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_window_start": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"web_sql": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"data_lens": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"metrika": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"serverless": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"data_transfer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"yandex_query": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"zookeeper": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"sql_user_management": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"sql_database_management": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"embedded_keeper": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"cloud_storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceYandexMDBClickHouseClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.ClickhouseClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source ClickHouse Cluster by name: %v", err)
		}

		d.Set("cluster_id", clusterID)
	}

	d.SetId(clusterID)
	return resourceYandexMDBClickHouseClusterRead(d, meta)
}
