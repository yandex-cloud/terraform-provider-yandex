---
layout: "yandex"
page_title: "Yandex: yandex_compute_disk"
sidebar_current: "docs-yandex-compute-disk"
description: |-
  Persistent disks are durable storage devices that function similarly to
  the physical disks in a desktop or a server.
---

# yandex\_compute\_disk

Persistent disks are used for data storage and function similarly to physical hard and solid state drives.

A disk can be attached or detached from the virtual machine and can be located locally. A disk can be moved between virtual machines within the same availability zone. Each disk can be attached to only one virtual machine at a time.

To get more information about Disk, see:

* [Documentation](https://cloud.yandex.com/docs/compute/concepts/disk)
* How-to Guides
    * [Attach and detach a disk](https://cloud.yandex.com/docs/compute/concepts/disk#attach-detach)
    * [Backup operation](https://cloud.yandex.com/docs/compute/concepts/disk#backup)

## Example Usage

```hcl
resource "yandex_compute_disk" "default" {
  name     = "disk"
  type     = "network-nvme"
  zone     = "ru-central1-a"
  image_id = "ubuntu-16.04-v20180727"
  labels {
    environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:


* `name` -
  (Optional)
  Name of the disk. Provide this property when
  you create a resource.

* `description` -
  (Optional) Description of the disk. Provide this property when
  you create the resource.

* `labels` -
  (Optional)
  Labels to assign to this disk. A list of key/value pairs.

* `size` -
  (Optional)
  Size of the persistent disk, specified in GB. You can specify this
  field when creating a persistent disk using the `image_id` or `snapshot_id`
  parameter, or specify it individually to create an empty persistent disk.
  If you specify this field along with `image_id` or `snapshot_id`,
  the value of size must not be less than the size of the source image
  or the size of the snapshot.

* `type` - (Optional) Type of the disk that is being created. Provide this when creating the disk.

* `image_id` -  (Optional) The source image to use for disk creation.

* `snapshot_id` - (Optional) The source snapshot to use for disk creation.

~> **NOTE:** Either `image_id` or `snapshot_id` must be specified.

* `zone` -
  (Optional)
  Availability zone where the disk will reside.

* `folder_id` - (Optional) The ID of the folder that the disk belongs to.
    If it is not provided, the default provider folder is used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `source_image_id` -
  The ID value of the image that is used to create this disk. This value
  identifies the exact image that was used to create this persistent
  disk. For example, if you created the persistent disk from an image
  that was later deleted and recreated under the same name, the source
  image ID would identify the exact version of the image that was used.

* `source_snapshot_id` -
  The unique ID of the snapshot that is used to create this disk. This value
  identifies the exact snapshot that was used to create this persistent
  disk. For example, if you created the persistent disk from a snapshot
  that was later deleted and recreated under the same name, the source
  snapshot ID would identify the exact version of the snapshot that was
  used.

## Timeouts

This resource provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

A disk can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_disk.default {{id}}
```
