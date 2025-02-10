---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  A VPC route table is a virtual version of the traditional route table on router device.
---

# {{.Name}} ({{.Type}})

Manages a route table within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)

## Example usage

{{ tffile "examples/vpc_route_table/r_vpc_route_table_1.tf" }}

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network this route table belongs to.

---

* `name` - (Optional) Name of the route table. Provided by the client when the route table is created.

* `description` - (Optional) An optional description of the route table. Provide this property when you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this route table. A list of key/value pairs.

* `static_route` - (Optional) A list of static route records for the route table. The structure is documented below.

The `static_route` block supports:

* `destination_prefix` - Route prefix in CIDR notation.

* `next_hop_address` - Address of the next hop.

* `gateway_id` - ID of the gateway used ad next hop.

~> Only one of `next_hop_address` or `gateway_id` should be specified.

## Attributes Reference

* `created_at` - Creation timestamp of the route table.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/vpc_route_table/import.sh" }}
