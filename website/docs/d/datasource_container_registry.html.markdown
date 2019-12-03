---
layout: "yandex"
page_title: "Yandex: yandex_container_registry"
sidebar_current: "docs-yandex-datasource-container-registry"
description: |-
  Get information about a Yandex Container Registry.
---

# yandex\_container\_registry

Get information about a Yandex Container Registry. For more information, see
[the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/registry)

## Example Usage

```hcl
data "yandex_container_registry" "source" {
  registry_id = "some_registry_id"
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) The ID of a specific registry.
* `name` - (Optional) Name of the registry.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.


## Attributes Reference

* `status` - Status of the registry.
* `labels` - Labels to assign to this registry.
* `created_at` - Creation timestamp of this registry.