---
subcategory: "Data Transfer"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Data Transfer endpoint within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a Data Transfer endpoint. For more information, see [the official documentation](https://yandex.cloud/docs/data-transfer/).

## Example usage

{{ tffile "examples/datatransfer_endpoint/r_datatransfer_endpoint_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the endpoint.
* `settings` - (Required) Settings for the endpoint. The structure is documented below.
* `description` - (Optional) Arbitrary description text for the endpoint.
* `folder_id` - (Optional) ID of the folder to create the endpoint in. If it is not provided, the default provider folder is used.
* `labels` - (Optional) A set of key/value label pairs to assign to the Data Transfer endpoint.

The `settings` block supports:

* `clickhouse_source` - (Optional) Settings specific to the ClickHouse source endpoint.
* `clickhouse_target` - (Optional) Settings specific to the ClickHouse target endpoint.
* `kafka_source` - (Optional) Settings specific to the Kafka source endpoint.
* `kafka_target` - (Optional) Settings specific to the Kafka target endpoint.
* `mongo_source` - (Optional) Settings specific to the MongoDB source endpoint.
* `mongo_target` - (Optional) Settings specific to the MongoDB target endpoint.
* `postgres_source` - (Optional) Settings specific to the PostgreSQL source endpoint.
* `postgres_target` - (Optional) Settings specific to the PostgreSQL target endpoint.
* `mysql_source` - (Optional) Settings specific to the MySQL source endpoint.
* `mysql_target` - (Optional) Settings specific to the MySQL target endpoint.
* `ydb_source` - (Optional) Settings specific to the YDB source endpoint.
* `ydb_target` - (Optional) Settings specific to the YDB target endpoint.
* `yds_source` - (Optional) Settings specific to the YDS source endpoint.
* `yds_target` - (Optional) Settings specific to the YDS target endpoint.

For the documentation of the specific endpoint settings see below.

---

The `postgres_source` block supports:

* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `service_schema` - (Optional) Name of the database schema in which auxiliary tables needed for the transfer will be created. Empty `service_schema` implies schema "public".
* `include_tables` - List of tables to transfer, formatted as `schemaname.tablename`. If omitted or an empty list is specified, all tables will be transferred.
* `exclude_tables` - List of tables which will not be transfered, formatted as `schemaname.tablename`.
* `object_transfer_settings` - (Optional) Defines which database schema objects should be transferred, e.g. views, functions, etc.
* `slot_gigabyte_lag_limit` - (Optional) Maximum WAL size held by the replication slot, in gigabytes. Exceeding this limit will result in a replication failure and deletion of the replication slot. Unlimited by default.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `postgres_target` block supports:

* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `connection` block supports exactly one of the following attributes:
* `mdb_cluster_id` - Identifier of the Managed PostgreSQL cluster.
* `on_premise` - Connection settings of the on-premise PostgreSQL server.

The `on_premise` block supports:
* `hosts` - (Required) List of host names of the PostgreSQL server. Exactly one host is expected currently.
* `port` - (Required) Port for the database connection.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `object_transfer_settings` block supports the following attributes:
* `sequence`
* `sequence_owned_by`
* `table`
* `primary_key`
* `fk_constraint`
* `default_values`
* `constraint`
* `index`
* `view`
* `function`
* `trigger`
* `type`
* `rule`
* `collation`
* `policy`
* `cast`
All of the attrubutes are optional and should be either "BEFORE_DATA", "AFTER_DATA" or "NEVER".

---

The `mysql_source` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `include_tables_regex` - (Optional) List of regular expressions of table names which should be transferred. A table name is formatted as schemaname.tablename. For example, a single regular expression may look like `^mydb.employees$`.
* `exclude_tables_regex` - (Optional) Opposite of `include_table_regex`. The tables matching the specified regular expressions will not be transferred.
* `timezone` - (Optional) Timezone to use for parsing timestamps for saving source timezones. Accepts values from IANA timezone database. Default: local timezone.
* `object_transfer_settings` - (Optional) Defines which database schema objects should be transferred, e.g. views, routines, etc.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `mysql_target` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `sql_mode` - (Optional) [sql_mode](https://dev.mysql.com/doc/refman/5.7/en/sql-mode.html) to use when interacting with the server. Defaults to "NO_AUTO_VALUE_ON_ZERO,NO_DIR_IN_CREATE,NO_ENGINE_SUBSTITUTION".
* `skip_constraint_checks` - (Optional) When true, disables foreign key checks. See [foreign_key_checks](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_foreign_key_checks). False by default.
* `timezone` - (Optional) Timezone to use for parsing timestamps for saving source timezones. Accepts values from IANA timezone database. Default: local timezone.
* `service_database` - (Optional) The name of the database where technical tables (`__tm_keeper`, `__tm_gtid_keeper`) will be created. Default is the value of the attribute `database`.
* `cleanup_policy` - (Optional) How to clean tables when activating the transfer. One of "DISABLED", "DROP" or "TRUNCATE".
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `connection` block supports exactly one of the following attributes:
* `mdb_cluster_id` - Identifier of the Managed MySQL cluster.
* `on_premise` - Connection settings of the on-premise MySQL server.

The `on_premise` block supports:
* `hosts` - (Required) List of host names of the MySQL server. Exactly one host is expected currently.
* `port` - (Required) Port for the database connection.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `object_transfer_settings` block supports the following attributes:
* `view`
* `routine`
* `trigger`
All of the attrubutes are optional and should be either "BEFORE_DATA", "AFTER_DATA" or "NEVER".

---

The `mongo_source` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `collections` - (Optional) The list of the MongoDB collections that should be transferred. If omitted, all available collections will be transferred. The structure of the list item is documented below.
* `excluded_collections` - (Optional) The list of the MongoDB collections that should not be transferred.
* `secondary_preferred_mode` - (Optional) whether the secondary server should be preferred to the primary when copying data.

The `mongo_target` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `cleanup_policy` - (Optional) How to clean collections when activating the transfer. One of "DISABLED", "DROP" or "TRUNCATE".
* `database` - (Optional) If not empty, then all the data will be written to the database with the specified name; otherwise the database name is the same as in the source endpoint.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `connection` block supports exactly one of the following attributes:
* `connection_options` (Required) Connection options. The structure is documented below.

The `connection_options` block supports the following attributes:
* `mdb_cluster_id` - (Optional) Identifier of the Managed MongoDB cluster.
* `on_premise` - (Optional) Connection settings of the on-premise MongoDB server.
* `auth_source` - (Required) Name of the database associated with the credentials.
* `user` - (Required) User for database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.

The `on_premise` block supports the following attributes:
* `hosts` - (Required) Host names of the replica set.
* `port` - (Required) TCP Port number.
* `replica_set` - (Optional) Replica set name.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.

The `collections` block supports the following attributes:
* `database_name` - (Required) Database name.
* `collection_name` - (Required) Collection name.

---

The `clickhouse_source` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `exclude_tables` - (Optional) The list of tables that should not be transferred.
* `include_tables` - (Optional) The list of tables that should be transferred. Leave empty if all tables should be transferred.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `clickhouse_target` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `cleanup_policy` - (Optional) How to clean collections when activating the transfer. One of "CLICKHOUSE_CLEANUP_POLICY_DISABLED" or "CLICKHOUSE_CLEANUP_POLICY_DROP".
* `clickhouse_cluster_name` - (Optional) Name of the ClickHouse cluster. For managed ClickHouse clusters defaults to managed cluster ID.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `alt_names` - (Optional) Table renaming rules. The structure is documented below.
* `sharding` - (Optional) Shard selection rules for the data being transferred. The structure is documented below.
* `custom_mapping` (Optional) A custom shard mapping by the value of the specified column. The structure is documented below.
* `round_robin` (Optional) Distribute incoming rows between ClickHouse shards in a round-robin manner. Specify as an empty block to enable.

The `connection` block supports the following attributes:
* `connection_options` (Required) Connection options. The structure is documented below.

The `connection_options` block supports the following attributes:
* `mdb_cluster_id` - (Optional) Identifier of the Managed ClickHouse cluster.
* `on_premise` - (Optional) Connection settings of the on-premise ClickHouse server.
* `database` - (Required) Database name.
* `user` - (Required) User for database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.

The `on_premise` block supports the following attributes:
* `http_port` - (Required) TCP port number for the HTTP interface of the ClickHouse server.
* `native_port` - (Required) TCP port number for the native interface of the ClickHouse server.
* `shards` - (Required) The list of ClickHouse shards. The structure is documented below.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.

The `shards` block supports the following attributes:
* `name` - (Required) Arbitrary shard name. This name may be used in `sharding` block to specify custom sharding rules.
* `hosts` - (Required) List of ClickHouse server host names.

The `sharding` block supports exactly one of the following attributes:
* `column_value_hash` - Shard data by the hash value of the specified column. The structure is documented below.
* `transfer_id` - Shard data by ID of the transfer.

The `column_value_hash` block supports:
* `column_name` - The name of the column to calculate hash from.

The `custom_mapping` block supports:
* `column_name` - (Required) The name of the column to inspect when deciding the shard to chose for an incoming row.
* `mapping` - (Required) The mapping of the specified column values to the shard names. The structure is documented below.

The `mapping` block supports:
* `column_value` - (Required) The value of the column. Currently only the string columns are supported. The structure is documented below.
* `shard_name` - (Required) The name of the shard into which all the rows with the specified `column_value` will be written.

The `column_value` block supports:
* `string_value` - (Optional) The string value of the column.

---

The `kafka_source` block supports:
* `connection` - (Required) Connection settings.
* `auth` - (Required) Authentication data.
* `topic_name` - (Optional) Deprecated. Please use `topic_names` instead.
* `topic_names` - (Optional) The list of full source topic names.
* `transformer` - (Optional) Transform data with a custom Cloud Function.
* `parser` - (Optional) Data parsing parameters. If not set, the source messages are read in raw.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `kafka_target` block supports:
* `connection` - (Required) Connection settings.
* `auth` - (Required) Authentication data.
* `topic_settings` - (Required) Target topic settings.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.
* `serializer` - (Required) Data serialization settings.

The `topic_settings` block supports exactly one of the following attributes:
* `topic_prefix` - Topic name prefix. Messages will be sent to topic with name <topic_prefix>.<schema>.<table_name>.
* `topic` - All messages will be sent to one topic. The structure is documented below.

The `topic` block supports:
* `topic_name` - Full topic name
* `save_tx_order` - Not to split events queue into separate per-table queues.

The `connection` block supports exactly one of the following attributes:
* `cluster_id` - Identifier of the Managed Kafka cluster.
* `on_premise` - Connection settings of the on-premise Kafka server.

The `on_premise` block supports:
* `broker_urls` (Required) List of Kafka broker URLs.
* `subnet_id` (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `tls_mode`  (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.

The `auth` block supports exactly one of the following attributes:
* `no_auth` - Connection without authentication data.
* `sasl` - Authentication using sasl.

---

The `parser` block supports exactly one of the following attributes:
* `audit_trails_v1_parser` - Parse Audit Trails data. Empty struct.
* `cloud_logging_parser` - Parse Cloud Logging data. Empty struct.
* `json_parser` - Parse data in json format.
* `tskv_parser` - Parse data if tskv format.

The `json_parser` and `tskv_parser` blocks supports:
* `data_schema` - (Required) Data parsing scheme.The structure is documented below.
* `add_rest_column` - Add fields, that are not in the schema, into the _rest column.
* `null_keys_allowed` - Allow null keys. If `false` - null keys will be putted to unparsed data
* `unescape_string_values` - Allow unescape string values.

The `data_schema` block supports exactly one of the following attributes:
* `fields`  - Description of the data schema in the array of `fields` structure (documented below).
* `json_fields`- Description of the data schema as JSON specification.

The `fields` block supports:
* `name` - (Required) Field name.
* `type` - (Required) Field type, one of: `INT64`, `INT32`, `INT16`, `INT8`, `UINT64`, `UINT32`, `UINT16`, `UINT8`, `DOUBLE`, `BOOLEAN`, `STRING`, `UTF8`, `ANY`, `DATETIME`.
* `path` - Path to the field.
* `required` - Mark field as required.
* `key` -Mark field as Primary Key.

---

The `tls_mode` block supports exactly one of the following attributes:
* `disabled` - Empty block designating that the connection is not secured, i.e. plaintext connection.
* `enabled` - If this attribute is not an empty block, then TLS is used for the server connection. The structure is documented below.

The `enabled` block supports:
* `ca_certificate` - (Optional) X.509 certificate of the certificate authority which issued the server's certificate, in PEM format. If empty, the server's certificate must be signed by a well-known CA.

The `alt_names` block supports:
* `from_name`
* `to_name`

---

The `serializer` block supports exactly one of the following attributes:
* `serializer_auto` - Empty block. Select data serialization format automatically.
* `serializer_json` - Empty block. Serialize data in json format.
* `serializer_debezium` - Serialize data in json format. The structure is documented below.

The `serializer_debezium` block supports:
* `serializer_parameters` - A list of debezium parameters set by the structure of the `key` and `value` string fields.
## Attributes Reference

* `id` - (Computed) Identifier of a new Data Transfer endpoint.
* `created_at` - (Computed) Data Transfer endpoint creation timestamp.
* `author` - (Computed) Identifier of the IAM user account of the user who created the endpoint.

---

The `ydb_source` block supports:
* `database` -- (Required) Database path in YDB where tables are stored. Example: "/ru/transfer_manager/prod/data-transfer-yt".
* `instance` -- (Optional) Instance of YDB. Example: "my-cute-ydb.yandex.cloud:2135".
* `service_account_id` -- (Optional) Service account ID for interaction with database.
* `paths` -- (Optional) A list of paths which should be uploaded. When not specified, all available tables are uploaded.
* `subnet_id` -- (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `security_groups` -- (Optional) List of security groups that the transfer associated with this endpoint should use.
* `sa_key_content` -- (Optional, Sensitive) Authentication key.
* `changefeed_custom_name` -- (Optional) Custom name for changefeed.

The `ydb_target` block supports:
* `database` -- (Required) Database path in YDB where tables are stored. Example: "/ru/transfer_manager/prod/data-transfer-yt".
* `instance` -- (Optional) Instance of YDB. Example: "my-cute-ydb.yandex.cloud:2135".
* `service_account_id` -- (Optional) Service account ID for interaction with database.
* `path` -- (Optional) A path where resulting tables are stored.
* `subnet_id` -- (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `security_groups` -- (Optional) List of security groups that the transfer associated with this endpoint should use.
* `sa_key_content` -- (Optional, Sensitive) Authentication key.
* `cleanup_policy` -- (Optional) How to clean collections when activating the transfer. One of "YDB_CLEANUP_POLICY_DISABLED" or "YDB_CLEANUP_POLICY_DROP".
* `is_table_column_oriented` -- (Optional) Whether a column-oriented (i.e. OLAP) tables should be created. Default is `false` (create row-oriented OLTP tables).
* `default_compression` -- (Optional) Compression that will be used for default columns family on YDB table creation One of "YDB_DEFAULT_COMPRESSION_UNSPECIFIED", "YDB_DEFAULT_COMPRESSION_DISABLED", "YDB_DEFAULT_COMPRESSION_LZ4".

* `connection` - (Required) Connection settings.
* `auth` - (Required) Authentication data.
* `topic_settings` - (Required) Target topic settings.
* `security_groups` - (Optional) List of security groups that the transfer associated with this endpoint should use.

The `topic_settings` block supports exactly one of the following attributes:
* `topic_prefix` - Topic name prefix. Messages will be sent to topic with name <topic_prefix>.<schema>.<table_name>.
* `topic` - All messages will be sent to one topic. The structure is documented below.

The `topic` block supports:
* `topic_name` - Full topic name
* `save_tx_order` - Not to split events queue into separate per-table queues.

The `connection` block supports exactly one of the following attributes:
* `cluster_id` - Identifier of the Managed Kafka cluster.
* `on_premise` - Connection settings of the on-premise Kafka server.

The `on_premise` block supports:
* `broker_urls` (Required) List of Kafka broker URLs.
* `subnet_id` (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `tls_mode`  (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.

The `auth` block supports exactly one of the following attributes:
* `no_auth` - Connection without authentication data.
* `sasl` - Authentication using sasl.

---

The `yds_source` block supports:
* `database` -- (Required) Database.
* `endpoint` -- (Optional) YDS Endpoint.
* `service_account_id` -- (Required) Service account ID for interaction with database.
* `subnet_id` -- (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `security_groups` -- (Optional) List of security groups that the transfer associated with this endpoint should use.
* `stream` -- (Optional) Stream.
* `consumer` -- (Optional) Consumer.
* `parser` -- (Optional) Data parsing rules.
* `supported_codecs` -- (Optional) List of supported compression codec.
* `allow_ttl_rewind` -- (Optional) Should continue working, if consumer read lag exceed TTL of topic.

The `yds_target` block supports:
* `database` -- (Required) Database.
* `endpoint` -- (Optional) YDS Endpoint.
* `service_account_id` -- (Required) Service account ID for interaction with database.
* `subnet_id` -- (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.
* `security_groups` -- (Optional) List of security groups that the transfer associated with this endpoint should use.
* `stream` -- (Optional) Stream.
* `serializer` -- (Optional) Data serialization format.
* `save_tx_order` -- (Optional) Save transaction order

---

The `parser` block supports exactly one of the following attributes:
* `audit_trails_v1_parser` - Parse Audit Trails data. Empty struct.
* `cloud_logging_parser` - Parse Cloud Logging data. Empty struct.
* `json_parser` - Parse data in json format.
* `tskv_parser` - Parse data if tskv format.

The `json_parser` and `tskv_parser` blocks supports:
* `data_schema` - (Required) Data parsing scheme.The structure is documented below.
* `add_rest_column` - Add fields, that are not in the schema, into the _rest column.
* `null_keys_allowed` - Allow null keys. If `false` - null keys will be putted to unparsed data
* `unescape_string_values` - Allow unescape string values.

The `data_schema` block supports exactly one of the following attributes:
* `fields`  - Description of the data schema in the array of `fields` structure (documented below).
* `json_fields`- Description of the data schema as JSON specification.

The `fields` block supports:
* `name` - (Required) Field name.
* `type` - (Required) Field type, one of: `INT64`, `INT32`, `INT16`, `INT8`, `UINT64`, `UINT32`, `UINT16`, `UINT8`, `DOUBLE`, `BOOLEAN`, `STRING`, `UTF8`, `ANY`, `DATETIME`.
* `path` - Path to the field.
* `required` - Mark field as required.
* `key` -Mark field as Primary Key.

---

The `serializer` block supports exactly one of the following attributes:
* `serializer_auto` - Empty block. Select data serialization format automatically.
* `serializer_json` - Empty block. Serialize data in json format.
* `serializer_debezium` - Serialize data in json format. The structure is documented below.

The `serializer_debezium` block supports:
* `serializer_parameters` - A list of debezium parameters set by the structure of the `key` and `value` string fields.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/datatransfer_endpoint/import.sh" }}
