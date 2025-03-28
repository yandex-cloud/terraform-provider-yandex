---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute disk.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute disk. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/disk).

## Example usage

{{ tffile "examples/compute_disk/d_compute_disk_1.tf" }}

## Argument Reference

The following arguments are supported:

* `disk_id` - (Optional) The ID of a specific disk.

* `name` - (Optional) Name of the disk.

~> One of `disk_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - Optional description of this disk.
* `folder_id` - ID of the folder that the disk belongs to.
* `zone` - ID of the zone where the disk resides.
* `size` - Size of the disk, specified in Gb.
* `block_size` - The block size of the disk in bytes.
* `image_id` - ID of the source image that was used to create this disk.
* `snapshot_id` - Source snapshot that was used to create this disk.
* `type` - Type of the disk.
* `status` - Status of the disk.
* `labels` - Map of labels applied to this disk.
* `product_ids` - License IDs that indicate which licenses are attached to this disk.
* `instance_ids` - IDs of instances to which this disk is attached.
* `created_at` - Disk creation timestamp.
* `kms_key_id` - ID of KMS symmetric key used to encrypt disk.
