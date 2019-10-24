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
  version     = "1.14"

  labels = {
    "key" = "value"
  }

  instance_template {
    platform_id = "standard-v2"
    nat         = true

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
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the Kubernetes cluster that this node group belongs to.
* `name` - (Optional) Name of a specific Kubernetes node group.
* `description` - (Optional) A description of the Kubernetes node group.
* `labels` - (Optional) A set of key/value label pairs assigned to the Kubernetes node group.
* `version` - (Optional) Version of Kubernetes that will be used for Kubernetes node group.
* `instance_template` - (Required) Template used to create compute instances in this Kubernetes node group.

The structure is documented below.

* `scale_policy` - (Required) Scale policy of the node group.
 
The structure is documented below.

* `allocation_policy` - This argument specify subnets (zones), that will be used by node group compute instances.

The structure is documented below.

* `instance_group_id` - (Computed) ID of instance group that is used to manage this Kubernetes node group.

* `maintenance_policy` - (Computed) Information about maintenance policy for this Kubernetes node group.

The structure is documented below.

* `version_info` - (Computed) Information about Kubernetes node group version.

The structure is documented below.

---

The `instance_template` block supports:

* `platform_id` - The ID of the hardware platform configuration for the node group compute instances.
* `nat` - Boolean flag, enables NAT for node group compute instances.
* `metadata` - The set of metadata `key:value` pairs assigned to this instance template. This includes custom metadata and predefined keys.
* `resources.0.memory` - The memory size allocated to the instance.
* `resources.0.cores` - Number of CPU cores allocated to the instance.
* `resources.0.core_fraction` - Baseline core performance as a percent.

* `boot_disk` - The specifications for boot disks that will be attached to the instance.

The structure is documented below.

* `scheduling_policy` - The scheduling policy for the instances in node group.

The structure is documented below.

* `status` - (Computed) Status of the Kubernetes node group.
* `created_at` - (Computed) The Kubernetes node group creation timestamp.

---

The `boot_disk` block supports:

* `size` - The size of the disk in GB. Allowed minimal size: 64 GB.
* `type` - The disk type.

---

The `scheduling_policy` block supports:

* `preemptible` - Specifies if the instance is preemptible. Defaults to false.

---

The `scale_policy` block supports:

* `fixed_scale` - The fixed scaling policy of the instance group.

The structure is documented below.

---

The `fixed_scale` block supports:

* `size` - The number of instances in the node group.

---

The `allocation_policy` block supports:

* `location` - Repeated field, that specify subnets (zones), that will be used by node group compute instances.

The structure is documented below.   

---

The `location` block supports:

* `zone` - ID of the availability zone where for one compute instance in node group.
* `subnet_id` - ID of the subnet, that will be used by one compute instance in node group.

Subnet specified by `subnet_id` should be allocated in zone specified by 'zone' argument 

---

The `maintenance_policy` block supports:

* `auto_upgrade` - Boolean flag.
* `auto_repair`- Boolean flag.

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

## Import

A Yandex Kubernetes Node Group can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_kubernetes_node_group.default node_group_id
```
