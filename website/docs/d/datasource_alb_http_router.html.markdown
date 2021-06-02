---
layout: "yandex"
page_title: "Yandex: yandex_alb_http_router"
sidebar_current: "docs-yandex-datasource-alb-http-router"
description: |-
Get information about a Yandex Application Load Balancer HTTP Router.
---

# yandex\_alb\_http\_router

Get information about a Yandex Application Load Balancer HTTP Router. For more information, see
[Yandex.Cloud Application Load Balancer](https://cloud.yandex.com/en/docs/application-load-balancer/quickstart).

```hcl
data "yandex_alb_http_router" "tf-router" {
  http_router_id = "my-http-router-id"
}
```

This data source is used to define [Application Load Balancer HTTP Router] that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `http_router_id` (Optional) - HTTP Router ID.

* `name` - (Optional) - Name of the HTTP Router.

~> **NOTE:** One of `http_router_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the HTTP Router.
* `labels` - Labels to assign to this HTTP Router.
* `created_at` - Creation timestamp of this HTTP Router.

[Application Load Balancer HTTP Router]: https://cloud.yandex.com/en/docs/application-load-balancer/concepts/http-router