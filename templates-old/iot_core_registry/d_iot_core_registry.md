---
subcategory: "IoT Core"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud IoT Core Registry.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex IoT Core Registry. For more information IoT Core, see [Yandex Cloud IoT Registry](https://yandex.cloud/docs/iot-core/quickstart).

## Example usage

{{ tffile "examples/iot_core_registry/d_iot_core_registry_1.tf" }}

This data source is used to define [Yandex Cloud IoT Registry](https://yandex.cloud/docs/iot-core/quickstart) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `registry_id` (Optional) - IoT Core Registry id used to define registry

* `name` (Optional) - IoT Core Registry name used to define registry

* `folder_id` (Optional) - Folder ID for the IoT Core Registry

~> Either `registry_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the IoT Core Registry
* `labels` - A set of key/value label pairs to assign to the IoT Core Registry.
* `created_at` - Creation timestamp of the IoT Core Registry
* `certificates` - A set of certificate's fingerprints for the IoT Core Registry
* `passwords` - A set of passwords's id for the IoT Core Registry
* `log_options` - Options for logging for IoT Core Registry
