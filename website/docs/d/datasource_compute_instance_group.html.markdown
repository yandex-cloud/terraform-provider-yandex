---
layout: "yandex"
page_title: "Yandex: yandex_compute_instance_group"
sidebar_current: "docs-yandex-datasource-compute-instance-group"
description: |-
  Get information about a Yandex Compute Instance Group.
---

# yandex\_compute\_instance\_group

Get information about a Yandex Compute instance group.

## Example Usage

```hcl
data "yandex_compute_instance_group" "my_group" {
  instance_group_id = "some_instance_group_id"
}

output "instance_external_ip" {
  value = "${data.yandex_compute_instance_group.my_group.network_interface.0.nat_ip_address}"
}
```

## Argument Reference

The following arguments are supported:

* `instance_group_id` - (Required) The ID of a specific instance group.

## Attributes Reference


* `name` - The name of the instance group.
* `description` - A description of the instance group.
* `folder_id` - The ID of the folder that the instance group belongs to.
* `labels` - A set of key/value label pairs to assign to the instance group.
* `health_check` - Health check specification.
The structure is documented below.
* `load_balancer` - Load balancing specification.
The structure is documented below.
* `deploy_policy` - The deployment policy of the instance group.
The structure is documented below.
* `allocation_policy` - The allocation policy of the instance group by zone and region.
The structure is documented below.
* `instances` - A list of instances in the specified instance group.
The structure is documented below.
* `instance_template` - The instance template that the instance group belongs to.
The structure is documented below.
* `service_account_id` - The ID of the service account authorized for this instance group. 
* `scale_policy` - The scaling policy of the instance group.
The structure is documented below.
* `load_balancer_state` - Information about which entities can be attached to this load balancer.
The structure is documented below.
* `created_at` - The instance group creation timestamp.
---

The `load_balancer_state` block supports:

* `target_group_id` - The ID of the target group used for load balancing.
* `status_message` - The status message of the target group.
---

The `scale_policy` block supports:

* `fixed_scale` - The fixed scaling policy of the instance group.
The structure is documented below.
---

The `fixed_scale` block supports:

* `size` - The number of instances in the instance group.
---

The `instance_template` block supports:

* `description` - A description of the instance template.
* `platform_id` - The ID of the hardware platform configuration for the instance.
* `service_account_id` - The service account ID for the instance.
* `metadata` - The set of metadata `key:value` pairs assigned to this instance template. This includes custom metadata and predefined keys.
* `labels` - A map of labels applied to this instance.
* `resources.0.memory` - The memory size allocated to the instance.
* `resources.0.cores` - Number of CPU cores allocated to the instance.
* `resources.0.core_fraction` - Baseline core performance as a percent.
* `scheduling_policy` - The scheduling policy for the instance.
The structure is documented below.
* `network_interface` - An array with the network interfaces that will be attached to the instance.
The structure is documented below.
* `secondary_disk` - An array with the secondary disks that will be attached to the instance.
The structure is documented below.
* `boot_disk` - The specifications for boot disks that will be attached to the instance.
The structure is documented below.
---

The `boot_disk` block supports:

* `device_name` - This value can be used to reference the device under `/dev/disk/by-id/`.
* `mode` - The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.
* `initialize_params` - The parameters used for creating a disk alongside the instance.
The structure is documented below.
---

The `initialize_params` block supports:

* `description` - A description of the boot disk.
* `size` - The size of the disk in GB.
* `type` - The disk type.
* `image_id` - The disk image to initialize this disk from.
* `snapshot_id` - The snapshot to initialize this disk from.
---

The `secondary_disk` block supports:

* `device_name` - This value can be used to reference the device under `/dev/disk/by-id/`.
* `mode` - The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.
* `initialize_params` - The parameters used for creating a disk alongside the instance.
The structure is documented below.
---

The `initialize_params` block supports:

* `description` - A description of the boot disk.
* `size` - The size of the disk in GB.
* `type` - The disk type.
* `image_id` - The disk image to initialize this disk from.
* `snapshot_id` - The snapshot to initialize this disk from.
---

The `network_interface` block supports:

* `network_id` - The ID of the network.
* `subnet_ids` - The IDs of the subnets.
* `nat` - A public address that can be used to access the internet over NAT.
---

The `scheduling_policy` block supports:

* `preemptible` - Specifies if the instance is preemptible. Defaults to false.
---

The `instances` block supports:

* `instance_id` - The ID of the instance.
* `name` - The name of the managed instance.
* `fqdn` - The Fully Qualified Domain Name.
* `status` - The status of the instance.
* `status_message` - The status message of the instance.
* `zone_id` - The ID of the availability zone where the instance resides.
* `network_interface` - An array with the network interfaces attached to the managed instance.
The structure is documented below.
---

The `network_interface` block supports:

* `index` - The index of the network interface as generated by the server.
* `mac_address` - The MAC address assigned to the network interface.
* `ip_address` - The private IP address to assign to the instance. If empty, the address is automatically assigned from the specified subnet.
* `subnet_id` - The ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.
* `nat` - The instance's public address for accessing the internet over NAT.
* `nat_ip_address` - The public IP address of the instance.
* `nat_ip_version` - The IP version for the public address.

---

The `allocation_policy` block supports:

* `zones` - A list of availability zones.
---

The `deploy_policy` block supports:

* `max_unavailable` - The maximum number of running instances that can be taken offline (stopped or deleted) at the same time
during the update process.
* `max_expansion` - The maximum number of instances that can be temporarily allocated above the group's target size during the update process.
* `max_deleting` - The maximum number of instances that can be deleted at the same time.
* `max_creating` - The maximum number of instances that can be created at the same time.
* `startup_duration` - The amount of time in seconds to allow for an instance to start.
Instance will be considered up and running (and start receiving traffic) only after the startup_duration
has elapsed and all health checks are passed.
---

The `load_balancer` block supports:

* `target_group_name` - The name of the target group.
* `target_group_description` - A description of the target group.
* `target_group_labels` - A set of key/value label pairs.
---

The `health_check` block supports:

* `interval` - The interval between health checks in seconds.
* `timeout` - Timeout for the managed instance to return a response for the health check in seconds.
* `healthy_threshold` - The number of successful health checks before the managed instance is declared healthy.
* `unhealthy_threshold` - The number of failed health checks before the managed instance is declared unhealthy.
* `tcp_options` - TCP check options.
The structure is documented below.
* `http_options` - HTTP check options.
The structure is documented below.
---

The `http_options` block supports:

* `port` - The port used for HTTP health checks.
* `path` - The URL path used for health check requests.
---

The `tcp_options` block supports:

* `port` - The port to use for TCP health checks.
