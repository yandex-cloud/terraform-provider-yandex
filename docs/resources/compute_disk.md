---
subcategory: "Compute"
page_title: "Yandex: yandex_compute_disk"
description: |-
  Persistent disks are durable storage devices that function similarly to
  the physical disks in a desktop or a server.
---


# yandex_compute_disk




Persistent disks are used for data storage and function similarly to physical hard and solid state drives.

A disk can be attached or detached from the virtual machine and can be located locally. A disk can be moved between virtual machines within the same availability zone. Each disk can be attached to only one virtual machine at a time.

For more information about disks in Yandex.Cloud, see:

* [Documentation](https://cloud.yandex.com/docs/compute/concepts/disk)
* How-to Guides
  * [Attach and detach a disk](https://cloud.yandex.com/docs/compute/concepts/disk#attach-detach)
  * [Backup operation](https://cloud.yandex.com/docs/compute/concepts/disk#backup)

```terraform
resource "yandex_compute_snapshot_schedule" "schedule1" {
  schedule_policy {
    expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
    description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}

resource "yandex_compute_snapshot_schedule_iam_binding" "editor" {
  snapshot_schedule_id = data.yandex_compute_snapshot_schedule.schedule1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

```terraform
resource "yandex_compute_disk" "nr" {
  name = "non-replicated-disk-name"
  size = 93 // NB size must be divisible by 93  
  type = "network-ssd-nonreplicated"
  zone = "ru-central1-b"

  disk_placement_policy {
    disk_placement_group_id = yandex_compute_disk_placement_group.this.id
  }
}

resource "yandex_compute_disk_placement_group" "this" {
  zone = "ru-central1-b"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the disk. Provide this property when you create a resource.

* `description` - (Optional) Description of the disk. Provide this property when you create a resource.

* `folder_id` - (Optional) The ID of the folder that the disk belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to assign to this disk. A list of key/value pairs.

* `zone` - (Optional) Availability zone where the disk will reside.

* `size` - (Optional) Size of the persistent disk, specified in GB. You can specify this field when creating a persistent disk using the `image_id` or `snapshot_id` parameter, or specify it alone to create an empty persistent disk. If you specify this field along with `image_id` or `snapshot_id`, the size value must not be less than the size of the source image or the size of the snapshot.

* `block_size` - (Optional) Block size of the disk, specified in bytes.

* `type` - (Optional) Type of disk to create. Provide this when creating a disk.

* `disk_placement_policy` - (Optional) Disk placement policy configuration. The structure is documented below.

* `image_id` - (Optional) The source image to use for disk creation.

* `snapshot_id` - (Optional) The source snapshot to use for disk creation.

* `hardware_generation` - (Optional) Hardware generation and its features,
  which will be applied to the instance when this disk is used as a boot
  disk. Provide this property if you wish to override this value, which
  otherwise is inherited from the source. The structure is documented below.

The `disk_placement_policy` block supports:

* `disk_placement_group_id` - (Required) Specifies Disk Placement Group id.

The `hardware_generation` consists of one of the following blocks:

* `legacy_features` - Defines the first known hardware generation and its features, which are:
  * `pci_topology` - A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.
* `generation2_features` - A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.

~> **NOTE:** Only one of `image_id` or `snapshot_id` can be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the disk.
* `created_at` - Creation timestamp of the disk.

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

A disk can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_disk.default disk_id
```
