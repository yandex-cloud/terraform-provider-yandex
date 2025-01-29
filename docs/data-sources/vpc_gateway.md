---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex VPC gateway.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex VPC gateway. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts).

## Example usage

{{ tffile "examples/vpc_gateway/d_vpc_gateway_1.tf" }}

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
