---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute Snapshot.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute snapshot. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/snapshot).

## Example usage

{{ tffile "examples/compute_snapshot/d_compute_snapshot_1.tf" }}

## Argument Reference

The following arguments are supported:

* `snapshot_id` - (Optional) The ID of a specific snapshot.

* `name` - (Optional) The name of the snapshot.

~> One of `snapshot_id` or `name` should be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `description` - An optional description of this snapshot.
* `folder_id` - ID of the folder that the snapshot belongs to.
* `storage_size` - The size of the snapshot, specified in Gb.
* `status` - The status of the snapshot.
* `disk_size` - Minimum required size of the disk which is created from this snapshot.
* `source_disk_id` - ID of the source disk.
* `labels` - A map of labels applied to this snapshot.
* `product_ids` - License IDs that indicate which licenses are attached to this snapshot.
* `created_at` - Snapshot creation timestamp.
* `kms_key_id` - ID of KMS symmetric key used to encrypt snapshot.
