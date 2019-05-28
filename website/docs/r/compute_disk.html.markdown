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

For more information about disks in Yandex.Cloud, see:

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

  labels = {
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
  you create a resource.

* `folder_id` - 
  (Optional) The ID of the folder that the disk belongs to.
  If it is not provided, the default provider folder is used.

* `labels` -
  (Optional)
  Labels to assign to this disk. A list of key/value pairs.

* `zone` -
  (Optional)
  Availability zone where the disk will reside.

* `size` -
  (Optional)
  Size of the persistent disk, specified in GB. You can specify this
  field when creating a persistent disk using the `image_id` or `snapshot_id`
  parameter, or specify it alone to create an empty persistent disk.
  If you specify this field along with `image_id` or `snapshot_id`,
  the size value must not be less than the size of the source image
  or the size of the snapshot.

* `type` - (Optional) Type of disk to create. Provide this when creating a disk. 
  One of `network-hdd` (default) or `network-nvme`.

* `image_id` -  (Optional) The source image to use for disk creation.

* `snapshot_id` - (Optional) The source snapshot to use for disk creation.

~> **NOTE:** Only one of `image_id` or `snapshot_id` can be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:
  
* `status` - The status of the disk.  
* `created_at` - Creation timestamp of the disk.

## Timeouts

This resource provides the following configuration options for
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

A disk can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_disk.default disk_id
```
