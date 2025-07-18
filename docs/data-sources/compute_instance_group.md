---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_instance_group"
description: |-
  Get information about a Yandex Compute Instance Group.
---

# yandex_compute_instance_group (Data Source)

Get information about a Yandex Compute instance group.

## Example usage

```terraform
//
// Get information about existing Compute Instance Group (IG)
//
data "yandex_compute_instance_group" "my_group" {
  instance_group_id = "some_instance_group_id"
}

output "instance_external_ip" {
  value = data.yandex_compute_instance_group.my_group.instances.*.network_interface.0.nat_ip_address
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_group_id` (String) The ID of a specific instance group.

### Read-Only

- `allocation_policy` (List of Object) (see [below for nested schema](#nestedatt--allocation_policy))
- `application_balancer_state` (List of Object) (see [below for nested schema](#nestedatt--application_balancer_state))
- `application_load_balancer` (List of Object) (see [below for nested schema](#nestedatt--application_load_balancer))
- `created_at` (String) The creation timestamp of the resource.
- `deletion_protection` (Boolean) The `true` value means that resource is protected from accidental deletion.
- `deploy_policy` (List of Object) (see [below for nested schema](#nestedatt--deploy_policy))
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `health_check` (List of Object) (see [below for nested schema](#nestedatt--health_check))
- `id` (String) The ID of this resource.
- `instance_template` (List of Object) (see [below for nested schema](#nestedatt--instance_template))
- `instances` (List of Object) (see [below for nested schema](#nestedatt--instances))
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `load_balancer` (List of Object) (see [below for nested schema](#nestedatt--load_balancer))
- `load_balancer_state` (List of Object) (see [below for nested schema](#nestedatt--load_balancer_state))
- `max_checking_health_duration` (Number) Timeout for waiting for the VM to become healthy. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.
- `name` (String) The resource name.
- `scale_policy` (List of Object) (see [below for nested schema](#nestedatt--scale_policy))
- `service_account_id` (String) [Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) which linked to the resource.
- `status` (String) The status of the instance.
- `variables` (Map of String) A set of key/value variables pairs to assign to the instance group.

<a id="nestedatt--allocation_policy"></a>
### Nested Schema for `allocation_policy`

Read-Only:

- `instance_tags_pool` (Block List) Array of availability zone IDs with list of instance tags. (see [below for nested schema](#nestedobjatt--allocation_policy--instance_tags_pool))

- `zones` (Set of String) A list of [availability zones](https://yandex.cloud/docs/overview/concepts/geo-scope).


<a id="nestedobjatt--allocation_policy--instance_tags_pool"></a>
### Nested Schema for `allocation_policy.instance_tags_pool`

Read-Only:

- `tags` (List of String) List of tags for instances in zone.

- `zone` (String) Availability zone.




<a id="nestedatt--application_balancer_state"></a>
### Nested Schema for `application_balancer_state`

Read-Only:

- `status_message` (String)
- `target_group_id` (String)


<a id="nestedatt--application_load_balancer"></a>
### Nested Schema for `application_load_balancer`

Read-Only:

- `ignore_health_checks` (Boolean) Do not wait load balancer health checks.

- `max_opening_traffic_duration` (Number) Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

- `status_message` (String) The status message of the instance.

- `target_group_description` (String) A description of the target group.

- `target_group_id` (String) The ID of the target group.

- `target_group_labels` (Map of String) A set of key/value label pairs.

- `target_group_name` (String) The name of the target group.



<a id="nestedatt--deploy_policy"></a>
### Nested Schema for `deploy_policy`

Read-Only:

- `max_creating` (Number) The maximum number of instances that can be created at the same time.

- `max_deleting` (Number) The maximum number of instances that can be deleted at the same time.

- `max_expansion` (Number) The maximum number of instances that can be temporarily allocated above the group's target size during the update process.

- `max_unavailable` (Number) The maximum number of running instances that can be taken offline (stopped or deleted) at the same time during the update process.

- `startup_duration` (Number) The amount of time in seconds to allow for an instance to start. Instance will be considered up and running (and start receiving traffic) only after the startup_duration has elapsed and all health checks are passed.

- `strategy` (String) Affects the lifecycle of the instance during deployment. If set to `proactive` (default), Instance Groups can forcefully stop a running instance. If `opportunistic`, Instance Groups does not stop a running instance. Instead, it will wait until the instance stops itself or becomes unhealthy.



<a id="nestedatt--health_check"></a>
### Nested Schema for `health_check`

Read-Only:

- `healthy_threshold` (Number) The number of successful health checks before the managed instance is declared healthy.

- `http_options` (Block List, Max: 1) HTTP check options. (see [below for nested schema](#nestedobjatt--health_check--http_options))

- `interval` (Number) The interval to wait between health checks in seconds.

- `tcp_options` (Block List, Max: 1) TCP check options. (see [below for nested schema](#nestedobjatt--health_check--tcp_options))

- `timeout` (Number) The length of time to wait for a response before the health check times out in seconds.

- `unhealthy_threshold` (Number) The number of failed health checks before the managed instance is declared unhealthy.


<a id="nestedobjatt--health_check--http_options"></a>
### Nested Schema for `health_check.http_options`

Read-Only:

- `path` (String) The URL path used for health check requests.

- `port` (Number) The port used for HTTP health checks.



<a id="nestedobjatt--health_check--tcp_options"></a>
### Nested Schema for `health_check.tcp_options`

Read-Only:

- `port` (Number) The port used for TCP health checks.




<a id="nestedatt--instance_template"></a>
### Nested Schema for `instance_template`

Read-Only:

- `boot_disk` (Block List, Min: 1, Max: 1) Boot disk specifications for the instance. (see [below for nested schema](#nestedobjatt--instance_template--boot_disk))

- `description` (String) A description of the instance.

- `filesystem` (Block Set) List of filesystems to attach to the instance. (see [below for nested schema](#nestedobjatt--instance_template--filesystem))

- `hostname` (String) Hostname template for the instance. This field is used to generate the FQDN value of instance. The `hostname` must be unique within the network and region. If not specified, the hostname will be equal to `id` of the instance and FQDN will be `<id>.auto.internal`. Otherwise FQDN will be `<hostname>.<region_id>.internal`.

- `labels` (Map of String) A set of key/value label pairs to assign to the instance.

- `metadata` (Map of String) A set of metadata key/value pairs to make available from within the instance.

- `metadata_options` (Block List, Max: 1) Options allow user to configure access to managed instances metadata (see [below for nested schema](#nestedobjatt--instance_template--metadata_options))

- `name` (String) Name template of the instance.

- `network_interface` (Block List, Min: 1) Network specifications for the instance. This can be used multiple times for adding multiple interfaces. (see [below for nested schema](#nestedobjatt--instance_template--network_interface))

- `network_settings` (Block List) Network acceleration type for instance. (see [below for nested schema](#nestedobjatt--instance_template--network_settings))

- `placement_policy` (Block List, Max: 1) The placement policy configuration. (see [below for nested schema](#nestedobjatt--instance_template--placement_policy))

- `platform_id` (String) The ID of the hardware platform configuration for the instance.

- `resources` (Block List, Min: 1, Max: 1) Compute resource specifications for the instance. (see [below for nested schema](#nestedobjatt--instance_template--resources))

- `scheduling_policy` (Block List, Max: 1) The scheduling policy configuration. (see [below for nested schema](#nestedobjatt--instance_template--scheduling_policy))

- `secondary_disk` (Block List) A list of disks to attach to the instance. (see [below for nested schema](#nestedobjatt--instance_template--secondary_disk))

- `service_account_id` (String) The ID of the service account authorized for this instance.


<a id="nestedobjatt--instance_template--boot_disk"></a>
### Nested Schema for `instance_template.boot_disk`

Read-Only:

- `device_name` (String) This value can be used to reference the device under `/dev/disk/by-id/`.

- `disk_id` (String) The ID of the existing disk (such as those managed by yandex_compute_disk) to attach as a boot disk.

- `initialize_params` (Block List, Max: 1) Parameters for creating a disk alongside the instance. (see [below for nested schema](#nestedobjatt--instance_template--boot_disk--initialize_params))

- `mode` (String) The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.

- `name` (String) When set can be later used to change DiskSpec of actual disk.


<a id="nestedobjatt--instance_template--boot_disk--initialize_params"></a>
### Nested Schema for `instance_template.boot_disk.initialize_params`

Read-Only:

- `description` (String) A description of the boot disk.

- `image_id` (String) The disk image to initialize this disk from.

- `size` (Number) The size of the disk in GB.

- `snapshot_id` (String) The snapshot to initialize this disk from.

- `type` (String) The disk type.




<a id="nestedobjatt--instance_template--filesystem"></a>
### Nested Schema for `instance_template.filesystem`

Read-Only:

- `device_name` (String) Name of the device representing the filesystem on the instance.

- `filesystem_id` (String) ID of the filesystem that should be attached.

- `mode` (String) Mode of access to the filesystem that should be attached. By default, filesystem is attached in `READ_WRITE` mode.



<a id="nestedobjatt--instance_template--metadata_options"></a>
### Nested Schema for `instance_template.metadata_options`

Read-Only:

- `aws_v1_http_endpoint` (Number) Enables access to AWS flavored metadata (IMDSv1). Possible values: `0`, `1` for `enabled` and `2` for `disabled`.

- `aws_v1_http_token` (Number) Enables access to IAM credentials with AWS flavored metadata (IMDSv1). Possible values: `0`, `1` for `enabled` and `2` for `disabled`.

- `gce_http_endpoint` (Number) Enables access to GCE flavored metadata. Possible values: `0`, `1` for `enabled` and `2` for `disabled`.

- `gce_http_token` (Number) Enables access to IAM credentials with GCE flavored metadata. Possible values: `0`, `1` for `enabled` and `2` for `disabled`.



<a id="nestedobjatt--instance_template--network_interface"></a>
### Nested Schema for `instance_template.network_interface`

Read-Only:

- `dns_record` (Block List) List of DNS records. (see [below for nested schema](#nestedobjatt--instance_template--network_interface--dns_record))

- `ip_address` (String) Manual set static IP address.

- `ipv4` (Boolean) Allocate an IPv4 address for the interface. The default value is `true`.

- `ipv6` (Boolean) If `true`, allocate an IPv6 address for the interface. The address will be automatically assigned from the specified subnet.

- `ipv6_address` (String) Manual set static IPv6 address.

- `ipv6_dns_record` (Block List) List of IPv6 DNS records. (see [below for nested schema](#nestedobjatt--instance_template--network_interface--ipv6_dns_record))

- `nat` (Boolean) Flag for using NAT.

- `nat_dns_record` (Block List) List of NAT DNS records. (see [below for nested schema](#nestedobjatt--instance_template--network_interface--nat_dns_record))

- `nat_ip_address` (String) A public address that can be used to access the internet over NAT. Use `variables` to set.

- `network_id` (String) The ID of the network.

- `security_group_ids` (Set of String) Security group (SG) `IDs` for network interface.

- `subnet_ids` (Set of String) The ID of the subnets to attach this interface to.


<a id="nestedobjatt--instance_template--network_interface--dns_record"></a>
### Nested Schema for `instance_template.network_interface.dns_record`

Read-Only:

- `dns_zone_id` (String) DNS zone id (if not set, private zone used).

- `fqdn` (String) DNS record FQDN (must have dot at the end).

- `ptr` (Boolean) When set to `true`, also create PTR DNS record.

- `ttl` (Number) DNS record TTL.



<a id="nestedobjatt--instance_template--network_interface--ipv6_dns_record"></a>
### Nested Schema for `instance_template.network_interface.ipv6_dns_record`

Read-Only:

- `dns_zone_id` (String) DNS zone id (if not set, private zone used).

- `fqdn` (String) DNS record FQDN (must have dot at the end).

- `ptr` (Boolean) When set to `true`, also create PTR DNS record.

- `ttl` (Number) DNS record TTL.



<a id="nestedobjatt--instance_template--network_interface--nat_dns_record"></a>
### Nested Schema for `instance_template.network_interface.nat_dns_record`

Read-Only:

- `dns_zone_id` (String) DNS zone id (if not set, private zone used).

- `fqdn` (String) DNS record FQDN (must have dot at the end).

- `ptr` (Boolean) When set to `true`, also create PTR DNS record.

- `ttl` (Number) DNS record TTL.




<a id="nestedobjatt--instance_template--network_settings"></a>
### Nested Schema for `instance_template.network_settings`

Read-Only:

- `type` (String) Network acceleration type. By default a network is in `STANDARD` mode.



<a id="nestedobjatt--instance_template--placement_policy"></a>
### Nested Schema for `instance_template.placement_policy`

Read-Only:

- `placement_group_id` (String) Specifies the id of the Placement Group to assign to the instances.



<a id="nestedobjatt--instance_template--resources"></a>
### Nested Schema for `instance_template.resources`

Read-Only:

- `core_fraction` (Number) If provided, specifies baseline core performance as a percent.

- `cores` (Number) The number of CPU cores for the instance.

- `gpus` (Number) If provided, specifies the number of GPU devices for the instance.

- `memory` (Number) The memory size in GB.



<a id="nestedobjatt--instance_template--scheduling_policy"></a>
### Nested Schema for `instance_template.scheduling_policy`

Read-Only:

- `preemptible` (Boolean) Specifies if the instance is preemptible. Defaults to `false`.



<a id="nestedobjatt--instance_template--secondary_disk"></a>
### Nested Schema for `instance_template.secondary_disk`

Read-Only:

- `device_name` (String) This value can be used to reference the device under `/dev/disk/by-id/`.

- `disk_id` (String) ID of the existing disk. To set use variables.

- `initialize_params` (Block List, Max: 1) Parameters used for creating a disk alongside the instance. (see [below for nested schema](#nestedobjatt--instance_template--secondary_disk--initialize_params))

- `mode` (String) The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.

- `name` (String) When set can be later used to change DiskSpec of actual disk.


<a id="nestedobjatt--instance_template--secondary_disk--initialize_params"></a>
### Nested Schema for `instance_template.secondary_disk.initialize_params`

Read-Only:

- `description` (String) A description of the boot disk.

- `image_id` (String) The disk image to initialize this disk from.

- `size` (Number) The size of the disk in GB.

- `snapshot_id` (String) The snapshot to initialize this disk from.

- `type` (String) The disk type.





<a id="nestedatt--instances"></a>
### Nested Schema for `instances`

Read-Only:

- `fqdn` (String)
- `instance_id` (String)
- `instance_tag` (String)
- `name` (String)
- `network_interface` (List of Object) (see [below for nested schema](#nestedobjatt--instances--network_interface)) (see [below for nested schema](#nestedobjatt--instances--network_interface))

- `status` (String)
- `status_changed_at` (String)
- `status_message` (String)
- `zone_id` (String)

<a id="nestedobjatt--instances--network_interface"></a>
### Nested Schema for `instances.network_interface`

Read-Only:

- `index` (Number)
- `ip_address` (String)
- `ipv4` (Boolean)
- `ipv6` (Boolean)
- `ipv6_address` (String)
- `mac_address` (String)
- `nat` (Boolean)
- `nat_ip_address` (String)
- `nat_ip_version` (String)
- `subnet_id` (String)



<a id="nestedatt--load_balancer"></a>
### Nested Schema for `load_balancer`

Read-Only:

- `ignore_health_checks` (Boolean) Do not wait load balancer health checks.

- `max_opening_traffic_duration` (Number) Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

- `status_message` (String) The status message of the target group.

- `target_group_description` (String) A description of the target group.

- `target_group_id` (String) The ID of the target group.

- `target_group_labels` (Map of String) A set of key/value label pairs.

- `target_group_name` (String) The name of the target group.



<a id="nestedatt--load_balancer_state"></a>
### Nested Schema for `load_balancer_state`

Read-Only:

- `status_message` (String)
- `target_group_id` (String)


<a id="nestedatt--scale_policy"></a>
### Nested Schema for `scale_policy`

Read-Only:

- `auto_scale` (Block List, Max: 1) The auto scaling policy of the instance group. (see [below for nested schema](#nestedobjatt--scale_policy--auto_scale))

- `fixed_scale` (Block List, Max: 1) The fixed scaling policy of the instance group. (see [below for nested schema](#nestedobjatt--scale_policy--fixed_scale))

- `test_auto_scale` (Block List, Max: 1) The test auto scaling policy of the instance group. Use it to test how the auto scale works. (see [below for nested schema](#nestedobjatt--scale_policy--test_auto_scale))


<a id="nestedobjatt--scale_policy--auto_scale"></a>
### Nested Schema for `scale_policy.auto_scale`

Read-Only:

- `auto_scale_type` (String) Autoscale type, can be `ZONAL` or `REGIONAL`. By default `ZONAL` type is used.

- `cpu_utilization_target` (Number) Target CPU load level.

- `custom_rule` (Block List) A list of custom rules. (see [below for nested schema](#nestedobjatt--scale_policy--auto_scale--custom_rule))

- `initial_size` (Number) The initial number of instances in the instance group.

- `max_size` (Number) The maximum number of virtual machines in the group.

- `measurement_duration` (Number) The amount of time, in seconds, that metrics are averaged for. If the average value at the end of the interval is higher than the `cpu_utilization_target`, the instance group will increase the number of virtual machines in the group.

- `min_zone_size` (Number) The minimum number of virtual machines in a single availability zone.

- `stabilization_duration` (Number) The minimum time interval, in seconds, to monitor the load before an instance group can reduce the number of virtual machines in the group. During this time, the group will not decrease even if the average load falls below the value of `cpu_utilization_target`.

- `warmup_duration` (Number) The warm-up time of the virtual machine, in seconds. During this time, traffic is fed to the virtual machine, but load metrics are not taken into account.


<a id="nestedobjatt--scale_policy--auto_scale--custom_rule"></a>
### Nested Schema for `scale_policy.auto_scale.custom_rule`

Read-Only:

- `folder_id` (String) If specified, sets the folder id to fetch metrics from. By default, it is the ID of the folder the group belongs to.

- `labels` (Map of String) Metrics [labels](https://yandex.cloud/en/docs/monitoring/concepts/data-model#label) from Monitoring.

- `metric_name` (String) Name of the metric in Monitoring.

- `metric_type` (String) Type of metric, can be `GAUGE` or `COUNTER`. `GAUGE` metric reflects the value at particular time point. `COUNTER` metric exhibits a monotonous growth over time.

- `rule_type` (String) The metric rule type (UTILIZATION, WORKLOAD). UTILIZATION for metrics describing resource utilization per VM instance. WORKLOAD for metrics describing total workload on all VM instances.

- `service` (String) If specified, sets the service name to fetch metrics. The default value is `custom`. You can use a label to specify service metrics, e.g., `service` with the `compute` value for Compute Cloud.

- `target` (Number) Target metric value by which Instance Groups calculates the number of required VM instances.




<a id="nestedobjatt--scale_policy--fixed_scale"></a>
### Nested Schema for `scale_policy.fixed_scale`

Read-Only:

- `size` (Number) The number of instances in the instance group.



<a id="nestedobjatt--scale_policy--test_auto_scale"></a>
### Nested Schema for `scale_policy.test_auto_scale`

Read-Only:

- `auto_scale_type` (String) Autoscale type, can be `ZONAL` or `REGIONAL`. By default `ZONAL` type is used.

- `cpu_utilization_target` (Number) Target CPU load level.

- `custom_rule` (Block List) A list of custom rules. (see [below for nested schema](#nestedobjatt--scale_policy--test_auto_scale--custom_rule))

- `initial_size` (Number) The initial number of instances in the instance group.

- `max_size` (Number) The maximum number of virtual machines in the group.

- `measurement_duration` (Number) The amount of time, in seconds, that metrics are averaged for. If the average value at the end of the interval is higher than the `cpu_utilization_target`, the instance group will increase the number of virtual machines in the group.

- `min_zone_size` (Number) The minimum number of virtual machines in a single availability zone.

- `stabilization_duration` (Number) The minimum time interval, in seconds, to monitor the load before an instance group can reduce the number of virtual machines in the group. During this time, the group will not decrease even if the average load falls below the value of `cpu_utilization_target`.

- `warmup_duration` (Number) The warm-up time of the virtual machine, in seconds. During this time, traffic is fed to the virtual machine, but load metrics are not taken into account.


<a id="nestedobjatt--scale_policy--test_auto_scale--custom_rule"></a>
### Nested Schema for `scale_policy.test_auto_scale.custom_rule`

Read-Only:

- `folder_id` (String) Folder ID of custom metric in Yandex Monitoring that should be used for scaling.

- `labels` (Map of String) A map of labels of metric.

- `metric_name` (String) The name of metric.

- `metric_type` (String) Metric type, `GAUGE` or `COUNTER`.

- `rule_type` (String) Rule type: `UTILIZATION` - This type means that the metric applies to one instance. First, Instance Groups calculates the average metric value for each instance, then averages the values for instances in one availability zone. This type of metric must have the `instance_id` label. `WORKLOAD` - This type means that the metric applies to instances in one availability zone. This type of metric must have the `zone_id` label.

- `service` (String) Service of custom metric in Yandex Monitoring that should be used for scaling.

- `target` (Number) Target metric value level.

