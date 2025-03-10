---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex VPC route table.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex VPC route table. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts).

## Example usage

{{ tffile "examples/vpc_route_table/d_vpc_route_table_1.tf" }}

This data source is used to define [VPC Route Table](https://yandex.cloud/docs/vpc/concepts/) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `route_table_id` (Optional) - Route table ID.
* `name` - (Optional) - Name of the route table.

~> One of `route_table_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the route table.
* `network_id` - ID of the network this route table belongs to.
* `labels` - Labels to assign to this route table.
* `static_route` - List of static route records of the route table. Structure is documented below.
* `created_at` - Creation timestamp of this route table.

The `static_route` block supports:

* `destination_prefix` - Route prefix in CIDR notation.
* `next_hop_address` - Address of the next hop.
* `gateway_id` - ID of the gateway used as next hop.

~> Only one of `next_hop_address` or `gateway_id` should be specified.
