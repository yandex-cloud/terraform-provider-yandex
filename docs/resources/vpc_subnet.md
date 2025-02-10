---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_subnet"
description: |-
  A VPC network is a virtual version of the traditional physical networks that exist within and between physical data centers.
---

# yandex_vpc_subnet (Resource)

Manages a subnet within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/network#subnet).

* How-to Guides
  * [Cloud Networking](https://yandex.cloud/docs/vpc/)
  * [VPC Addressing](https://yandex.cloud/docs/vpc/concepts/address)

## Example usage

```terraform
//
// Create a new VPC Subnet.
//
resource "yandex_vpc_subnet" "my_subnet" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.lab-net.id
}

resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}
```

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network this subnet belongs to. Only networks that are in the distributed mode can have subnets.

* `v4_cidr_blocks` - (Required) A list of blocks of internal IPv4 addresses that are owned by this subnet. Provide this property when you create the subnet. For example, 10.0.0.0/22 or 192.168.0.0/16. Blocks of addresses must be unique and non-overlapping within a network. Minimum subnet size is /28, and maximum subnet size is /16. Only IPv4 is supported.

* `zone` - (Required) Name of the Yandex Cloud zone for this subnet.

---

* `name` - (Optional) Name of the subnet. Provided by the client when the subnet is created.

* `description` - (Optional) An optional description of the subnet. Provide this property when you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this subnet. A list of key/value pairs.

* `route_table_id` - (Optional) The ID of the route table to assign to this subnet. Assigned route table should belong to the same network as this subnet.

* `dhcp_options` - (Optional) Options for DHCP client. The structure is documented below.

---

The `dhcp_options` block supports:

* `domain_name` - (Optional) Domain name.
* `domain_name_servers` - (Optional) Domain name server IP addresses.
* `ntp_servers` - (Optional) NTP server IP addresses.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `v6_cidr_blocks` - An optional list of blocks of IPv6 addresses that are owned by this subnet.

## Timeouts

This resource provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 3 minute.
- `update` - Default is 3 minute.
- `delete` - Default is 3 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_vpc_subnet.<resource Name> <resource Id>
terraform import yandex_vpc_subnet.my_subnet ...
```
