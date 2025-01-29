---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Placement group resource.
---

# {{.Name}} ({{.Type}})

A Placement group resource. For more information, see [the official documentation](https://cloud.yandex.com/docs/compute/concepts/placement-groups).

## Example usage

{{ tffile "examples/compute_placement_group/r_compute_placement_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `name` - (Optional) The name of the Placement Group.

* `description` - (Optional) A description of the Placement Group.

* `labels` - (Optional) A set of key/value label pairs to assign to the Placement Group.

* `placement_strategy_spread` - A placement strategy with spread policy of the Placement Group. Should be true or unset (conflicts with placement_strategy_partitions).

* `placement_strategy_partitions` - A number of partitions in the placement strategy with partitions policy of the Placement Group (conflicts with placement_strategy_spread).

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

A Placement Group can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_placement_group.default placement_group_id
```
