---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_route_table"
description: |-
  A VPC route table is a virtual version of the traditional route table on router device.
---

# yandex_vpc_route_table (Resource)

Manages a route table within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)

## Example usage

```terraform
//
// Create a new VPC Route Table.
//
resource "yandex_vpc_route_table" "lab-rt-a" {
  network_id = yandex_vpc_network.lab-net.id

  static_route {
    destination_prefix = "10.2.0.0/16"
    next_hop_address   = "172.16.10.10"
  }

  static_route {
    destination_prefix = "0.0.0.0/0"
    gateway_id         = yandex_vpc_gateway.egress-gateway.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_gateway" "egress-gateway" {
  name = "egress-gateway"
  shared_egress_gateway {}
}
```

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

```shell
# terraform import yandex_vpc_route_table.<resource Name> <resource Id>
terraform import yandex_vpc_route_table.lab-rt-a ...
```
