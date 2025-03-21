---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Container Registry.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Container Registry. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/registry)

## Example usage

{{ tffile "examples/container_registry/d_container_registry_1.tf" }}

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) The ID of a specific registry.
* `name` - (Optional) Name of the registry.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

* `status` - Status of the registry.
* `labels` - Labels to assign to this registry.
* `created_at` - Creation timestamp of this registry.
