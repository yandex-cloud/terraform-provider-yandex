---
subcategory: "Cloud Content Delivery Network (CDN)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud CDN Origin Groups.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud CDN Origin Groups](https://yandex.cloud/docs/cdn/concepts/origins).

~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`

## Example usage

{{ tffile "examples/cdn_origin_group/r_cdn_origin_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` (Required) - CDN Origin Group name used to define device.

* `use_next` (Optional) - If the option is active (has true value), in case the origin responds with 4XX or 5XX codes, use the next origin from the list.

* `origins` - A set of available origins, an origins group must contain at least one enabled origin with fields:
  - source (Required) - IP address or Domain name of your origin and the port;
  - enabled (Optional) - the origin is enabled and used as a source for the CDN. Default is enabled.
  - backup (Optional) - specifies whether the origin is used in its origin group as backup. A backup origin is used when one of active origins becomes unavailable.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the IoT Core Device

## Timeouts

This resource provides the following configuration options for [timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 5 minutes.
- `update` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/cdn_origin_group/import.sh" }}
