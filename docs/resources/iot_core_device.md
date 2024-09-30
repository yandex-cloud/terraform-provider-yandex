---
subcategory: "IoT Core"
page_title: "Yandex: yandex_iot_core_device"
description: |-
  Allows management of a Yandex.Cloud IoT Core Device.
---


# yandex_iot_core_device




Allows management of [Yandex.Cloud IoT Device](https://cloud.yandex.com/docs/iot-core/quickstart).

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

* `registry_id` - IoT Core Registry ID for the IoT Core Device

* `name` (Required) - IoT Core Device name used to define device

* `description` (Optional) - Description of the IoT Core Device

* `aliases` - A set of key/value aliases pairs to assign to the IoT Core Device

* `certificates` - A set of certificate's fingerprints for the IoT Core Device

* `passwords` - A set of passwords's id for the IoT Core Device

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the IoT Core Device
