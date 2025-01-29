---
subcategory: "Cloud Content Delivery Network (CDN)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex CDN Origin Group.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex CDN Origin Group. For more information, see [the official documentation](https://yandex.cloud/docs/cdn/concepts/origins).

~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`

## Example usage

{{ tffile "examples/cdn_origin_group/d_cdn_origin_group_1.tf" }}

## Argument Reference

The following arguments are supported:

* `origin_group_id` - (Optional) The ID of a specific origin group.
* `name` - (Optional) Name of the origin group.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `origins` -A set of available origins, an origins group must contain at least one enabled origin with fields:
  * `source` (Required) - IP address or Domain name of your origin and the port;
  * `enabled` (Optional) - the origin is enabled and used as a source for the CDN. Default is enabled.
  * `backup` (Optional) - specifies whether the origin is used in its origin group as backup. A backup origin is used when one of active origins becomes unavailable.

~> One of `origin_group_id` or `name` should be specified.
