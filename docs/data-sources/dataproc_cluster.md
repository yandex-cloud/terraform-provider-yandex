---
subcategory: "Data Processing"
page_title: "Yandex: yandex_dataproc_cluster"
description: |-
  Get information about a Yandex Data Processing cluster
---

# yandex_dataproc_cluster (Data Source)



## Example usage

```terraform
//
// Get information about existing Data Processing Cluster.
//
data "yandex_dataproc_cluster" "foo" {
  name = "test"
}

output "service_account_id" {
  value = data.yandex_dataproc_cluster.foo.service_account_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `cluster_id` (String) The ID of the Yandex Data Processing cluster.
- `name` (String) The resource name.

### Read-Only

- `autoscaling_service_account_id` (String) Service account to be used for managing hosts in an autoscaled subcluster.
- `bucket` (String) Name of the Object Storage bucket to use for Yandex Data Processing jobs. Yandex Data Processing Agent saves output of job driver's process to specified bucket. In order for this to work service account (specified by the `service_account_id` argument) should be given permission to create objects within this bucket.
- `cluster_config` (List of Object) Configuration and resources for hosts that should be created with the cluster. (see [below for nested schema](#nestedatt--cluster_config))
- `created_at` (String) The creation timestamp of the resource.
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `description` (String) The resource description.
- `environment` (String) Deployment environment of the cluster. Can be either `PRESTABLE` or `PRODUCTION`. The default is `PRESTABLE`.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `host_group_ids` (Set of String) A list of host group IDs to place VMs of the cluster on.
- `id` (String) The ID of this resource.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `log_group_id` (String) ID of the cloud logging group for cluster logs.
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `service_account_id` (String) Service account to be used by the Yandex Data Processing agent to access resources of Yandex Cloud. Selected service account should have `mdb.dataproc.agent` role on the folder where the Yandex Data Processing cluster will be located.
- `ui_proxy` (Boolean) Whether to enable UI Proxy feature.
- `zone_id` (String) The [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.

<a id="nestedatt--cluster_config"></a>
### Nested Schema for `cluster_config`

Read-Only:

- `hadoop` (Block List, Max: 1) Yandex Data Processing specific options. (see [below for nested schema](#nestedobjatt--cluster_config--hadoop))

- `subcluster_spec` (Block List, Min: 1) Configuration of the Yandex Data Processing subcluster. (see [below for nested schema](#nestedobjatt--cluster_config--subcluster_spec))

- `version_id` (String) Version of Yandex Data Processing image.


<a id="nestedobjatt--cluster_config--hadoop"></a>
### Nested Schema for `cluster_config.hadoop`

Read-Only:

- `initialization_action` (Block List) List of initialization scripts. (see [below for nested schema](#nestedobjatt--cluster_config--hadoop--initialization_action))

- `oslogin` (Boolean) Whether to enable authorization via OS Login.

- `properties` (Map of String) A set of key/value pairs that are used to configure cluster services.

- `services` (Set of String) List of services to run on Yandex Data Processing cluster.

- `ssh_public_keys` (Set of String) List of SSH public keys to put to the hosts of the cluster. For information on how to connect to the cluster, see [the official documentation](https://yandex.cloud/docs/data-proc/operations/connect).


<a id="nestedobjatt--cluster_config--hadoop--initialization_action"></a>
### Nested Schema for `cluster_config.hadoop.initialization_action`

Read-Only:

- `args` (List of String) List of arguments of the initialization script.

- `timeout` (String) Script execution timeout, in seconds.

- `uri` (String) Script URI.




<a id="nestedobjatt--cluster_config--subcluster_spec"></a>
### Nested Schema for `cluster_config.subcluster_spec`

Read-Only:

- `assign_public_ip` (Boolean) If `true` then assign public IP addresses to the hosts of the subclusters.

- `autoscaling_config` (Block List, Max: 1) Autoscaling configuration for compute subclusters. (see [below for nested schema](#nestedobjatt--cluster_config--subcluster_spec--autoscaling_config))

- `hosts_count` (Number) Number of hosts within Yandex Data Processing subcluster.

- `id` (String) ID of the subcluster.

- `name` (String) Name of the Yandex Data Processing subcluster.

- `resources` (Block List, Min: 1, Max: 1) Resources allocated to each host of the Yandex Data Processing subcluster. (see [below for nested schema](#nestedobjatt--cluster_config--subcluster_spec--resources))

- `role` (String) Role of the subcluster in the Yandex Data Processing cluster.

- `subnet_id` (String) The ID of the subnet, to which hosts of the subcluster belong. Subnets of all the subclusters must belong to the same VPC network.


<a id="nestedobjatt--cluster_config--subcluster_spec--autoscaling_config"></a>
### Nested Schema for `cluster_config.subcluster_spec.autoscaling_config`

Read-Only:

- `cpu_utilization_target` (String) Defines an autoscaling rule based on the average CPU utilization of the instance group. If not set default autoscaling metric will be used.

- `decommission_timeout` (String) Timeout to gracefully decommission nodes during downscaling. In seconds.

- `max_hosts_count` (Number) Maximum number of nodes in autoscaling subclusters.

- `measurement_duration` (String) Time in seconds allotted for averaging metrics.

- `preemptible` (Boolean) Use preemptible compute instances. Preemptible instances are stopped at least once every 24 hours, and can be stopped at any time if their resources are needed by Compute. For more information, see [Preemptible Virtual Machines](https://yandex.cloud/docs/compute/concepts/preemptible-vm).

- `stabilization_duration` (String) Minimum amount of time in seconds allotted for monitoring before Instance Groups can reduce the number of instances in the group. During this time, the group size doesn't decrease, even if the new metric values indicate that it should.

- `warmup_duration` (String) The warmup time of the instance in seconds. During this time, traffic is sent to the instance, but instance metrics are not collected.



<a id="nestedobjatt--cluster_config--subcluster_spec--resources"></a>
### Nested Schema for `cluster_config.subcluster_spec.resources`

Read-Only:

- `disk_size` (Number) Volume of the storage available to a host, in gigabytes.

- `disk_type_id` (String) Type of the storage of a host. One of `network-hdd` (default) or `network-ssd`.

- `resource_preset_id` (String) The ID of the preset for computational resources available to a host. All available presets are listed in the [documentation](https://yandex.cloud/docs/data-proc/concepts/instance-types).

