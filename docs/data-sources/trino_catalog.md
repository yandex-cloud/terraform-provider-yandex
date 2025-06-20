---
subcategory: "Managed Service for Trino"
page_title: "Yandex: yandex_trino_catalog"
description: |-
  Get information about Trino catalog.
---

# yandex_trino_catalog (Data Source)

Catalog for Managed Trino cluster.

## Example usage

```terraform
//
// Get information about Trino catalog by name
//
data "yandex_trino_catalog" "trino_catalog_by_name" {
  cluster_id = yandex_trino_cluster.trino.id
  name       = "catalog"
}

//
// Get information about Trino catalog by id
//
data "yandex_trino_catalog" "trino_catalog_by_id" {
  cluster_id = yandex_trino_cluster.trino.id
  id         = "<tirno-catalog-id>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_id` (String) ID of the Trino cluster.

### Optional

- `id` (String) The resource identifier.
- `name` (String) The resource name.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `clickhouse` (Attributes) Configuration for Clickhouse connector. (see [below for nested schema](#nestedatt--clickhouse))
- `delta_lake` (Attributes) Configuration for Delta Lake connector. (see [below for nested schema](#nestedatt--delta_lake))
- `description` (String) The resource description.
- `hive` (Attributes) Configuration for Hive connector. (see [below for nested schema](#nestedatt--hive))
- `iceberg` (Attributes) Configuration for Iceberg connector. (see [below for nested schema](#nestedatt--iceberg))
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `oracle` (Attributes) Configuration for Oracle connector. (see [below for nested schema](#nestedatt--oracle))
- `postgresql` (Attributes) Configuration for Postgresql connector. (see [below for nested schema](#nestedatt--postgresql))
- `sqlserver` (Attributes) Configuration for SQLServer connector. (see [below for nested schema](#nestedatt--sqlserver))
- `tpcds` (Attributes) Configuration for TPCDS connector. (see [below for nested schema](#nestedatt--tpcds))
- `tpch` (Attributes) Configuration for TPCH connector. (see [below for nested schema](#nestedatt--tpch))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.


<a id="nestedatt--clickhouse"></a>
### Nested Schema for `clickhouse`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `connection_manager` (Attributes) Configuration for connection manager connection. (see [below for nested schema](#nestedatt--clickhouse--connection_manager))
- `on_premise` (Attributes) Configuration for on-premise connection. (see [below for nested schema](#nestedatt--clickhouse--on_premise))

<a id="nestedatt--clickhouse--connection_manager"></a>
### Nested Schema for `clickhouse.connection_manager`

Read-Only:

- `connection_id` (String) Connection ID.
- `connection_properties` (Map of String) Additional connection properties.
- `database` (String) Database.


<a id="nestedatt--clickhouse--on_premise"></a>
### Nested Schema for `clickhouse.on_premise`

Read-Only:

- `connection_url` (String) Connection URL.
- `password` (String, Sensitive) Password of the user.
- `user_name` (String) Name of the user.



<a id="nestedatt--delta_lake"></a>
### Nested Schema for `delta_lake`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `file_system` (Attributes) File system configuration. (see [below for nested schema](#nestedatt--delta_lake--file_system))
- `metastore` (Attributes) Metastore configuration. (see [below for nested schema](#nestedatt--delta_lake--metastore))

<a id="nestedatt--delta_lake--file_system"></a>
### Nested Schema for `delta_lake.file_system`

Read-Only:

- `external_s3` (Attributes) Describes External S3 compatible file system. (see [below for nested schema](#nestedatt--delta_lake--file_system--external_s3))
- `s3` (Attributes) Describes YandexCloud native S3 file system. (see [below for nested schema](#nestedatt--delta_lake--file_system--s3))

<a id="nestedatt--delta_lake--file_system--external_s3"></a>
### Nested Schema for `delta_lake.file_system.external_s3`

Read-Only:

- `aws_access_key` (String, Sensitive) AWS access key ID for S3 authentication.
- `aws_endpoint` (String) AWS S3 compatible endpoint URL.
- `aws_region` (String) AWS region for S3 storage.
- `aws_secret_key` (String, Sensitive) AWS secret access key for S3 authentication.


<a id="nestedatt--delta_lake--file_system--s3"></a>
### Nested Schema for `delta_lake.file_system.s3`



<a id="nestedatt--delta_lake--metastore"></a>
### Nested Schema for `delta_lake.metastore`

Read-Only:

- `uri` (String) The resource description.



<a id="nestedatt--hive"></a>
### Nested Schema for `hive`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `file_system` (Attributes) File system configuration. (see [below for nested schema](#nestedatt--hive--file_system))
- `metastore` (Attributes) Metastore configuration. (see [below for nested schema](#nestedatt--hive--metastore))

<a id="nestedatt--hive--file_system"></a>
### Nested Schema for `hive.file_system`

Read-Only:

- `external_s3` (Attributes) Describes External S3 compatible file system. (see [below for nested schema](#nestedatt--hive--file_system--external_s3))
- `s3` (Attributes) Describes YandexCloud native S3 file system. (see [below for nested schema](#nestedatt--hive--file_system--s3))

<a id="nestedatt--hive--file_system--external_s3"></a>
### Nested Schema for `hive.file_system.external_s3`

Read-Only:

- `aws_access_key` (String, Sensitive) AWS access key ID for S3 authentication.
- `aws_endpoint` (String) AWS S3 compatible endpoint URL.
- `aws_region` (String) AWS region for S3 storage.
- `aws_secret_key` (String, Sensitive) AWS secret access key for S3 authentication.


<a id="nestedatt--hive--file_system--s3"></a>
### Nested Schema for `hive.file_system.s3`



<a id="nestedatt--hive--metastore"></a>
### Nested Schema for `hive.metastore`

Read-Only:

- `uri` (String) The resource description.



<a id="nestedatt--iceberg"></a>
### Nested Schema for `iceberg`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `file_system` (Attributes) File system configuration. (see [below for nested schema](#nestedatt--iceberg--file_system))
- `metastore` (Attributes) Metastore configuration. (see [below for nested schema](#nestedatt--iceberg--metastore))

<a id="nestedatt--iceberg--file_system"></a>
### Nested Schema for `iceberg.file_system`

Read-Only:

- `external_s3` (Attributes) Describes External S3 compatible file system. (see [below for nested schema](#nestedatt--iceberg--file_system--external_s3))
- `s3` (Attributes) Describes YandexCloud native S3 file system. (see [below for nested schema](#nestedatt--iceberg--file_system--s3))

<a id="nestedatt--iceberg--file_system--external_s3"></a>
### Nested Schema for `iceberg.file_system.external_s3`

Read-Only:

- `aws_access_key` (String, Sensitive) AWS access key ID for S3 authentication.
- `aws_endpoint` (String) AWS S3 compatible endpoint URL.
- `aws_region` (String) AWS region for S3 storage.
- `aws_secret_key` (String, Sensitive) AWS secret access key for S3 authentication.


<a id="nestedatt--iceberg--file_system--s3"></a>
### Nested Schema for `iceberg.file_system.s3`



<a id="nestedatt--iceberg--metastore"></a>
### Nested Schema for `iceberg.metastore`

Read-Only:

- `uri` (String) The resource description.



<a id="nestedatt--oracle"></a>
### Nested Schema for `oracle`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `on_premise` (Attributes) Configuration for on-premise connection. (see [below for nested schema](#nestedatt--oracle--on_premise))

<a id="nestedatt--oracle--on_premise"></a>
### Nested Schema for `oracle.on_premise`

Read-Only:

- `connection_url` (String) Connection URL.
- `password` (String, Sensitive) Password of the user.
- `user_name` (String) Name of the user.



<a id="nestedatt--postgresql"></a>
### Nested Schema for `postgresql`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `connection_manager` (Attributes) Configuration for connection manager connection. (see [below for nested schema](#nestedatt--postgresql--connection_manager))
- `on_premise` (Attributes) Configuration for on-premise connection. (see [below for nested schema](#nestedatt--postgresql--on_premise))

<a id="nestedatt--postgresql--connection_manager"></a>
### Nested Schema for `postgresql.connection_manager`

Read-Only:

- `connection_id` (String) Connection ID.
- `connection_properties` (Map of String) Additional connection properties.
- `database` (String) Database.


<a id="nestedatt--postgresql--on_premise"></a>
### Nested Schema for `postgresql.on_premise`

Read-Only:

- `connection_url` (String) Connection URL.
- `password` (String, Sensitive) Password of the user.
- `user_name` (String) Name of the user.



<a id="nestedatt--sqlserver"></a>
### Nested Schema for `sqlserver`

Read-Only:

- `additional_properties` (Map of String) Additional properties.
- `on_premise` (Attributes) Configuration for on-premise connection. (see [below for nested schema](#nestedatt--sqlserver--on_premise))

<a id="nestedatt--sqlserver--on_premise"></a>
### Nested Schema for `sqlserver.on_premise`

Read-Only:

- `connection_url` (String) Connection URL.
- `password` (String, Sensitive) Password of the user.
- `user_name` (String) Name of the user.



<a id="nestedatt--tpcds"></a>
### Nested Schema for `tpcds`

Read-Only:

- `additional_properties` (Map of String) Additional properties.


<a id="nestedatt--tpch"></a>
### Nested Schema for `tpch`

Read-Only:

- `additional_properties` (Map of String) Additional properties.

