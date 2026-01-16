package mdb_clickhouse_cluster_v2

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	clickhouse "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

var (
	clusterId = "cluster-id"

	minimalConfig = types.ObjectValueMust(
		models.ClusterAttrTypes,
		map[string]attr.Value{
			"id":                        types.StringValue(clusterId),
			"folder_id":                 types.StringValue("test-folder-1"),
			"created_at":                types.StringNull(),
			"name":                      types.StringValue("test-cluster"),
			"description":               types.StringNull(),
			"labels":                    types.MapNull(types.StringType),
			"environment":               types.StringValue("PRESTABLE"),
			"network_id":                types.StringValue("test-network"),
			"version":                   types.StringValue("24.8"),
			"maintenance_window":        types.ObjectNull(models.MaintenanceWindowAttrTypes),
			"clickhouse":                types.ObjectNull(models.ClickhouseAttrTypes),
			"zookeeper":                 types.ObjectNull(models.ZookeeperAttrTypes),
			"security_group_ids":        types.SetNull(types.StringType),
			"backup_window_start":       types.ObjectNull(models.BackupWindowStartAttrTypes),
			"access":                    types.ObjectNull(models.AccessAttrTypes),
			"cloud_storage":             types.ObjectNull(models.CloudStorageAttrTypes),
			"sql_database_management":   types.BoolNull(),
			"sql_user_management":       types.BoolNull(),
			"admin_password":            types.StringNull(),
			"embedded_keeper":           types.BoolNull(),
			"backup_retain_period_days": types.Int64Null(),
			"deletion_protection":       types.BoolNull(),
			"service_account_id":        types.StringNull(),
			"disk_encryption_key_id":    types.StringNull(),
			"ml_model":                  types.SetNull(types.ObjectType{AttrTypes: models.MLModelAttrTypes}),
			"format_schema":             types.SetNull(types.ObjectType{AttrTypes: models.FormatSchemaAttrTypes}),
			"shards":                    types.MapNull(types.ObjectType{AttrTypes: models.ShardAttrTypes}),
			"shard_group":               types.ListNull(types.ObjectType{AttrTypes: models.ShardGroupAttrTypes}),
			"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
				"host1": types.StringValue("host1"),
				"host2": types.StringValue("host2"),
			}),
			"timeouts":                 timeouts.Value{},
			"copy_schema_on_new_hosts": types.BoolNull(),
		},
	)

	maximalConfig = types.ObjectValueMust(
		models.ClusterAttrTypes,
		map[string]attr.Value{
			"id":          types.StringValue(clusterId),
			"folder_id":   types.StringValue("test-folder-2"),
			"created_at":  types.StringNull(),
			"name":        types.StringValue("test-cluster"),
			"description": types.StringValue("test-description"),
			"labels": types.MapValueMust(types.StringType, map[string]attr.Value{
				"key": types.StringValue("value"),
			}),
			"environment": types.StringValue("PRESTABLE"),
			"network_id":  types.StringValue("test-network"),
			"version":     types.StringValue("25.3"),
			"maintenance_window": types.ObjectValueMust(
				mdbcommon.MaintenanceWindowType.AttrTypes,
				map[string]attr.Value{
					"type": types.StringValue("ANYTIME"),
					"day":  types.StringValue("MON"),
					"hour": types.Int64Value(1),
				},
			),
			"clickhouse": types.ObjectValueMust(
				models.ClickhouseAttrTypes,
				map[string]attr.Value{
					"config": types.ObjectValueMust(
						models.ClickhouseConfigAttrTypes,
						map[string]attr.Value{
							"log_level":            types.StringValue("TRACE"),
							"background_pool_size": types.Int64Value(32),
							"background_merges_mutations_concurrency_ratio": types.Int64Value(64),
							"background_schedule_pool_size":                 types.Int64Value(64),
							"background_fetches_pool_size":                  types.Int64Value(48),
							"background_move_pool_size":                     types.Int64Value(32),
							"background_distributed_schedule_pool_size":     types.Int64Value(128),
							"background_buffer_flush_schedule_pool_size":    types.Int64Value(64),
							"background_message_broker_schedule_pool_size":  types.Int64Value(96),
							"background_common_pool_size":                   types.Int64Value(32),
							"dictionaries_lazy_load":                        types.BoolValue(true),
							"query_log_retention_size":                      types.Int64Value(16),
							"query_log_retention_time":                      types.Int64Value(32),
							"query_thread_log_enabled":                      types.BoolValue(false),
							"query_thread_log_retention_size":               types.Int64Value(64),
							"query_thread_log_retention_time":               types.Int64Value(32),
							"part_log_retention_size":                       types.Int64Value(256),
							"part_log_retention_time":                       types.Int64Value(128),
							"metric_log_enabled":                            types.BoolValue(true),
							"metric_log_retention_size":                     types.Int64Value(64),
							"metric_log_retention_time":                     types.Int64Value(32),
							"trace_log_enabled":                             types.BoolValue(false),
							"trace_log_retention_size":                      types.Int64Value(128),
							"trace_log_retention_time":                      types.Int64Value(64),
							"text_log_enabled":                              types.BoolValue(true),
							"text_log_retention_size":                       types.Int64Value(256),
							"text_log_retention_time":                       types.Int64Value(32),
							"text_log_level":                                types.StringValue("DEBUG"),
							"opentelemetry_span_log_enabled":                types.BoolValue(false),
							"opentelemetry_span_log_retention_size":         types.Int64Null(),
							"opentelemetry_span_log_retention_time":         types.Int64Null(),
							"query_views_log_enabled":                       types.BoolValue(false),
							"query_views_log_retention_size":                types.Int64Null(),
							"query_views_log_retention_time":                types.Int64Null(),
							"asynchronous_metric_log_enabled":               types.BoolValue(true),
							"asynchronous_metric_log_retention_size":        types.Int64Value(64),
							"asynchronous_metric_log_retention_time":        types.Int64Value(16),
							"session_log_enabled":                           types.BoolValue(true),
							"session_log_retention_size":                    types.Int64Value(128),
							"session_log_retention_time":                    types.Int64Value(32),
							"zookeeper_log_enabled":                         types.BoolValue(false),
							"zookeeper_log_retention_size":                  types.Int64Null(),
							"zookeeper_log_retention_time":                  types.Int64Null(),
							"asynchronous_insert_log_enabled":               types.BoolValue(false),
							"asynchronous_insert_log_retention_size":        types.Int64Null(),
							"asynchronous_insert_log_retention_time":        types.Int64Null(),
							"processors_profile_log_enabled":                types.BoolValue(true),
							"processors_profile_log_retention_size":         types.Int64Value(256),
							"processors_profile_log_retention_time":         types.Int64Value(128),
							"error_log_enabled":                             types.BoolValue(true),
							"error_log_retention_size":                      types.Int64Value(512),
							"error_log_retention_time":                      types.Int64Value(256),
							"access_control_improvements": types.ObjectValueMust(
								models.AccessControlImprovementsAttrTypes,
								map[string]attr.Value{
									"select_from_system_db_requires_grant":          types.BoolValue(true),
									"select_from_information_schema_requires_grant": types.BoolValue(false),
								},
							),
							"max_connections":                         types.Int64Value(1024),
							"max_concurrent_queries":                  types.Int64Value(512),
							"max_table_size_to_drop":                  types.Int64Value(256),
							"max_partition_size_to_drop":              types.Int64Value(128),
							"keep_alive_timeout":                      types.Int64Value(64),
							"uncompressed_cache_size":                 types.Int64Value(32),
							"timezone":                                types.StringValue("MSK"),
							"geobase_enabled":                         types.BoolValue(false),
							"geobase_uri":                             types.StringValue("geobase_uri"),
							"default_database":                        types.StringValue("default_database"),
							"total_memory_profiler_step":              types.Int64Value(16),
							"total_memory_tracker_sample_probability": types.Float64Value(10),
							"async_insert_threads":                    types.Int64Value(128),
							"backup_threads":                          types.Int64Value(512),
							"restore_threads":                         types.Int64Value(64),
							"merge_tree": types.ObjectValueMust(
								models.MergeTreeConfigAttrTypes,
								map[string]attr.Value{
									"replicated_deduplication_window":                           types.Int64Value(100),
									"replicated_deduplication_window_seconds":                   types.Int64Value(200),
									"parts_to_delay_insert":                                     types.Int64Value(10),
									"parts_to_throw_insert":                                     types.Int64Value(20),
									"max_replicated_merges_in_queue":                            types.Int64Value(30),
									"number_of_free_entries_in_pool_to_lower_max_size_of_merge": types.Int64Value(5),
									"max_bytes_to_merge_at_min_space_in_pool":                   types.Int64Value(111),
									"max_bytes_to_merge_at_max_space_in_pool":                   types.Int64Value(222),
									"inactive_parts_to_delay_insert":                            types.Int64Value(7),
									"inactive_parts_to_throw_insert":                            types.Int64Value(8),
									"min_bytes_for_wide_part":                                   types.Int64Value(1024),
									"min_rows_for_wide_part":                                    types.Int64Value(2048),
									"ttl_only_drop_parts":                                       types.BoolValue(true),
									"merge_with_ttl_timeout":                                    types.Int64Value(300),
									"merge_with_recompression_ttl_timeout":                      types.Int64Value(400),
									"max_parts_in_total":                                        types.Int64Value(500),
									"max_number_of_merges_with_ttl_in_pool":                     types.Int64Value(3),
									"cleanup_delay_period":                                      types.Int64Value(600),
									"number_of_free_entries_in_pool_to_execute_mutation":        types.Int64Value(4),
									"max_avg_part_size_for_too_many_parts":                      types.Int64Value(700),
									"min_age_to_force_merge_seconds":                            types.Int64Value(800),
									"min_age_to_force_merge_on_partition_only":                  types.BoolValue(false),
									"merge_selecting_sleep_ms":                                  types.Int64Value(900),
									"check_sample_column_is_correct":                            types.BoolValue(true),
									"merge_max_block_size":                                      types.Int64Value(1000),
									"max_merge_selecting_sleep_ms":                              types.Int64Value(1100),
									"max_cleanup_delay_period":                                  types.Int64Value(1200),
									"deduplicate_merge_projection_mode":                         types.StringValue("DEDUPLICATE_MERGE_PROJECTION_MODE_DROP"),
									"lightweight_mutation_projection_mode":                      types.StringValue("LIGHTWEIGHT_MUTATION_PROJECTION_MODE_REBUILD"),
									"materialize_ttl_recalculate_only":                          types.BoolValue(true),
									"fsync_after_insert":                                        types.BoolValue(true),
									"fsync_part_directory":                                      types.BoolValue(false),
									"min_compressed_bytes_to_fsync_after_fetch":                 types.Int64Value(1300),
									"min_compressed_bytes_to_fsync_after_merge":                 types.Int64Value(1400),
									"min_rows_to_fsync_after_merge":                             types.Int64Value(1500),
								},
							),
							"compression": types.ListValueMust(
								types.ObjectType{AttrTypes: models.CompressionAttrTypes},
								[]attr.Value{
									types.ObjectValueMust(
										models.CompressionAttrTypes,
										map[string]attr.Value{
											"method":              types.StringValue("LZ4"),
											"min_part_size":       types.Int64Value(1024),
											"min_part_size_ratio": types.NumberValue(big.NewFloat(0.5)),
											"level":               types.Int64Value(3),
										},
									),
									types.ObjectValueMust(
										models.CompressionAttrTypes,
										map[string]attr.Value{
											"method":              types.StringValue("ZSTD"),
											"min_part_size":       types.Int64Value(2048),
											"min_part_size_ratio": types.NumberValue(big.NewFloat(0.75)),
											"level":               types.Int64Value(5),
										},
									),
								},
							),
							"graphite_rollup": types.ListValueMust(
								types.ObjectType{AttrTypes: models.GraphiteRollupAttrTypes},
								[]attr.Value{
									types.ObjectValueMust(
										models.GraphiteRollupAttrTypes,
										map[string]attr.Value{
											"name": types.StringValue("rollup_1"),
											"patterns": types.ListValueMust(
												types.ObjectType{AttrTypes: models.PatternAttrTypes},
												[]attr.Value{
													types.ObjectValueMust(
														models.PatternAttrTypes,
														map[string]attr.Value{
															"regexp":   types.StringValue("^cpu\\."),
															"function": types.StringValue("max"),
															"retention": types.ListValueMust(
																types.ObjectType{AttrTypes: models.RetentionAttrTypes},
																[]attr.Value{
																	types.ObjectValueMust(
																		models.RetentionAttrTypes,
																		map[string]attr.Value{
																			"age":       types.Int64Value(60),
																			"precision": types.Int64Value(10),
																		},
																	),
																	types.ObjectValueMust(
																		models.RetentionAttrTypes,
																		map[string]attr.Value{
																			"age":       types.Int64Value(600),
																			"precision": types.Int64Value(60),
																		},
																	),
																},
															),
														},
													),
													types.ObjectValueMust(
														models.PatternAttrTypes,
														map[string]attr.Value{
															"regexp":   types.StringValue("^disk\\."),
															"function": types.StringValue("avg"),
															"retention": types.ListValueMust(
																types.ObjectType{AttrTypes: models.RetentionAttrTypes},
																[]attr.Value{
																	types.ObjectValueMust(
																		models.RetentionAttrTypes,
																		map[string]attr.Value{
																			"age":       types.Int64Value(300),
																			"precision": types.Int64Value(30),
																		},
																	),
																},
															),
														},
													),
												},
											),
											"path_column_name":    types.StringValue("path"),
											"time_column_name":    types.StringValue("time"),
											"value_column_name":   types.StringValue("value"),
											"version_column_name": types.StringValue("version"),
										},
									),
									types.ObjectValueMust(
										models.GraphiteRollupAttrTypes,
										map[string]attr.Value{
											"name": types.StringValue("rollup_2"),
											"patterns": types.ListValueMust(
												types.ObjectType{AttrTypes: models.PatternAttrTypes},
												[]attr.Value{
													types.ObjectValueMust(
														models.PatternAttrTypes,
														map[string]attr.Value{
															"regexp":   types.StringValue(".*"),
															"function": types.StringValue("sum"),
															"retention": types.ListValueMust(
																types.ObjectType{AttrTypes: models.RetentionAttrTypes},
																[]attr.Value{
																	types.ObjectValueMust(
																		models.RetentionAttrTypes,
																		map[string]attr.Value{
																			"age":       types.Int64Value(120),
																			"precision": types.Int64Value(30),
																		},
																	),
																},
															),
														},
													),
												},
											),
											"path_column_name":    types.StringValue("p"),
											"time_column_name":    types.StringValue("t"),
											"value_column_name":   types.StringValue("v"),
											"version_column_name": types.StringValue("ver"),
										},
									),
								},
							),
							"kafka": types.ObjectValueMust(
								models.KafkaAttrTypes,
								map[string]attr.Value{
									"security_protocol":                   types.StringValue("SECURITY_PROTOCOL_PLAINTEXT"),
									"sasl_mechanism":                      types.StringValue("SASL_MECHANISM_PLAIN"),
									"sasl_username":                       types.StringValue("kafka-user"),
									"sasl_password":                       types.StringValue("kafka-pass"),
									"enable_ssl_certificate_verification": types.BoolValue(true),
									"max_poll_interval_ms":                types.Int64Value(300000),
									"session_timeout_ms":                  types.Int64Value(60000),
									"debug":                               types.StringValue("DEBUG_ALL"),
									"auto_offset_reset":                   types.StringValue("AUTO_OFFSET_RESET_EARLIEST"),
								},
							),
							"rabbitmq": types.ObjectValueMust(
								models.RabbitmqAttrTypes,
								map[string]attr.Value{
									"username": types.StringValue("rmq-user"),
									"password": types.StringValue("rmq-pass"),
									"vhost":    types.StringValue("rmq-vhost"),
								},
							),
							"query_masking_rules": types.ListValueMust(
								types.ObjectType{AttrTypes: models.QueryMaskingRuleAttrTypes},
								[]attr.Value{
									types.ObjectValueMust(
										models.QueryMaskingRuleAttrTypes,
										map[string]attr.Value{
											"name":    types.StringValue("mask_passwords"),
											"regexp":  types.StringValue("(?i)password\\s*=?\\s*'[^']*'"),
											"replace": types.StringValue("password='***'"),
										},
									),
									types.ObjectValueMust(
										models.QueryMaskingRuleAttrTypes,
										map[string]attr.Value{
											"name":    types.StringValue("mask_tokens"),
											"regexp":  types.StringValue("(?i)token\\s*=?\\s*'[^']*'"),
											"replace": types.StringValue("token='***'"),
										},
									),
								},
							),
							"query_cache": types.ObjectValueMust(
								models.QueryCacheAttrTypes,
								map[string]attr.Value{
									"max_size_in_bytes":       types.Int64Value(1_000_000),
									"max_entries":             types.Int64Value(10_000),
									"max_entry_size_in_bytes": types.Int64Value(50_000),
									"max_entry_size_in_rows":  types.Int64Value(1_000),
								},
							),
							"jdbc_bridge": types.ObjectValueMust(
								models.JdbcBridgeAttrTypes,
								map[string]attr.Value{
									"host": types.StringValue("jdbc-bridge-host"),
									"port": types.Int64Value(9019),
								},
							),
							"mysql_protocol": types.BoolValue(true),
							"custom_macros": types.ListValueMust(
								types.ObjectType{AttrTypes: models.MacroAttrTypes},
								[]attr.Value{
									types.ObjectValueMust(
										models.MacroAttrTypes,
										map[string]attr.Value{
											"name":  types.StringValue("shard"),
											"value": types.StringValue("shard_01"),
										},
									),
									types.ObjectValueMust(
										models.MacroAttrTypes,
										map[string]attr.Value{
											"name":  types.StringValue("replica"),
											"value": types.StringValue("replica_01"),
										},
									),
								},
							),
						},
					),
					"resources": types.ObjectValueMust(
						models.ResourcesAttrTypes,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("s2.micro"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(16),
						},
					),
				},
			),
			"zookeeper": types.ObjectValueMust(
				models.ZookeeperAttrTypes,
				map[string]attr.Value{
					"resources": types.ObjectValueMust(
						models.ResourcesAttrTypes,
						map[string]attr.Value{
							"resource_preset_id": types.StringValue("b3-c1-m4"),
							"disk_type_id":       types.StringValue("network-ssd"),
							"disk_size":          types.Int64Value(10),
						},
					),
				},
			),
			"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("test-sg"),
			}),
			"backup_window_start": types.ObjectValueMust(
				models.BackupWindowStartAttrTypes,
				map[string]attr.Value{
					"hours":   types.Int64Value(3),
					"minutes": types.Int64Value(15),
				},
			),
			"access": types.ObjectValueMust(
				models.AccessAttrTypes,
				map[string]attr.Value{
					"data_lens":     types.BoolValue(false),
					"data_transfer": types.BoolValue(true),
					"metrika":       types.BoolValue(false),
					"serverless":    types.BoolValue(true),
					"web_sql":       types.BoolValue(false),
					"yandex_query":  types.BoolValue(true),
				},
			),
			"cloud_storage": types.ObjectValueMust(
				models.CloudStorageAttrTypes,
				map[string]attr.Value{
					"enabled":             types.BoolValue(true),
					"move_factor":         types.NumberValue(big.NewFloat(3)),
					"data_cache_enabled":  types.BoolValue(true),
					"data_cache_max_size": types.Int64Value(32),
					"prefer_not_to_merge": types.BoolValue(false),
				},
			),
			"format_schema": types.SetValueMust(
				types.ObjectType{AttrTypes: models.FormatSchemaAttrTypes},
				[]attr.Value{
					types.ObjectValueMust(
						models.FormatSchemaAttrTypes,
						map[string]attr.Value{
							"name": types.StringValue("schema1"),
							"type": types.StringValue("FORMAT_SCHEMA_TYPE_PROTOBUF"),
							"uri":  types.StringValue("s3://bucket/schema1.proto"),
						},
					),
					types.ObjectValueMust(
						models.FormatSchemaAttrTypes,
						map[string]attr.Value{
							"name": types.StringValue("schema2"),
							"type": types.StringValue("FORMAT_SCHEMA_TYPE_CAPNPROTO"),
							"uri":  types.StringValue("s3://bucket/schema2.capnp"),
						},
					),
				},
			),
			"ml_model": types.SetValueMust(
				types.ObjectType{AttrTypes: models.MLModelAttrTypes},
				[]attr.Value{
					types.ObjectValueMust(
						models.MLModelAttrTypes,
						map[string]attr.Value{
							"name": types.StringValue("model1"),
							"type": types.StringValue("ML_MODEL_TYPE_CATBOOST"),
							"uri":  types.StringValue("s3://bucket/model1.cbm"),
						},
					),
					types.ObjectValueMust(
						models.MLModelAttrTypes,
						map[string]attr.Value{
							"name": types.StringValue("model2"),
							"type": types.StringValue("ML_MODEL_TYPE_CATBOOST"),
							"uri":  types.StringValue("s3://bucket/model2.tf"),
						},
					),
				},
			),
			"shard_group": types.ListValueMust(
				types.ObjectType{AttrTypes: models.ShardGroupAttrTypes},
				[]attr.Value{
					types.ObjectValueMust(
						models.ShardGroupAttrTypes,
						map[string]attr.Value{
							"name":        types.StringValue("group1"),
							"description": types.StringValue("first group"),
							"shard_names": types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("shard1"),
									types.StringValue("shard2"),
								},
							),
						},
					),
					types.ObjectValueMust(
						models.ShardGroupAttrTypes,
						map[string]attr.Value{
							"name":        types.StringValue("group2"),
							"description": types.StringValue("second group"),
							"shard_names": types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("shard3"),
								},
							),
						},
					),
				},
			),
			"shards": types.MapValueMust(
				types.ObjectType{AttrTypes: models.ShardAttrTypes},
				map[string]attr.Value{
					"shard1": types.ObjectValueMust(
						models.ShardAttrTypes,
						map[string]attr.Value{
							"weight": types.Int64Value(10),
							"resources": types.ObjectValueMust(
								models.ResourcesAttrTypes,
								map[string]attr.Value{
									"resource_preset_id": types.StringValue("s2.small"),
									"disk_type_id":       types.StringValue("network-ssd"),
									"disk_size":          types.Int64Value(20),
								},
							),
						},
					),
					"shard2": types.ObjectValueMust(
						models.ShardAttrTypes,
						map[string]attr.Value{
							"weight": types.Int64Value(20),
							"resources": types.ObjectValueMust(
								models.ResourcesAttrTypes,
								map[string]attr.Value{
									"resource_preset_id": types.StringValue("s2.medium"),
									"disk_type_id":       types.StringValue("network-ssd"),
									"disk_size":          types.Int64Value(40),
								},
							),
						},
					),
				},
			),
			"sql_database_management":   types.BoolValue(true),
			"sql_user_management":       types.BoolValue(true),
			"admin_password":            types.StringNull(),
			"embedded_keeper":           types.BoolValue(false),
			"backup_retain_period_days": types.Int64Value(14),
			"deletion_protection":       types.BoolValue(true),
			"service_account_id":        types.StringValue("sa-id"),
			"disk_encryption_key_id":    types.StringValue("test-key"),
			"hosts": types.MapValueMust(types.StringType, map[string]attr.Value{
				"host1": types.StringValue("host1"),
				"host2": types.StringValue("host2"),
			}),
			"timeouts":                 timeouts.Value{},
			"copy_schema_on_new_hosts": types.BoolNull(),
		},
	)
)

func TestYandexProvider_MDBClickHouseClusterPrepareCreateRequests(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname                     string
		reqVal                       types.Object
		hostSpecs                    []*clickhouse.HostSpec
		expectedClusterRequest       *clickhouse.CreateClusterRequest
		expectedFormatSchemaRequests []*clickhouse.CreateFormatSchemaRequest
		expectedMlModelRequests      []*clickhouse.CreateMlModelRequest
		expectedShardGroupRequests   []*clickhouse.CreateClusterShardGroupRequest
		expectedError                bool
	}{
		{
			testname: "MinimalConfigCheck",
			reqVal:   minimalConfig,
			hostSpecs: []*clickhouse.HostSpec{
				{ZoneId: "ru-central1-a", ShardName: "shard1", Type: clickhouse.Host_CLICKHOUSE},
				{ZoneId: "ru-central1-b", ShardName: "shard2", Type: clickhouse.Host_CLICKHOUSE},
			},
			expectedClusterRequest: &clickhouse.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "",
				Environment: clickhouse.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				FolderId:    "test-folder-1",
				HostSpecs: []*clickhouse.HostSpec{
					{ZoneId: "ru-central1-a", ShardName: "shard1", Type: clickhouse.Host_CLICKHOUSE},
					{ZoneId: "ru-central1-b", ShardName: "shard2", Type: clickhouse.Host_CLICKHOUSE},
				},
				ServiceAccountId: "",
				ConfigSpec: &clickhouse.ConfigSpec{
					Version:                "24.8",
					Clickhouse:             nil,
					Zookeeper:              nil,
					BackupWindowStart:      &timeofday.TimeOfDay{},
					Access:                 nil,
					CloudStorage:           nil,
					SqlDatabaseManagement:  nil,
					SqlUserManagement:      nil,
					AdminPassword:          "",
					EmbeddedKeeper:         nil,
					BackupRetainPeriodDays: nil,
				},
				DeletionProtection:  false,
				SecurityGroupIds:    nil,
				MaintenanceWindow:   nil,
				DiskEncryptionKeyId: nil,
			},
		},
		{
			testname: "MaximalConfigCheck",
			reqVal:   maximalConfig,
			hostSpecs: []*clickhouse.HostSpec{
				{ZoneId: "ru-central1-a", ShardName: "shard1", Type: clickhouse.Host_CLICKHOUSE},
				{ZoneId: "ru-central1-b", ShardName: "shard2", Type: clickhouse.Host_CLICKHOUSE},
			},
			expectedClusterRequest: &clickhouse.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "test-description",
				Labels: map[string]string{
					"key": "value",
				},
				Environment: clickhouse.Cluster_PRESTABLE,
				NetworkId:   "test-network",
				FolderId:    "test-folder-2",
				HostSpecs: []*clickhouse.HostSpec{
					{ZoneId: "ru-central1-a", ShardName: "shard1", Type: clickhouse.Host_CLICKHOUSE},
					{ZoneId: "ru-central1-b", ShardName: "shard2", Type: clickhouse.Host_CLICKHOUSE},
				},
				ServiceAccountId: "sa-id",
				ConfigSpec: &clickhouse.ConfigSpec{
					Version: "25.3",
					Clickhouse: &clickhouse.ConfigSpec_Clickhouse{
						Config: &clickhouseConfig.ClickhouseConfig{
							LogLevel:           clickhouseConfig.ClickhouseConfig_TRACE,
							BackgroundPoolSize: wrapperspb.Int64(32),
							BackgroundMergesMutationsConcurrencyRatio: wrapperspb.Int64(64),
							BackgroundSchedulePoolSize:                wrapperspb.Int64(64),
							BackgroundFetchesPoolSize:                 wrapperspb.Int64(48),
							BackgroundMovePoolSize:                    wrapperspb.Int64(32),
							BackgroundDistributedSchedulePoolSize:     wrapperspb.Int64(128),
							BackgroundBufferFlushSchedulePoolSize:     wrapperspb.Int64(64),
							BackgroundMessageBrokerSchedulePoolSize:   wrapperspb.Int64(96),
							BackgroundCommonPoolSize:                  wrapperspb.Int64(32),
							DictionariesLazyLoad:                      wrapperspb.Bool(true),
							QueryLogRetentionSize:                     wrapperspb.Int64(16),
							QueryLogRetentionTime:                     wrapperspb.Int64(32),
							QueryThreadLogEnabled:                     wrapperspb.Bool(false),
							QueryThreadLogRetentionSize:               wrapperspb.Int64(64),
							QueryThreadLogRetentionTime:               wrapperspb.Int64(32),
							PartLogRetentionSize:                      wrapperspb.Int64(256),
							PartLogRetentionTime:                      wrapperspb.Int64(128),
							MetricLogEnabled:                          wrapperspb.Bool(true),
							MetricLogRetentionSize:                    wrapperspb.Int64(64),
							MetricLogRetentionTime:                    wrapperspb.Int64(32),
							TraceLogEnabled:                           wrapperspb.Bool(false),
							TraceLogRetentionSize:                     wrapperspb.Int64(128),
							TraceLogRetentionTime:                     wrapperspb.Int64(64),
							TextLogEnabled:                            wrapperspb.Bool(true),
							TextLogRetentionSize:                      wrapperspb.Int64(256),
							TextLogRetentionTime:                      wrapperspb.Int64(32),
							TextLogLevel:                              clickhouseConfig.ClickhouseConfig_DEBUG,
							OpentelemetrySpanLogEnabled:               wrapperspb.Bool(false),
							OpentelemetrySpanLogRetentionSize:         nil,
							OpentelemetrySpanLogRetentionTime:         nil,
							QueryViewsLogEnabled:                      wrapperspb.Bool(false),
							QueryViewsLogRetentionSize:                nil,
							QueryViewsLogRetentionTime:                nil,
							AsynchronousMetricLogEnabled:              wrapperspb.Bool(true),
							AsynchronousMetricLogRetentionSize:        wrapperspb.Int64(64),
							AsynchronousMetricLogRetentionTime:        wrapperspb.Int64(16),
							SessionLogEnabled:                         wrapperspb.Bool(true),
							SessionLogRetentionSize:                   wrapperspb.Int64(128),
							SessionLogRetentionTime:                   wrapperspb.Int64(32),
							ZookeeperLogEnabled:                       wrapperspb.Bool(false),
							ZookeeperLogRetentionSize:                 nil,
							ZookeeperLogRetentionTime:                 nil,
							AsynchronousInsertLogEnabled:              wrapperspb.Bool(false),
							AsynchronousInsertLogRetentionSize:        nil,
							AsynchronousInsertLogRetentionTime:        nil,
							ProcessorsProfileLogEnabled:               wrapperspb.Bool(true),
							ProcessorsProfileLogRetentionSize:         wrapperspb.Int64(256),
							ProcessorsProfileLogRetentionTime:         wrapperspb.Int64(128),
							ErrorLogEnabled:                           wrapperspb.Bool(true),
							ErrorLogRetentionSize:                     wrapperspb.Int64(512),
							ErrorLogRetentionTime:                     wrapperspb.Int64(256),
							AccessControlImprovements: &clickhouseConfig.ClickhouseConfig_AccessControlImprovements{
								SelectFromSystemDbRequiresGrant:          wrapperspb.Bool(true),
								SelectFromInformationSchemaRequiresGrant: wrapperspb.Bool(false),
							},
							MaxConnections:                      wrapperspb.Int64(1024),
							MaxConcurrentQueries:                wrapperspb.Int64(512),
							MaxTableSizeToDrop:                  wrapperspb.Int64(256),
							MaxPartitionSizeToDrop:              wrapperspb.Int64(128),
							KeepAliveTimeout:                    wrapperspb.Int64(64),
							UncompressedCacheSize:               wrapperspb.Int64(32),
							Timezone:                            "MSK",
							GeobaseEnabled:                      wrapperspb.Bool(false),
							GeobaseUri:                          "geobase_uri",
							DefaultDatabase:                     wrapperspb.String("default_database"),
							TotalMemoryProfilerStep:             wrapperspb.Int64(16),
							TotalMemoryTrackerSampleProbability: wrapperspb.Double(10),
							AsyncInsertThreads:                  wrapperspb.Int64(128),
							BackupThreads:                       wrapperspb.Int64(512),
							RestoreThreads:                      wrapperspb.Int64(64),
							MergeTree: &clickhouseConfig.ClickhouseConfig_MergeTree{
								ReplicatedDeduplicationWindow:                  wrapperspb.Int64(100),
								ReplicatedDeduplicationWindowSeconds:           wrapperspb.Int64(200),
								PartsToDelayInsert:                             wrapperspb.Int64(10),
								PartsToThrowInsert:                             wrapperspb.Int64(20),
								MaxReplicatedMergesInQueue:                     wrapperspb.Int64(30),
								NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge: wrapperspb.Int64(5),
								MaxBytesToMergeAtMinSpaceInPool:                wrapperspb.Int64(111),
								MaxBytesToMergeAtMaxSpaceInPool:                wrapperspb.Int64(222),
								InactivePartsToDelayInsert:                     wrapperspb.Int64(7),
								InactivePartsToThrowInsert:                     wrapperspb.Int64(8),
								MinBytesForWidePart:                            wrapperspb.Int64(1024),
								MinRowsForWidePart:                             wrapperspb.Int64(2048),
								TtlOnlyDropParts:                               wrapperspb.Bool(true),
								MergeWithTtlTimeout:                            wrapperspb.Int64(300),
								MergeWithRecompressionTtlTimeout:               wrapperspb.Int64(400),
								MaxPartsInTotal:                                wrapperspb.Int64(500),
								MaxNumberOfMergesWithTtlInPool:                 wrapperspb.Int64(3),
								CleanupDelayPeriod:                             wrapperspb.Int64(600),
								NumberOfFreeEntriesInPoolToExecuteMutation:     wrapperspb.Int64(4),
								MaxAvgPartSizeForTooManyParts:                  wrapperspb.Int64(700),
								MinAgeToForceMergeSeconds:                      wrapperspb.Int64(800),
								MinAgeToForceMergeOnPartitionOnly:              wrapperspb.Bool(false),
								MergeSelectingSleepMs:                          wrapperspb.Int64(900),
								CheckSampleColumnIsCorrect:                     wrapperspb.Bool(true),
								MergeMaxBlockSize:                              wrapperspb.Int64(1000),
								MaxMergeSelectingSleepMs:                       wrapperspb.Int64(1100),
								MaxCleanupDelayPeriod:                          wrapperspb.Int64(1200),
								DeduplicateMergeProjectionMode:                 clickhouseConfig.ClickhouseConfig_MergeTree_DEDUPLICATE_MERGE_PROJECTION_MODE_DROP,
								LightweightMutationProjectionMode:              clickhouseConfig.ClickhouseConfig_MergeTree_LIGHTWEIGHT_MUTATION_PROJECTION_MODE_REBUILD,
								MaterializeTtlRecalculateOnly:                  wrapperspb.Bool(true),
								FsyncAfterInsert:                               wrapperspb.Bool(true),
								FsyncPartDirectory:                             wrapperspb.Bool(false),
								MinCompressedBytesToFsyncAfterFetch:            wrapperspb.Int64(1300),
								MinCompressedBytesToFsyncAfterMerge:            wrapperspb.Int64(1400),
								MinRowsToFsyncAfterMerge:                       wrapperspb.Int64(1500),
							},
							Compression: []*clickhouseConfig.ClickhouseConfig_Compression{
								{
									Method:           clickhouseConfig.ClickhouseConfig_Compression_LZ4,
									MinPartSize:      1024,
									MinPartSizeRatio: 0.5,
									Level:            wrapperspb.Int64(3),
								},
								{
									Method:           clickhouseConfig.ClickhouseConfig_Compression_ZSTD,
									MinPartSize:      2048,
									MinPartSizeRatio: 0.75,
									Level:            wrapperspb.Int64(5),
								},
							},
							GraphiteRollup: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup{
								{
									Name: "rollup_1",
									Patterns: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
										{
											Regexp:   "^cpu\\.",
											Function: "max",
											Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
												{
													Age:       60,
													Precision: 10,
												},
												{
													Age:       600,
													Precision: 60,
												},
											},
										},
										{
											Regexp:   "^disk\\.",
											Function: "avg",
											Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
												{
													Age:       300,
													Precision: 30,
												},
											},
										},
									},
									PathColumnName:    "path",
									TimeColumnName:    "time",
									ValueColumnName:   "value",
									VersionColumnName: "version",
								},
								{
									Name: "rollup_2",
									Patterns: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
										{
											Regexp:   ".*",
											Function: "sum",
											Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
												{
													Age:       120,
													Precision: 30,
												},
											},
										},
									},
									PathColumnName:    "p",
									TimeColumnName:    "t",
									ValueColumnName:   "v",
									VersionColumnName: "ver",
								},
							},
							Kafka: &clickhouseConfig.ClickhouseConfig_Kafka{
								SecurityProtocol:                 clickhouseConfig.ClickhouseConfig_Kafka_SECURITY_PROTOCOL_PLAINTEXT,
								SaslMechanism:                    clickhouseConfig.ClickhouseConfig_Kafka_SASL_MECHANISM_PLAIN,
								SaslUsername:                     "kafka-user",
								SaslPassword:                     "kafka-pass",
								EnableSslCertificateVerification: wrapperspb.Bool(true),
								MaxPollIntervalMs:                wrapperspb.Int64(300000),
								SessionTimeoutMs:                 wrapperspb.Int64(60000),
								Debug:                            clickhouseConfig.ClickhouseConfig_Kafka_DEBUG_ALL,
								AutoOffsetReset:                  clickhouseConfig.ClickhouseConfig_Kafka_AUTO_OFFSET_RESET_EARLIEST,
							},
							Rabbitmq: &clickhouseConfig.ClickhouseConfig_Rabbitmq{
								Username: "rmq-user",
								Password: "rmq-pass",
								Vhost:    "rmq-vhost",
							},
							QueryMaskingRules: []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule{
								{
									Name:    "mask_passwords",
									Regexp:  "(?i)password\\s*=?\\s*'[^']*'",
									Replace: "password='***'",
								},
								{
									Name:    "mask_tokens",
									Regexp:  "(?i)token\\s*=?\\s*'[^']*'",
									Replace: "token='***'",
								},
							},
							QueryCache: &clickhouseConfig.ClickhouseConfig_QueryCache{
								MaxSizeInBytes:      wrapperspb.Int64(1_000_000),
								MaxEntries:          wrapperspb.Int64(10_000),
								MaxEntrySizeInBytes: wrapperspb.Int64(50_000),
								MaxEntrySizeInRows:  wrapperspb.Int64(1_000),
							},
							JdbcBridge: &clickhouseConfig.ClickhouseConfig_JdbcBridge{
								Host: "jdbc-bridge-host",
								Port: wrapperspb.Int64(9019),
							},
							MysqlProtocol: wrapperspb.Bool(true),
							CustomMacros: []*clickhouseConfig.ClickhouseConfig_Macro{
								{
									Name:  "shard",
									Value: "shard_01",
								},
								{
									Name:  "replica",
									Value: "replica_01",
								},
							},
						},
						Resources: &clickhouse.Resources{
							ResourcePresetId: "s2.micro",
							DiskTypeId:       "network-ssd",
							DiskSize:         datasize.ToBytes(16),
						},
					},
					Zookeeper: &clickhouse.ConfigSpec_Zookeeper{
						Resources: &clickhouse.Resources{
							ResourcePresetId: "b3-c1-m4",
							DiskTypeId:       "network-ssd",
							DiskSize:         datasize.ToBytes(10),
						},
					},
					BackupWindowStart: &timeofday.TimeOfDay{
						Hours:   3,
						Minutes: 15,
					},
					Access: &clickhouse.Access{
						DataLens:     false,
						WebSql:       false,
						Metrika:      false,
						Serverless:   true,
						DataTransfer: true,
						YandexQuery:  true,
					},
					CloudStorage: &clickhouse.CloudStorage{
						Enabled:          true,
						MoveFactor:       wrapperspb.Double(3),
						DataCacheEnabled: wrapperspb.Bool(true),
						DataCacheMaxSize: wrapperspb.Int64(32),
						PreferNotToMerge: wrapperspb.Bool(false),
					},
					SqlDatabaseManagement:  wrapperspb.Bool(true),
					SqlUserManagement:      wrapperspb.Bool(true),
					AdminPassword:          "",
					EmbeddedKeeper:         wrapperspb.Bool(false),
					BackupRetainPeriodDays: wrapperspb.Int64(14),
				},
				ShardSpecs: []*clickhouse.ShardSpec{
					{
						Name: "shard1",
						ConfigSpec: &clickhouse.ShardConfigSpec{
							Clickhouse: &clickhouse.ShardConfigSpec_Clickhouse{
								Weight: wrapperspb.Int64(10),
								Resources: &clickhouse.Resources{
									ResourcePresetId: "s2.small",
									DiskTypeId:       "network-ssd",
									DiskSize:         datasize.ToBytes(20),
								},
							},
						},
					},
					{
						Name: "shard2",
						ConfigSpec: &clickhouse.ShardConfigSpec{
							Clickhouse: &clickhouse.ShardConfigSpec_Clickhouse{
								Weight: wrapperspb.Int64(20),
								Resources: &clickhouse.Resources{
									ResourcePresetId: "s2.medium",
									DiskTypeId:       "network-ssd",
									DiskSize:         datasize.ToBytes(40),
								},
							},
						},
					},
				},
				DeletionProtection: true,
				SecurityGroupIds:   []string{"test-sg"},
				MaintenanceWindow: &clickhouse.MaintenanceWindow{
					Policy: &clickhouse.MaintenanceWindow_Anytime{
						Anytime: &clickhouse.AnytimeMaintenanceWindow{},
					},
				},
				DiskEncryptionKeyId: wrapperspb.String("test-key"),
			},
			expectedFormatSchemaRequests: []*clickhouse.CreateFormatSchemaRequest{
				{
					ClusterId:        clusterId,
					FormatSchemaName: "schema1",
					Type:             clickhouse.FormatSchemaType_FORMAT_SCHEMA_TYPE_PROTOBUF,
					Uri:              "s3://bucket/schema1.proto",
				},
				{
					ClusterId:        clusterId,
					FormatSchemaName: "schema2",
					Type:             clickhouse.FormatSchemaType_FORMAT_SCHEMA_TYPE_CAPNPROTO,
					Uri:              "s3://bucket/schema2.capnp",
				},
			},
			expectedMlModelRequests: []*clickhouse.CreateMlModelRequest{
				{
					ClusterId:   clusterId,
					MlModelName: "model1",
					Type:        clickhouse.MlModelType_ML_MODEL_TYPE_CATBOOST,
					Uri:         "s3://bucket/model1.cbm",
				},
				{
					ClusterId:   clusterId,
					MlModelName: "model2",
					Type:        clickhouse.MlModelType_ML_MODEL_TYPE_CATBOOST,
					Uri:         "s3://bucket/model2.tf",
				},
			},
			expectedShardGroupRequests: []*clickhouse.CreateClusterShardGroupRequest{
				{
					ClusterId:      clusterId,
					ShardGroupName: "group1",
					Description:    "first group",
					ShardNames:     []string{"shard1", "shard2"},
				},
				{
					ClusterId:      clusterId,
					ShardGroupName: "group2",
					Description:    "second group",
					ShardNames:     []string{"shard3"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.testname, func(t *testing.T) {
			t.Parallel()

			cluster := &models.Cluster{}
			diags := c.reqVal.As(ctx, cluster, datasize.DefaultOpts)
			if diags.HasError() {
				t.Errorf(
					"Unexpected diagnostics in As() for %s: %v",
					c.testname,
					diags.Errors(),
				)
				return
			}

			// Check create cluster request
			req := prepareClusterCreateRequest(ctx, cluster, &config.State{}, &diags, c.hostSpecs)
			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected diagnostics status %s: expectedError=%t, actual=%t, errors=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
				return
			}
			utils.AssertProtoEqual(t, c.testname, c.expectedClusterRequest, req)

			// Check create format schema requests
			fsReqs := prepareFormatSchemasCreateRequests(ctx, cluster, &diags)
			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected diagnostics status %s: expectedError=%t, actual=%t, errors=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
				return
			}

			if len(fsReqs) != len(c.expectedFormatSchemaRequests) {
				t.Errorf(
					"Unexpected number of format schema requests %s: expected=%d, got=%d",
					c.testname,
					len(c.expectedFormatSchemaRequests),
					len(fsReqs),
				)
				return
			}

			mapFormatSchemaNameRequest := map[string]*clickhouse.CreateFormatSchemaRequest{}
			for _, req := range fsReqs {
				mapFormatSchemaNameRequest[req.FormatSchemaName] = req
			}

			for _, expReq := range c.expectedFormatSchemaRequests {
				actReq, ok := mapFormatSchemaNameRequest[expReq.FormatSchemaName]
				if !ok {
					t.Errorf("Missing format schema request for %q in %s", expReq.FormatSchemaName, c.testname)
					return
				}
				utils.AssertProtoEqual(t, c.testname, expReq, actReq)
			}

			// Check create ml model requests
			mlReqs := prepareMlModelsCreateRequests(ctx, cluster, &diags)
			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected diagnostics status %s: expectedError=%t, actual=%t, errors=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
				return
			}

			if len(mlReqs) != len(c.expectedMlModelRequests) {
				t.Errorf(
					"Unexpected number of ml models requests %s: expected=%d, got=%d",
					c.testname,
					len(c.expectedMlModelRequests),
					len(mlReqs),
				)
				return
			}

			mapMlModelNameRequest := map[string]*clickhouse.CreateMlModelRequest{}
			for _, req := range mlReqs {
				mapMlModelNameRequest[req.MlModelName] = req
			}

			for _, expReq := range c.expectedMlModelRequests {
				actReq, ok := mapMlModelNameRequest[expReq.MlModelName]
				if !ok {
					t.Errorf("Missing ml model request for %q in %s", expReq.MlModelName, c.testname)
					return
				}
				utils.AssertProtoEqual(t, c.testname, expReq, actReq)
			}

			// Check create shard group requests
			sgReqs := prepareShardGroupsCreateRequests(ctx, cluster, &diags)
			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected diagnostics status %s: expectedError=%t, actual=%t, errors=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
				return
			}

			if len(sgReqs) != len(c.expectedShardGroupRequests) {
				t.Errorf(
					"Unexpected number of ml models requests %s: expected=%d, got=%d",
					c.testname,
					len(c.expectedShardGroupRequests),
					len(sgReqs),
				)
				return
			}

			mapShardGroupNameRequest := map[string]*clickhouse.CreateClusterShardGroupRequest{}
			for _, req := range sgReqs {
				mapShardGroupNameRequest[req.ShardGroupName] = req
			}

			for _, expReq := range c.expectedShardGroupRequests {
				actReq, ok := mapShardGroupNameRequest[expReq.ShardGroupName]
				if !ok {
					t.Errorf("Missing shard group request for %q in %s", expReq.ShardGroupName, c.testname)
					return
				}
				utils.AssertProtoEqual(t, c.testname, expReq, actReq)
			}
		})
	}
}
