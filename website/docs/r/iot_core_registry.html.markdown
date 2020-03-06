---
layout: "yandex"
page_title: "Yandex: yandex_iot_core_registry"
sidebar_current: "docs-yandex-iot-core-registry"
description: |-
 Allows management of a Yandex.Cloud IoT Core Registry.
---

# yandex\_iot\_registry

Allows management of [Yandex.Cloud IoT Registry](https://cloud.yandex.com/docs/iot-core/quickstart).

## Example Usage

```hcl
resource "yandex_iot_core_registry" "my_registry" {
  name        = "some_name"
  description = "any description"
  labels = {
    my-label = "my-label-value"
  }
  passwords = [
    "my-password1",
    "my-password2"
  ]
  certificates = [
    "public part of certificate1",
    "public part of certificate2"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` (Required) - IoT Core Device name used to define registry

* `description` (Optional) - Description of the IoT Core Registry

* `labels` - A set of key/value label pairs to assign to the IoT Core Registry.

* `certificates` - A set of certificate's fingerprints for the IoT Core Registry

* `passwords` - A set of passwords's id for the IoT Core Registry


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `folder_id` - Folder ID for the IoT Core Registry

* `created_at` - Creation timestamp of the IoT Core Registry