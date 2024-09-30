---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_registry"
description: |-
  Get information about a Yandex Container Registry.
---


# yandex_container_registry




Get information about a Yandex Container Registry. For more information, see [the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/registry)

```terraform
resource "yandex_container_registry" "default" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
}

data "yandex_container_registry_ip_permission" "my_ip_permission_by_id" {
  registry_id = yandex_container_registry.default.id
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
