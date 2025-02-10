---
subcategory: "Container Registry"
page_title: "Yandex: yandex_container_registry"
description: |-
  Creates a new container registry.
---

# yandex_container_registry (Resource)

Creates a new container registry. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/registry)

## Example usage

```terraform
//
// Create a new Container Registry.
//
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

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_container_registry.<resource Name> <resource Id>
terraform import yandex_container_registry.my_registry crps9**********k9psn
```
