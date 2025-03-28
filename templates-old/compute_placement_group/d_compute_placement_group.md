---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Compute Placement Group.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Compute Placement group. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/placement-groups).

## Example usage

{{ tffile "examples/compute_placement_group/d_compute_placement_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The ID of a specific group.
* `name` - (Optional) Name of the group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `placement_strategy_spread` - placement strategy with spread policy
* `placement_strategy_partitions` - placement strategy with partitions policy

~> One of `group_id` or `name` should be specified.

## Attributes Reference

* `created_at` - Placement group creation timestamp.
* `description` - Description of the group.
* `labels` - A set of key/value label pairs assigned to the group.
* `placement_strategy` - Placement strategy set for group.
