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
  value = "${data.yandex_compute_instance_group.my_group.instances.*.network_interface.0.nat_ip_address}"
}
```

## Argument Reference

The following arguments are supported:

* `instance_group_id` - The ID of a specific instance group.

## Attributes Reference

* `name` - The name of the instance group.
* `description` - A description of the instance group.
* `folder_id` - The ID of the folder that the instance group belongs to.
* `labels` - A set of key/value label pairs to assign to the instance group.
* `health_check` - Health check specification. The structure is documented below.
* `max_checking_health_duration` - Timeout for waiting for the VM to become healthy. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

* `load_balancer` - Load balancing specification. The structure is documented below.

* `application_load_balancer` - Application Load balancing (L7) specifications. The structure is documented below.

* `deploy_policy` - The deployment policy of the instance group. The structure is documented below.

* `allocation_policy` - The allocation policy of the instance group by zone and region. The structure is documented below.

* `instances` - A list of instances in the specified instance group. The structure is documented below.

* `instance_template` - The instance template that the instance group belongs to. The structure is documented below.

* `service_account_id` - The ID of the service account authorized for this instance group. 
* `scale_policy` - The scaling policy of the instance group. The structure is documented below.

* `load_balancer_state` - Information about which entities can be attached to this load balancer. The structure is documented below.

* `application_load_balancer_state` - Information about which entities can be attached to this application load balancer. The structure is documented below.
  
* `created_at` - The instance group creation timestamp.

* `variables` - A set of key/value  variables pairs to assign to the instance group.

* `status` - Status of the instance group.

* `deletion_protection` - Flag that protects the instance group from accidental deletion.

---

The `application_load_balancer_state` block supports:

* `target_group_id` - The ID of the target group used for load balancing.
* `status_message` - The status message of the target group.

---
The `load_balancer_state` block supports:

* `target_group_id` - The ID of the target group used for load balancing.
* `status_message` - The status message of the target group.

---

The `scale_policy` block supports:

* `fixed_scale` - The fixed scaling policy of the instance group. The structure is documented below.

* `auto_scale` - The auto scaling policy of the instance group. The structure is documented below.

* `test_auto_scale` - The test auto scaling policy of the instance group. Use it to test how the auto scale works. The structure is documented below.

---

The `fixed_scale` block supports:

* `size` - The number of instances in the instance group.

---

The `auto_scale` block supports:

* `initial_size` - The initial number of instances in the instance group.

* `measurement_duration` - The amount of time, in seconds, that metrics are averaged for.
If the average value at the end of the interval is higher than the `cpu_utilization_target`,
the instance group will increase the number of virtual machines in the group.

* `min_zone_size` - The minimum number of virtual machines in a single availability zone.

* `max_size` - The maximum number of virtual machines in the group.

* `warmup_duration` - The warm-up time of the virtual machine, in seconds. During this time,
traffic is fed to the virtual machine, but load metrics are not taken into account.

* `stabilization_duration` - The minimum time interval, in seconds, to monitor the load before
an instance group can reduce the number of virtual machines in the group. During this time, the group
will not decrease even if the average load falls below the value of `cpu_utilization_target`.

* `cpu_utilization_target` - Target CPU load level.

* `custom_rule` - A list of custom rules. The structure is documented below.

---

The `test_auto_scale` block supports:

* `initial_size` - The initial number of instances in the instance group.

* `measurement_duration` - The amount of time, in seconds, that metrics are averaged for.
If the average value at the end of the interval is higher than the `cpu_utilization_target`,
the instance group will increase the number of virtual machines in the group.

* `min_zone_size` - The minimum number of virtual machines in a single availability zone.

* `max_size` - The maximum number of virtual machines in the group.

* `warmup_duration` - The warm-up time of the virtual machine, in seconds. During this time,
traffic is fed to the virtual machine, but load metrics are not taken into account.

* `stabilization_duration` - The minimum time interval, in seconds, to monitor the load before
an instance group can reduce the number of virtual machines in the group. During this time, the group
will not decrease even if the average load falls below the value of `cpu_utilization_target`.

* `cpu_utilization_target` - Target CPU load level.

* `custom_rule` - A list of custom rules. The structure is documented below.

---
The `custom_rule` block supports:

* `rule_type` - Rule type: `UTILIZATION` - This type means that the metric applies to one instance.
First, Instance Groups calculates the average metric value for each instance,
then averages the values for instances in one availability zone.
This type of metric must have the `instance_id` label. `WORKLOAD` - This type means that the metric applies to instances in one availability zone.
This type of metric must have the `zone_id` label.

* `metric_type` - Metric type, `GAUGE` or `COUNTER`.

* `metric_name` - The name of metric.

* `target` - Target metric value level.

* `labels` - A map of labels of metric.

* `folder_id` - Folder ID of custom metric in Yandex Monitoring that should be used for scaling.

* `service` - Service of custom metric in Yandex Monitoring that should be used for scaling.

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
* `resources.0.gpus` - Number of GPU cores allocated to the instance.
* `scheduling_policy` - The scheduling policy for the instance. The structure is documented below.
* `network_interface` - An array with the network interfaces that will be attached to the instance. The structure is documented below.
* `secondary_disk` - An array with the secondary disks that will be attached to the instance. The structure is documented below.
* `boot_disk` - The specifications for boot disk that will be attached to the instance. The structure is documented below.
* `network_settings` - Network acceleration settings. The structure is documented below.
* `name` - Name template of the instance.
* `hostname` - Hostname temaplate for the instance.

---

The `boot_disk` block supports:

* `device_name` - This value can be used to reference the device under `/dev/disk/by-id/`.
* `mode` - The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.
* `disk_id` - ID of the existing disk. To set use variables.
* `initialize_params` - The parameters used for creating a disk alongside the instance. The structure is documented below.

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
* `initialize_params` - The parameters used for creating a disk alongside the instance. The structure is documented below.

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
* `ipv4` - Is IPv4 address assigned.
* `nat` - Flag for using NAT.
* `nat_ip_address` - A public address that can be used to access the internet over NAT. Use `variables` to set.
* `security_group_ids` - Security group ids for network interface.
* `ip_address` - Manual set static IP address.
* `ipv6_address` - Manual set static IPv6 address.
* `dns_record` - List of dns records.  The structure is documented below.
* `ipv6_dns_record` - List of ipv6 dns records.  The structure is documented below.
* `nat_dns_record` - List of nat dns records.  The structure is documented below.

---

The `dns_record` block supports:

* `fqdn` - DNS record fqdn.
* `dns_zone_id` - DNS zone id (if not set, private zone is used).
* `ttl` - DNS record TTL.
* `ptr` - When set to true, also create PTR DNS record.

---

The `ipv6_dns_record` block supports:

* `fqdn` - DNS record fqdn.
* `dns_zone_id` - DNS zone id (if not set, private zone is used).
* `ttl` - DNS record TTL.
* `ptr` - When set to true, also create PTR DNS record.

---

The `nat_dns_record` block supports:

* `fqdn` - DNS record fqdn.
* `dns_zone_id` - DNS zone id (if not set, private zone is used).
* `ttl` - DNS record TTL.
* `ptr` - When set to true, also create PTR DNS record.

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
* `network_interface` - An array with the network interfaces attached to the managed instance. The structure is documented below.
* `status_changed_at` -The timestamp when the status of the managed instance was last changed.

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
* `strategy` - Affects the lifecycle of the instance during deployment. If set to `proactive` (default), Instance Groups
  can forcefully stop a running instance. If `opportunistic`, Instance Groups does not stop a running instance. Instead,
  it will wait until the instance stops itself or becomes unhealthy.

Instance will be considered up and running (and start receiving traffic) only after the startup_duration
has elapsed and all health checks are passed.

---

The `application_load_balancer` block supports:

* `target_group_name` - The name of the target group.
* `target_group_description` - A description of the target group.
* `target_group_labels` - A set of key/value label pairs.
* `target_group_id` - The ID of the target group.
* `status_message` - The status message of the target group.
* `max_opening_traffic_duration` - Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.


---

The `load_balancer` block supports:

* `target_group_name` - The name of the target group.
* `target_group_description` - A description of the target group.
* `target_group_labels` - A set of key/value label pairs.
* `target_group_id` - The ID of the target group.
* `status_message` - The status message of the target group.
* `max_opening_traffic_duration` - Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

---

The `health_check` block supports:

* `interval` - The interval between health checks in seconds.
* `timeout` - Timeout for the managed instance to return a response for the health check in seconds.
* `healthy_threshold` - The number of successful health checks before the managed instance is declared healthy.
* `unhealthy_threshold` - The number of failed health checks before the managed instance is declared unhealthy.
* `tcp_options` - TCP check options. The structure is documented below.
* `http_options` - HTTP check options. The structure is documented below.

---

The `http_options` block supports:

* `port` - The port used for HTTP health checks.
* `path` - The URL path used for health check requests.

---

The `tcp_options` block supports:

* `port` - The port to use for TCP health checks.

---

The `network_settings` block supports:

* `type` - Network acceleration type. By default a network is in `STANDARD` mode.
