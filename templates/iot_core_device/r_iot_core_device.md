---
subcategory: "IoT Core"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud IoT Core Device.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud IoT Device](https://yandex.cloud/docs/iot-core/quickstart).

## Example usage

{{ tffile "examples/iot_core_device/r_iot_core_device_1.tf" }}

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

## Import

~> Import for this resource is not implemented yet.
