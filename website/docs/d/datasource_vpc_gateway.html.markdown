---
layout: "yandex"
page_title: "Yandex: yandex_vpc_gateway"
sidebar_current: "docs-yandex-datasource-vpc-gateway"
description: |-
  Get information about a Yandex VPC gateway.
---

# yandex\_vpc\_gateway

Get information about a Yandex VPC gateway. For more information, see
[Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/index).

```hcl
data "yandex_vpc_gateway" "default" {
  gateway_id = "my-gateway-id"
}
```

This data source is used to define [VPC Gateways] that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `gateway_id` (Optional) - ID of the VPC Gateway.
* `name` (Optional) - Name of the VPC Gateway.

~> **NOTE:** One of `gateway_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the VPC Gateway.
* `labels` - Labels assigned to this VPC Gateway.
* `shared_egress_gateway` - Shared egress gateway configuration
* `created_at` - Creation timestamp of this VPC Gateway.

The `shared_egress_gateway` currently does not support any attributes.

[VPC Gateways]: https://cloud.yandex.com/docs/vpc/concepts/gateway
