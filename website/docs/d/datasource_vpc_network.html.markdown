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

* `network_id` (Required) - ID of the network.

## Attributes Reference

The following attribute is exported:

* `description` - Description of the network.
* `name` - Name of the network.
* `folder_id` - ID of the folder that the resource belongs to.
* `labels` - Labels assigned to this network.
* `created_at` - Creation timestamp of this network.

[VPC Networks]: https://cloud.yandex.com/docs/vpc/concepts/network
