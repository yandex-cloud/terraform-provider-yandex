---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Application Load Balancer HTTP Router.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Application Load Balancer HTTP Router. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).

## Example usage

{{ tffile "examples/alb_http_router/d_alb_http_router_1.tf" }}

This data source is used to define [Application Load Balancer HTTP Router](https://yandex.cloud/docs/application-load-balancer/concepts/http-router) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `http_router_id` (Optional) - HTTP Router ID.

* `name` - (Optional) - Name of the HTTP Router.

~> One of `http_router_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the HTTP Router.
* `labels` - Labels to assign to this HTTP Router.
* `created_at` - Creation timestamp of this HTTP Router.
