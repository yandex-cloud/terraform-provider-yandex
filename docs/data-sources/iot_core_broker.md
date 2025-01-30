---
subcategory: "IoT Core"
page_title: "Yandex: yandex_iot_core_broker"
description: |-
  Get information about a Yandex Cloud IoT Core Broker.
---

# yandex_iot_core_broker (Data Source)

Get information about a Yandex IoT Core Broker. For more information IoT Core, see [Yandex Cloud IoT Broker](https://yandex.cloud/docs/iot-core/quickstart).

## Example usage

```terraform
data "yandex_iot_core_broker" "my_broker" {
  broker_id = "are1samplebroker11"
}
```

This data source is used to define [Yandex Cloud IoT Broker](https://cloud.yandex.com/docs/iot-core/quickstart) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `broker_id` (Optional) - IoT Core Broker id used to define broker

* `name` (Optional) - IoT Core Broker name used to define broker

* `folder_id` (Optional) - Folder ID for the IoT Core Broker

~> Either `broker_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the IoT Core Broker
* `labels` - A set of key/value label pairs to assign to the IoT Core Broker.
* `created_at` - Creation timestamp of the IoT Core Broker
* `certificates` - A set of certificates fingerprints for the IoT Core Broker
* `log_options` - Options for logging for IoT Core Broker
