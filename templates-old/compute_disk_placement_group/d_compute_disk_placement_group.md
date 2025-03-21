---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute Disk Placement Group.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute Disk Placement group. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/disk#nr-disks).

## Example usage

{{ tffile "examples/compute_disk_placement_group/d_compute_disk_placement_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The ID of a specific group.
* `name` - (Optional) Name of the group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> One of `group_id` or `name` should be specified.

## Attributes Reference

* `created_at` - The creation timestamp of the Disk Placement Group.
* `description` - Description of the Disk Placement Group.
* `labels` - A set of key/value label pairs assigned to the Disk Placement Group.
* `zone` - ID of the zone where the Disk Placement Group resides.
* `status` - Status of the Disk Placement Group.
