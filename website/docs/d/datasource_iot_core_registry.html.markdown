---
layout: "yandex"
page_title: "Yandex: yandex_iot_core_registry"
sidebar_current: "docs-yandex-datasource-iot-core-registry"
description: |-
  Get information about a Yandex.Cloud IoT Core Registry.
---

# yandex\_iot\_registry

Get information about a Yandex IoT Core Registry. For more information IoT Core, see 
[Yandex.Cloud IoT Registry](https://cloud.yandex.com/docs/iot-core/quickstart).

```hcl
data "yandex_iot_core_registry" "my_registry" {
  registry_id = "are1sampleregistry11"
}
```

This data source is used to define [Yandex.Cloud IoT Registry](https://cloud.yandex.com/docs/iot-core/quickstart) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `registry_id` (Optional) - IoT Core Registry id used to define registry

* `name` (Optional) - IoT Core Registry name used to define registry

* `folder_id` (Optional) - Folder ID for the IoT Core Registry

~> **NOTE:** Either `registry_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the IoT Core Registry
* `labels` - A set of key/value label pairs to assign to the IoT Core Registry.
* `created_at` - Creation timestamp of the IoT Core Registry
* `certificates` - A set of certificate's fingerprints for the IoT Core Registry
* `passwords` - A set of passwords's id for the IoT Core Registry

