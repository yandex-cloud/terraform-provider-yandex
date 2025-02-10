---
subcategory: "Cloud Logging"
page_title: "Yandex: yandex_logging_group"
description: |-
  Manages Yandex Cloud Logging group.
---

# yandex_logging_group (Resource)

Yandex Cloud Logging group resource. For more information, see [the official documentation](https://yandex.cloud/docs/logging/concepts/log-group).

## Example usage

```terraform
//
// Create a new Logging Group.
//
resource "yandex_logging_group" "group1" {
  name      = "test-logging-group"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name for the Yandex Cloud Logging group.
* `location_id` - (Optional) Location ID for the Yandex Cloud Logging group.
* `folder_id` - (Optional) ID of the folder that the Yandex Cloud Logging group belongs to. It will be deduced from provider configuration if not set explicitly.
* `retention_period` - (Optional) Log entries retention period for the Yandex Cloud Logging group.
* `description` - (Optional) A description for the Yandex Cloud Logging group.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Cloud Logging group.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The Yandex Cloud Logging group ID.
* `cloud_id` - ID of the cloud that the Yandex Cloud Logging group belong to.
* `created_at` - The Yandex Cloud Logging group creation timestamp.
* `status` - The Yandex Cloud Logging group status.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_logging_group.<resource Name> <resource Id>
terraform import yandex_logging_group.group1 ...
```
