---
subcategory: "Managed Service for YDB"
page_title: "Yandex: yandex_ydb_table_changefeed"
description: |-
  Manages Yandex Database dedicated cluster.
---


# yandex_ydb_table_changefeed




Yandex Database [table changefeed](https://ydb.tech/en/docs/concepts/cdc), or Change Data Capture (CDC) resource, keeps you informed about changes in a given table. When you add, update, or delete a table row, the CDC mechanism generates a change record where it specifies the primary key of the row and writes it to the topic partition corresponding to this key. A [topic](https://ydb.tech/en/docs/concepts/topic) is an entity for storing unstructured messages and delivering them to multiple subscribers. Basically, a topic is a named set of messages.

## Example Usage

```tf
resource "yandex_ydb_table_changefeed" "ydb_changefeed" {
  table_id = yandex_ydb_table.test_table_2.id
  name = "changefeed"
  mode = "NEW_IMAGE"
  format = "JSON"

  consumer {
    name = "test_consumer"
  }
}
```

We used the following parameters to create a table changefeed:
* `table_id`: ID of the table for which we create the changefeed.
* `name`: Changefeed name.
* `mode`: Changefeed operating mode. The available changefeed operating modes are presented in the [documentation](https://ydb.tech/en/docs/yql/reference/syntax/alter_table#changefeed-options).
* `format`: Changefeed format. Only JSON format is available.

This table describes all the `"yandex_ydb_table_changefeed"` resource parameters:

* `table_path` - (Required) Table path

* `connection_string` - (Required) Connection string, conflicts with `table_id`

* `database_id` - (Required) Database ID, conflicts with `table_path` and `connection_string`

* `table_id` - (Required) Terraform ID of the table

* `name` - (Required) Changefeed name

* `mode` - (Required) [Changefeed mode](https://ydb.tech/en/docs/yql/reference/syntax/alter_table#changefeed-options)

* `format` - (Required) Changefeed format

* `virtual_timestamps` - (Optional) Use [virtual timestamps](https://ydb.tech/en/docs/concepts/cdc#virtual-timestamps)

* `retention_period` - (Optional) Time of data retention in the topic, [ISO 8601](https://ru.wikipedia.org/wiki/ISO_8601) format

* `consumer` - (Optional) Changefeed [consumers](https://ydb.tech/en/docs/concepts/topic#consumer) - named entities for reading data from the topic.

When initializing the `"yandex_ydb_table_changefeed"` resource, you can specify a single connection parameter: `connection_string`, `table_path`, or `table_id`. If you specify multiple connection parameters, they will come into conflict. For this reason, specify a single connection parameter. For example, `table_id` with a relative link in the format: `<resource>.<ID>.<parameter>`: `yandex_ydb_table.test_table_2.id`.

The `consumer` section supports:

* `name` - (Required) Consumer name. It is used in the SDK or CLI to [read data](https://ydb.tech/en/docs/best_practices/cdc#read) from the topic.

* `supported_codecs` - (Optional) Supported data encodings

* `starting_message_timestamp_ms` - (Optional) Timestamp in the UNIX timestamp format, from which the consumer will start reading data
