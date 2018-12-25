---
layout: "yandex"
page_title: "Yandex: yandex_compute_subnet"
sidebar_current: "docs-yandex-vpc-subnet"
description: |-
  A VPC network is a virtual version of the traditional physical networks that exist within and between physical data centers.
---

# yandex\_vpc\_subnet

Manages a subnet within the Yandex Cloud. For more information see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts/network#subnet).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)
    * [VPC Addressing](https://cloud.yandex.com/docs/vpc/concepts/address)

## Example Usage

```hcl
resource "yandex_compute_network" "custom-net" {
  name = "test-network"
}

resource "yandex_compute_subnet" "custom-subnet" {
  v4_cidr_blocks = "10.2.0.0/16"
  zone           = "ru-central1-a"
  network        = "${yandex_compute_network.custom-net.id}"
}
```

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network this subnet belongs to.
  Only networks that are in the distributed mode can have subnets.

* `zone` - (Required) Name of the Yandex Cloud zone for this subnet.

- - -

* `v4_cidr_blocks` - (Optional)  The range of internal addresses that are owned by this subnet.
  Provide this property when you create the subnet. For example,
  10.0.0.0/22 or 192.168.0.0/16. Ranges must be unique and non-overlapping within a network. Minimum subnet size is /28, and maximum subnet size is /16. Only IPv4 is supported.

* `v6_cidr_blocks` - (Optional)  The range of internal IPv6 addresses that are owned by this subnet.

* `name` - (Optional) Name of the subnet. Provided by the client when the subnet is created.

* `description` - (Optional) An optional description of the subnet. Provide this property when
  you create the resource.

* `folder_id` - (Optional) The ID of the folder in which the resource belongs.
    If it is not provided, the provider folder is used.

* `labels` - (Optional) Labels to assign to this subnet. A list of key/value pairs.

~> **Note:** `v6_cidr_blocks` attribute is currently not supported. It will be available in the future.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `region_id` - ID of the availability zone where the subnet resides.

## Timeouts

This resource provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

Subnetwork can be imported using any of these accepted formats:

```
$ terraform import yandex_compute_subnet.default subnet_id
```
