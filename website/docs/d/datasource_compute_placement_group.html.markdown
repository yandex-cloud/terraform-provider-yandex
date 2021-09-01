---
layout: "yandex"
page_title: "Yandex: yandex_compute_placement_group"
sidebar_current: "docs-yandex-datasource-compute-placement-group"
description: |-
  Get information about a Yandex Compute Placement Group.
---

# yandex\_compute\_placement_group

Get information about a Yandex Compute Placement group. For more information, see
[the official documentation](https://cloud.yandex.com/docs/compute/concepts/placement-groups).

## Example Usage

```hcl
data "yandex_compute_placement_group" "my_group" {
  group_id = "some_group_id"
}

output "placement_group_name" {
  value = "${data.yandex_compute_placement_group.my_group.name}"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The ID of a specific group.
* `name` - (Optional) Name of the group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> **NOTE:** One of `group_id` or `name` should be specified.

## Attributes Reference

* `created_at` - Placement group creation timestamp.
* `description` - Description of the group.
* `labels` - A set of key/value label pairs assigned to the group.
