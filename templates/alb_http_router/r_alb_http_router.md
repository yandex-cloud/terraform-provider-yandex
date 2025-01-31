---
subcategory: "Application Load Balancer (ALB)"
page_title: "Yandex: {{.Name}}"
description: |-
  The HTTP router defines the routing rules for HTTP requests to backend groups.
---

# {{.Name}} ({{.Type}})

Creates an HTTP Router in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/application-load-balancer/concepts/http-router).

## Example usage

{{ tffile "examples/alb_http_router/r_alb_http_router_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the HTTP Router. Provided by the client when the HTTP Router is created.

* `description` - (Optional) An optional description of the HTTP Router. Provide this property when you create the resource.

* `folder_id` - (Optional) The ID of the folder to which the resource belongs. If omitted, the provider folder is used.

* `labels` - (Optional) Labels to assign to this HTTP Router. A list of key/value pairs.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the HTTP Router.

* `created_at` - The HTTP Router creation timestamp.

## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 5 minute.
- `update` - Default is 5 minute.
- `delete` - Default is 5 minute.

## Import

An HTTP Router can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_alb_http_router.default http_router_id
```
