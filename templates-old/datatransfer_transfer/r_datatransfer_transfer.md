---
subcategory: "Data Transfer"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Data Transfer transfer within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a Data Transfer transfer. For more information, see [the official documentation](https://yandex.cloud/docs/data-transfer/).

## Example usage

{{ tffile "examples/datatransfer_transfer/r_datatransfer_transfer_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the transfer.
* `type` - (Required) Type of the transfer. One of "SNAPSHOT_ONLY", "INCREMENT_ONLY", "SNAPSHOT_AND_INCREMENT".
* `source_id` - (Optional) ID of the source endpoint for the transfer.
* `target_id` - (Optional) ID of the target endpoint for the transfer.
* `description` - (Optional) Arbitrary description text for the transfer.
* `folder_id` - (Optional) ID of the folder to create the transfer in. If it is not provided, the default provider folder is used.
* `labels` - (Optional) A set of key/value label pairs to assign to the Data Transfer transfer.
* `runtime` - (Optional) Runtime parameters for the transfer.
* `transformation` - (Optional) Transformation for the transfer.
* `on_create_activate_mode` - (Optional) Activation action on create a new incremental transfer.
It is not part of the transfer parameter and is used only on create.
One of "sync_activate", "async_activate", "dont_activate". The default is "sync_activate".

For the documentation of the runtime and transformation see below.

---

The `runtime` block supports:

* `yc_runtime` - (Optional) YC Runtime parameters for the transfer.

The `yc_runtime` block supports:

* `job_count` - (Optional) Number of workers in parallel replication.
* `upload_shard_params` - (Optional) Parallel snapshot parameters.

The `upload_shard_params` block supports:

* `job_count` - (Optional) Number of workers.
* `process_count` - (Optional) Number of threads.

---

The `transformation` block supports:

* `transformers` - (Optional) A list of transformers. You can specify exactly 1 transformer in each element of list.

---

The `transformers` block supports:

* `mask_field` - (Optional) Mask field transformer allows you to hash data.
* `filter_columns` - (Optional) Set up a list of table columns to transfer.
* `rename_tables` - (Optional) Set rules for renaming tables by specifying the current names of the tables in the source and new names for these tables in the target.
* `replace_primary_key` - (Optional) Override primary keys.
* `convert_to_string` - (Optional) Convert column values to strings.
* `sharder_transformer` - (Optional) Set the number of shards for particular tables and a list of columns whose values will be used for calculating a hash to determine a shard.
* `table_splitter_transformer` - (Optional) Splits the X table into multiple tables (X_1, X_2, ..., X_n) based on data.
* `filter_rows` - (Optional) This filter only applies to transfers with queues (Apache Kafka®) as a data source. When running a transfer, only the strings meeting the specified criteria remain in a changefeed.

---

The `mask_field` block supports:

* `tables` - (Optional) Table filter.
* `columns` - (Optional) List of strings that specify the name of the column for data masking (a regular expression).
* `function` - (Optional) Mask function.

---

The `function` block supports:

* `mask_function_hash` - (Optional) Hash mask function.

---

The `mask_function_hash` block supports:

* `user_defined_salt` - (Optional) This string will be used in the HMAC(sha256, salt) function applied to the column data.

---

The `filter_columns` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `columns` - (Optional) List of the columns to transfer to the target tables using lists of included and excluded columns (see block documentation below).

---

The `rename_tables` block supports:

* `rename_tables` - (Optional) List of renaming rules.

---

The `rename_tables` block supports:

* `original_name` - (Optional) Specify the current names of the table in the source.
* `new_name` - (Optional) Specify the new names for this table in the target.

---

The `replace_primary_key` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `keys` - (Optional) List of columns to be used as primary keys.

---

The `convert_to_string` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `columns` - (Optional) List of the columns to transfer to the target tables using lists of included and excluded columns (see block documentation below).

---

The `sharder_transformer` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `columns` - (Optional) List of the columns to transfer to the target tables using lists of included and excluded columns (see block documentation below).
* `shards_count` - (Optional) Number of shards.

---

The `table_splitter_transformer` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `columns` - (Optional) List of strings that specify the columns in the tables to be partitioned.
* `splitter` - (Optional) Specify the split string to be used for merging components in a new table name.

---

The `filter_rows` block supports:

* `tables` - (Optional) Table filter (see block documentation below).
* `filter` - (Optional) Filtering criterion. This can be comparison operators for numeric, string, and Boolean values, comparison to NULL, and checking whether a substring is part of a string. Details here: https://yandex.cloud/docs/data-transfer/concepts/data-transformation#append-only-sources

---

The `columns` block supports:

* `include_columns` - (Optional) List of columns that will be included to transfer.
* `exclude_columns` - (Optional) List of columns that will be excluded to transfer.

---

The `tables` block supports:

* `include_tables` - (Optional) List of tables that will be included to transfer.
* `exclude_tables` - (Optional) List of tables that will be excluded to transfer.

## Attributes Reference

* `id` - (Computed) Identifier of a new Data Transfer transfer.
* `warning` - (Computed) Error description if transfer has any errors.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/datatransfer_transfer/import.sh" }}
