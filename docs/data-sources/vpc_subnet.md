---
subcategory: "Virtual Private Cloud (VPC)"
page_title: "Yandex: yandex_vpc_subnet"
description: |-
  Get information about a Yandex VPC subnet.
---


# yandex_vpc_subnet




Get information about a Yandex VPC subnet. For more information, see [Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/index).

## Example usage

```terraform
data "yandex_vpc_subnet" "admin" {
  subnet_id = "my-subnet-id"
}
```

This data source is used to define [VPC Subnets](https://cloud.yandex.com/docs/vpc/concepts/network#subnet) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `subnet_id` (Optional) - Subnet ID.
* `name` - (Optional) - Name of the subnet.

~> **NOTE:** One of `subnet_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the subnet.
* `network_id` - ID of the network this subnet belongs to.
* `labels` - Labels to assign to this subnet.
* `zone` - Name of the availability zone for this subnet.
* `route_table_id` - ID of the route table to assign to this subnet.
* `v4_cidr_blocks` - The blocks of internal IPv4 addresses owned by this subnet.
* `v6_cidr_blocks` - The blocks of internal IPv6 addresses owned by this subnet.
* `dhcp_options` - Options for DHCP client. The structure is documented below.
* `created_at` - Creation timestamp of this subnet.

~> **Note:** `v6_cidr_blocks` attribute is currently not supported. It will be available in the future.

---

The `dhcp_options` block supports:

* `domain_name` - Domain name.
* `domain_name_servers` - Domain name server IP addresses.
* `ntp_servers` - NTP server IP addresses.
