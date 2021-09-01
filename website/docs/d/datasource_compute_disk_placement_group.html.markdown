---
layout: "yandex"
page_title: "Yandex: yandex_compute_disk_placement_group"
sidebar_current: "docs-yandex-datasource-compute-disk_placement-group"
description: |-
  Get information about a Yandex Compute Disk Placement Group.
---

# yandex\_compute\_disk_\placement_group

Get information about a Yandex Compute Disk Placement group. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/disk#nr-disks).

## Example Usage

```hcl
data "yandex_compute_disk_placement_group" "my_group" {
  group_id = "some_group_id"
}

output "placement_group_name" {
  value = "${data.yandex_compute_disk_placement_group.my_group.name}"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The ID of a specific group.
* `name` - (Optional) Name of the group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> **NOTE:** One of `group_id` or `name` should be specified.

## Attributes Reference

* `created_at` - The creation timestamp of the Disk Placement Group.
* `description` - Description of the Disk Placement Group.
* `labels` - A set of key/value label pairs assigned to the Disk Placement Group.
* `zone` - ID of the zone where the Disk Placement Group resides.
* `status` - Status of the Disk Placement Group.
