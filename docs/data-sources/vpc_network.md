---
subcategory: "VPC (Virtual Private Cloud)"
page_title: "Yandex: yandex_vpc_network"
description: |-
  Get information about a Yandex VPC network.
---


# yandex_vpc_network




Get information about a Yandex VPC network. For more information, see [Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/index).

## Example usage

```terraform
data "yandex_vpc_network" "admin" {
  network_id = "my-network-id"
}
```

This data source is used to define [VPC Networks](https://cloud.yandex.com/docs/vpc/concepts/network) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `network_id` (Optional) - ID of the network.
* `name` (Optional) - Name of the network.

~> **NOTE:** One of `network_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `subnet_ids` - List of subnet ids.
* `description` - Description of the network.
* `labels` - Labels assigned to this network.
* `default_security_group_id` - ID of default Security Group of this network.
* `created_at` - Creation timestamp of this network.
