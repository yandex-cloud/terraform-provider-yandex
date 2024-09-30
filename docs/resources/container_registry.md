---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_registry"
description: |-
  Creates a new container registry.
---


# yandex_container_registry




Creates a new container registry. For more information, see [the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/registry)

```terraform
resource "yandex_container_registry" "my_registry" {
  name = "test-registry"
}

resource "yandex_container_registry_ip_permission" "my_ip_permission" {
  registry_id = yandex_container_registry.my_registry.id
  push        = ["10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"]
  pull        = ["10.1.0.0/16", "10.5.0/16"]
}
```

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

A registry can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_container_registry.default registry_id
```
