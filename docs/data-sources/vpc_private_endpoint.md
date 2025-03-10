---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_private_endpoint"
description: |-
  Get information about a Yandex VPC Private Endpoint.
---

# yandex_vpc_private_endpoint (Data Source)

Get information about a Yandex VPC Private Endpoint. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts/index).

## Example usage

```terraform
//
// Get information about existing VPC Private Endpoint.
//
data "yandex_vpc_private_endpoint" "pe" {
  private_endpoint_id = "my-private-endpoint-id"
}
```

This data source is used to define [VPC Private Endpoint](https://yandex.cloud/docs/vpc/concepts/private-endpoint) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `private_endpoint_id` (Optional) - ID of the private endpoint.
* `name` (Optional) - Name of the private endpoint.

~> One of `private_endpoint_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the private endpoint.
* `labels` - Labels assigned to this private endpoint.
* `created_at` - Creation timestamp of this private endpoint.
* `network_id` - ID of the network which private endpoint belongs to.
* `endpoint_address` - Address information of private endpoint.
* `dns_options` - DNS options of private endpoint.

---

The `endpoint_address` block supports:

* `address_id` - ID of the address.
* `subnet_id` - Subnet of the IP address.
* `address` - IP address.

---

The `dns_options` block supports:

* `private_dns_records_enabled` - `true` if private dns records enabled.
