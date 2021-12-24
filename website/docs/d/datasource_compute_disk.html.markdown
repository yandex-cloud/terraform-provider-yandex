---
layout: "yandex"
page_title: "Yandex: yandex_compute_disk"
sidebar_current: "docs-yandex-datasource-compute-disk"
description: |-
  Get information about a Yandex Compute disk.
---

# yandex\_compute\_disk

Get information about a Yandex Compute disk. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/disk).

## Example Usage

```hcl
data "yandex_compute_disk" "my_disk" {
  disk_id = "some_disk_id"
}

resource "yandex_compute_instance" "default" {
  ...

  secondary_disk {
    disk_id = "${data.yandex_compute_disk.my_disk.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Optional) The ID of a specific disk.

* `name` - (Optional) Name of the disk.

~> **NOTE:** One of `disk_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `description` - Optional description of this disk.
* `folder_id` - ID of the folder that the disk belongs to.
* `zone` - ID of the zone where the disk resides.
* `size` - Size of the disk, specified in Gb.
* `block_size` - The block size of the disk in bytes.
* `image_id` - ID of the source image that was used to create this disk.
* `snapshot_id` - Source snapshot that was used to create this disk. 
* `type` - Type of the disk.
* `status` - Status of the disk.
* `labels` -  Map of labels applied to this disk.
* `product_ids` - License IDs that indicate which licenses are attached to this disk.
* `instance_ids` - IDs of instances to which this disk is attached.  
* `created_at` - Disk creation timestamp.
