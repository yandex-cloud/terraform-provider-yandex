---
layout: "yandex"
page_title: "Yandex: yandex_iot_core_broker"
sidebar_current: "docs-yandex-iot-core-broker"
description: |-
 Allows management of a Yandex.Cloud IoT Core Broker.
---

# yandex\_iot\_broker

Allows management of [Yandex.Cloud IoT Broker](https://cloud.yandex.com/docs/iot-core/quickstart).

The service is at the Preview stage.

## Example Usage

```hcl
resource "yandex_iot_core_broker" "my_broker" {
  name        = "some_name"
  description = "any description"
  labels = {
    my-label = "my-label-value"
  }
  certificates = [
    "public part of certificate1",
    "public part of certificate2"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` (Required) - IoT Core Broker name used to define broker

* `description` (Optional) - Description of the IoT Core Broker

* `labels` - A set of key/value label pairs to assign to the IoT Core Broker.

* `certificates` - A set of certificate's fingerprints for the IoT Core Broker


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `folder_id` - Folder ID for the IoT Core Broker

* `created_at` - Creation timestamp of the IoT Core Broker
