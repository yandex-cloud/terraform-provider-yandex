---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: yandex_alb_target_group"
description: |-
  Get information about a Yandex Application Load Balancer target group.
---

# yandex_alb_target_group (Data Source)

Get information about a Yandex Application Load Balancer target group. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).

## Example usage

```terraform
//
// Get information about existing ALB Target Group
//
data "yandex_alb_target_group" "foo" {
  target_group_id = "my-target-group-id"
}
```

This data source is used to define [Application Load Balancer Target Groups](https://yandex.cloud/docs/application-load-balancer/concepts/target-group) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `target_group_id` (Optional) - Target Group ID.

* `name` - (Optional) - Name of the Target Group.

~> One of `target_group_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the target group.
* `labels` - Labels to assign to this target group.
* `target.0.ip_address` - IP address of the target.
* `target.0.subnet_id` - ID of the subnet that targets are connected to.
* `created_at` - Creation timestamp of this target group.
