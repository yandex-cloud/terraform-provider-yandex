---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex VPC address.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex VPC address. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/address).

## Example usage

{{ tffile "examples/vpc_address/d_vpc_address_1.tf" }}

This data source is used to define [VPC Address](https://yandex.cloud/docs/vpc/concepts/address) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `address_id` (Optional) - ID of the address.
* `name` (Optional) - Name of the address.

~> One of `address_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the address.
* `labels` - Labels assigned to this address.
* `created_at` - Creation timestamp of this address.
* `used` - `true` if address is used.
* `reserved` - `false` means that address is ephemeral.
* `external_ipv4_address` - spec of IP v4 address.
* `deletion_protection` - Flag that protects the address from accidental deletion.

---

The `external_ipv4_address` block supports:

* `address` - IP address.
* `zone_id` - Zone for allocating address.
* `ddos_protection_provider` - DDOS protection provider.
* `outgoing_smtp_capability` - Outgoing smtp capability.

---
