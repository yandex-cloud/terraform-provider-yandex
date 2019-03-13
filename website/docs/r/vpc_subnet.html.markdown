---
layout: "yandex"
page_title: "Yandex: yandex_vpc_subnet"
sidebar_current: "docs-yandex-vpc-subnet"
description: |-
  A VPC network is a virtual version of the traditional physical networks that exist within and between physical data centers.
---

# yandex\_vpc\_subnet

Manages a subnet within the Yandex Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts/network#subnet).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)
    * [VPC Addressing](https://cloud.yandex.com/docs/vpc/concepts/address)

## Example Usage

```hcl
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_subnet" "lab-subnet-a" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.lab-net.id}"
}
```

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network this subnet belongs to.
  Only networks that are in the distributed mode can have subnets.

* `v4_cidr_blocks` - (Required) A list of blocks of internal IPv4 addresses that are owned by this subnet.
  Provide this property when you create the subnet. For example, 10.0.0.0/22 or 192.168.0.0/16. 
  Blocks of addresses must be unique and non-overlapping within a network. 
  Minimum subnet size is /28, and maximum subnet size is /16. Only IPv4 is supported.

* `zone` - (Required) Name of the Yandex Cloud zone for this subnet.

- - -

* `name` - (Optional) Name of the subnet. Provided by the client when the subnet is created.

* `description` - (Optional) An optional description of the subnet. Provide this property when
  you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs.
    If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this subnet. A list of key/value pairs.

* `v6_cidr_blocks` - (Optional) An optional list of blocks of IPv6 addresses that are owned by this subnet.

~> **Note:** The `v6_cidr_blocks` attribute is currently not supported. It will be available in the future.

## Timeouts

This resource provides the following configuration options for 
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

A subnet can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_vpc_subnet.default subnet_id
```
