---
layout: "yandex"
page_title: "Yandex: yandex_vpc_route_table"
sidebar_current: "docs-yandex-vpc-route-table"
description: |-
  A VPC route table is a virtual version of the traditional route table on router device.
---

# yandex\_vpc\_route\_table

Manages a route table within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)

## Example Usage

```hcl
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_gateway" "egress-gateway" {
  name = "egress-gateway"
  shared_egress_gateway {}
}

resource "yandex_vpc_route_table" "lab-rt-a" {
  network_id = "${yandex_vpc_network.lab-net.id}"

  static_route {
    destination_prefix = "10.2.0.0/16"
    next_hop_address   = "172.16.10.10"
  }

  static_route {
    destination_prefix = "0.0.0.0/0"
    gateway_id         = "${yandex_vpc_gateway.egress-gateway.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network this route table belongs to.

- - -

* `name` - (Optional) Name of the route table. Provided by the client when the route table is created.

* `description` - (Optional) An optional description of the route table. Provide this property when
  you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs.
    If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this route table. A list of key/value pairs.

* `static_route` - (Optional) A list of static route records for the route table. The structure is documented below.

The `static_route` block supports:

* `destination_prefix` - Route prefix in CIDR notation.

* `next_hop_address` - Address of the next hop.

* `gateway_id` - ID of the gateway used ad next hop.

~> **NOTE:** Only one of `next_hop_address` or `gateway_id` should be specified.

## Attributes Reference

* `created_at` - Creation timestamp of the route table.

## Import

A route table can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_vpc_route_table.default route_table_id
```
