---
layout: "yandex"
page_title: "Yandex: yandex_alb_target_group"
sidebar_current: "docs-yandex-datasource-alb-target-group"
description: |-
  Get information about a Yandex Application Load Balancer target group.
---

# yandex\_alb\_target\_group

Get information about a Yandex Application Load Balancer target group. For more information, see
[Yandex.Cloud Application Load Balancer](https://cloud.yandex.com/en/docs/application-load-balancer/quickstart).

```hcl
data "yandex_alb_target_group" "foo" {
  target_group_id = "my-target-group-id"
}
```

This data source is used to define [Application Load Balancer Target Groups] that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `target_group_id` (Optional) - Target Group ID.

* `name` - (Optional) - Name of the Target Group.

~> **NOTE:** One of `target_group_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the target group.
* `labels` - Labels to assign to this target group.
* `target.0.ip_address` - IP address of the target.
* `target.0.subnet_id` - ID of the subnet that targets are connected to.
* `created_at` - Creation timestamp of this target group.

[Application Load Balancer Target Groups]: https://cloud.yandex.com/en/docs/application-load-balancer/concepts/target-group