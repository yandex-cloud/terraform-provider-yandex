package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 60 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 90 * time.Minute
)

var schemaResources = map[string]*schema.Schema{
	"resource_preset_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"disk_size": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"disk_type_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
}
var schemaConfig = map[string]*schema.Schema{
	"log_level":                       {Type: schema.TypeString, Optional: true, Computed: true},
	"max_connections":                 {Type: schema.TypeInt, Optional: true, Computed: true},
	"max_concurrent_queries":          {Type: schema.TypeInt, Optional: true, Computed: true},
	"keep_alive_timeout":              {Type: schema.TypeInt, Optional: true, Computed: true},
	"uncompressed_cache_size":         {Type: schema.TypeInt, Optional: true, Computed: true},
	"mark_cache_size":                 {Type: schema.TypeInt, Optional: true, Computed: true},
	"max_table_size_to_drop":          {Type: schema.TypeInt, Optional: true, Computed: true},
	"max_partition_size_to_drop":      {Type: schema.TypeInt, Optional: true, Computed: true},
	"timezone":                        {Type: schema.TypeString, Optional: true, Computed: true},
	"geobase_uri":                     {Type: schema.TypeString, Optional: true, Computed: true},
	"query_log_retention_size":        {Type: schema.TypeInt, Optional: true, Computed: true},
	"query_log_retention_time":        {Type: schema.TypeInt, Optional: true, Computed: true},
	"query_thread_log_enabled":        {Type: schema.TypeBool, Optional: true, Computed: true},
	"query_thread_log_retention_size": {Type: schema.TypeInt, Optional: true, Computed: true},
	"query_thread_log_retention_time": {Type: schema.TypeInt, Optional: true, Computed: true},
	"part_log_retention_size":         {Type: schema.TypeInt, Optional: true, Computed: true},
	"part_log_retention_time":         {Type: schema.TypeInt, Optional: true, Computed: true},
	"metric_log_enabled":              {Type: schema.TypeBool, Optional: true, Computed: true},
	"metric_log_retention_size":       {Type: schema.TypeInt, Optional: true, Computed: true},
	"metric_log_retention_time":       {Type: schema.TypeInt, Optional: true, Computed: true},
	"trace_log_enabled":               {Type: schema.TypeBool, Optional: true, Computed: true},
	"trace_log_retention_size":        {Type: schema.TypeInt, Optional: true, Computed: true},
	"trace_log_retention_time":        {Type: schema.TypeInt, Optional: true, Computed: true},
	"text_log_enabled":                {Type: schema.TypeBool, Optional: true, Computed: true},
	"text_log_retention_size":         {Type: schema.TypeInt, Optional: true, Computed: true},
	"text_log_retention_time":         {Type: schema.TypeInt, Optional: true, Computed: true},
	"text_log_level":                  {Type: schema.TypeString, Optional: true, Computed: true},
	"background_pool_size":            {Type: schema.TypeInt, Optional: true, Computed: true},
	"background_schedule_pool_size":   {Type: schema.TypeInt, Optional: true, Computed: true},
	"background_fetches_pool_size":    {Type: schema.TypeInt, Optional: true, Computed: true},
	"default_database":                {Type: schema.TypeString, Optional: true, Computed: true},
	"total_memory_profiler_step":      {Type: schema.TypeInt, Optional: true, Computed: true},
	"merge_tree": {
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"replicated_deduplication_window":                           {Type: schema.TypeInt, Optional: true, Computed: true},
				"replicated_deduplication_window_seconds":                   {Type: schema.TypeInt, Optional: true, Computed: true},
				"parts_to_delay_insert":                                     {Type: schema.TypeInt, Optional: true, Computed: true},
				"parts_to_throw_insert":                                     {Type: schema.TypeInt, Optional: true, Computed: true},
				"max_replicated_merges_in_queue":                            {Type: schema.TypeInt, Optional: true, Computed: true},
				"number_of_free_entries_in_pool_to_lower_max_size_of_merge": {Type: schema.TypeInt, Optional: true, Computed: true},
				"max_bytes_to_merge_at_min_space_in_pool":                   {Type: schema.TypeInt, Optional: true, Computed: true},
				"min_bytes_for_wide_part":                                   {Type: schema.TypeInt, Optional: true, Computed: true},
				"min_rows_for_wide_part":                                    {Type: schema.TypeInt, Optional: true, Computed: true},
				"ttl_only_drop_parts":                                       {Type: schema.TypeBool, Optional: true, Computed: true},
			},
		},
	},
	"kafka": {
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"security_protocol": {Type: schema.TypeString, Optional: true, Computed: true},
				"sasl_mechanism":    {Type: schema.TypeString, Optional: true, Computed: true},
				"sasl_username":     {Type: schema.TypeString, Optional: true, Computed: true},
				"sasl_password":     {Type: schema.TypeString, Optional: true, Sensitive: true, Computed: true},
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
				"settings": {
					Type:     schema.TypeList,
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
				"username": {Type: schema.TypeString, Optional: true, Computed: true},
				"password": {Type: schema.TypeString, Optional: true, Sensitive: true, Computed: true},
				"vhost":    {Type: schema.TypeString, Optional: true, Computed: true},
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
							"regexp":   {Type: schema.TypeString, Optional: true, Computed: true},
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
}

func resourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBClickHouseClusterCreate,
		Read:   resourceYandexMDBClickHouseClusterRead,
		Update: resourceYandexMDBClickHouseClusterUpdate,
		Delete: resourceYandexMDBClickHouseClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBClickHouseClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBClickHouseClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBClickHouseClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseClickHouseEnv),
			},
			"clickhouse": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: schemaConfig,
							},
						},
						"resources": {
							Type:             schema.TypeList,
							MaxItems:         1,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: compareClusterResources,
							Elem: &schema.Resource{
								Schema: schemaResources,
							},
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      clickHouseUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Set:      clickHouseUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Required: true,
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
									"max_concurrent_queries_for_user":                    {Type: schema.TypeInt, Optional: true, Computed: true},
									"memory_profiler_step":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"memory_profiler_sample_probability":                 {Type: schema.TypeFloat, Optional: true, Computed: true},
									"insert_null_as_default":                             {Type: schema.TypeBool, Optional: true, Computed: true},
									"allow_suspicious_low_cardinality_types":             {Type: schema.TypeBool, Optional: true, Computed: true},
									"connect_timeout_with_failover":                      {Type: schema.TypeInt, Optional: true, Computed: true},
									"allow_introspection_functions":                      {Type: schema.TypeBool, Optional: true, Computed: true},
									"async_insert":                                       {Type: schema.TypeBool, Optional: true, Computed: true},
									"async_insert_threads":                               {Type: schema.TypeInt, Optional: true, Computed: true},
									"wait_for_async_insert":                              {Type: schema.TypeBool, Optional: true, Computed: true},
									"wait_for_async_insert_timeout":                      {Type: schema.TypeInt, Optional: true, Computed: true},
									"async_insert_max_data_size":                         {Type: schema.TypeInt, Optional: true, Computed: true},
									"async_insert_busy_timeout":                          {Type: schema.TypeInt, Optional: true, Computed: true},
									"async_insert_stale_timeout":                         {Type: schema.TypeInt, Optional: true, Computed: true},
									"timeout_before_checking_execution_speed":            {Type: schema.TypeInt, Optional: true, Computed: true},
									"cancel_http_readonly_queries_on_client_close":       {Type: schema.TypeBool, Optional: true, Computed: true},
									"flatten_nested":                                     {Type: schema.TypeBool, Optional: true, Computed: true},
									"max_http_get_redirects":                             {Type: schema.TypeInt, Optional: true, Computed: true},
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
			"shard": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      clickHouseShardHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"resources": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: schemaResources,
							},
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      clickHouseDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"copy_schema_on_new_hosts": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"host": {
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateParsableValue(parseClickHouseHostType),
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"shard_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
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
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"shard_names": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"format_schema": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ml_model": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"backup_window_start": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"web_sql": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"data_lens": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"metrika": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"serverless": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"data_transfer": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"yandex_query": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"zookeeper": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressZooKeeperResourcesDIff,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
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
				Optional: true,
				Computed: true,
			},
			"admin_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"sql_user_management": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"sql_database_management": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"embedded_keeper": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cloud_storage": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"move_factor": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"data_cache_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"data_cache_max_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							ValidateFunc: validateParsableValue(parseClickHouseWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexMDBClickHouseClusterCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] create started")
	backupOriginalClusterResource(d)
	config := meta.(*Config)

	req, shardsToAdd, shardsFromSpec, err := prepareCreateClickHouseCreateRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create ClickHouse Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting ClickHouse create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*clickhouse.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create ClickHouse Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("ClickHouse Cluster creation failed: %s", err)
	}

	for shardName, shardHosts := range shardsToAdd {
		var shardSpec *clickhouse.ShardConfigSpec
		if v, ok := shardsFromSpec[shardName]; ok {
			shardSpec = v
		}
		err = createClickHouseShard(ctx, config, d, shardName, shardHosts, shardSpec)
		if err != nil {
			return err
		}
		delete(shardsFromSpec, shardName)
	}

	for shardNameFromSpec, shardConfigFromSpec := range shardsFromSpec {
		if req.ShardName != shardNameFromSpec {
			log.Printf("[ERROR] trying to update non-existent shard, name=%s\n", shardNameFromSpec)
			continue
		}
		log.Printf("[DEBUG] update exists shard=%s\n", shardNameFromSpec)
		if err := updateClickHouseShard(ctx, config, d, shardNameFromSpec, shardConfigFromSpec); err != nil {
			return err
		}
	}

	shardGroups, err := expandClickHouseShardGroups(d)
	if err != nil {
		return err
	}

	for _, group := range shardGroups {
		err = createClickHouseShardGroup(ctx, config, d, group)
		if err != nil {
			return err
		}
	}

	formatSchemas, err := expandClickHouseFormatSchemas(d)
	if err != nil {
		return err
	}

	for _, formatSchema := range formatSchemas {
		err = createClickHouseFormatSchema(ctx, config, d, formatSchema)
		if err != nil {
			return err
		}
	}

	mlModels, err := expandClickHouseMlModels(d)
	if err != nil {
		return err
	}

	for _, mlModel := range mlModels {
		err = createClickHouseMlModel(ctx, config, d, mlModel)
		if err != nil {
			return err
		}
	}

	return resourceYandexMDBClickHouseClusterRead(d, meta)
}

// Returns request for creating the Cluster and the map of the remaining shards to add.
func prepareCreateClickHouseCreateRequest(d *schema.ResourceData, meta *Config) (*clickhouse.CreateClusterRequest, map[string][]*clickhouse.HostSpec, map[string]*clickhouse.ShardConfigSpec, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding labels on ClickHouse Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting folder ID while creating ClickHouse Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseClickHouseEnv(e)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error resolving environment while creating ClickHouse Cluster: %s", err)
	}

	dbSpecs, err := expandClickHouseDatabases(d)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding databases on ClickHouse Cluster create: %s", err)
	}

	users, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding user specs on ClickHouse Cluster create: %s", err)
	}

	hosts, err := expandClickHouseHosts(d)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding hosts on ClickHouse Cluster create: %s", err)
	}

	_, toAdd, _ := clickHouseHostsDiff(nil, hosts)
	log.Printf("[DEBUG] hosts to add: %v\n", toAdd)

	firstHosts := toAdd["zk"]
	firstShard := ""
	delete(toAdd, "zk")
	for shardName, shardHosts := range toAdd {
		firstShard = shardName
		firstHosts = append(firstHosts, shardHosts...)
		delete(toAdd, shardName)
		break
	}

	clickhouseConfigSpec, err := expandClickHouseSpec(d)
	if err != nil {
		return nil, nil, nil, err
	}

	configSpec := &clickhouse.ConfigSpec{
		Version:           d.Get("version").(string),
		Clickhouse:        clickhouseConfigSpec,
		Zookeeper:         expandClickHouseZookeeperSpec(d),
		BackupWindowStart: expandClickHouseBackupWindowStart(d),
		Access:            expandClickHouseAccess(d),
		CloudStorage:      expandClickHouseCloudStorage(d),
	}

	shardsFromSpec, err := expandClickhouseShardSpecs(d)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding shard specs on ClickHouse Cluster create: %s", err)
	}

	if val, ok := d.GetOk("admin_password"); ok {
		configSpec.SetAdminPassword(val.(string))
	}

	if val, ok := d.GetOk("sql_user_management"); ok {
		configSpec.SetSqlUserManagement(&wrappers.BoolValue{Value: val.(bool)})
	}

	if val, ok := d.GetOk("sql_database_management"); ok {
		configSpec.SetSqlDatabaseManagement(&wrappers.BoolValue{Value: val.(bool)})
	}

	if val, ok := d.GetOk("embedded_keeper"); ok {
		configSpec.SetEmbeddedKeeper(&wrappers.BoolValue{Value: val.(bool)})
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while expanding network id on ClickHouse Cluster create: %s", err)
	}

	mw, err := expandClickHouseMaintenanceWindow(d)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creation error while expand clickhouse maintenance_window: %s", err)
	}

	req := clickhouse.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		DatabaseSpecs:      dbSpecs,
		ConfigSpec:         configSpec,
		HostSpecs:          firstHosts,
		UserSpecs:          users,
		Labels:             labels,
		ShardName:          firstShard,
		SecurityGroupIds:   securityGroupIds,
		ServiceAccountId:   d.Get("service_account_id").(string),
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  mw,
	}
	return &req, toAdd, shardsFromSpec, nil
}

func resourceYandexMDBClickHouseClusterRead(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] cluster read started")
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Clickhouse().Cluster().Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}
	chResources, err := flattenClickHouseResources(cluster.Config.Clickhouse.Resources)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] read cluster clickhouse resources: %v\n", chResources)
	chConfig, err := flattenClickHouseConfig(d, cluster.Config.Clickhouse.Config)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] read cluster clickhouse config: %v\n", chConfig)
	err = d.Set("clickhouse", []map[string]interface{}{
		{
			"resources": chResources,
			"config":    chConfig,
		},
	})
	if err != nil {
		return err
	}

	zkResources, err := flattenClickHouseResources(cluster.Config.Zookeeper.Resources)
	if err != nil {
		return err
	}
	err = d.Set("zookeeper", []map[string]interface{}{
		{
			"resources": zkResources,
		},
	})
	if err != nil {
		return err
	}

	bws := flattenClickHouseBackupWindowStart(cluster.Config.BackupWindowStart)
	if err := d.Set("backup_window_start", bws); err != nil {
		return err
	}

	ac := flattenClickHouseAccess(cluster.GetConfig().GetAccess())
	if err := d.Set("access", ac); err != nil {
		return err
	}

	mw := flattenClickHouseMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	hosts, err := listClickHouseHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	dHosts, err := expandClickHouseHosts(d)
	if err != nil {
		return err
	}

	hosts = sortClickHouseHosts(hosts, dHosts)
	hs, err := flattenClickHouseHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := setShardsToSchema(ctx, config, d); err != nil {
		return err
	}

	groups, err := listClickHouseShardGroups(ctx, config, d.Id())
	if err != nil {
		return err
	}

	sg, err := flattenClickHouseShardGroups(groups)
	if err != nil {
		return err
	}

	if err := d.Set("shard_group", sg); err != nil {
		return err
	}

	formatSchemas, err := listClickHouseFormatSchemas(ctx, config, d.Id())
	if err != nil {
		return err
	}
	fs, err := flattenClickHouseFormatSchemas(formatSchemas)
	if err != nil {
		return err
	}
	if err := d.Set("format_schema", fs); err != nil {
		return err
	}

	mlModels, err := listClickHouseMlModels(ctx, config, d.Id())
	if err != nil {
		return err
	}
	ml, err := flattenClickHouseMlModels(mlModels)
	if err != nil {
		return err
	}
	if err := d.Set("ml_model", ml); err != nil {
		return err
	}

	databases, err := listClickHouseDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}
	dbs := flattenClickHouseDatabases(databases)
	if err := d.Set("database", dbs); err != nil {
		return err
	}

	dUsers, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return err
	}
	passwords := clickHouseUsersPasswords(dUsers)

	users, err := listClickHouseUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	us := flattenClickHouseUsers(users, passwords)
	if err := d.Set("user", us); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("version", cluster.Config.Version)
	d.Set("sql_user_management", cluster.Config.GetSqlUserManagement().GetValue())
	d.Set("sql_database_management", cluster.Config.GetSqlDatabaseManagement().GetValue())
	d.Set("embedded_keeper", cluster.Config.GetEmbeddedKeeper().GetValue())
	d.Set("service_account_id", cluster.ServiceAccountId)
	d.Set("deletion_protection", cluster.DeletionProtection)

	cs := flattenClickHouseCloudStorage(cluster.Config.CloudStorage)
	if err := d.Set("cloud_storage", cs); err != nil {
		return err
	}

	log.Printf("[DEBUG] cluster read finished: schema after read=%+v\n", d)
	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBClickHouseClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Started update ClickHouse Cluster %q", d.Id())
	backupOriginalClusterResource(d)

	d.Partial(true)

	if err := setClickHouseFolderID(d, meta); err != nil {
		return err
	}

	if err := updateClickHouseClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("database") {
		if err := updateClickHouseClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateClickHouseClusterUsers(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		log.Println("[DEBUG] host configuration change detected.")
		if err := updateClickHouseClusterHosts(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("shard") {
		log.Println("[DEBUG] shard configuration changes detected.")
		if err := updateClickHouseClusterShards(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("shard_group") {
		if err := updateClickHouseClusterShardGroups(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("format_schema") {
		if err := updateClickHouseFormatSchemas(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("ml_model") {
		if err := updateClickHouseMlModels(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	log.Printf("[DEBUG] Finished updating ClickHouse Cluster %q", d.Id())
	return resourceYandexMDBClickHouseClusterRead(d, meta)
}

var mdbClickHouseUpdateFieldsMap = map[string]string{
	"name":                    "name",
	"description":             "description",
	"labels":                  "labels",
	"version":                 "config_spec.version",
	"access":                  "config_spec.access",
	"backup_window_start":     "config_spec.backup_window_start",
	"clickhouse":              "config_spec.clickhouse",
	"admin_password":          "config_spec.admin_password",
	"sql_user_management":     "config_spec.sql_user_management",
	"sql_database_management": "config_spec.sql_database_management",
	"cloud_storage":           "config_spec.cloud_storage",
	"security_group_ids":      "security_group_ids",
	"service_account_id":      "service_account_id",
	"maintenance_window":      "maintenance_window",
	"deletion_protection":     "deletion_protection",
}

func updateClickHouseClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	if d.HasChanges("resources", "version") {
		oldVersion, newVersion := d.GetChange("version")
		log.Printf("[DEBUG] Pre-updating ClickHouse Cluster %q version %q -> %q", d.Id(), oldVersion, newVersion)

		req := &clickhouse.UpdateClusterRequest{
			ClusterId: d.Id(),
			ConfigSpec: &clickhouse.ConfigSpec{
				Version: d.Get("version").(string),
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"config_spec.version"},
			},
		}

		ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
		defer cancel()

		op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Update(ctx, req))
		if err != nil {
			return fmt.Errorf("error while requesting API to update ClickHouse Cluster version %q: %s", d.Id(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while updating ClickHouse Cluster version %q: %s", d.Id(), err)
		}
	}

	req, err := getClickHouseClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbClickHouseUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {

			})
		}
	}

	// We only can apply this if ZK subcluster already exists
	if d.HasChange("zookeeper") {
		ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
		defer cancel()

		currHosts, err := listClickHouseHosts(ctx, config, d.Id())
		if err != nil {
			return err
		}

		for _, h := range currHosts {
			if h.Type == clickhouse.Host_ZOOKEEPER {
				updatePath = append(updatePath, "config_spec.zookeeper")
				onDone = append(onDone, func() {

				})
				break
			}
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating ClickHouse Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func getClickHouseClusterUpdateRequest(d *schema.ResourceData) (*clickhouse.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating ClickHouse cluster: %s", err)
	}

	clickhouseConfigSpec, err := expandClickHouseSpec(d)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] resource for cluster update request: %+v\n", clickhouseConfigSpec.Resources)
	mw, err := expandClickHouseMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("update error while expand clickhouse maintenance_window: %s", err)
	}

	req := &clickhouse.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		ConfigSpec: &clickhouse.ConfigSpec{
			Version:               d.Get("version").(string),
			Clickhouse:            clickhouseConfigSpec,
			Zookeeper:             expandClickHouseZookeeperSpec(d),
			BackupWindowStart:     expandClickHouseBackupWindowStart(d),
			Access:                expandClickHouseAccess(d),
			SqlUserManagement:     &wrappers.BoolValue{Value: d.Get("sql_user_management").(bool)},
			SqlDatabaseManagement: &wrappers.BoolValue{Value: d.Get("sql_database_management").(bool)},
			CloudStorage:          expandClickHouseCloudStorage(d),
		},
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		ServiceAccountId:   d.Get("service_account_id").(string),
		MaintenanceWindow:  mw,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	if pass, ok := d.GetOk("admin_password"); ok {
		req.ConfigSpec.SetAdminPassword(pass.(string))
	}

	return req, nil
}

func updateClickHouseClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listClickHouseDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandClickHouseDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := clickHouseDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteClickHouseDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createClickHouseDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateClickHouseClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currUsers, err := listClickHouseUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := clickHouseUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deleteClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	changedUsers := clickHouseChangedUsers(oldSpecs.(*schema.Set), newSpecs.(*schema.Set), d)
	for _, u := range changedUsers {
		err := updateClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateClickHouseClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := listClickHouseHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetHosts, err := expandClickHouseHosts(d)
	if err != nil {
		return err
	}

	currZkHosts := []*clickhouse.Host{}
	for _, h := range currHosts {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			currZkHosts = append(currZkHosts, h)
		}
	}
	targetZkHosts := []*clickhouse.HostSpec{}
	for _, h := range targetHosts {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			targetZkHosts = append(targetZkHosts, h)
		}
	}

	toDelete, toAdd, toUpdate := clickHouseHostsDiff(currHosts, targetHosts)

	log.Printf("[DEBUG] hosts to delete: %v\n", toDelete)
	log.Printf("[DEBUG] hosts to add: %v\n", toAdd)
	log.Printf("[DEBUG] hosts to update: %v\n", toUpdate)

	for _, h := range toUpdate {
		err = updateClickHouseHost(ctx, config, d, h)
		if err != nil {
			return err
		}
	}

	// Check if any shard has HA-configuration (2+ hosts)
	needZk := false
	m := map[string][]struct{}{}
	for _, h := range targetHosts {
		if h.Type == clickhouse.Host_CLICKHOUSE {
			shardName := "shard1"
			if h.ShardName != "" {
				shardName = h.ShardName
			}
			m[shardName] = append(m[shardName], struct{}{})
			if len(m[shardName]) > 1 {
				needZk = true
				break
			}
		}
	}

	// We need to create a ZooKeeper subcluster first
	if len(currZkHosts) == 0 && (needZk || len(toAdd["zk"]) > 0) {
		zkSpecs := toAdd["zk"]
		delete(toAdd, "zk")
		zk := expandClickHouseZookeeperSpec(d)

		err = createClickHouseZooKeeper(ctx, config, d, zk.Resources, zkSpecs)
		if err != nil {
			return err
		}
	}

	// Do not remove implicit ZooKeeper subcluster.
	if len(currZkHosts) > 1 && len(targetZkHosts) == 0 {
		delete(toDelete, "zk")
	}

	currShards, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return err
	}

	for shardName, specs := range toAdd {
		shardExists := false
		for _, s := range currShards {
			if s.Name == shardName {
				shardExists = true
			}
		}

		if shardName != "" && shardName != "zk" && !shardExists {
			err = createClickHouseShard(ctx, config, d, shardName, specs, nil)
			if err != nil {
				return err
			}
		} else {
			for _, h := range specs {
				err := createClickHouseHost(ctx, config, d, h)
				if err != nil {
					return err
				}
			}
		}
	}

	for shardName, fqdns := range toDelete {
		deleteShard := true
		for _, th := range targetHosts {
			if th.ShardName == shardName {
				deleteShard = false
			}
		}
		if shardName != "zk" && deleteShard {
			err = deleteClickHouseShard(ctx, config, d, shardName)
			if err != nil {
				return err
			}
		} else {
			for _, h := range fqdns {
				err := deleteClickHouseHost(ctx, config, d, h)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func updateClickHouseClusterShards(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	shardsOnCluster, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] before update shards got shards from cluster: %+v\n", shardsOnCluster)

	shardsFromSpec, err := expandClickhouseShardSpecs(d)
	if err != nil {
		return nil
	}

	log.Printf("[DEBUG] before update shards got shards from schema: %+v\n", shardsFromSpec)

	for _, shard := range shardsOnCluster {
		if shardSpec, ok := shardsFromSpec[shard.Name]; ok {
			if err = updateClickHouseShard(ctx, config, d, shard.Name, shardSpec); err != nil {
				return fmt.Errorf("failed update shard from config: %s", err)
			}
		}
	}

	return nil
}

func updateClickHouseClusterShardGroups(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currGroups, err := listClickHouseShardGroups(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetGroups, err := expandClickHouseShardGroups(d)
	if err != nil {
		return err
	}

	shardGroupDiff := clickHouseShardGroupDiff(currGroups, targetGroups)
	for _, g := range shardGroupDiff.toDelete {
		err := deleteClickHouseShardGroup(ctx, config, d, g)
		if err != nil {
			return err
		}
	}

	for _, g := range shardGroupDiff.toAdd {
		err := createClickHouseShardGroup(ctx, config, d, g)
		if err != nil {
			return err
		}
	}

	for _, g := range shardGroupDiff.toUpdate {
		err := updateClickHouseShardGroup(ctx, config, d, g)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateClickHouseFormatSchemas(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currSchemas, err := listClickHouseFormatSchemas(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetSchemas, err := expandClickHouseFormatSchemas(d)
	if err != nil {
		return err
	}

	formatSchemaDiff := clickHouseFormatSchemaDiff(currSchemas, targetSchemas)
	for _, fs := range formatSchemaDiff.toDelete {
		err := deleteClickHouseFormatSchema(ctx, config, d, fs)
		if err != nil {
			return err
		}
	}

	for _, fs := range formatSchemaDiff.toAdd {
		err := createClickHouseFormatSchema(ctx, config, d, fs)
		if err != nil {
			return err
		}
	}

	for _, fs := range formatSchemaDiff.toUpdate {
		err := updateClickHouseFormatSchema(ctx, config, d, fs)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateClickHouseMlModels(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currModels, err := listClickHouseMlModels(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetModels, err := expandClickHouseMlModels(d)
	if err != nil {
		return err
	}

	mlModelDiff := clickHouseMlModelDiff(currModels, targetModels)
	for _, ml := range mlModelDiff.toDelete {
		err := deleteClickHouseMlModel(ctx, config, d, ml)
		if err != nil {
			return err
		}
	}

	for _, ml := range mlModelDiff.toAdd {
		err := createClickHouseMlModel(ctx, config, d, ml)
		if err != nil {
			return err
		}
	}

	for _, ml := range mlModelDiff.toUpdate {
		err := updateClickHouseMlModel(ctx, config, d, ml)
		if err != nil {
			return err
		}
	}

	return nil
}

func createClickHouseDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Database().Create(ctx, &clickhouse.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &clickhouse.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Database().Delete(ctx, &clickhouse.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Create(ctx, &clickhouse.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Delete(ctx, &clickhouse.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Update(ctx, &clickhouse.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
			Settings:    user.Settings,
			Quotas:      user.Quotas,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddHosts(ctx, &clickhouse.AddClusterHostsRequest{
			ClusterId:  d.Id(),
			HostSpecs:  []*clickhouse.HostSpec{spec},
			CopySchema: &wrappers.BoolValue{Value: d.Get("copy_schema_on_new_hosts").(bool)},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add host to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding host to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *clickhouse.UpdateHostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().UpdateHosts(ctx, &clickhouse.UpdateClusterHostsRequest{
			ClusterId:       d.Id(),
			UpdateHostSpecs: []*clickhouse.UpdateHostSpec{spec},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update host of ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host of ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseHost(ctx context.Context, config *Config, d *schema.ResourceData, fqdn string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteHosts(ctx, &clickhouse.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{fqdn},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, name string, specs []*clickhouse.HostSpec, shardSpec *clickhouse.ShardConfigSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddShard(ctx, &clickhouse.AddClusterShardRequest{
			ClusterId:  d.Id(),
			ShardName:  name,
			ConfigSpec: shardSpec,
			HostSpecs:  specs,
			CopySchema: &wrappers.BoolValue{Value: d.Get("copy_schema_on_new_hosts").(bool)},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func isShardResourceDiskSizeChanged(fromCluster, fromSpec *clickhouse.Resources) bool {
	if fromCluster != nil && fromSpec == nil {
		log.Printf("[DEBUG] shard's weight is removed from configuration. set default value.")
		return true
	}
	if fromCluster.DiskSize != fromSpec.DiskSize {
		log.Printf("[DEBUG] change shard's weight according to the configuration tf file.")
		return true
	}
	log.Printf("[DEBUG] no change in shard's weight.")
	return false
}

func isShardResourceResourcePresetIdChanged(fromCluster, fromSpec *clickhouse.Resources) bool {
	if fromCluster != nil && fromSpec == nil {
		log.Printf("[DEBUG] shard's ResourcePresetId is removed from configuration. set default value.")
		return true
	}
	if fromCluster.ResourcePresetId != fromSpec.ResourcePresetId {
		log.Printf("[DEBUG] change shard's ResourcePresetId according to the configuration tf file.")
		return true
	}
	log.Printf("[DEBUG] no change in shard's ResourcePresetId.")
	return false
}

func isShardResourceDiskTypeIdChanged(fromCluster, fromSpec *clickhouse.Resources) bool {
	if fromCluster != nil && fromSpec == nil {
		log.Printf("[DEBUG] shard's DiskTypeId is removed from configuration. set default value.")
		return true
	}
	if fromCluster.DiskTypeId != fromSpec.DiskTypeId {
		log.Printf("[DEBUG] change shard's DiskTypeId according to the configuration tf file.")
		return true
	}
	log.Printf("[DEBUG] no change in shard's DiskTypeId.")
	return false
}

func updateClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string, shardSpec *clickhouse.ShardConfigSpec) error {
	resp, err := config.sdk.MDB().Clickhouse().Cluster().GetShard(context.Background(), &clickhouse.GetClusterShardRequest{
		ClusterId: d.Id(),
		ShardName: shardName,
	})
	if err != nil {
		return fmt.Errorf("eerror while requesting API to get shard's config, shard name=%s. Error=%s", shardName, err)
	}

	updateRequired := false
	var updatePath []string

	log.Println("[DEBUG] start compute updating fields")
	if resp.Config.Clickhouse.Weight.Value != shardSpec.Clickhouse.Weight.Value {
		log.Printf("[DEBUG] shard=%s has wegith=%d, update to %d\n", shardName, resp.Config.Clickhouse.Weight.Value, shardSpec.Clickhouse.Weight.Value)
		updateRequired = true
		updatePath = append(updatePath, "config_spec.clickhouse.weight")
	}

	if shardSpec.Clickhouse.Resources != nil {
		if isShardResourceDiskSizeChanged(resp.Config.Clickhouse.Resources, shardSpec.Clickhouse.Resources) {
			log.Printf("[DEBUG] shard=%s has disk_size=%d, update to %d\n", shardName, resp.Config.Clickhouse.Resources.GetDiskSize(), shardSpec.Clickhouse.Resources.GetDiskSize())
			updateRequired = true
			updatePath = append(updatePath, "config_spec.clickhouse.resources.disk_size")
		}

		if isShardResourceResourcePresetIdChanged(resp.Config.Clickhouse.Resources, shardSpec.Clickhouse.Resources) {
			log.Printf("[DEBUG] shard=%s has resource_preset_id=%s, update to %s\n", shardName, resp.Config.Clickhouse.Resources.GetResourcePresetId(), shardSpec.Clickhouse.Resources.GetResourcePresetId())
			updateRequired = true
			updatePath = append(updatePath, "config_spec.clickhouse.resources.ResourcePresetId")
		}

		if isShardResourceDiskTypeIdChanged(resp.Config.Clickhouse.Resources, shardSpec.Clickhouse.Resources) {
			log.Printf("[DEBUG] shard=%s has disk_type_id=%s, update to %s\n", shardName, resp.Config.Clickhouse.Resources.GetDiskTypeId(), shardSpec.Clickhouse.Resources.GetDiskTypeId())
			updateRequired = true
			updatePath = append(updatePath, "config_spec.clickhouse.resources.DiskTypeId")
		}
	}

	if !updateRequired {
		return nil
	}

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().UpdateShard(ctx, &clickhouse.UpdateClusterShardRequest{
			ClusterId:  d.Id(),
			ShardName:  shardName,
			ConfigSpec: shardSpec,
			UpdateMask: &field_mask.FieldMask{Paths: updatePath},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func deleteClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteShard(ctx, &clickhouse.DeleteClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete shard from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting shard from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseShardGroup(ctx context.Context, config *Config, d *schema.ResourceData, group *clickhouse.ShardGroup) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().CreateShardGroup(ctx, &clickhouse.CreateClusterShardGroupRequest{
			ClusterId:      d.Id(),
			ShardGroupName: group.Name,
			Description:    group.Description,
			ShardNames:     group.ShardNames,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add shard group to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding shard group to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseShardGroup(ctx context.Context, config *Config, d *schema.ResourceData, group *clickhouse.ShardGroup) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().UpdateShardGroup(ctx, &clickhouse.UpdateClusterShardGroupRequest{
			ClusterId:      d.Id(),
			ShardGroupName: group.Name,
			Description:    group.Description,
			ShardNames:     group.ShardNames,
			UpdateMask:     &field_mask.FieldMask{Paths: []string{"description", "shard_names"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update shard group to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating shard group to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseShardGroup(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteShardGroup(ctx, &clickhouse.DeleteClusterShardGroupRequest{
			ClusterId:      d.Id(),
			ShardGroupName: name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete shard group from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting shard group from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseFormatSchema(ctx context.Context, config *Config, d *schema.ResourceData, schema *clickhouse.FormatSchema) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().FormatSchema().Create(ctx, &clickhouse.CreateFormatSchemaRequest{
			ClusterId:        d.Id(),
			FormatSchemaName: schema.Name,
			Type:             schema.Type,
			Uri:              schema.Uri,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create format schema in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating format schema in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseFormatSchema(ctx context.Context, config *Config, d *schema.ResourceData, schema *clickhouse.FormatSchema) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().FormatSchema().Update(ctx, &clickhouse.UpdateFormatSchemaRequest{
			ClusterId:        d.Id(),
			FormatSchemaName: schema.Name,
			Uri:              schema.Uri,
			UpdateMask:       &field_mask.FieldMask{Paths: []string{"uri"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update format schema in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating format schema in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseFormatSchema(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().FormatSchema().Delete(ctx, &clickhouse.DeleteFormatSchemaRequest{
			ClusterId:        d.Id(),
			FormatSchemaName: name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete format schema from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting format schema from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseMlModel(ctx context.Context, config *Config, d *schema.ResourceData, model *clickhouse.MlModel) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().MlModel().Create(ctx, &clickhouse.CreateMlModelRequest{
			ClusterId:   d.Id(),
			MlModelName: model.Name,
			Type:        model.Type,
			Uri:         model.Uri,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add ml model to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding ml model to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseMlModel(ctx context.Context, config *Config, d *schema.ResourceData, model *clickhouse.MlModel) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().MlModel().Update(ctx, &clickhouse.UpdateMlModelRequest{
			ClusterId:   d.Id(),
			MlModelName: model.Name,
			Uri:         model.Uri,
			UpdateMask:  &field_mask.FieldMask{Paths: []string{"uri"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update ml model in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating ml model in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseMlModel(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().MlModel().Delete(ctx, &clickhouse.DeleteMlModelRequest{
			ClusterId:   d.Id(),
			MlModelName: name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete shard group from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting ml model from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseZooKeeper(ctx context.Context, config *Config, d *schema.ResourceData, resources *clickhouse.Resources, specs []*clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddZookeeper(ctx, &clickhouse.AddClusterZookeeperRequest{
			ClusterId: d.Id(),
			Resources: resources,
			HostSpecs: specs,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *clickhouse.MaintenanceWindow) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().Update(ctx, &clickhouse.UpdateClusterRequest{
			ClusterId:         d.Id(),
			MaintenanceWindow: mw,
			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update maintenance window in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating maintenance window in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listClickHouseHosts(ctx context.Context, config *Config, id string) ([]*clickhouse.Host, error) {
	hosts := []*clickhouse.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListHosts(ctx, &clickhouse.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of hosts for '%s': %s", id, err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func listClickHouseUsers(ctx context.Context, config *Config, id string) ([]*clickhouse.User, error) {
	users := []*clickhouse.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().User().List(ctx, &clickhouse.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		users = append(users, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return users, nil
}

func listClickHouseDatabases(ctx context.Context, config *Config, id string) ([]*clickhouse.Database, error) {
	dbs := []*clickhouse.Database{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Database().List(ctx, &clickhouse.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of databases for '%s': %s", id, err)
		}
		dbs = append(dbs, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return dbs, nil
}

func listClickHouseShards(ctx context.Context, config *Config, id string) ([]*clickhouse.Shard, error) {
	shards := []*clickhouse.Shard{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListShards(ctx, &clickhouse.ListClusterShardsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of shards for cluster_id='%s': %s", id, err)
		}
		shards = append(shards, resp.Shards...)
		if resp.NextPageToken == "" || resp.NextPageToken == pageToken {
			break
		}
		pageToken = resp.NextPageToken
	}
	return shards, nil
}

func resourceYandexMDBClickHouseClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting ClickHouse Cluster %q", d.Id())

	req := &clickhouse.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ClickHouse Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting ClickHouse Cluster %q", d.Id())
	return nil
}

func listClickHouseShardGroups(ctx context.Context, config *Config, id string) ([]*clickhouse.ShardGroup, error) {
	var groups []*clickhouse.ShardGroup
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListShardGroups(ctx, &clickhouse.ListClusterShardGroupsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of shard groups for '%s': %s", id, err)
		}

		groups = append(groups, resp.ShardGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return groups, nil
}

func listClickHouseFormatSchemas(ctx context.Context, config *Config, id string) ([]*clickhouse.FormatSchema, error) {
	var formatSchema []*clickhouse.FormatSchema
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().FormatSchema().List(ctx, &clickhouse.ListFormatSchemasRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of format schemas for '%s': %s", id, err)
		}

		formatSchema = append(formatSchema, resp.FormatSchemas...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return formatSchema, nil
}

func listClickHouseMlModels(ctx context.Context, config *Config, id string) ([]*clickhouse.MlModel, error) {
	var groups []*clickhouse.MlModel
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().MlModel().List(ctx, &clickhouse.ListMlModelsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of ml models for '%s': %s", id, err)
		}

		groups = append(groups, resp.MlModels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return groups, nil
}

func suppressZooKeeperResourcesDIff(k, old, new string, d *schema.ResourceData) bool {
	for _, host := range d.Get("host").([]interface{}) {
		if hostType, ok := host.(map[string]interface{})["type"]; ok && hostType == "ZOOKEEPER" {
			return false
		}
	}

	return true
}

func setShardsToSchema(ctx context.Context, config *Config, d *schema.ResourceData) error {
	shardsOnCluster, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return fmt.Errorf("read cluster: failed to get list of current shards: %s", err)
	}

	shards, err := flattenClickHouseShards(shardsOnCluster)
	if err != nil {
		return fmt.Errorf("read cluster: failed to flat current shards: %s", err)
	}

	if err := d.Set("shard", shards); err != nil {
		return fmt.Errorf("read cluster: failed to set shards in schema: %s", err)
	}
	fmt.Printf("[DEBUG] read data for fill schema: shards=%v\n", shards)

	return nil
}

func compareResourcesFields(key string, old, new string, configSpec *clickhouse.ShardConfigSpec) bool {
	// need only such as: clickhouse.0.resources.0.disk_size
	rawPartKey := strings.Split(key, ".0.")
	if len(rawPartKey) != 3 {
		log.Printf("[DEBUG] wrong format key: %s\n", key)
		return defaultResourcesCompare(old, new)
	}
	resources := configSpec.Clickhouse.Resources
	log.Printf("[DEBUG] current resource from first shard = %v\n", resources)

	switch k := rawPartKey[len(rawPartKey)-1]; k {
	case "disk_size":
		oldGb, err := strconv.Atoi(old)
		if err != nil {
			log.Printf("[ERROR] failed parse value=%s, err=%s\n", old, err)
			break
		}
		oldBytes := toBytes(oldGb)
		log.Printf("[DEBUG] smart compare disk_size: shard=%v, old=%d\n", resources.GetDiskSize(), oldBytes)
		return resources.GetDiskSize() == oldBytes
	case "resource_preset_id":
		log.Printf("[DEBUG] smart compare resource_preset_id: shard=%v, old=%s, new=%s\n", resources.GetResourcePresetId(), old, new)
		return resources.GetResourcePresetId() == old
	case "disk_type_id":
		log.Printf("[DEBUG] smart compare disk_type_id: shard=%v, old=%s\n", resources.GetDiskTypeId(), old)
		return resources.GetDiskTypeId() == old
	default:
		log.Printf("[ERROR] wrong key=%s\n", k)
	}
	return defaultResourcesCompare(old, new)
}

func defaultResourcesCompare(old, new string) bool {
	log.Println("[DEBUG] default compare resources with value from cluster spec")
	return old == new
}

func dropShardsWithDefaultResources(clusterResources *clickhouse.Resources, shardsResources map[string]*clickhouse.ShardConfigSpec) {
	log.Println("[DEBUG] try to drop shards with default cluster resources.")
	for shardName, shardSpec := range shardsResources {
		log.Printf("[DEBUG] check shard: shard_name=%s, shard_spec=%v\n", shardName, shardSpec)

		shardResources := shardSpec.Clickhouse.Resources
		if shardResources == nil {
			log.Printf("[DEBUG] shard_name=%s, shard_spec is empty. skip.\n", shardName)
			continue
		}

		log.Printf("[DEBUG] shard_name=%s: compare resources: resources from cluster=%v, resources from shard=%v\n", shardName, clusterResources, shardResources)
		if isEqualResources(clusterResources, shardResources) {
			log.Printf("[DEBUG] shard_name=%s is default. drop.\n", shardName)
			delete(shardsResources, shardName)
			continue
		}
		log.Printf("[DEBUG] shard_name=%s isn't default. skip.\n", shardName)
	}
}

func compareClusterResources(k, old, new string, updatedSchema *schema.ResourceData) bool {
	if len(old) == 0 {
		return defaultResourcesCompare(old, new)
	}
	log.Printf("[DEBUG] compareClusterResources: old=%s, new=%s, key=%s\n", old, new, k)
	log.Printf("[DEBUG] original cluster schema: cluster=%v\n", originalClusterResources)

	// originalClusterResources is nil for terraform plan and for first apply.
	if originalClusterResources == nil {
		log.Println("[DEBUG] original cluster resources is nil. default compare.")
		return defaultResourcesCompare(old, new)
	}

	updatedClusterResources := expandClickHouseResources(updatedSchema, "clickhouse.0.resources.0")
	log.Printf("[DEBUG] updated cluster schema: cluster=%v\n", updatedClusterResources)

	updatedShardsResources, _ := expandClickhouseShardSpecs(updatedSchema)
	log.Printf("[DEBUG] updated shards schema: shards=%v\n", updatedShardsResources)

	dropShardsWithDefaultResources(originalClusterResources, updatedShardsResources)

	log.Printf("[DEBUG] current shards schema after drop shards with default resources: shards=%v\n", updatedShardsResources)

	hosts, err := expandClickHouseHosts(updatedSchema)
	if err != nil || len(hosts) == 0 {
		log.Printf("[DEBUG] compareClusterResources: error expand hosts schema: %s\n", err)
		return defaultResourcesCompare(old, new)
	}

	if configSpec, ok := updatedShardsResources[hosts[0].ShardName]; ok && configSpec != nil && configSpec.Clickhouse.Resources != nil {
		log.Println("[DEBUG] compareClusterResources: shard for first host specify in current shard resources schema. smart compare.")
		return compareResourcesFields(k, old, new, configSpec)
	}
	log.Println("[DEBUG] compareClusterResources: shard for first host not specify in current shard resources schema. default compare.")
	return defaultResourcesCompare(old, new)
}

func setClickHouseFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Clickhouse().Cluster().Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	folderID, ok := d.GetOk("folder_id")
	if !ok {
		return nil
	}
	if folderID == "" {
		return nil
	}

	if cluster.FolderId != folderID {
		request := &clickhouse.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending ClickHouse cluster move request: %+v", request)
			return config.sdk.MDB().Clickhouse().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move ClickHouse Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving ClickHouse Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving ClickHouse Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
