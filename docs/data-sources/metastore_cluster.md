---
subcategory: "Managed Service for Hive Metastore"
page_title: "Yandex: yandex_metastore_cluster"
description: |-
  Get information about Hive Metastore cluster.
---

# yandex_metastore_cluster (Data Source)

Managed Metastore cluster.

## Example usage

```terraform
//
// Get information about Hive Metastore cluster by name
//
data "yandex_metastore_cluster" "metastore_cluster_by_name" {
  name = "metastore-created-with-terraform"
}

//
// Get information about Hive Metastore cluster by id
//
data "yandex_metastore_cluster" "metastore_cluster_by_id" {
  id = "<metastore-cluster-id>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) The resource identifier.
- `name` (String) The resource name.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `cluster_config` (Attributes) Hive Metastore cluster configuration. (see [below for nested schema](#nestedatt--cluster_config))
- `created_at` (String) The creation timestamp of the resource.
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion. By default is set to `false`.
- `description` (String) The resource description.
- `endpoint_ip` (String) IP address of Metastore server balancer endpoint.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `health` (String) Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `logging` (Attributes) Cloud Logging configuration. (see [below for nested schema](#nestedatt--logging))
- `maintenance_window` (Attributes) Configuration of window for maintenance operations. (see [below for nested schema](#nestedatt--maintenance_window))
- `network_id` (String) VPC network identifier which resource is attached.
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `service_account_id` (String) [Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) with role `managed-metastore.integrationProvider`. For more information, see [documentation](https://yandex.cloud/docs/metadata-hub/concepts/metastore-impersonation).
- `status` (String) Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
- `subnet_ids` (Set of String) The list of VPC subnets identifiers which resource is attached.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.


<a id="nestedatt--cluster_config"></a>
### Nested Schema for `cluster_config`

Read-Only:

- `resource_preset_id` (String) The identifier of the preset for computational resources available to an instance (CPU, memory etc.).


<a id="nestedatt--logging"></a>
### Nested Schema for `logging`

Read-Only:

- `enabled` (Boolean) Enables delivery of logs generated by Metastore to [Cloud Logging](https://yandex.cloud/docs/logging/).
- `folder_id` (String) Logs will be written to **default log group** of specified folder. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.
- `log_group_id` (String) Logs will be written to the **specified log group**. Exactly one of the attributes `folder_id` or `log_group_id` should be specified.
- `min_level` (String) Minimum level of messages that will be sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. If not set then server default is applied (currently `INFO`).


<a id="nestedatt--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Read-Only:

- `day` (String) Day of week for maintenance window. One of `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.
- `hour` (Number) Hour of day in UTC time zone (1-24) for maintenance window.
- `type` (String) Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. If `WEEKLY`, day and hour must be specified.
