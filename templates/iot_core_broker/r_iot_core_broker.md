---
subcategory: "IoT Core"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud IoT Core Broker.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud IoT Broker](https://cloud.yandex.com/docs/iot-core/quickstart).

## Example usage

{{ tffile "examples/iot_core_broker/r_iot_core_broker_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` (Required) - IoT Core Broker name used to define broker

* `description` (Optional) - Description of the IoT Core Broker

* `labels` - A set of key/value label pairs to assign to the IoT Core Broker.

* `certificates` - A set of certificate's fingerprints for the IoT Core Broker

* `log_options` - Options for logging for IoT Core Broker

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `folder_id` - Folder ID for the IoT Core Broker

* `created_at` - Creation timestamp of the IoT Core Broker

---

The `log_options` block supports:
* `disabled` - Is logging for broker disabled
* `log_group_id` - Log entries are written to specified log group
* `folder_id` - Log entries are written to default log group for specified folder
* `min_level` - Minimum log entry level
