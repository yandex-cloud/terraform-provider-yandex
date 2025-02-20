---
subcategory: "Container Registry"
page_title: "Yandex: {{.Name}}"
description: |-
  Creates a new container registry.
---

# {{.Name}} ({{.Type}})

Creates a new container registry. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/registry)

## Example usage

{{ tffile "examples/container_registry/r_container_registry_1.tf" }}

## Argument Reference

The following arguments are supported:

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `name` - (Optional) A name of the registry.

* `labels` - (Optional) A set of key/value label pairs to assign to the registry.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - Status of the registry.
* `created_at` - Creation timestamp of the registry.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/container_registry/import.sh" }}
