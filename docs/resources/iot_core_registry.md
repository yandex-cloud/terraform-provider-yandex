---
subcategory: "IoT Core"
page_title: "Yandex: yandex_iot_core_registry"
description: |-
  Allows management of a Yandex.Cloud IoT Core Registry.
---


# yandex_iot_core_registry




Allows management of [Yandex.Cloud IoT Registry](https://cloud.yandex.com/docs/iot-core/quickstart).

## Example usage

```terraform
resource "yandex_iot_core_registry" "my_registry" {
  name        = "some_name"
  description = "any description"
  labels = {
    my-label = "my-label-value"
  }
  log_options {
    log_group_id = "log-group-id"
    min_level    = "ERROR"
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

* `log_options` - Options for logging for IoT Core Registry

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `folder_id` - Folder ID for the IoT Core Registry

* `created_at` - Creation timestamp of the IoT Core Registry

---

The `log_options` block supports:
* `disabled` - Is logging for registry disabled
* `log_group_id` - Log entries are written to specified log group
* `folder_id` - Log entries are written to default log group for specified folder
* `min_level` - Minimum log entry level
