---
layout: "yandex"
page_title: "Yandex: yandex_logging_group"
sidebar_current: "docs-yandex-logging-group"
description: |-
  Manages Yandex Cloud Logging group.
---

# yandex\_logging\_group

Yandex Cloud Logging group resource. For more information, see
[the official documentation](https://cloud.yandex.com/en/docs/logging/concepts/log-group).

## Example Usage

```hcl
resource "yandex_logging_group" "group1" {
  name      = "test-logging-group"
  folder_id = "${data.yandex_resourcemanager_folder.test_folder.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name for the Yandex Cloud Logging group.
* `location_id` - (Optional) Location ID for the Yandex Cloud Logging group.
* `folder_id` - (Optional) ID of the folder that the Yandex Cloud Logging group belongs to.
  It will be deduced from provider configuration if not set explicitly.
* `retention_period` - (Optional) Log entries retention period for the Yandex Cloud Logging group.
* `description` - (Optional) A description for the Yandex Cloud Logging group.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Cloud Logging group.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The Yandex Cloud Logging group ID.
* `cloud_id` - ID of the cloud that the Yandex Cloud Logging group belong to.
* `created_at` - The Yandex Cloud Logging group creation timestamp.
* `status` - The Yandex Cloud Logging group status.
