---
layout: "yandex"
page_title: "Yandex: yandex_vpc_address"
sidebar_current: "docs-yandex-vpc-address"
description: |-
  Manages a VPC address within Yandex.Cloud.
---

# yandex\_vpc\_address

Manages a address within the Yandex.Cloud. You can only create a reserved (static) address via this resource. An ephemeral address could be obtained via implicit creation at a compute instance creation only. For more information, see [the official documentation](https://cloud.yandex.com/docs/vpc/concepts/address).

* How-to Guides
    * [Cloud Networking](https://cloud.yandex.com/docs/vpc/)
    * [VPC Addressing](https://cloud.yandex.com/docs/vpc/concepts/address)

## Example Usage

### External ipv4 address

```hcl
resource "yandex_vpc_address" "addr" {
  name = "exampleAddress"

  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}
```

### Address with DDoS protection

```hcl
resource "yandex_vpc_address" "vpnaddr" {
  name = "vpnaddr"

  external_ipv4_address {
    zone_id                  = "ru-central1-a"
    ddos_protection_provider = "qrator"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the address. Provided by the client when the address is created.
* `description` - (Optional) An optional description of this resource. Provide this property when
  you create the resource.
* `folder_id` - (Optional) ID of the folder that the resource belongs to. If it
    is not provided, the default provider folder is used.
* `labels` - (Optional) Labels to apply to this resource. A list of key/value pairs.

---

* `external_ipv4_address` - (Optional) spec of IP v4 address
---

The `external_ipv4_address` block supports:

* `zone_id` - Zone for allocating address.
* `ddos_protection_provider` - (Optional) Enable DDOS protection. Possible values are: "qrator"
* `outgoing_smtp_capability` - (Optional) Wanted outgoing smtp capability.

~> **NOTE:** Either one `address` or `zone_id` arguments can be specified.
~> **NOTE:** Either one `ddos_protection_provider` or `outgoing_smtp_capability` arguments can be specified.
~> **NOTE:** Change any argument in `external_ipv4_address` will cause an address recreate

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `address` - Allocated IP address.
* `created_at` - Creation timestamp of the key.
* `reserved` - `false` means that address is ephemeral.
* `used` - `true` if address is used.

## Import

A address can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_vpc_address.addr address_id
```
