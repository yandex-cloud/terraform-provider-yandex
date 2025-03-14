---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_gateway"
description: |-
  Get information about a Yandex VPC gateway.
---

# yandex_vpc_gateway (Data Source)

Get information about a Yandex VPC gateway. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts).

## Example usage

```terraform
//
// Get information about existing VPC NAT Gateway.
//
data "yandex_vpc_gateway" "default" {
  gateway_id = "my-gateway-id"
}
```

This data source is used to define [VPC Gateways](https://yandex.cloud/docs/vpc/concepts/gateways) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `gateway_id` (Optional) - ID of the VPC Gateway.
* `name` (Optional) - Name of the VPC Gateway.

~> One of `gateway_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the VPC Gateway.
* `labels` - Labels assigned to this VPC Gateway.
* `shared_egress_gateway` - Shared egress gateway configuration
* `created_at` - Creation timestamp of this VPC Gateway.

The `shared_egress_gateway` currently does not support any attributes.
