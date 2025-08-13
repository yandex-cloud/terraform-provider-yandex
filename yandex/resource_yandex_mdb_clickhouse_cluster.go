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
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 60 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 90 * time.Minute
	yandexMDBClickHouseClusterPollInterval  = 10 * time.Second
)

var yandexMDBClickhouseRetryOperationConfig = &OperationRetryConfig{
	retriableCodes: []codes.Code{codes.Internal, codes.Unavailable, codes.DeadlineExceeded},
	retryCount:     3,
	retryInterval:  2 * time.Minute,
	pollInterval:   5 * time.Second,
}

var schemaResources = map[string]*schema.Schema{
	"resource_preset_id": {
		Type:        schema.TypeString,
		Description: "The ID of the preset for computational resources available to a ClickHouse host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",
		Optional:    true,
		Computed:    true,
	},
	"disk_size": {
		Type:        schema.TypeInt,
		Description: "Volume of the storage available to a ClickHouse host, in gigabytes.",
		Optional:    true,
		Computed:    true,
	},
	"disk_type_id": {
		Type:        schema.TypeString,
		Description: "Type of the storage of ClickHouse hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
		Optional:    true,
		Computed:    true,
	},
}
var schemaConfig = map[string]*schema.Schema{
	"log_level":                                     {Type: schema.TypeString, Optional: true, Computed: true, Description: "Logging level."},
	"max_connections":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Max server connections."},
	"max_concurrent_queries":                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit on total number of concurrently executed queries."},
	"keep_alive_timeout":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The number of seconds that ClickHouse waits for incoming requests for HTTP protocol before closing the connection."},
	"uncompressed_cache_size":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Cache size (in bytes) for uncompressed data used by table engines from the MergeTree family. Zero means disabled."},
	"mark_cache_size":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum size of cache for marks "},
	"max_table_size_to_drop":                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Restriction on deleting tables."},
	"max_partition_size_to_drop":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Restriction on dropping partitions."},
	"timezone":                                      {Type: schema.TypeString, Optional: true, Computed: true, Description: "The server's time zone."},
	"geobase_uri":                                   {Type: schema.TypeString, Optional: true, Computed: true, Description: "Address of the archive with the user geobase in Object Storage."},
	"geobase_enabled":                               {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable geobase."},
	"query_log_retention_size":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that query_log can grow to before old data will be removed."},
	"query_log_retention_time":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that query_log records will be retained before removal."},
	"query_thread_log_enabled":                      {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable query_thread_log system table."},
	"query_thread_log_retention_size":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that query_thread_log can grow to before old data will be removed."},
	"query_thread_log_retention_time":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that query_thread_log records will be retained before removal."},
	"part_log_retention_size":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that part_log can grow to before old data will be removed."},
	"part_log_retention_time":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that part_log records will be retained before removal."},
	"metric_log_enabled":                            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable metric_log system table."},
	"metric_log_retention_size":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that metric_log can grow to before old data will be removed."},
	"metric_log_retention_time":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that metric_log records will be retained before removal."},
	"trace_log_enabled":                             {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable trace_log system table."},
	"trace_log_retention_size":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that trace_log can grow to before old data will be removed."},
	"trace_log_retention_time":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that trace_log records will be retained before removal."},
	"text_log_enabled":                              {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable text_log system table."},
	"text_log_retention_size":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that text_log can grow to before old data will be removed."},
	"text_log_retention_time":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that text_log records will be retained before removal."},
	"opentelemetry_span_log_enabled":                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable opentelemetry_span_log system table."},
	"opentelemetry_span_log_retention_size":         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that opentelemetry_span_log can grow to before old data will be removed."},
	"opentelemetry_span_log_retention_time":         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that opentelemetry_span_log records will be retained before removal."},
	"query_views_log_enabled":                       {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable query_views_log system table."},
	"query_views_log_retention_size":                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that query_views_log can grow to before old data will be removed."},
	"query_views_log_retention_time":                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that query_views_log records will be retained before removal."},
	"asynchronous_metric_log_enabled":               {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable asynchronous_metric_log system table."},
	"asynchronous_metric_log_retention_size":        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that asynchronous_metric_log can grow to before old data will be removed."},
	"asynchronous_metric_log_retention_time":        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that asynchronous_metric_log records will be retained before removal."},
	"session_log_enabled":                           {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable session_log system table."},
	"session_log_retention_size":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that session_log can grow to before old data will be removed."},
	"session_log_retention_time":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that session_log records will be retained before removal."},
	"zookeeper_log_enabled":                         {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable zookeeper_log system table."},
	"zookeeper_log_retention_size":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that zookeeper_log can grow to before old data will be removed."},
	"zookeeper_log_retention_time":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that zookeeper_log records will be retained before removal."},
	"asynchronous_insert_log_enabled":               {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable or disable asynchronous_insert_log system table."},
	"asynchronous_insert_log_retention_size":        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size that asynchronous_insert_log can grow to before old data will be removed."},
	"asynchronous_insert_log_retention_time":        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum time that asynchronous_insert_log records will be retained before removal."},
	"text_log_level":                                {Type: schema.TypeString, Optional: true, Computed: true, Description: "Logging level for text_log system table."},
	"background_pool_size":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the number of threads performing background merges and mutations for MergeTree-engine tables."},
	"background_schedule_pool_size":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for constantly executing some lightweight periodic operations for replicated tables, Kafka streaming, and DNS cache updates."},
	"background_fetches_pool_size":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for fetching data parts from another replica for MergeTree-engine tables in a background."},
	"background_move_pool_size":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for moving data parts to another disk or volume for MergeTree-engine tables in a background."},
	"background_distributed_schedule_pool_size":     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for executing distributed sends."},
	"background_buffer_flush_schedule_pool_size":    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for performing flush operations for Buffer-engine tables in the background."},
	"background_message_broker_schedule_pool_size":  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for executing background operations for message streaming."},
	"background_common_pool_size":                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads that will be used for performing a variety of operations (mostly garbage collection) for MergeTree-engine tables in a background."},
	"background_merges_mutations_concurrency_ratio": {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets a ratio between the number of threads and the number of background merges and mutations that can be executed concurrently."},
	"default_database":                              {Type: schema.TypeString, Optional: true, Computed: true, Description: "Default database name."},
	"total_memory_profiler_step":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Whenever server memory usage becomes larger than every next step in number of bytes the memory profiler will collect the allocating stack trace."},
	"dictionaries_lazy_load":                        {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Lazy loading of dictionaries. If true, then each dictionary is loaded on the first use."},

	"merge_tree": {
		Type:        schema.TypeList,
		Description: "MergeTree engine configuration.",
		MaxItems:    1,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"replicated_deduplication_window":                           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Replicated deduplication window: Number of recent hash blocks that ZooKeeper will store (the old ones will be deleted)."},
				"replicated_deduplication_window_seconds":                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Replicated deduplication window seconds: Time during which ZooKeeper stores the hash blocks (the old ones wil be deleted)."},
				"parts_to_delay_insert":                                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Parts to delay insert: Number of active data parts in a table, on exceeding which ClickHouse starts artificially reduce the rate of inserting data into the table"},
				"parts_to_throw_insert":                                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Parts to throw insert: Threshold value of active data parts in a table, on exceeding which ClickHouse throws the 'Too many parts ...' exception."},
				"inactive_parts_to_delay_insert":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If the number of inactive parts in a single partition in the table at least that many the inactive_parts_to_delay_insert value, an INSERT artificially slows down. It is useful when a server fails to clean up parts quickly enough."},
				"inactive_parts_to_throw_insert":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If the number of inactive parts in a single partition more than the inactive_parts_to_throw_insert value, INSERT is interrupted with the `Too many inactive parts (N). Parts cleaning are processing significantly slower than inserts` exception."},
				"max_replicated_merges_in_queue":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Max replicated merges in queue: Maximum number of merge tasks that can be in the ReplicatedMergeTree queue at the same time."},
				"number_of_free_entries_in_pool_to_lower_max_size_of_merge": {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Number of free entries in pool to lower max size of merge: Threshold value of free entries in the pool. If the number of entries in the pool falls below this value, ClickHouse reduces the maximum size of a data part to merge. This helps handle small merges faster, rather than filling the pool with lengthy merges."},
				"max_bytes_to_merge_at_min_space_in_pool":                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Max bytes to merge at min space in pool: Maximum total size of a data part to merge when the number of free threads in the background pool is minimum."},
				"max_bytes_to_merge_at_max_space_in_pool":                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum total parts size (in bytes) to be merged into one part, if there are enough resources available. max_bytes_to_merge_at_max_space_in_pool -- roughly corresponds to the maximum possible part size created by an automatic background merge."},
				"min_bytes_for_wide_part":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimum number of bytes in a data part that can be stored in Wide format. You can set one, both or none of these settings."},
				"min_rows_for_wide_part":                                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimum number of rows in a data part that can be stored in Wide format. You can set one, both or none of these settings."},
				"ttl_only_drop_parts":                                       {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables zero-copy replication when a replica is located on a remote filesystem."},
				"allow_remote_fs_zero_copy_replication":                     {Type: schema.TypeBool, Optional: true, Computed: true, Description: "When this setting has a value greater than zero only a single replica starts the merge immediately if merged part on shared storage and allow_remote_fs_zero_copy_replication is enabled."},
				"merge_with_ttl_timeout":                                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimum delay in seconds before repeating a merge with delete TTL. Default value: 14400 seconds (4 hours)."},
				"merge_with_recompression_ttl_timeout":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimum delay in seconds before repeating a merge with recompression TTL. Default value: 14400 seconds (4 hours)."},
				"max_parts_in_total":                                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum number of parts in all partitions."},
				"max_number_of_merges_with_ttl_in_pool":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "When there is more than specified number of merges with TTL entries in pool, do not assign new merge with TTL."},
				"cleanup_delay_period":                                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimum period to clean old queue logs, blocks hashes and parts."},
				"number_of_free_entries_in_pool_to_execute_mutation":        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "When there is less than specified number of free entries in pool, do not execute part mutations. This is to leave free threads for regular merges and avoid `Too many parts`. Default value: 20."},
				"max_avg_part_size_for_too_many_parts":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The `too many parts` check according to `parts_to_delay_insert` and `parts_to_throw_insert` will be active only if the average part size (in the relevant partition) is not larger than the specified threshold. If it is larger than the specified threshold, the INSERTs will be neither delayed or rejected. This allows to have hundreds of terabytes in a single table on a single server if the parts are successfully merged to larger parts. This does not affect the thresholds on inactive parts or total parts."},
				"min_age_to_force_merge_seconds":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Merge parts if every part in the range is older than the value of `min_age_to_force_merge_seconds`."},
				"min_age_to_force_merge_on_partition_only":                  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Whether min_age_to_force_merge_seconds should be applied only on the entire partition and not on subset."},
				"merge_selecting_sleep_ms":                                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sleep time for merge selecting when no part is selected. A lower setting triggers selecting tasks in background_schedule_pool frequently, which results in a large number of requests to ClickHouse Keeper in large-scale clusters."},
				"merge_max_block_size":                                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The number of rows that are read from the merged parts into memory. Default value: 8192."},
				"check_sample_column_is_correct":                            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables the check at table creation, that the data type of a column for sampling or sampling expression is correct. The data type must be one of unsigned integer types: UInt8, UInt16, UInt32, UInt64. Default value: true."},
				"max_merge_selecting_sleep_ms":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum sleep time for merge selecting, a lower setting will trigger selecting tasks in background_schedule_pool frequently which result in large amount of requests to zookeeper in large-scale clusters. Default value: 60000 milliseconds (60 seconds)."},
				"max_cleanup_delay_period":                                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum period to clean old queue logs, blocks hashes and parts. Default value: 300 seconds."},
			},
		},
	},
	"kafka": {
		Type:        schema.TypeList,
		Description: "Kafka connection configuration.",
		MaxItems:    1,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"security_protocol":                   {Type: schema.TypeString, Optional: true, Computed: true, Description: "Security protocol used to connect to kafka server."},
				"sasl_mechanism":                      {Type: schema.TypeString, Optional: true, Computed: true, Description: "SASL mechanism used in kafka authentication."},
				"sasl_username":                       {Type: schema.TypeString, Optional: true, Computed: true, Description: "Username on kafka server."},
				"sasl_password":                       {Type: schema.TypeString, Optional: true, Sensitive: true, Computed: true, Description: "User password on kafka server."},
				"enable_ssl_certificate_verification": {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable verification of SSL certificates."},
				"max_poll_interval_ms":                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum allowed time between calls to consume messages (e.g., `rd_kafka_consumer_poll()` for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member."},
				"session_timeout_ms":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance."},
				"debug":                               {Type: schema.TypeString, Optional: true, Computed: true, Description: "A comma-separated list of debug contexts to enable."},
				"auto_offset_reset":                   {Type: schema.TypeString, Optional: true, Computed: true, Description: "Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error (ERR__AUTO_OFFSET_RESET) which is retrieved by consuming messages and checking 'message->err'."},
			},
		},
	},
	"kafka_topic": {
		Type:        schema.TypeList,
		Description: "Kafka topic connection configuration.",
		MinItems:    0,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {Type: schema.TypeString, Required: true, Description: "Kafka topic name."},
				"settings": {
					Type:        schema.TypeList,
					Description: "Kafka connection settings.",
					MinItems:    0,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"security_protocol":                   {Type: schema.TypeString, Optional: true, Description: "Security protocol used to connect to kafka server."},
							"sasl_mechanism":                      {Type: schema.TypeString, Optional: true, Description: "SASL mechanism used in kafka authentication."},
							"sasl_username":                       {Type: schema.TypeString, Optional: true, Description: "Username on kafka server."},
							"sasl_password":                       {Type: schema.TypeString, Optional: true, Sensitive: true, Description: "User password on kafka server."},
							"enable_ssl_certificate_verification": {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable verification of SSL certificates."},
							"max_poll_interval_ms":                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum allowed time between calls to consume messages (e.g., `rd_kafka_consumer_poll()` for high-level consumers. If this interval is exceeded the consumer is considered failed and the group will rebalance in order to reassign the partitions to another consumer group member."},
							"session_timeout_ms":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Client group session and failure detection timeout. The consumer sends periodic heartbeats (heartbeat.interval.ms) to indicate its liveness to the broker. If no hearts are received by the broker for a group member within the session timeout, the broker will remove the consumer from the group and trigger a rebalance."},
							"debug":                               {Type: schema.TypeString, Optional: true, Computed: true, Description: "A comma-separated list of debug contexts to enable."},
							"auto_offset_reset":                   {Type: schema.TypeString, Optional: true, Computed: true, Description: "Action to take when there is no initial offset in offset store or the desired offset is out of range: 'smallest','earliest' - automatically reset the offset to the smallest offset, 'largest','latest' - automatically reset the offset to the largest offset, 'error' - trigger an error (ERR__AUTO_OFFSET_RESET) which is retrieved by consuming messages and checking 'message->err'."},
						},
					},
				},
			},
		},
	},
	"rabbitmq": {
		Type:        schema.TypeList,
		Description: "RabbitMQ connection configuration.",
		MaxItems:    1,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {Type: schema.TypeString, Optional: true, Computed: true, Description: "RabbitMQ username."},
				"password": {Type: schema.TypeString, Optional: true, Sensitive: true, Computed: true, Description: "RabbitMQ user password."},
				"vhost":    {Type: schema.TypeString, Optional: true, Computed: true, Description: "RabbitMQ vhost. Default: `\\`."},
			},
		},
	},
	"compression": {
		Type:        schema.TypeList,
		Description: "Data compression configuration.",
		MinItems:    0,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"method":              {Type: schema.TypeString, Required: true, Description: "Compression method. Two methods are available: `LZ4` and `zstd`."},
				"min_part_size":       {Type: schema.TypeInt, Required: true, Description: "Min part size: Minimum size (in bytes) of a data part in a table. ClickHouse only applies the rule to tables with data parts greater than or equal to the Min part size value."},
				"min_part_size_ratio": {Type: schema.TypeFloat, Required: true, Description: "Min part size ratio: Minimum table part size to total table size ratio. ClickHouse only applies the rule to tables in which this ratio is greater than or equal to the Min part size ratio value."},
				"level":               {Type: schema.TypeInt, Optional: true, Description: " Compression level for `ZSTD` method."},
			},
		},
	},
	"graphite_rollup": {
		Type:        schema.TypeList,
		Description: "Graphite rollup configuration.",
		MinItems:    0,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":                {Type: schema.TypeString, Required: true, Description: "Graphite rollup configuration name."},
				"path_column_name":    {Type: schema.TypeString, Optional: true, Computed: true, Description: "The name of the column storing the metric name (Graphite sensor). Default value: Path."},
				"time_column_name":    {Type: schema.TypeString, Optional: true, Computed: true, Description: "The name of the column storing the time of measuring the metric. Default value: Time."},
				"value_column_name":   {Type: schema.TypeString, Optional: true, Computed: true, Description: "The name of the column storing the value of the metric at the time set in `time_column_name`. Default value: Value."},
				"version_column_name": {Type: schema.TypeString, Optional: true, Computed: true, Description: "The name of the column storing the version of the metric. Default value: Timestamp."},
				"pattern": {
					Type:        schema.TypeList,
					Description: "Set of thinning rules.",
					MinItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"regexp":   {Type: schema.TypeString, Optional: true, Computed: true, Description: "Regular expression that the metric name must match."},
							"function": {Type: schema.TypeString, Required: true, Description: "Aggregation function name."},
							"retention": {
								Type:        schema.TypeList,
								Description: "Retain parameters.",
								MinItems:    0,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"age":       {Type: schema.TypeInt, Required: true, Description: "Minimum data age in seconds."},
										"precision": {Type: schema.TypeInt, Required: true, Description: "Accuracy of determining the age of the data in seconds."},
									},
								},
							},
						},
					},
				},
			},
		},
	},
	"query_masking_rules": {
		Type:        schema.TypeList,
		Description: "Query masking rules configuration.",
		MinItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Optional: true, Computed: true, Description: "Name for the rule."},
				"regexp":  {Type: schema.TypeString, Required: true, Description: "RE2 compatible regular expression."},
				"replace": {Type: schema.TypeString, Optional: true, Computed: true, Description: "Substitution string for sensitive data. Default value: six asterisks."},
			},
		},
	},
	"query_cache": {
		Type:        schema.TypeList,
		Description: "Query cache configuration.",
		MaxItems:    1,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"max_size_in_bytes":       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum cache size in bytes. 0 means the query cache is disabled. Default value: 1073741824 (1 GiB)."},
				"max_entries":             {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of SELECT query results stored in the cache. Default value: 1024."},
				"max_entry_size_in_bytes": {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size in bytes SELECT query results may have to be saved in the cache. Default value: 1048576 (1 MiB)."},
				"max_entry_size_in_rows":  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of rows SELECT query results may have to be saved in the cache. Default value: 30000000 (30 mil)."},
			},
		},
	},
	"jdbc_bridge": {
		Type:        schema.TypeList,
		Description: "JDBC bridge configuration.",
		MaxItems:    1,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host": {Type: schema.TypeString, Required: true, Description: "Host of jdbc bridge."},
				"port": {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Port of jdbc bridge. Default value: 9019."},
			},
		},
	},
}

func resourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a ClickHouse cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",

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
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The cluster identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the ClickHouse cluster. Can be either `PRESTABLE` or `PRODUCTION`.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseClickHouseEnv),
			},
			"clickhouse": {
				Type:        schema.TypeList,
				Description: "Configuration of the ClickHouse subcluster.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"config": {
							Type:        schema.TypeList,
							Description: "ClickHouse server parameters. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/settings-list).",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: schemaConfig,
							},
						},
						"resources": {
							Type:             schema.TypeList,
							Description:      "Resources allocated to hosts of the ClickHouse subcluster.",
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
				Type:        schema.TypeSet,
				Description: "A user of the ClickHouse cluster.",
				Optional:    true,
				Set:         clickHouseUserHash,
				Deprecated:  useResourceInstead("user", "yandex_mdb_clickhouse_user"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the user.",
							Required:    true,
						},
						"password": {
							Type:        schema.TypeString,
							Description: "The password of the user.",
							Optional:    true,
							Sensitive:   true,
						},
						"permission": {
							Type:        schema.TypeSet,
							Description: "Set of permissions granted to the user.",
							Optional:    true,
							Computed:    true,
							Set:         clickHouseUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: "The name of the database that the permission grants access to.",
										Required:    true,
									},
								},
							},
						},
						"connection_manager": {
							Type:        schema.TypeMap,
							Description: "Connection Manager connection configuration. Filled in by the server automatically.",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"generate_password": {
							Type:        schema.TypeBool,
							Description: "Generate password using Connection Manager. Allowed values: `true` or `false`. It's used only during user creation and is ignored during updating.\n\n~> **Must specify either password or generate_password**.\n",
							Optional:    true,
							Default:     false,
						},
						"settings": {
							Type:        schema.TypeList,
							Description: "Custom settings for user.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"readonly":                      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Restricts permissions for reading data, write data and change settings queries."},
									"allow_ddl":                     {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Allows or denies DDL queries."},
									"insert_quorum":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Enables the quorum writes."},
									"connect_timeout":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Connect timeout in milliseconds on the socket used for communicating with the client."},
									"receive_timeout":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Receive timeout in milliseconds on the socket used for communicating with the client."},
									"send_timeout":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Send timeout in milliseconds on the socket used for communicating with the client."},
									"insert_quorum_timeout":         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Write to a quorum timeout in milliseconds."},
									"insert_quorum_parallel":        {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables parallelism for quorum INSERT queries."},
									"select_sequential_consistency": {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables sequential consistency for SELECT queries."},
									"deduplicate_blocks_in_dependent_materialized_views": {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables the deduplication check for materialized views that receive data from `Replicated` tables."},
									"max_replica_delay_for_distributed_queries":          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Disables lagging replicas for distributed queries."},
									"fallback_to_stale_replicas_for_distributed_queries": {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Forces a query to an out-of-date replica if updated data is not available."},
									"replication_alter_partitions_sync":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "For ALTER ... ATTACH|DETACH|DROP queries, you can use the replication_alter_partitions_sync setting to set up waiting."},
									"distributed_product_mode":                           {Type: schema.TypeString, Optional: true, Computed: true, Description: "Changes the behavior of distributed subqueries."},
									"distributed_aggregation_memory_efficient":           {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Determine the behavior of distributed subqueries."},
									"distributed_ddl_task_timeout":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout for DDL queries, in milliseconds."},
									"skip_unavailable_shards":                            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables silently skipping of unavailable shards."},
									"compile":                                            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enable compilation of queries."},
									"min_count_to_compile":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "How many times to potentially use a compiled chunk of code before running compilation."},
									"compile_expressions":                                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Turn on expression compilation."},
									"min_count_to_compile_expression":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "A query waits for expression compilation process to complete prior to continuing execution."},
									"max_block_size":                                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "A recommendation for what size of the block (in a count of rows) to load from tables."},
									"min_insert_block_size_rows":                         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the minimum number of rows in the block which can be inserted into a table by an INSERT query."},
									"min_insert_block_size_bytes":                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the minimum number of bytes in the block which can be inserted into a table by an INSERT query."},
									"max_insert_block_size":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The size of blocks (in a count of rows) to form for insertion into a table."},
									"min_bytes_to_use_direct_io":                         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The minimum data volume required for using direct I/O access to the storage disk."},
									"use_uncompressed_cache":                             {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Whether to use a cache of uncompressed blocks."},
									"merge_tree_max_rows_to_use_cache":                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If ClickHouse should read more than merge_tree_max_rows_to_use_cache rows in one query, it doesn’t use the cache of uncompressed blocks."},
									"merge_tree_max_bytes_to_use_cache":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If ClickHouse should read more than merge_tree_max_bytes_to_use_cache bytes in one query, it doesn’t use the cache of uncompressed blocks."},
									"merge_tree_min_rows_for_concurrent_read":            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If the number of rows to be read from a file of a MergeTree table exceeds merge_tree_min_rows_for_concurrent_read then ClickHouse tries to perform a concurrent reading from this file on several threads."},
									"merge_tree_min_bytes_for_concurrent_read":           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "If the number of bytes to read from one file of a MergeTree-engine table exceeds merge_tree_min_bytes_for_concurrent_read, then ClickHouse tries to concurrently read from this file in several threads."},
									"max_bytes_before_external_group_by":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit in bytes for using memory for GROUP BY before using swap on disk."},
									"max_bytes_before_external_sort":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "This setting is equivalent of the max_bytes_before_external_group_by setting, except for it is for sort operation (ORDER BY), not aggregation."},
									"group_by_two_level_threshold":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the threshold of the number of keys, after that the two-level aggregation should be used."},
									"group_by_two_level_threshold_bytes":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the threshold of the number of bytes, after that the two-level aggregation should be used."},
									"priority":                                           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Query priority."},
									"max_threads":                                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of query processing threads, excluding threads for retrieving data from remote servers."},
									"max_memory_usage":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum memory usage (in bytes) for processing queries on a single server."},
									"max_memory_usage_for_user":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum memory usage (in bytes) for processing of user's queries on a single server."},
									"max_network_bandwidth":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the speed of the data exchange over the network in bytes per second."},
									"max_network_bandwidth_for_user":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the speed of the data exchange over the network in bytes per second."},
									"force_index_by_date":                                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Disables query execution if the index can’t be used by date."},
									"force_primary_key":                                  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Disables query execution if indexing by the primary key is not possible."},
									"max_rows_to_read":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of rows that can be read from a table when running a query."},
									"max_bytes_to_read":                                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of bytes (uncompressed data) that can be read from a table when running a query."},
									"read_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow while read. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_rows_to_group_by":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of unique keys received from aggregation function."},
									"group_by_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow while GROUP BY operation. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n* `any` - perform approximate GROUP BY operation by continuing aggregation for the keys that got into the set, but don’t add new keys to the set.\n"},
									"max_rows_to_sort":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of rows that can be read from a table for sorting."},
									"max_bytes_to_sort":                                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of bytes (uncompressed data) that can be read from a table for sorting."},
									"sort_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow while sort. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_result_rows":                                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the number of rows in the result."},
									"max_result_bytes":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the number of bytes in the result."},
									"result_overflow_mode":                               {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow in result. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_rows_in_distinct":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of different rows when using DISTINCT."},
									"max_bytes_in_distinct":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum size of a hash table in bytes (uncompressed data) when using DISTINCT."},
									"distinct_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow when using DISTINCT. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_rows_to_transfer":                               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of rows that can be passed to a remote server or saved in a temporary table when using GLOBAL IN."},
									"max_bytes_to_transfer":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of bytes (uncompressed data) that can be passed to a remote server or saved in a temporary table when using GLOBAL IN."},
									"transfer_overflow_mode":                             {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_execution_time":                                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum query execution time in milliseconds."},
									"timeout_overflow_mode":                              {Type: schema.TypeString, Optional: true, Computed: true, Description: " Sets behavior on overflow. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_rows_in_set":                                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit on the number of rows in the set resulting from the execution of the IN section."},
									"max_bytes_in_set":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit on the number of bytes in the set resulting from the execution of the IN section."},
									"set_overflow_mode":                                  {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow in the set resulting. Possible values:\n  * `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"max_rows_in_join":                                   {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit on maximum size of the hash table for JOIN, in rows."},
									"max_bytes_in_join":                                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limit on maximum size of the hash table for JOIN, in bytes."},
									"join_overflow_mode":                                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Sets behavior on overflow in JOIN. Possible values:\n* `throw` - abort query execution, return an error.\n* `break` - stop query execution, return partial result.\n"},
									"join_algorithm": {
										Type:        schema.TypeList,
										Description: "Specifies which JOIN algorithm is used. Possible values:\n* `hash` - hash join algorithm is used. The most generic implementation that supports all combinations of kind and strictness and multiple join keys that are combined with OR in the JOIN ON section.\n* `parallel_hash` - a variation of hash join that splits the data into buckets and builds several hash tables instead of one concurrently to speed up this process.\n* `partial_merge` - a variation of the sort-merge algorithm, where only the right table is fully sorted.\n* `direct` - this algorithm can be applied when the storage for the right table supports key-value requests.\n* `auto` - when set to auto, hash join is tried first, and the algorithm is switched on the fly to another algorithm if the memory limit is violated.\n* `full_sorting_merge` - sort-merge algorithm with full sorting joined tables before joining.\n* `prefer_partial_merge` - clickHouse always tries to use partial_merge join if possible, otherwise, it uses hash. Deprecated, same as partial_merge,hash.\n",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Optional:    true,
										Computed:    true,
									},
									"any_join_distinct_right_table_keys":            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables legacy ClickHouse server behavior in ANY INNER|LEFT JOIN operations."},
									"max_columns_to_read":                           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of columns that can be read from a table in a single query."},
									"max_temporary_columns":                         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, including constant columns."},
									"max_temporary_non_const_columns":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of temporary columns that must be kept in RAM at the same time when running a query, excluding constant columns."},
									"max_query_size":                                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum part of a query that can be taken to RAM for parsing with the SQL parser."},
									"max_ast_depth":                                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum abstract syntax tree depth."},
									"max_ast_elements":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum abstract syntax tree elements."},
									"max_expanded_ast_elements":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum abstract syntax tree depth after after expansion of aliases."},
									"min_execution_speed":                           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimal execution speed in rows per second."},
									"min_execution_speed_bytes":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Minimal execution speed in bytes per second."},
									"count_distinct_implementation":                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Specifies which of the uniq* functions should be used to perform the COUNT(DISTINCT …) construction."},
									"input_format_values_interpret_expressions":     {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables the full SQL parser if the fast stream parser can’t parse the data."},
									"input_format_defaults_for_omitted_fields":      {Type: schema.TypeBool, Optional: true, Computed: true, Description: "When performing INSERT queries, replace omitted input column values with default values of the respective columns."},
									"input_format_null_as_default":                  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables the initialization of NULL fields with default values, if data type of these fields is not nullable."},
									"input_format_with_names_use_header":            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables checking the column order when inserting data."},
									"output_format_json_quote_64bit_integers":       {Type: schema.TypeBool, Optional: true, Computed: true, Description: "If the value is true, integers appear in quotes when using JSON* Int64 and UInt64 formats (for compatibility with most JavaScript implementations); otherwise, integers are output without the quotes."},
									"output_format_json_quote_denormals":            {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables +nan, -nan, +inf, -inf outputs in JSON output format."},
									"low_cardinality_allow_in_native_format":        {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Allows or restricts using the LowCardinality data type with the Native format."},
									"empty_result_for_aggregation_by_empty_set":     {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Allows to return empty result."},
									"joined_subquery_requires_alias":                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Require aliases for subselects and table functions in FROM that more than one table is present."},
									"join_use_nulls":                                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Sets the type of JOIN behavior. When merging tables, empty cells may appear. ClickHouse fills them differently based on this setting."},
									"transform_null_in":                             {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables equality of NULL values for IN operator."},
									"http_connection_timeout":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout for HTTP connection in milliseconds."},
									"http_receive_timeout":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout for HTTP connection in milliseconds."},
									"http_send_timeout":                             {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout for HTTP connection in milliseconds."},
									"enable_http_compression":                       {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables data compression in the response to an HTTP request."},
									"send_progress_in_http_headers":                 {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables `X-ClickHouse-Progress` HTTP response headers in clickhouse-server responses."},
									"http_headers_progress_interval":                {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets minimal interval between notifications about request process in HTTP header X-ClickHouse-Progress."},
									"add_http_cors_header":                          {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Include CORS headers in HTTP responses."},
									"quota_mode":                                    {Type: schema.TypeString, Optional: true, Computed: true, Description: "Quota accounting mode."},
									"max_concurrent_queries_for_user":               {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of concurrent requests per user. Default value: 0 (no limit)."},
									"memory_profiler_step":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Memory profiler step (in bytes). If the next query step requires more memory than this parameter specifies, the memory profiler collects the allocating stack trace. Values lower than a few megabytes slow down query processing. Default value: 4194304 (4 MB). Zero means disabled memory profiler."},
									"memory_profiler_sample_probability":            {Type: schema.TypeFloat, Optional: true, Computed: true, Description: "Collect random allocations and deallocations and write them into system.trace_log with 'MemorySample' trace_type. The probability is for every alloc/free regardless to the size of the allocation. Possible values: from 0 to 1. Default: 0."},
									"insert_null_as_default":                        {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables the insertion of default values instead of NULL into columns with not nullable data type. Default value: true."},
									"allow_suspicious_low_cardinality_types":        {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Allows specifying LowCardinality modifier for types of small fixed size (8 or less) in CREATE TABLE statements. Enabling this may increase merge times and memory consumption."},
									"connect_timeout_with_failover":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The timeout in milliseconds for connecting to a remote server for a Distributed table engine, if the ‘shard’ and ‘replica’ sections are used in the cluster definition. If unsuccessful, several attempts are made to connect to various replicas. Default value: 50."},
									"allow_introspection_functions":                 {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables introspections functions for query profiling."},
									"async_insert":                                  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables asynchronous inserts. Disabled by default."},
									"async_insert_threads":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads for background data parsing and insertion. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 16."},
									"wait_for_async_insert":                         {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables waiting for processing of asynchronous insertion. If enabled, server returns OK only after the data is inserted."},
									"wait_for_async_insert_timeout":                 {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The timeout (in seconds) for waiting for processing of asynchronous insertion. Value must be at least 1000 (1 second)."},
									"async_insert_max_data_size":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size of the unparsed data in bytes collected per query before being inserted. If the parameter is set to 0, asynchronous insertions are disabled. Default value: 100000."},
									"async_insert_busy_timeout":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum timeout in milliseconds since the first INSERT query before inserting collected data. If the parameter is set to 0, the timeout is disabled. Default value: 200."},
									"async_insert_stale_timeout":                    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum timeout in milliseconds since the last INSERT query before dumping collected data. If enabled, the settings prolongs the async_insert_busy_timeout with every INSERT query as long as async_insert_max_data_size is not exceeded."},
									"timeout_before_checking_execution_speed":       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout (in seconds) between checks of execution speed. It is checked that execution speed is not less that specified in min_execution_speed parameter. Must be at least 1000."},
									"cancel_http_readonly_queries_on_client_close":  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Cancels HTTP read-only queries (e.g. SELECT) when a client closes the connection without waiting for the response. Default value: false."},
									"flatten_nested":                                {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Sets the data format of a nested columns."},
									"format_regexp":                                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Regular expression (for Regexp format)."},
									"format_regexp_skip_unmatched":                  {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Skip lines unmatched by regular expression."},
									"max_http_get_redirects":                        {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits the maximum number of HTTP GET redirect hops for URL-engine tables."},
									"input_format_import_nested_json":               {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables the insertion of JSON data with nested objects."},
									"input_format_parallel_parsing":                 {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables or disables order-preserving parallel parsing of data formats. Supported only for TSV, TKSV, CSV and JSONEachRow formats."},
									"max_final_threads":                             {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Sets the maximum number of parallel threads for the SELECT query data read phase with the FINAL modifier."},
									"max_read_buffer_size":                          {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum size of the buffer to read from the filesystem."},
									"local_filesystem_read_method":                  {Type: schema.TypeString, Optional: true, Computed: true, Description: "Method of reading data from local filesystem. Possible values:\n* `read` - abort query execution, return an error.\n* `pread` - abort query execution, return an error.\n* `pread_threadpool` - stop query execution, return partial result. If the parameter is set to 0 (default), no hops is allowed.\n"},
									"remote_filesystem_read_method":                 {Type: schema.TypeString, Optional: true, Computed: true, Description: "Method of reading data from remote filesystem, one of: `read`, `threadpool`."},
									"insert_keeper_max_retries":                     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The setting sets the maximum number of retries for ClickHouse Keeper (or ZooKeeper) requests during insert into replicated MergeTree. Only Keeper requests which failed due to network error, Keeper session timeout, or request timeout are considered for retries."},
									"max_temporary_data_on_disk_size_for_user":      {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running user queries. Zero means unlimited."},
									"max_temporary_data_on_disk_size_for_query":     {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum amount of data consumed by temporary files on disk in bytes for all concurrently running queries. Zero means unlimited."},
									"max_parser_depth":                              {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Limits maximum recursion depth in the recursive descent parser. Allows controlling the stack size. Zero means unlimited."},
									"memory_overcommit_ratio_denominator":           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "It represents soft memory limit in case when hard limit is reached on user level. This value is used to compute overcommit ratio for the query. Zero means skip the query."},
									"memory_overcommit_ratio_denominator_for_user":  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "It represents soft memory limit in case when hard limit is reached on global level. This value is used to compute overcommit ratio for the query. Zero means skip the query."},
									"memory_usage_overcommit_max_wait_microseconds": {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Maximum time thread will wait for memory to be freed in the case of memory overcommit on a user level. If the timeout is reached and memory is not freed, an exception is thrown."},
									"log_query_threads":                             {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Setting up query threads logging. Query threads log into the system.query_thread_log table. This setting has effect only when log_queries is true. Queries’ threads run by ClickHouse with this setup are logged according to the rules in the query_thread_log server configuration parameter. Default value: `true`."},
									"max_insert_threads":                            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The maximum number of threads to execute the INSERT SELECT query. Default value: 0."},
									"use_hedged_requests":                           {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables hedged requests logic for remote queries. It allows to establish many connections with different replicas for query. New connection is enabled in case existent connection(s) with replica(s) were not established within hedged_connection_timeout or no data was received within receive_data_timeout. Query uses the first connection which send non empty progress packet (or data packet, if allow_changing_replica_until_first_data_packet); other connections are cancelled. Queries with max_parallel_replicas > 1 are supported. Default value: true."},
									"idle_connection_timeout":                       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Timeout to close idle TCP connections after specified number of seconds. Default value: 3600 seconds."},
									"hedged_connection_timeout_ms":                  {Type: schema.TypeInt, Optional: true, Computed: true, Description: "Connection timeout for establishing connection with replica for Hedged requests. Default value: 50 milliseconds."},
									"load_balancing":                                {Type: schema.TypeString, Optional: true, Computed: true, Description: "Specifies the algorithm of replicas selection that is used for distributed query processing, one of: random, nearest_hostname, in_order, first_or_random, round_robin. Default value: random."},
									"prefer_localhost_replica":                      {Type: schema.TypeBool, Optional: true, Computed: true, Description: "Enables/disables preferable using the localhost replica when processing distributed queries. Default value: true."},
									"date_time_input_format":                        {Type: schema.TypeString, Optional: true, Computed: true, Description: "Allows choosing a parser of the text representation of date and time, one of: `best_effort`, `basic`, `best_effort_us`. Default value: `basic`. Cloud default value: `best_effort`."},
									"date_time_output_format":                       {Type: schema.TypeString, Optional: true, Computed: true, Description: "Allows choosing different output formats of the text representation of date and time, one of: `simple`, `iso`, `unix_timestamp`. Default value: `simple`."},
								},
							},
						},
						"quota": {
							Type:        schema.TypeSet,
							Description: "Set of user quotas.",
							Optional:    true,
							Computed:    true,
							Set:         clickHouseUserQuotaHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval_duration": {Type: schema.TypeInt, Required: true, Description: "Duration of interval for quota in milliseconds."},
									"queries":           {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The total number of queries."},
									"errors":            {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The number of queries that threw exception."},
									"result_rows":       {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The total number of rows given as the result."},
									"read_rows":         {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The total number of source rows read from tables for running the query, on all remote servers."},
									"execution_time":    {Type: schema.TypeInt, Optional: true, Computed: true, Description: "The total query execution time, in milliseconds (wall time)."},
								},
							},
						},
					},
				},
			},
			"shard": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Set:         clickHouseShardHash,
				Description: "A shard of the ClickHouse cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of shard.",
							Required:    true,
						},
						"weight": {
							Type:        schema.TypeInt,
							Description: "The weight of shard.",
							Optional:    true,
							Computed:    true,
						},
						"resources": {
							Type:        schema.TypeList,
							Description: "Resources allocated to host of the shard. The resources specified for the shard takes precedence over the resources specified for the cluster.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: schemaResources,
							},
						},
					},
				},
			},
			"database": {
				Type:        schema.TypeSet,
				Description: "A database of the ClickHouse cluster.",
				Deprecated:  useResourceInstead("database", "yandex_mdb_clickhouse_database"),
				Optional:    true,
				Set:         clickHouseDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the database.",
							Required:    true,
						},
					},
				},
			},
			"copy_schema_on_new_hosts": {
				Type:        schema.TypeBool,
				Description: "Whether to copy schema on new ClickHouse hosts.",
				Optional:    true,
				Default:     true,
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the ClickHouse cluster.",
				MinItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Required:    true,
						},
						"type": {
							Type:         schema.TypeString,
							Description:  "The type of the host to be deployed. Can be either `CLICKHOUSE` or `ZOOKEEPER`.",
							Required:     true,
							ValidateFunc: validateParsableValue(parseClickHouseHostType),
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Sets whether the host should get a public IP address on creation. Can be either `true` or `false`.",
							Optional:    true,
							Default:     false,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
							Optional:    true,
							Computed:    true,
						},
						"shard_name": {
							Type:         schema.TypeString,
							Description:  "The name of the shard to which the host belongs.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
					},
				},
			},
			"shard_group": {
				Type:        schema.TypeList,
				Description: "A group of clickhouse shards.",
				MinItems:    0,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the shard group, used as cluster name in Distributed tables.",
							Required:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Description of the shard group.",
							Optional:    true,
						},
						"shard_names": {
							Type:        schema.TypeList,
							Description: "List of shards names that belong to the shard group.",
							MinItems:    1,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"format_schema": {
				Type:        schema.TypeSet,
				Description: "A set of `protobuf` or `capnproto` format schemas.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the format schema.",
							Required:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type of the format schema.",
							Required:    true,
						},
						"uri": {
							Type:        schema.TypeString,
							Description: "Format schema file URL. You can only use format schemas stored in Yandex Object Storage.",
							Required:    true,
						},
					},
				},
			},
			"ml_model": {
				Type:        schema.TypeSet,
				Description: "A group of machine learning models.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the ml model.",
							Required:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type of the model.",
							Required:    true,
						},
						"uri": {
							Type:        schema.TypeString,
							Description: "Model file URL. You can only use models stored in Yandex Object Storage.",
							Required:    true,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"version": {
				Type:         schema.TypeString,
				Description:  "Version of the ClickHouse server software.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: "Time to start the daily backup, in the UTC timezone.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:         schema.TypeInt,
							Description:  "The hour at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minutes": {
							Type:         schema.TypeInt,
							Description:  "The minute at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the ClickHouse cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"web_sql": {
							Type:        schema.TypeBool,
							Description: "Allow access for Web SQL.",
							Optional:    true,
							Default:     false,
						},
						"data_lens": {
							Type:        schema.TypeBool,
							Description: "Allow access for DataLens.",
							Optional:    true,
							Default:     false,
						},
						"metrika": {
							Type:        schema.TypeBool,
							Description: "Allow access for Yandex.Metrika.",
							Optional:    true,
							Default:     false,
						},
						"serverless": {
							Type:        schema.TypeBool,
							Description: "Allow access for Serverless.",
							Optional:    true,
							Default:     false,
						},
						"data_transfer": {
							Type:        schema.TypeBool,
							Description: "Allow access for DataTransfer.",
							Optional:    true,
							Default:     false,
						},
						"yandex_query": {
							Type:        schema.TypeBool,
							Description: "Allow access for YandexQuery.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"zookeeper": {
				Type:             schema.TypeList,
				Description:      "Configuration of the ZooKeeper subcluster.",
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressZooKeeperResourcesDIff,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:        schema.TypeList,
							Description: "Resources allocated to hosts of the ZooKeeper subcluster.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:        schema.TypeString,
										Description: "The ID of the preset for computational resources available to a ZooKeeper host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).",
										Optional:    true,
										Computed:    true,
									},
									"disk_size": {
										Type:        schema.TypeInt,
										Description: "Volume of the storage available to a ZooKeeper host, in gigabytes.",
										Optional:    true,
										Computed:    true,
									},
									"disk_type_id": {
										Type:        schema.TypeString,
										Description: "Type of the storage of ZooKeeper hosts. For more information see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts/storage).",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Aggregated health of the cluster. Can be `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-clickhouse/api-ref/Cluster/).",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster. Can be `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-clickhouse/api-ref/Cluster/).",
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"admin_password": {
				Type:        schema.TypeString,
				Description: "A password used to authorize as user `admin` when `sql_user_management` enabled.",
				Optional:    true,
				Sensitive:   true,
			},
			"sql_user_management": {
				Type:        schema.TypeBool,
				Description: "Enables `admin` user with user management permission.",
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"sql_database_management": {
				Type:        schema.TypeBool,
				Description: "Grants `admin` user database management permission.",
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"embedded_keeper": {
				Type:        schema.TypeBool,
				Description: "Whether to use ClickHouse Keeper as a coordination system and place it on the same hosts with ClickHouse. If not, it's used ZooKeeper with placement on separate hosts.",
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Optional:    true,
			},
			"cloud_storage": {
				Type:        schema.TypeList,
				Description: "Cloud Storage settings.",
				Computed:    true,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Whether to use Yandex Object Storage for storing ClickHouse data. Can be either `true` or `false`.",
							Required:    true,
						},
						"move_factor": {
							Type:        schema.TypeFloat,
							Description: "Sets the minimum free space ratio in the cluster storage. If the free space is lower than this value, the data is transferred to Yandex Object Storage. Acceptable values are 0 to 1, inclusive.",
							Optional:    true,
							Computed:    true,
						},
						"data_cache_enabled": {
							Type:        schema.TypeBool,
							Description: "Enables temporary storage in the cluster repository of data requested from the object repository.",
							Optional:    true,
							Computed:    true,
						},
						"data_cache_max_size": {
							Type:        schema.TypeInt,
							Description: "Defines the maximum amount of memory (in bytes) allocated in the cluster storage for temporary storage of data requested from the object storage.",
							Optional:    true,
							Computed:    true,
						},
						"prefer_not_to_merge": {
							Type:        schema.TypeBool,
							Description: "Disables merging of data parts in `Yandex Object Storage`.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Description:  "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							Description:  "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
							ValidateFunc: validateParsableValue(parseClickHouseWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							Description:  "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"disk_encryption_key_id": {
				Type:        schema.TypeString,
				Description: "ID of the KMS key for cluster disk encryption.",
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"backup_retain_period_days": {
				Type:        schema.TypeInt,
				Description: "The period in days during which backups are stored.",
				Optional:    true,
				Default:     7,
			},
		},
	}
}

func resourceYandexMDBClickHouseClusterCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] create started")
	backupOriginalClusterResource(d)
	config := meta.(*Config)

	req, shardsToAdd, err := prepareCreateClickHouseCreateRequest(d, config)

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

	// Will add all other ClickHouse shards, except of the first one(shard1 by default)
	err = addClickHouseShards(ctx, config, d, shardsToAdd)
	if err != nil {
		return fmt.Errorf("error while adding shards to ClickHouse Cluster: %s", err)
	}

	// First shard will always be added with default weight and cluster resources, have to check and update weight
	err = updateClickHouseFirstShard(ctx, config, d)
	if err != nil {
		return err
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
func prepareCreateClickHouseCreateRequest(d *schema.ResourceData, meta *Config) (*clickhouse.CreateClusterRequest, map[string][]*clickhouse.HostSpec, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding labels on ClickHouse Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting folder ID while creating ClickHouse Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseClickHouseEnv(e)
	if err != nil {
		return nil, nil, fmt.Errorf("Error resolving environment while creating ClickHouse Cluster: %s", err)
	}

	dbSpecs, err := expandClickHouseDatabases(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding databases on ClickHouse Cluster create: %s", err)
	}

	users, err := expandClickHouseUserSpecs(d, true)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding user specs on ClickHouse Cluster create: %s", err)
	}

	hosts, err := expandClickHouseHosts(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding hosts on ClickHouse Cluster create: %s", err)
	}

	_, toAdd, _ := clickHouseHostsDiff(nil, hosts)
	log.Printf("[DEBUG] hosts to add: %v\n", toAdd)

	firstHosts := toAdd["zk"]
	delete(toAdd, "zk")

	clickhouseConfigSpec, err := expandClickHouseSpec(d)
	if err != nil {
		return nil, nil, err
	}

	shardSpecs, err := expandClickhouseShardSpecs(d)
	if err != nil {
		return nil, nil, err
	}

	var firstShardName = ""
	// try to use default shard name as first shard
	for shardName, shardHosts := range toAdd {
		if shardName == "shard1" {
			firstHosts = append(firstHosts, shardHosts...)
			delete(toAdd, shardName)
			firstShardName = shardName
			break
		}
	}

	if firstShardName == "" {
		for shardName, shardHosts := range toAdd {
			firstHosts = append(firstHosts, shardHosts...)
			delete(toAdd, shardName)
			firstShardName = shardName
			break
		}
	}

	if firstShardSpecs, ok := shardSpecs[firstShardName]; ok {
		if !isEqualResources(clickhouseConfigSpec.Resources, firstShardSpecs.Clickhouse.Resources) {
			return nil, nil, fmt.Errorf("cluster resources should be equal to first shard resources %s", firstShardName)
		}
		if firstShardSpecs.Clickhouse.Weight.GetValue() == 0 {
			return nil, nil, fmt.Errorf("weight of first shard %s should be greater than zero", firstShardName)
		}
	}

	cloudStorage, err := expandClickHouseCloudStorage(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding cloud storage on ClickHouse Cluster create: %s", err)
	}

	configSpec := &clickhouse.ConfigSpec{
		Version:                d.Get("version").(string),
		Clickhouse:             clickhouseConfigSpec,
		Zookeeper:              expandClickHouseZookeeperSpec(d),
		BackupWindowStart:      expandClickHouseBackupWindowStart(d),
		Access:                 expandClickHouseAccess(d),
		CloudStorage:           cloudStorage,
		BackupRetainPeriodDays: expandClickhouseBackupRetainPeriodDays(d),
	}

	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding shard specs on ClickHouse Cluster create: %s", err)
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
		return nil, nil, fmt.Errorf("error while expanding network id on ClickHouse Cluster create: %s", err)
	}

	mw, err := expandClickHouseMaintenanceWindow(d)
	if err != nil {
		return nil, nil, fmt.Errorf("creation error while expand clickhouse maintenance_window: %s", err)
	}

	var diskEncryptionKeyId *wrapperspb.StringValue
	if val, ok := d.GetOk("disk_encryption_key_id"); ok {
		diskEncryptionKeyId = &wrapperspb.StringValue{
			Value: val.(string),
		}
	}

	req := clickhouse.CreateClusterRequest{
		FolderId:            folderID,
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		NetworkId:           networkID,
		Environment:         env,
		DatabaseSpecs:       dbSpecs,
		ConfigSpec:          configSpec,
		HostSpecs:           firstHosts,
		UserSpecs:           users,
		Labels:              labels,
		SecurityGroupIds:    securityGroupIds,
		ServiceAccountId:    d.Get("service_account_id").(string),
		DeletionProtection:  d.Get("deletion_protection").(bool),
		MaintenanceWindow:   mw,
		DiskEncryptionKeyId: diskEncryptionKeyId,
	}

	return &req, toAdd, nil
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

	dUsers, err := expandClickHouseUserSpecs(d, false)
	if err != nil {
		return err
	}
	passwords := clickHouseUsersPasswords(dUsers)
	generatePasswordsFlags := clickHouseUsersGeneratePasswords(dUsers)

	users, err := listClickHouseUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	us := flattenClickHouseUsers(users, passwords, generatePasswordsFlags)
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

	if cluster.DiskEncryptionKeyId != nil {
		if err = d.Set("disk_encryption_key_id", cluster.DiskEncryptionKeyId.GetValue()); err != nil {
			return err
		}
	}

	if err := d.Set("backup_retain_period_days", cluster.Config.BackupRetainPeriodDays.Value); err != nil {
		return err
	}

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
	"name":                      "name",
	"description":               "description",
	"labels":                    "labels",
	"access":                    "config_spec.access",
	"backup_window_start":       "config_spec.backup_window_start",
	"admin_password":            "config_spec.admin_password",
	"sql_user_management":       "config_spec.sql_user_management",
	"sql_database_management":   "config_spec.sql_database_management",
	"cloud_storage":             "config_spec.cloud_storage",
	"backup_retain_period_days": "config_spec.backup_retain_period_days",
	"security_group_ids":        "security_group_ids",
	"service_account_id":        "service_account_id",
	"network_id":                "network_id",
	"maintenance_window":        "maintenance_window",
	"deletion_protection":       "deletion_protection",
}

var mdbClickHouseConfigUpdateFieldsMaps = []string{
	"rabbitmq",
	"compression",
	"dictionaries",
	"graphite_rollup",
	"background_pool_size",
	"background_common_pool_size",
	"background_schedule_pool_size",
	"background_fetches_pool_size",
	"background_move_pool_size",
	"background_distributed_schedule_pool_size",
	"background_buffer_flush_schedule_pool_size",
	"background_message_broker_schedule_pool_size",
	"background_merges_mutations_concurrency_ratio",
	"log_level",
	"query_log_retention_size",
	"query_log_retention_time",
	"query_thread_log_enabled",
	"query_thread_log_retention_size",
	"query_thread_log_retention_time",
	"part_log_retention_size",
	"part_log_retention_time",
	"metric_log_enabled",
	"metric_log_retention_size",
	"metric_log_retention_time",
	"trace_log_enabled",
	"trace_log_retention_size",
	"trace_log_retention_time",
	"text_log_enabled",
	"text_log_retention_size",
	"text_log_retention_time",
	"text_log_level",
	"opentelemetry_span_log_enabled",
	"opentelemetry_span_log_retention_size",
	"opentelemetry_span_log_retention_time",
	"session_log_enabled",
	"session_log_retention_size",
	"session_log_retention_time",
	"zookeeper_log_enabled",
	"zookeeper_log_retention_size",
	"zookeeper_log_retention_time",
	"asynchronous_insert_log_enabled",
	"asynchronous_insert_log_retention_size",
	"asynchronous_insert_log_retention_time",
	"asynchronous_metric_log_enabled",
	"asynchronous_metric_log_retention_size",
	"asynchronous_metric_log_retention_time",
	"query_views_log_enabled",
	"query_views_log_retention_size",
	"query_views_log_retention_time",
	"max_connections",
	"max_concurrent_queries",
	"keep_alive_timeout",
	"uncompressed_cache_size",
	"mark_cache_size",
	"max_table_size_to_drop",
	"max_partition_size_to_drop",
	"builtin_dictionaries_reload_interval",
	"timezone",
	"geobase_uri",
	"geobase_enabled",
	"default_database",
	"total_memory_profiler_step",
	"total_memory_tracker_sample_probability",
	"query_masking_rules",
	"dictionaries_lazy_load",
	"query_cache",
	"jdbc_bridge",
}
var mdbClickhouseMergeTreeUpdateFields = []string{
	"replicated_deduplication_window",
	"replicated_deduplication_window_seconds",
	"parts_to_delay_insert",
	"parts_to_throw_insert",
	"inactive_parts_to_delay_insert",
	"inactive_parts_to_throw_insert",
	"max_replicated_merges_in_queue",
	"number_of_free_entries_in_pool_to_lower_max_size_of_merge",
	"max_bytes_to_merge_at_min_space_in_pool",
	"max_bytes_to_merge_at_max_space_in_pool",
	"min_bytes_for_wide_part",
	"min_rows_for_wide_part",
	"ttl_only_drop_parts",
	"allow_remote_fs_zero_copy_replication",
	"merge_with_ttl_timeout",
	"merge_with_recompression_ttl_timeout",
	"max_parts_in_total",
	"max_number_of_merges_with_ttl_in_pool",
	"cleanup_delay_period",
	"number_of_free_entries_in_pool_to_execute_mutation",
	"max_avg_part_size_for_too_many_parts",
	"min_age_to_force_merge_seconds",
	"min_age_to_force_merge_on_partition_only",
	"merge_selecting_sleep_ms",
	"merge_max_block_size",
	"check_sample_column_is_correct",
	"max_merge_selecting_sleep_ms",
	"max_cleanup_delay_period",
}

func updateClickHouseClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	if d.HasChanges("version") {
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

		err = op.WaitInterval(ctx, yandexMDBClickHouseClusterPollInterval)
		if err != nil {
			return fmt.Errorf("error while updating ClickHouse Cluster version %q: %s", d.Id(), err)
		}
	}

	req, err := getClickHouseClusterUpdateRequest(d, config)
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

	if d.HasChange("clickhouse.0.resources") {
		updatePath = append(updatePath, "config_spec.clickhouse.resources")
	}

	if d.HasChange("clickhouse.0.config") {
		var rootClickhouseConfigTfPath = "clickhouse.0.config.0."
		// update clickhouse config settings, if there are a changes, except kafka*
		for _, item := range mdbClickHouseConfigUpdateFieldsMaps {
			if d.HasChange(rootClickhouseConfigTfPath + item) {
				updatePath = append(updatePath, "config_spec.clickhouse.config."+item)
			}
		}
		if d.HasChange(rootClickhouseConfigTfPath + "merge_tree") {
			for _, item := range mdbClickhouseMergeTreeUpdateFields {
				if d.HasChange(rootClickhouseConfigTfPath + "merge_tree.0." + item) {
					updatePath = append(updatePath, "config_spec.clickhouse.config.merge_tree."+item)
				}
			}
		}
	}
	if d.HasChange("clickhouse.0.config.0.kafka") {
		updatePath = append(updatePath, "config_spec.clickhouse.config.kafka")
	}
	if d.HasChange("clickhouse.0.config.0.kafka_topic") {
		// Update all kafka_topics even if there is change in one of them
		updatePath = append(updatePath, "config_spec.clickhouse.config.kafka_topics")
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

	log.Printf("[DEBUG] update_request: %v update_paths: %v\n", req, updatePath)

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

func getClickHouseClusterUpdateRequest(d *schema.ResourceData, config *Config) (*clickhouse.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating ClickHouse cluster: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, config)
	if err != nil {
		return nil, fmt.Errorf("update error while expand clickhouse network_id: %s", err)
	}

	clickhouseConfigSpec, err := expandClickHouseSpec(d)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] resource for cluster update request: %+v\n", clickhouseConfigSpec.Resources)

	cloudStorage, err := expandClickHouseCloudStorage(d)
	if err != nil {
		return nil, fmt.Errorf("update error while expand clickhouse cloud_storage: %s", err)
	}

	mw, err := expandClickHouseMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("update error while expand clickhouse maintenance_window: %s", err)
	}

	req := &clickhouse.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		NetworkId:   networkID,
		ConfigSpec: &clickhouse.ConfigSpec{
			Version:                d.Get("version").(string),
			Clickhouse:             clickhouseConfigSpec,
			Zookeeper:              expandClickHouseZookeeperSpec(d),
			BackupWindowStart:      expandClickHouseBackupWindowStart(d),
			Access:                 expandClickHouseAccess(d),
			SqlUserManagement:      &wrappers.BoolValue{Value: d.Get("sql_user_management").(bool)},
			SqlDatabaseManagement:  &wrappers.BoolValue{Value: d.Get("sql_database_management").(bool)},
			CloudStorage:           cloudStorage,
			BackupRetainPeriodDays: expandClickhouseBackupRetainPeriodDays(d),
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
	targetUsers, err := expandClickHouseUserSpecs(d, false)
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
		if !isValidClickhousePasswordConfiguration(u) {
			return fmt.Errorf("must specify either password or generate_password")
		}
		err := createClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")

	// We have to calculate changed users and fields of this users via UserSpec,
	// because schema HasChange returns true on every field of user because of new hash
	changedUsers, updatedPathsOfChangedUsers := clickHouseChangedUsers(oldSpecs.(*schema.Set), newSpecs.(*schema.Set), d)

	for i, u := range changedUsers {
		if !isValidClickhousePasswordConfiguration(u) {
			return fmt.Errorf("must specify either password or generate_password")
		}
		err := updateClickHouseUser(ctx, config, d, u, updatedPathsOfChangedUsers[i])
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
		delete(toDelete, "") // no shard == zk subcluster
	}

	currShards, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return err
	}

	hostSpecsToAddToExistingShards := []*clickhouse.HostSpec{}
	hostSpecsOfNewShards := map[string][]*clickhouse.HostSpec{}
	for shardName, specs := range toAdd {
		shardExists := false
		for _, s := range currShards {
			if s.Name == shardName {
				shardExists = true
			}
		}

		if shardName != "" && shardName != "zk" && !shardExists {
			hostSpecsOfNewShards[shardName] = specs
		} else {
			hostSpecsToAddToExistingShards = append(hostSpecsToAddToExistingShards, specs...)
		}
	}

	err = addClickHouseShards(ctx, config, d, hostSpecsOfNewShards)
	if err != nil {
		return err
	}

	if len(hostSpecsToAddToExistingShards) > 0 {
		err := createClickHouseHosts(ctx, config, d, hostSpecsToAddToExistingShards)
		if err != nil {
			return err
		}
	}

	hostFqdnsToDelete := []string{}
	shardNamesToDelete := []string{}
	for shardName, fqdns := range toDelete {
		deleteShard := true
		for _, th := range targetHosts {
			if th.ShardName == shardName {
				deleteShard = false
			}
		}
		if shardName != "zk" && shardName != "" && deleteShard {
			shardNamesToDelete = append(shardNamesToDelete, shardName)
		} else {
			hostFqdnsToDelete = append(hostFqdnsToDelete, fqdns...)
		}
	}

	if len(shardNamesToDelete) > 0 {
		err = deleteClickHouseShards(ctx, config, d, shardNamesToDelete)
		if err != nil {
			return err
		}
	}

	if len(hostFqdnsToDelete) > 0 {
		err := deleteClickHouseHosts(ctx, config, d, hostFqdnsToDelete)
		if err != nil {
			return err
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

func updateClickHouseFirstShard(ctx context.Context, config *Config, d *schema.ResourceData) error {
	shardsOnCluster, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return err
	}

	firstShardName := shardsOnCluster[0].Name

	log.Printf("[DEBUG] first shard name from from cluster: %s\n", firstShardName)

	shardsFromSpec, err := expandClickhouseShardSpecs(d)
	if err != nil {
		return nil
	}

	firstShardFromSpec, ok := shardsFromSpec[firstShardName]
	if !ok {
		log.Printf("[DEBUG] not found special specs for first shard: %s\n", firstShardName)
		return nil
	}

	log.Printf("[DEBUG] before update shards got shard from schema: %+v\n", firstShardFromSpec)

	if err = updateClickHouseShard(ctx, config, d, firstShardName, firstShardFromSpec); err != nil {
		return fmt.Errorf("failed update shard from config: %s", err)
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
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Database().Create(ctx, &clickhouse.CreateDatabaseRequest{
				ClusterId: d.Id(),
				DatabaseSpec: &clickhouse.DatabaseSpec{
					Name: dbName,
				},
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while adding database to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Database().Delete(ctx, &clickhouse.DeleteDatabaseRequest{
				ClusterId:    d.Id(),
				DatabaseName: dbName,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while deleting database from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().User().Create(ctx, &clickhouse.CreateUserRequest{
				ClusterId: d.Id(),
				UserSpec:  user,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while creating user for ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().User().Delete(ctx, &clickhouse.DeleteUserRequest{
				ClusterId: d.Id(),
				UserName:  userName,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while deleting user from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec, changedFields []string) error {

	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig, func() (*operation.Operation, error) {
		return config.sdk.MDB().Clickhouse().User().Update(ctx, &clickhouse.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
			Settings:    user.Settings,
			Quotas:      user.Quotas,
			UpdateMask:  &field_mask.FieldMask{Paths: changedFields},
		})
	})

	if err != nil {
		return fmt.Errorf("error while updating user in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseHosts(ctx context.Context, config *Config, d *schema.ResourceData, spec []*clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddHosts(ctx, &clickhouse.AddClusterHostsRequest{
			ClusterId:  d.Id(),
			HostSpecs:  spec,
			CopySchema: &wrappers.BoolValue{Value: d.Get("copy_schema_on_new_hosts").(bool)},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add hosts to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding hosts to ClickHouse Cluster %q: %s", d.Id(), err)
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
	err = op.WaitInterval(ctx, yandexMDBClickHouseClusterPollInterval)
	if err != nil {
		return fmt.Errorf("error while updating host of ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseHosts(ctx context.Context, config *Config, d *schema.ResourceData, fqdns []string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteHosts(ctx, &clickhouse.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: fqdns,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete hosts from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting hosts from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseShards(ctx context.Context, config *Config, d *schema.ResourceData, hostSpecs []*clickhouse.HostSpec, shardSpecs []*clickhouse.ShardSpec) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Cluster().AddShards(ctx, &clickhouse.AddClusterShardsRequest{
				ClusterId:  d.Id(),
				ShardSpecs: shardSpecs,
				HostSpecs:  hostSpecs,
				CopySchema: &wrappers.BoolValue{Value: d.Get("copy_schema_on_new_hosts").(bool)},
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while adding shards to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string, shardSpec *clickhouse.ShardConfigSpec) error {
	resp, err := config.sdk.MDB().Clickhouse().Cluster().GetShard(context.Background(), &clickhouse.GetClusterShardRequest{
		ClusterId: d.Id(),
		ShardName: shardName,
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to get shard's config, shard name=%s. Error=%s", shardName, err)
	}

	updateRequired := false
	var updatePath []string

	log.Println("[DEBUG] start compute updating fields")
	if resp.Config.Clickhouse.Weight.Value != shardSpec.Clickhouse.Weight.Value {
		log.Printf("[DEBUG] shard=%s has wegith=%d, update to %d\n", shardName, resp.Config.Clickhouse.Weight.Value, shardSpec.Clickhouse.Weight.Value)
		updateRequired = true
		updatePath = append(updatePath, "config_spec.clickhouse.weight")
	}

	if resp.Config.Clickhouse.Resources.GetDiskSize() != shardSpec.Clickhouse.Resources.GetDiskSize() {
		log.Printf("[DEBUG] shard=%s has disk_size=%d, update to %d\n", shardName, resp.Config.Clickhouse.Resources.GetDiskSize(), shardSpec.Clickhouse.Resources.GetDiskSize())
		updateRequired = true
		updatePath = append(updatePath, "config_spec.clickhouse.resources.disk_size")
	}

	if resp.Config.Clickhouse.Resources.GetResourcePresetId() != shardSpec.Clickhouse.Resources.GetResourcePresetId() {
		log.Printf("[DEBUG] shard=%s has resource_preset_id=%s, update to %s\n", shardName, resp.Config.Clickhouse.Resources.GetResourcePresetId(), shardSpec.Clickhouse.Resources.GetResourcePresetId())
		updateRequired = true
		updatePath = append(updatePath, "config_spec.clickhouse.resources.resource_preset_id")
	}

	if resp.Config.Clickhouse.Resources.GetDiskTypeId() != shardSpec.Clickhouse.Resources.GetDiskTypeId() {
		log.Printf("[DEBUG] shard=%s has disk_type_id=%s, update to %s\n", shardName, resp.Config.Clickhouse.Resources.GetDiskTypeId(), shardSpec.Clickhouse.Resources.GetDiskTypeId())
		updateRequired = true
		updatePath = append(updatePath, "config_spec.clickhouse.resources.disk_type_id")
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
		// It can happen on resources update and it easier to ignore this error rather than fix shard state
		// In the end we get shard in desirable state
		if strings.Contains(err.Error(), "no changes detected") {
			log.Printf("[DEBUG] ignored no changes error from API for shard %s of Cluster %q\n", shardName, d.Id())
			return nil
		}
		return fmt.Errorf("error while requesting API to update shard %s to ClickHouse Cluster %q: %s", shardName, d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating shard %s to ClickHouse Cluster %q: %s", shardName, d.Id(), err)
	}

	return nil
}

func deleteClickHouseShards(ctx context.Context, config *Config, d *schema.ResourceData, names []string) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Cluster().DeleteShards(ctx, &clickhouse.DeleteClusterShardsRequest{
				ClusterId:  d.Id(),
				ShardNames: names,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error while deleting shards from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func addClickHouseShards(ctx context.Context, config *Config, d *schema.ResourceData, hostSpecsByShard map[string][]*clickhouse.HostSpec) error {
	if len(hostSpecsByShard) == 0 {
		return nil
	}

	shardConfigSpecsMap, err := expandClickhouseShardSpecs(d)
	if err != nil {
		return err
	}

	var shardSpecs []*clickhouse.ShardSpec
	var hostSpecs []*clickhouse.HostSpec
	for shardName, shardHostSpecs := range hostSpecsByShard {
		hostSpecs = append(hostSpecs, shardHostSpecs...)

		shardSpec := &clickhouse.ShardSpec{
			Name: shardName,
		}
		if shardConfigSpec, ok := shardConfigSpecsMap[shardName]; ok {
			shardSpec.ConfigSpec = shardConfigSpec
		}

		shardSpecs = append(shardSpecs, shardSpec)
	}

	for _, shardSpec := range shardSpecs {
		if _, ok := hostSpecsByShard[shardSpec.Name]; !ok {
			return fmt.Errorf("no hosts defined for shard %s", shardSpec.Name)
		}
	}

	log.Printf("[DEBUG] Shard specs to add: %v\n", shardSpecs)
	log.Printf("[DEBUG] Host specs to add: %v\n", hostSpecs)

	err = createClickHouseShards(ctx, config, d, hostSpecs, shardSpecs)
	if err != nil {
		return err
	}

	return nil
}

func createClickHouseShardGroup(ctx context.Context, config *Config, d *schema.ResourceData, group *clickhouse.ShardGroup) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Cluster().CreateShardGroup(ctx, &clickhouse.CreateClusterShardGroupRequest{
				ClusterId:      d.Id(),
				ShardGroupName: group.Name,
				Description:    group.Description,
				ShardNames:     group.ShardNames,
			})
		},
	)
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
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().Cluster().DeleteShardGroup(ctx, &clickhouse.DeleteClusterShardGroupRequest{
				ClusterId:      d.Id(),
				ShardGroupName: name,
			})
		})
	if err != nil {
		return fmt.Errorf("error while deleting shard group from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseFormatSchema(ctx context.Context, config *Config, d *schema.ResourceData, schema *clickhouse.FormatSchema) error {
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().FormatSchema().Create(ctx, &clickhouse.CreateFormatSchemaRequest{
				ClusterId:        d.Id(),
				FormatSchemaName: schema.Name,
				Type:             schema.Type,
				Uri:              schema.Uri,
			})
		})
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
	err := waitOperationWithRetry(ctx, config, yandexMDBClickhouseRetryOperationConfig,
		func() (*operation.Operation, error) {
			return config.sdk.MDB().Clickhouse().FormatSchema().Delete(ctx, &clickhouse.DeleteFormatSchemaRequest{
				ClusterId:        d.Id(),
				FormatSchemaName: name,
			})
		})
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
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		return config.sdk.MDB().Clickhouse().Cluster().AddZookeeper(ctx, &clickhouse.AddClusterZookeeperRequest{
			ClusterId: d.Id(),
			Resources: resources,
			HostSpecs: specs,
		})
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to create ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

// TODO: deadcode
//func updateClickHouseMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *clickhouse.MaintenanceWindow) error {
//	op, err := config.sdk.WrapOperation(
//		config.sdk.MDB().Clickhouse().Cluster().Update(ctx, &clickhouse.UpdateClusterRequest{
//			ClusterId:         d.Id(),
//			MaintenanceWindow: mw,
//			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
//		}),
//	)
//	if err != nil {
//		return fmt.Errorf("error while requesting API to update maintenance window in ClickHouse Cluster %q: %s", d.Id(), err)
//	}
//	err = op.Wait(ctx)
//	if err != nil {
//		return fmt.Errorf("error while updating maintenance window in ClickHouse Cluster %q: %s", d.Id(), err)
//	}
//	return nil
//}

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

	req := &clickhouse.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Deleting ClickHouse Cluster %q", d.Id())
		return config.sdk.MDB().Clickhouse().Cluster().Delete(ctx, req)
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ClickHouse Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting Clickhouse cluster %q: %s", d.Id(), err)
	}

	_, err = op.Response()
	if err != nil {
		return fmt.Errorf("error while deleting Clickhouse cluster %q: %s", d.Id(), err)
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
	log.Printf("[DEBUG] read data for fill schema: shards=%v\n", shards)

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
