---
layout: "yandex"
page_title: "Yandex: yandex_vpc_network"
sidebar_current: "docs-yandex-datasource-vpc-network"
description: |-
  Get information about a Yandex VPC network.
---

# yandex\_vpc\_network

Get information about a Yandex VPC network. For more information, see
[Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/index).

```hcl
data "yandex_vpc_network" "admin" {
  network_id = "my-network-id"
}
```

This data source is used to define [VPC Networks] that can be used by other resources.

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

[VPC Networks]: https://cloud.yandex.com/docs/vpc/concepts/network
