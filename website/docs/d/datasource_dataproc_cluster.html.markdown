---
layout: "yandex"
page_title: "Yandex: yandex_dataproc_cluster"
sidebar_current: "docs-yandex-datasource-dataproc-cluster"
description: |-
  Get information about a Yandex Data Proc cluster.
---

# yandex\_dataproc\_cluster

Get information about a Yandex Data Proc cluster. For more information, see [the official documentation](https://cloud.yandex.com/docs/data-proc/).

## Example Usage

```hcl
data "yandex_dataproc_cluster" "foo" {
  name = "test"
}

output "service_account_id" {
  value = "${data.yandex_dataproc_cluster.foo.service_account_id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) The ID of the Data Proc cluster.
* `name` - (Optional) The name of the Data Proc cluster.

~> **NOTE:** Either `cluster_id` or `name` should be specified.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `bucket` - Name of the Object Storage bucket used for Data Proc jobs.
* `cluster_config` - Configuration and resources of the cluster. The structure is documented below.
* `created_at` - The Data Proc cluster creation timestamp.
* `description` - Description of the Data Proc cluster.
* `id` - Id of the Data Proc cluster.
* `labels` - A set of key/value label pairs assigned to the Data Proc cluster.
* `service_account_id` - Service account used by the Data Proc agent to access resources of Yandex.Cloud.
* `ui_proxy` - Whether UI proxy feature is enabled.
* `zone_id` - ID of the availability zone where the cluster resides.
* `host_group_ids` - A list of IDs of the host groups hosting VMs of the cluster.

---

The `cluster_config` block supports:

* `version_id` - Version of Data Proc image.
* `hadoop` - Data Proc specific options. The structure is documented below.
* `subcluster_spec` - Configuration of the Data Proc subcluster. The structure is documented below.

---

The `hadoop` block supports:

* `services` - List of services launched on Data Proc cluster.
* `properties` - A set of key/value pairs used to configure cluster services.
* `ssh_public_keys` - List of SSH public keys distributed to the hosts of the cluster.

---

The `subcluster_spec` block supports:

* `id` - ID of the Data Proc subcluster.
* `name` - Name of the Data Proc subcluster.
* `role` - Role of the subcluster in the Data Proc cluster.
* `resources` - Resources allocated to each host of the Data Proc subcluster. The structure is documented below.
* `subnet_id` - The ID of the subnet, to which hosts of the subcluster belong.
* `hosts_count` - Number of hosts within Data Proc subcluster.
* `assign_public_ip` - The hosts of the subclusters have public IP addresses.
* `autoscaling_config` - Optional autoscaling configuration for compute subclusters.

---

The `resources` block supports:

* `resource_preset_id` - The ID of the preset for computational resources available to a host. All available presets are listed in the [documentation](https://cloud.yandex.com/docs/data-proc/concepts/instance-types).
* `disk_size` - Volume of the storage available to a host, in gigabytes.
* `disk_type_id` - Type of the storage of a host.

---

The `autoscaling_config` block supports:

* `max_hosts_count` - Maximum number of nodes in autoscaling subclusters.
* `preemptible` - Bool flag -- whether to use preemptible compute instances. Preemptible instances are stopped at least once every 24 hours, and can be stopped at any time if their resources are needed by Compute. For more information, see [Preemptible Virtual Machines](https://cloud.yandex.com/docs/compute/concepts/preemptible-vm).
* `warmup_duration` - The warmup time of the instance in seconds. During this time, traffic is sent to the instance, but instance metrics are not collected.
* `stabilization_duration` - Minimum amount of time in seconds allotted for monitoring before Instance Groups can reduce the number of instances in the group. During this time, the group size doesn't decrease, even if the new metric values indicate that it should.
* `measurement_duration` - Time in seconds allotted for averaging metrics.
* `cpu_utilization_target` - Defines an autoscaling rule based on the average CPU utilization of the instance group. If not set default autoscaling metric will be used.
* `decommission_timeout` - Timeout to gracefully decommission nodes during downscaling. In seconds.
