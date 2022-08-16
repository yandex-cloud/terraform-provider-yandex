---
layout: "yandex"
page_title: "Yandex: yandex_vpc_gateway"
sidebar_current: "docs-yandex-vpc-gateway"
description: |-
  Manages a gateway within Yandex.Cloud.
---

# yandex\_vpc\_gateway

Manages a gateway within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts/gateway#gateway).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)

## Example Usage

```hcl
resource "yandex_vpc_gateway" "default" {
  name = "foobar"
  shared_egress_gateway {}
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the VPC Gateway. Provided by the client when the VPC Gateway is created.

* `description` - (Optional) An optional description of this resource. Provide this property when
  you create the resource.

* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.

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
