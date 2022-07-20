---
layout: "yandex"
page_title: "Yandex: yandex_kubernetes_node_group"
sidebar_current: "docs-yandex-kubernetes-node-group"
description: |-
   Allows management of Yandex Kubernetes Node Group. For more information, see
   [the official documentation](https://cloud.yandex.com/docs/managed-kubernetes/concepts/#node-group).
---

# yandex\_kubernetes\_node\_group

Creates a Yandex Kubernetes Node Group.

## Example Usage

```hcl
resource "yandex_kubernetes_node_group" "my_node_group" {
  cluster_id  = "${yandex_kubernetes_cluster.my_cluster.id}"
  name        = "name"
  description = "description"
  version     = "1.17"

  labels = {
    "key" = "value"
  }

  instance_template {
    platform_id = "standard-v2"

    network_interface {
      nat                = true
      subnet_ids         = ["${yandex_vpc_subnet.my_subnet.id}"]
    }

    resources {
      memory = 2
      cores  = 2
    }

    boot_disk {
      type = "network-hdd"
      size = 64
    }

    scheduling_policy {
      preemptible = false
    }

    container_runtime {
      type = "containerd"
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    location {
      zone = "ru-central1-a"
    }
  }

  maintenance_policy {
    auto_upgrade = true
    auto_repair  = true

    maintenance_window {
      day        = "monday"
      start_time = "15:00"
      duration   = "3h"
    }

    maintenance_window {
      day        = "friday"
      start_time = "10:00"
      duration   = "4h30m"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kubernetes cluster that this node group belongs to.
* `name` - (Optional) Name of a specific Kubernetes node group.
* `description` - (Optional) A description of the Kubernetes node group.
* `labels` - (Optional) A set of key/value label pairs assigned to the Kubernetes node group.
* `version` - (Optional) Version of Kubernetes that will be used for Kubernetes node group.
* `instance_template` - (Required) Template used to create compute instances in this Kubernetes node group. The structure is documented below.

* `scale_policy` - (Required) Scale policy of the node group. The structure is documented below.

* `allocation_policy` - This argument specify subnets (zones), that will be used by node group compute instances. The structure is documented below.

* `instance_group_id` - (Computed) ID of instance group that is used to manage this Kubernetes node group.

* `maintenance_policy` - (Optional) (Computed) Maintenance policy for this Kubernetes node group.
If policy is omitted, automatic revision upgrades are enabled and could happen at any time.
Revision upgrades are performed only within the same minor version, e.g. 1.13.
Minor version upgrades (e.g. 1.13->1.14) should be performed manually. The structure is documented below.

* `node_labels` - (Optional, Forces new resource) A set of key/value label pairs, that are assigned to all the nodes of this Kubernetes node group.

* `node_taints` - (Optional, Forces new resource) A list of Kubernetes taints, that are applied to all the nodes of this Kubernetes node group.

* `allowed_unsafe_sysctls` - (Optional, Forces new resource) A list of allowed unsafe sysctl parameters for this node group. For more details see [documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/).

* `version_info` - (Computed) Information about Kubernetes node group version. The structure is documented below.

* `deploy_policy` - Deploy policy of the node group. The structure is documented below.

---

The `instance_template` block supports:

* `platform_id` - The ID of the hardware platform configuration for the node group compute instances.
* `nat` - Boolean flag, enables NAT for node group compute instances.
* `metadata` - The set of metadata `key:value` pairs assigned to this instance template. This includes custom metadata and predefined keys.
* `resources.0.memory` - The memory size allocated to the instance.
* `resources.0.cores` - Number of CPU cores allocated to the instance.
* `resources.0.core_fraction` - Baseline core performance as a percent.
* `resources.0.gpus` - Number of GPU cores allocated to the instance.

* `boot_disk` - The specifications for boot disks that will be attached to the instance. The structure is documented below.

* `scheduling_policy` - The scheduling policy for the instances in node group. The structure is documented below.
* `placement_policy` - (Optional) The placement policy configuration. The structure is documented below.

* `network_interface` - An array with the network interfaces that will be attached to the instance. The structure is documented below.
* `network_acceleration_type` - (Optional) Type of network acceleration. Values: `standard`, `software_accelerated`.

* `container_runtime` - (Optional) Container runtime configuration. The structure is documented below.

* `name` - (Optional) Name template of the instance.
In order to be unique it must contain at least one of instance unique placeholders:   
{instance.short_id}   
{instance.index}   
combination of {instance.zone_id} and {instance.index_in_zone}   
Example: my-instance-{instance.index}  
If not set, default is used: {instance_group.id}-{instance.short_id}   
It may also contain another placeholders, see [Compute Instance group metadata doc](https://cloud.yandex.com/en-ru/docs/compute/api-ref/grpc/instance_group_service) for full list.

* `labels` - (Optional) Labels that will be assigned to compute nodes (instances), created by the Node Group.
---

The `boot_disk` block supports:

* `size` - The size of the disk in GB. Allowed minimal size: 64 GB.
* `type` - The disk type.

---

The `scheduling_policy` block supports:

* `preemptible` - Specifies if the instance is preemptible. Defaults to false.
---

The `placement_policy` block supports:

* `placement_group_id` - (Optional) Specifies the id of the Placement Group to assign to the instances.

---

The `network_interface` block supports:

* `subnet_ids` - The IDs of the subnets.
* `ipv4` - (Optional) Allocate an IPv4 address for the interface. The default value is `true`.
* `ipv6` - (Optional) If true, allocate an IPv6 address for the interface. The address will be automatically assigned from the specified subnet.
* `nat` - A public address that can be used to access the internet over NAT.
* `security_group_ids` - (Optional) Security group ids for network interface.
* `ipv4_dns_records` - (Optional) List of configurations for creating ipv4 DNS records. The structure is documented below.
* `ipv6_dns_records` - (Optional) List of configurations for creating ipv6 DNS records. The structure is documented below.

---

The `ipv4_dns_records` block supports:

* `fqdn` - (Required) DNS record FQDN.
* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone is used).
* `ttl` - (Optional) DNS record TTL (in seconds).
* `ptr` - (Optional) When set to true, also create a PTR DNS record.

---

The `ipv6_dns_records` block supports:

* `fqdn` - (Required) DNS record FQDN.
* `dns_zone_id` - (Optional) DNS zone ID (if not set, private zone is used).
* `ttl` - (Optional) DNS record TTL (in seconds).
* `ptr` - (Optional) When set to true, also create a PTR DNS record.

---

The `container_runtime` block supports:

* `type` - (Required) Type of container runtime. Values: `docker`, `containerd`.

---

The `scale_policy` block supports:

* `fixed_scale` - Scale policy for a fixed scale node group. The structure is documented below.
* `auto_scale` - Scale policy for an autoscaled node group. The structure is documented below.

---

The `fixed_scale` block supports:

* `size` - The number of instances in the node group.

---

The `auto_scale` block supports:

* `min` - Minimum number of instances in the node group.
* `max` - Maximum number of instances in the node group.
* `initial` - Initial number of instances in the node group.

---

The `allocation_policy` block supports:

* `location` - Repeated field, that specify subnets (zones), that will be used by node group compute instances. The structure is documented below.   

---

The `location` block supports:

* `zone` - ID of the availability zone where for one compute instance in node group.
* `subnet_id` - ID of the subnet, that will be used by one compute instance in node group.

Subnet specified by `subnet_id` should be allocated in zone specified by 'zone' argument 

---

The `maintenance_policy` block supports:

* `auto_upgrade` - (Required) Boolean flag that specifies if node group can be upgraded automatically. When omitted, default value is TRUE.
* `auto_repair`- (Required) Boolean flag that specifies if node group can be repaired automatically. When omitted, default value is TRUE.
* `maintenance_window` - (Optional) (Computed) Set of day intervals, when maintenance is allowed for this node group. When omitted, it defaults to any time. 

To specify time of day interval, for all days, one element should be provided, with two fields set, `start_time` and `duration`.

To allow maintenance only on specific days of week, please provide list of elements, with all fields set. Only one 
time interval is allowed for each day of week. Please see `my_node_group` config example.

---

The `version_info` block supports:

* `current_version` - Current Kubernetes version, major.minor (e.g. 1.15).
* `new_revision_available` - True/false flag.
Newer revisions may include Kubernetes patches (e.g 1.15.1 -> 1.15.2) as well
as some internal component updates - new features or bug fixes in yandex-specific
components either on the master or nodes.

* `new_revision_summary` - Human readable description of the changes to be applied
when updating to the latest revision. Empty if new_revision_available is false.
* `version_deprecated` - True/false flag. The current version is on the deprecation schedule,
component (master or node group) should be upgraded.

---

The `deploy_policy` block supports:

* `max_expansion` - The maximum number of instances that can be temporarily allocated above the group's target size during the update.
* `max_unavailable` - The maximum number of running instances that can be taken offline during update.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - (Computed) Status of the Kubernetes node group.
* `created_at` - (Computed) The Kubernetes node group creation timestamp.

## Timeouts

This resource provides the following configuration options for 
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 60 minute.
- `update` - Default is 60 minute.
- `delete` - Default is 20 minute.

## Import

A Yandex Kubernetes Node Group can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_kubernetes_node_group.default node_group_id
```
