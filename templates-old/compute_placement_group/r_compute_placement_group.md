---
subcategory: "Compute Cloud"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Placement group resource.
---

# {{.Name}} ({{.Type}})

A Placement group resource. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/placement-groups).

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

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/compute_placement_group/import.sh" }}
