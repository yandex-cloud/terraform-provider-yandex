---
layout: "yandex"
page_title: "Yandex: yandex_compute_instance_group"
sidebar_current: "docs-yandex-compute-instance-group"
description: |-
  Manages an Instance group resource.
---

# yandex\_compute\_instance\_group

An Instance group resource. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/instance-groups/).

## Example Usage

```hcl
resource "yandex_compute_instance_group" "group1" {
  name                = "test-ig"
  folder_id           = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id  = "${yandex_iam_service_account.test_account.id}"
  deletion_protection = true
  instance_template {
    platform_id = "standard-v1"
    resources {
      memory = 2
      cores  = 2
    }
    boot_disk {
      mode = "READ_WRITE"
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }
    network_interface {
      network_id = "${yandex_vpc_network.my-inst-group-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.my-inst-group-subnet.id}"]
    }
    labels = {
      label1 = "label1-value"
      label2 = "label2-value"
    }
    metadata = {
      foo      = "bar"
      ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
    }
    network_settings {
      type = "STANDARD"
    }
  }

  variables = {
    test_key1 = "test_value1"
    test_key2 = "test_value2"
  }

  scale_policy {
    fixed_scale {
      size = 3
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 2
    max_creating    = 2
    max_expansion   = 2
    max_deleting    = 2
  }
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) The ID of the folder that the resources belong to.

* `scale_policy` - (Required) The scaling policy of the instance group. The structure is documented below.

* `deploy_policy` - (Required) The deployment policy of the instance group. The structure is documented below.

* `service_account_id` - (Required) The ID of the service account authorized for this instance group.

* `instance_template` - (Required) The template for creating new instances. The structure is documented below.

* `allocation_policy` - (Required) The allocation policy of the instance group by zone and region. The structure is documented below.

* `name` - (Optional) The name of the instance group.

* `health_check` - (Optional) Health check specifications. The structure is documented below.

* `max_checking_health_duration` - (Optional) Timeout for waiting for the VM to become healthy. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

* `load_balancer` - (Optional) Load balancing specifications. The structure is documented below.

* `application_load_balancer` - (Optional) Application Load balancing (L7) specifications. The structure is documented below.

* `description` - (Optional) A description of the instance group.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance group.

* `variables` - (Optional) A set of key/value  variables pairs to assign to the instance group.

* `deletion_protection` - (Optional) Flag that protects the instance group from accidental deletion.

---

The `application_load_balancer` block supports:

* `target_group_name` - (Optional) The name of the target group.

* `target_group_description` - (Optional) A description of the target group.

* `target_group_labels` - (Optional) A set of key/value label pairs.

* `max_opening_traffic_duration` - (Optional) Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

---

The `load_balancer` block supports:

* `target_group_name` - (Optional) The name of the target group.

* `target_group_description` - (Optional) A description of the target group.

* `target_group_labels` - (Optional) A set of key/value label pairs.

* `max_opening_traffic_duration` - (Optional) Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.

---

The `health_check` block supports:

* `interval` - (Optional) The interval to wait between health checks in seconds.

* `timeout` - (Optional) The length of time to wait for a response before the health check times out in seconds.

* `healthy_threshold` - (Optional) The number of successful health checks before the managed instance is declared healthy.

* `unhealthy_threshold` - (Optional) The number of failed health checks before the managed instance is declared unhealthy.

* `tcp_options` - (Optional) TCP check options. The structure is documented below.

* `http_options` - (Optional) HTTP check options. The structure is documented below.

---

The `http_options` block supports:

* `port` - (Required) The port used for HTTP health checks.

* `path` - (Required) The URL path used for health check requests.

---

The `tcp_options` block supports:

* `port` - (Required) The port used for TCP health checks.

---

The `allocation_policy` block supports:

* `zones` - (Required) A list of availability zones.

---

The `instance_template` block supports:

* `boot_disk` - (Required) Boot disk specifications for the instance. The structure is documented below.

* `resources` - (Required) Compute resource specifications for the instance. The structure is documented below.

* `network_interface` - (Required) Network specifications for the instance. This can be used multiple times for adding multiple interfaces. The structure is documented below.

* `scheduling_policy` - (Optional) The scheduling policy configuration. The structure is documented below.

* `placement_policy` - (Optional) The placement policy configuration. The structure is documented below.

* `description` - (Optional) A description of the instance.

* `metadata` - (Optional) A set of metadata key/value pairs to make available from within the instance.

* `labels` - (Optional) A set of key/value label pairs to assign to the instance.

* `platform_id` - (Optional) The ID of the hardware platform configuration for the instance. The default is 'standard-v1'.

* `secondary_disk` - (Optional) A list of disks to attach to the instance. The structure is documented below.

* `service_account_id` - (Optional) The ID of the service account authorized for this instance.

* `network_settings` - (Optional) Network acceleration type for instance. The structure is documented below.

* `name` - (Optional) Name template of the instance.  
In order to be unique it must contain at least one of instance unique placeholders:   
{instance.short_id}   
{instance.index}   
combination of {instance.zone_id} and {instance.index_in_zone}   
Example: my-instance-{instance.index}  
If not set, default is used: {instance_group.id}-{instance.short_id}   
It may also contain another placeholders, see metadata doc for full list.
* `hostname` - (Optional) Hostname template for the instance.   
This field is used to generate the FQDN value of instance.   
The hostname must be unique within the network and region.   
If not specified, the hostname will be equal to id of the instance   
and FQDN will be `<id>.auto.internal`. Otherwise FQDN will be `<hostname>.<region_id>.internal`.   
In order to be unique it must contain at least on of instance unique placeholders:   
 {instance.short_id}   
 {instance.index}   
 combination of {instance.zone_id} and {instance.index_in_zone}   
Example: my-instance-{instance.index}   
If not set, `name` value will be used   
It may also contain another placeholders, see metadata doc for full list.

---

The `secondary_disk` block supports:

* `mode` - (Optional) The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.

* `disk_id` - (Optional) ID of the existing disk. To set use variables.

* `initialize_params` - (Optional) Parameters used for creating a disk alongside the instance. The structure is documented below.

- - -
* `device_name` - (Optional) This value can be used to reference the device under `/dev/disk/by-id/`.

---

The `initialize_params` block supports:

* `description` - (Optional) A description of the boot disk.

* `size` - (Optional) The size of the disk in GB.

* `type` - (Optional) The disk type.

* `image_id` - (Optional) The disk image to initialize this disk from.

* `snapshot_id` - (Optional) The snapshot to initialize this disk from.

~> **NOTE:** `image_id` or `snapshot_id` must be specified.

---

The `scheduling_policy` block supports:

* `preemptible` - (Optional) Specifies if the instance is preemptible. Defaults to false.

---

The `placement_policy` block supports:

* `placement_group_id` - (Optional) Specifies the id of the Placement Group to assign to the instances.

---

The `network_interface` block supports:

* `network_id` - (Optional) The ID of the network.

* `subnet_ids` - (Optional) The ID of the subnets to attach this interface to.

* `nat` - (Optional) Flag for using NAT.

* `nat_ip_address` - (Optional) A public address that can be used to access the internet over NAT. Use `variables` to set.
  
* `security_group_ids` - (Optional) Security group ids for network interface.

* `ip_address` - (Optional) Manual set static IP address.

* `ipv6_address` - (Optional) Manual set static IPv6 address.
  
* `dns_record` - (Optional) List of dns records.  The structure is documented below.

* `ipv6_dns_record` - (Optional) List of ipv6 dns records.  The structure is documented below.
  
* `nat_dns_record` - (Optional) List of nat dns records.  The structure is documented below.

---

The `dns_record` block supports:

* `fqdn` - (Required) DNS record fqdn (must have dot at the end).

* `dns_zone_id` - (Optional) DNS zone id (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL.

* `ptr` - (Optional) When set to true, also create PTR DNS record. ---

---

The `ipv6_dns_record` block supports:

* `fqdn` - (Required) DNS record fqdn (must have dot at the end).

* `dns_zone_id` - (Optional) DNS zone id (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL.

* `ptr` - (Optional) When set to true, also create PTR DNS record.

---

The `nat_dns_record` block supports:

* `fqdn` - (Required) DNS record fqdn (must have dot at the end).

* `dns_zone_id` - (Optional) DNS zone id (if not set, private zone used).

* `ttl` - (Optional) DNS record TTL.

* `ptr` - (Optional) When set to true, also create PTR DNS record.

---

The `resources` block supports:

* `memory` - (Required) The memory size in GB.

* `cores` - (Required) The number of CPU cores for the instance.

- - -
* `core_fraction` - (Optional) If provided, specifies baseline core performance as a percent.

---

The `boot_disk` block supports:

* `mode` - (Optional) The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.

* `initialize_params` - (Required) Parameters for creating a disk alongside the instance. The structure is documented below.

- - -
* `device_name` - (Optional) This value can be used to reference the device under `/dev/disk/by-id/`.

---

The `initialize_params` block supports:

* `description` - (Optional) A description of the boot disk.

* `size` - (Optional) The size of the disk in GB.

* `type` - (Optional) The disk type.

* `image_id` - (Optional) The disk image to initialize this disk from.

* `snapshot_id` - (Optional) The snapshot to initialize this disk from.

~> **NOTE:** `image_id` or `snapshot_id` must be specified.

---

The `deploy_policy` block supports:

* `max_unavailable` - (Required) The maximum number of running instances that can be taken offline (stopped or deleted) at the same time
during the update process.

* `max_expansion` - (Required) The maximum number of instances that can be temporarily allocated above the group's target size
during the update process.

- - -
* `max_deleting` - (Optional) The maximum number of instances that can be deleted at the same time.

* `max_creating` - (Optional) The maximum number of instances that can be created at the same time.

* `startup_duration` - (Optional) The amount of time in seconds to allow for an instance to start.
Instance will be considered up and running (and start receiving traffic) only after the startup_duration
has elapsed and all health checks are passed.

* `strategy` - (Optional) Affects the lifecycle of the instance during deployment. If set to `proactive` (default), Instance Groups
  can forcefully stop a running instance. If `opportunistic`, Instance Groups does not stop a running instance. Instead,
  it will wait until the instance stops itself or becomes unhealthy.
  
---

The `scale_policy` block supports:

* `fixed_scale` - (Optional) The fixed scaling policy of the instance group. The structure is documented below.

* `auto_scale` - (Optional) The auto scaling policy of the instance group. The structure is documented below.

~> **NOTE:** Either `fixed_scale` or `auto_scale` must be specified.

* `test_auto_scale` - (Optional) The test auto scaling policy of the instance group. Use it to test how the auto scale works. The structure is documented below.

---

The `fixed_scale` block supports:

* `size` - (Required) The number of instances in the instance group.

---

The `auto_scale` block supports:

* `initial_size` - (Required) The initial number of instances in the instance group.

* `measurement_duration` - (Required) The amount of time, in seconds, that metrics are averaged for.
If the average value at the end of the interval is higher than the `cpu_utilization_target`,
the instance group will increase the number of virtual machines in the group.

* `cpu_utilization_target` - (Required) Target CPU load level.

* `min_zone_size` - (Optional) The minimum number of virtual machines in a single availability zone.

* `max_size` - (Optional) The maximum number of virtual machines in the group.

* `warmup_duration` - (Optional) The warm-up time of the virtual machine, in seconds. During this time,
traffic is fed to the virtual machine, but load metrics are not taken into account.

* `stabilization_duration` - (Optional) The minimum time interval, in seconds, to monitor the load before
an instance group can reduce the number of virtual machines in the group. During this time, the group
will not decrease even if the average load falls below the value of `cpu_utilization_target`.

* `custom_rule` - (Optional) A list of custom rules. The structure is documented below.

---

The `test_auto_scale` block supports:

* `initial_size` - (Required) The initial number of instances in the instance group.

* `measurement_duration` - (Required) The amount of time, in seconds, that metrics are averaged for.
If the average value at the end of the interval is higher than the `cpu_utilization_target`,
the instance group will increase the number of virtual machines in the group.

* `cpu_utilization_target` - (Required) Target CPU load level.

* `min_zone_size` - (Optional) The minimum number of virtual machines in a single availability zone.

* `max_size` - (Optional) The maximum number of virtual machines in the group.

* `warmup_duration` - (Optional) The warm-up time of the virtual machine, in seconds. During this time,
traffic is fed to the virtual machine, but load metrics are not taken into account.

* `stabilization_duration` - (Optional) The minimum time interval, in seconds, to monitor the load before
an instance group can reduce the number of virtual machines in the group. During this time, the group
will not decrease even if the average load falls below the value of `cpu_utilization_target`.

* `custom_rule` - (Optional) A list of custom rules. The structure is documented below.

---

The `custom_rule` block supports:

* `rule_type` - (Required) Rule type: `UTILIZATION` - This type means that the metric applies to one instance.
First, Instance Groups calculates the average metric value for each instance,
then averages the values for instances in one availability zone.
This type of metric must have the `instance_id` label. `WORKLOAD` - This type means that the metric applies to instances in one availability zone.
This type of metric must have the `zone_id` label.

* `metric_type` - (Required) Metric type, `GAUGE` or `COUNTER`.

* `metric_name` - (Required) The name of metric.

* `target` - (Required) Target metric value level.

* `labels` - (Optional) A map of labels of metric.

* `folder_id` - (Optional) Folder ID of custom metric in Yandex Monitoring that should be used for scaling.

* `service` - (Optional) Service of custom metric in Yandex Monitoring that should be used for scaling.

---

The `network_settings` block supports:

* `type` - (Optional) Network acceleration type. By default a network is in `STANDARD` mode.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the instance group.

* `created_at` - The instance group creation timestamp.

* `load_balancer.0.target_group_id` - The ID of the target group.

* `load_balancer.0.status_message` - The status message of the target group.

The `instances` block supports:

* `instance_id` - The ID of the instance.
* `name` - The name of the managed instance.
* `fqdn` - The Fully Qualified Domain Name.
* `status` - The status of the instance.
* `status_message` - The status message of the instance.
* `zone_id` - The ID of the availability zone where the instance resides.
* `network_interface` - An array with the network interfaces attached to the managed instance.

---

The `network_interface` block supports:

* `index` - The index of the network interface as generated by the server.
* `mac_address` - The MAC address assigned to the network interface.
* `ipv4` - True if IPv4 address allocated for the network interface.
* `ip_address` - The private IP address to assign to the instance. If empty, the address is automatically assigned from the specified subnet.
* `subnet_id` - The ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.
* `nat` - The instance's public address for accessing the internet over NAT.
* `nat_ip_address` - The public IP address of the instance.
* `nat_ip_version` - The IP version for the public address.
