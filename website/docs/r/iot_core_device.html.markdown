---
layout: "yandex"
page_title: "Yandex: yandex_iot_core_device"
sidebar_current: "docs-yandex-iot-core-device"
description: |-
 Allows management of a Yandex.Cloud IoT Core Device.
---

# yandex\_iot\_device

Allows management of [Yandex.Cloud IoT Device](https://cloud.yandex.com/docs/iot-core/quickstart).

## Example Usage

```hcl
resource "yandex_iot_core_device" "my_device" {
  registry_id = "are1sampleregistryid11"
  name        = "some_name"
  description = "any description"
  aliases = {
    "some_alias1/subtopic" = "$devices/{id}/events/somesubtopic",
    "some_alias2/subtopic" = "$devices/{id}/events/aaa/bbb",
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
