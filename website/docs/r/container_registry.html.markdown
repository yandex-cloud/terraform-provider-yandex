---
layout: "yandex"
page_title: "Yandex: yandex_container_registry"
sidebar_current: "docs-yandex-container-registry"
description: |-
  Creates a new container registry.
---

# yandex\_container\_registry

Creates a new container registry. For more information, see
[the official documentation](https://cloud.yandex.com/docs/container-registry/concepts/registry)

## Example Usage

```hcl
resource "yandex_container_registry" "default" {
  name      = "test-registry"
  folder_id = "test_folder_id"

  labels = {
    my-label = "my-label-value"
  }
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