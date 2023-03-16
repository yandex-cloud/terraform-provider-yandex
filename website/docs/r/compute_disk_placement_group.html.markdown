---
layout: "yandex"
page_title: "Yandex: yandex_compute_disk_placement_group"
sidebar_current: "docs-yandex-compute-disk_placement-group"
description: |-
  Manages a Disk Placement Group resource.
---

# yandex\_compute\_disk\_placement\_group

A Disk Placement Group resource. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/disk#nr-disks).

## Example Usage

```hcl
resource "yandex_compute_disk_placement_group" "group1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `name` - (Optional) The name of the Disk Placement Group.

* `description` - (Optional) A description of the Disk Placement Group.

* `labels` - (Optional) A set of key/value label pairs to assign to the Disk Placement Group.

* `zone` - ID of the zone where the Disk Placement Group resides.

* `status` - Status of the Disk Placement Group.

## Timeouts

This resource provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.
