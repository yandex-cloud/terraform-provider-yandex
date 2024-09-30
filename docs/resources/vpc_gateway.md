---
subcategory: "VPC (Virtual Private Cloud)"
page_title: "Yandex: yandex_vpc_gateway"
description: |-
  Manages a gateway within Yandex.Cloud.
---


# yandex_vpc_gateway




Manages a gateway within the Yandex.Cloud. For more information, see [the official documentation](https://cloud.yandex.com/en/docs/vpc/concepts/gateways).

* How-to Guides
  * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)

```terraform
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_subnet" "lab-subnet-a" {
  v4_cidr_blocks = ["10.2.0.0/16"]
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.lab-net.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the VPC Gateway. Provided by the client when the VPC Gateway is created.

* `description` - (Optional) An optional description of this resource. Provide this property when you create the resource.

* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.

* `labels` - (Optional) Labels to apply to this VPC Gateway. A list of key/value pairs.

* `shared_egress_gateway` - Shared egress gateway configuration. Currently empty.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the key.

## Import

A gateway can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_vpc_gateway.default gateway_id
```
