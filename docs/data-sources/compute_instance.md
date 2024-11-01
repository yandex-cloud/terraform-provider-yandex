---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_instance"
description: |-
  Get information about a Yandex Compute Instance.
---


# yandex_compute_instance




Get information about a Yandex Compute instance. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/vm).

## Example usage

```terraform
data "yandex_compute_instance" "my_instance" {
  instance_id = "some_instance_id"
}

output "instance_external_ip" {
  value = data.yandex_compute_instance.my_instance.network_interface.0.nat_ip_address
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Optional) The ID of a specific instance.
* `name` - (Optional) Name of the instance.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> **NOTE:** One of `instance_id` or `name` should be specified.

## Attributes Reference

* `description` - Description of the instance.
* `fqdn` - FQDN of the instance.
* `zone` - Availability zone where the instance resides.
* `labels` - A set of key/value label pairs assigned to the instance.
* `metadata` - Metadata key/value pairs to make available from within the instance.
* `platform_id` - Type of virtual machine to create. Default is 'standard-v1'.
* `status` - Status of the instance.
* `resources.0.memory` - Memory size allocated for the instance.
* `resources.0.cores` - Number of CPU cores allocated for the instance.
* `resources.0.core_fraction` - Baseline performance for a core, set as a percent.
* `resources.0.gpus` - Number of GPU cores allocated for the instance.
* `boot_disk` - The boot disk for the instance. Structure is documented below.
* `network_acceleration_type` - Type of network acceleration. The default is `standard`. Values: `standard`, `software_accelerated`
* `network_interface` - The networks attached to the instance. Structure is documented below.
* `network_interface.0.ip_address` - An internal IP address of the instance, either manually or dynamically assigned.
* `network_interface.0.nat_ip_address` - An assigned external IP address if the instance has NAT enabled.
* `secondary_disk` - Set of secondary disks attached to the instance. Structure is documented below.
* `scheduling_policy` - Scheduling policy configuration. The structure is documented below.
* `service_account_id` - ID of the service account authorized for this instance.
* `created_at` - Instance creation timestamp.
* `placement_policy` - The placement policy configuration. The structure is documented below.
* `local_disk` - List of local disks that are attached to the instance. Structure is documented below.
* `gpu_cluster_id` - ID of GPU cluster if instance is part of it.
* `metadata_options` - Options allow user to configure access to instance's metadata
* `maintenance_policy` - Behaviour on maintenance events. The default is `unspecified`. Values: `unspecified`, `migrate`, `restart`.
* `maintenance_grace_period` - Time between notification via metadata service and maintenance. E.g., `60s`.
* `hardware_generation` - Instance hardware generation and its features. The structure is documented below.

---

The `boot_disk` block supports:

* `auto_delete` - Whether the disk is auto-deleted when the instance is deleted. The default value is false.
* `device_name` - Name that can be used to access an attached disk under `/dev/disk/by-id/`.
* `mode` - Access to the disk resource. By default a disk is attached in `READ_WRITE` mode.
* `disk_id` - ID of the attached disk.
* `initialize_params` - Parameters used for creating a disk alongside the instance. The structure is documented below.

The `initialize_params` block supports:

* `name` - Name of the boot disk.
* `description` - Description of the boot disk.
* `size` - Size of the disk in GB.
* `block_size` - The block size of the disk in bytes.
* `type` - Disk type.
* `image_id` - A disk image to initialize this disk from.
* `snapshot_id` - A snapshot to initialize this disk from.

The `network_interface` block supports:

* `index` - The index of the network interface, generated by the server.
* `mac_address` - MAC address that is assigned to the network interface.
* `ipv4` - Show if IPv4 address is assigned to the network interface.
* `ip_address` - The assignd private IP address to the network interface.
* `subnet_id` - ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.
* `nat` - Assigned for the instance's public address that is used to access the internet over NAT.
* `nat_ip_address` - Public IP address of the instance.
* `nat_ip_version` - IP version for the public address.
* `security_group_ids` - Security group ids for network interface.
* `dns_record` - List of configurations for creating ipv4 DNS records. The structure is documented below.
* `ipv6_dns_record` - List of configurations for creating ipv6 DNS records. The structure is documented below.
* `nat_dns_record` - List of configurations for creating ipv4 NAT DNS records. The structure is documented below.

The `dns_record` block supports:

* `fqdn` - DNS record FQDN.
* `dns_zone_id` - DNS zone ID (if not set, private zone is used).
* `ttl` - DNS record TTL. in seconds
* `ptr` - When set to true, also create a PTR DNS record.

The `ipv6_dns_record` block supports:

* `fqdn` - DNS record FQDN.
* `dns_zone_id` - DNS zone ID (if not set, private zone is used).
* `ttl` - DNS record TTL. in seconds
* `ptr` - When set to true, also create a PTR DNS record.

The `nat_dns_record` block supports:

* `fqdn` - DNS record FQDN.
* `dns_zone_id` - DNS zone ID (if not set, private zone is used).
* `ttl` - DNS record TTL. in seconds
* `ptr` - When set to true, also create a TR DNS record.

The `secondary_disk` block supports:

* `auto_delete` - Specifies whether the disk is auto-deleted when the instance is deleted.
* `device_name` - This value can be used to reference the device from within the instance for mounting, resizing, and so on.
* `mode` - Access to the Disk resource. By default, a disk is attached in `READ_WRITE` mode.
* `disk_id` - ID of the disk that is attached to the instance.

The `scheduling_policy` block supports:

* `preemptible` - (Optional) Specifies if the instance is preemptible. Defaults to false.

The `placement_policy` block supports:

* `placement_group_id` - Specifies the id of the Placement Group to assign to the instance.
* `placement_group_partition` - Specifies the number of partition in the Placement Group with the partition placement strategy.
* `host_affinity_rules` - List of host affinity rules. The structure is documented below.

The `host_affinity_rules` block supports:

* `key` - Affinity label or one of reserved values - `yc.hostId`, `yc.hostGroupId`.
* `op` - Affinity action. The only value supported is `IN`.
* `values` - List of values (host IDs or host group IDs).

The `local_disk` block supports:

* `size_bytes` - Size of the disk, specified in bytes.
* `device_name` - Name of the device.

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.