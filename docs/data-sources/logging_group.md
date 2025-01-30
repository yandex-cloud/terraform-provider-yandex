---
subcategory: "Cloud Logging"
page_title: "Yandex: yandex_logging_group"
description: |-
  Get information about a Yandex Cloud Logging group.
---

# yandex_logging_group (Data Source)

Get information about a Yandex Cloud Logging group. For more information, see [the official documentation](https://yandex.cloud/docs/logging/concepts/log-group).

## Example usage

```terraform
data "yandex_logging_group" "my_group" {
  group_id = "some_yandex_logging_group_id"
}

output "log_group_retention_period" {
  value = data.yandex_logging_group.my_group.retention_period
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The Yandex Cloud Logging group ID.
* `name` - (Optional) The Yandex Cloud Logging group name.
* `folder_id` - (Optional) ID of the folder that the Yandex Cloud Logging group belongs to. It will be deduced from provider configuration if not set explicitly.

~> If `group_id` is not specified `name` and `folder_id` will be used to designate Yandex Cloud Logging group.

## Attributes Reference

* `retention_period` - The Yandex Cloud Logging group log entries retention period.
* `description` - The Yandex Cloud Logging group description.
* `labels` - A set of key/value label pairs assigned to the Yandex Cloud Logging group.
* `cloud_id` - ID of the cloud that the Yandex Cloud Logging group belong to.
* `created_at` - The Yandex Cloud Logging group creation timestamp.
* `status` - The Yandex Cloud Logging group status.
