---
layout: "yandex"
page_title: "Yandex: yandex_vpc_network"
sidebar_current: "docs-yandex-vpc-network"
description: |-
  Manages a network within Yandex.Cloud.
---

# yandex\_vpc\_network

Manages a network within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts/network#network).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)
    * [VPC Addressing](https://cloud.yandex.com/docs/vpc/concepts/address)

## Example Usage

```hcl
resource "yandex_vpc_network" "default" {
  name = "foobar"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the network. Provided by the client when the network is created.

* `description` - (Optional) An optional description of this resource. Provide this property when
  you create the resource.

* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to apply to this network. A list of key/value pairs.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

* `default_security_group_id` - ID of default Security Group of this network.

## Import

A network can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_vpc_network.default network_id
```
