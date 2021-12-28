---
layout: "yandex"
page_title: "Yandex: yandex_kubernetes_node_group"
sidebar_current: "docs-yandex-datasource-kubernetes-node-group"
description: |-
  Get information about a Yandex Kubernetes Node Group.
---

# yandex\_kubernetes\_node\_group

Get information about a Yandex Kubernetes Node Group. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-kubernetes/concepts/#node-group).

## Example Usage

```hcl
data "yandex_kubernetes_node_group" "my_node_group" {
  node_group_id = "some_k8s_node_group_id"
}

output "my_node_group.status" {
  value = "${data.yandex_kubernetes_node_group.my_node_group.status}"
}
```

## Argument Reference

The following arguments are supported:

* `node_group_id` - (Optional) ID of a specific Kubernetes node group.

* `name` - (Optional) Name of a specific Kubernetes node group.

~> **NOTE:** One of `node_group_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

* `cluster_id` - The ID of the Kubernetes cluster that this node group belongs to.
* `description` - A description of the Kubernetes node group.
* `labels` - A set of key/value label pairs assigned to the Kubernetes node group.
* `created_at` - The Kubernetes node group creation timestamp.
* `status` - Status of the Kubernetes node group.

* `instance_template` - Template used to create compute instances in this Kubernetes node group. The structure is documented below.

* `scale_policy` - Scale policy of the node group. The structure is documented below.

* `allocation_policy` - This argument specify subnets (zones), that will be used by node group compute instances. The structure is documented below.

* `instance_group_id` - ID of instance group that is used to manage this Kubernetes node group.

* `maintenance_policy` - Information about maintenance policy for this Kubernetes node group. The structure is documented below.

* `node_labels` - A set of key/value label pairs, that are assigned to all the nodes of this Kubernetes node group.

* `node_taints` - A list of Kubernetes taints, that are applied to all the nodes of this Kubernetes node group.

* `allowed_unsafe_sysctls` - A list of allowed unsafe sysctl parameters for this node group. For more details see [documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/).

* `version_info` - Information about Kubernetes node group version. The structure is documented below.

* `deploy_policy` - Deploy policy of the node group. The structure is documented below.

---

The `instance_template` block supports:

* `platform_id` - The ID of the hardware platform configuration for the instance.
* `nat` - Boolean flag, when true, NAT for node group instances is enabled.
* `metadata` - The set of metadata `key:value` pairs assigned to this instance template. This includes custom metadata and predefined keys.
* `labels` - A map of labels applied to this instance.
* `resources.0.memory` - The memory size allocated to the instance.
* `resources.0.cores` - Number of CPU cores allocated to the instance.
* `resources.0.core_fraction` - Baseline core performance as a percent.
* `resources.0.gpus` - Number of GPU cores allocated to the instance.

* `network_interface` - An array with the network interfaces that will be attached to the instance. The structure is documented below.
* `network_acceleration_type` - Type of network acceleration. Values: `standard`, `software_accelerated`.

* `boot_disk` - The specifications for boot disks that will be attached to the instance. The structure is documented below.

* `scheduling_policy` - The scheduling policy for the instances in node group. The structure is documented below.
* `placement_policy` - (Optional) The placement policy configuration. The structure is documented below.

* `container_runtime` - Container runtime configuration. The structure is documented below.
---

The `network_interface` block supports:

* `subnet_ids` - The IDs of the subnets.
* `nat` - A public address that can be used to access the internet over NAT.
* `security_group_ids` - Security group ids for network interface.
* `ipv4` - Indicates whether the IPv4 address has been assigned.
* `ipv6` - Indicates whether the IPv6 address has been assigned.

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

The `container_runtime` block supports:

* `type` - Type of container runtime. Values: `docker`, `containerd`.

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

* `auto_upgrade` - Boolean flag.
* `auto_repair` - Boolean flag.
* `maintenance_window` - Set of day intervals, when maintenance is allowed for this node group.
When omitted, it defaults to any time.

Weekly maintenance policy expands to one element, with only two fields set: `start_time`, `duration`, and `day` field omitted.

Daily maintenance policy expands to list of elements, with all fields set, that specify time interval for selected days.
Only one interval is possible for any week day. Some days can be omitted, when there is no allowed interval for
maintenance specified.

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

---

