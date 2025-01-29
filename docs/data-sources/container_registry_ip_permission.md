---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Container Registry IP Permission.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Container Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/operations/registry/registry-access)

## Example usage

{{ tffile "examples/container_registry_ip_permission/d_container_registry_ip_permission_1.tf" }}

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) The ID of a specific Container Registry.

* `registry_name` - (Optional) The Name of specific Container Registry.

~> Either `registry_id` or `registry_name` must be specified.

## Attributes Reference

* `push` - List of configured CIDRs, from which push is allowed.

* `pull` - List of configured CIDRs, from which pull is allowed.
