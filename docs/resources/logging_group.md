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

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `data_stream` (String) Data Stream.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.
- `retention_period` (String) Log entries retention period for the Yandex Cloud Logging group.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `cloud_id` (String) The `Cloud ID` which resource belongs to. If it is not provided, the default provider `cloud-id` is used.
- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.
- `status` (String) The Yandex Cloud Logging group status.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `default` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_logging_group.<resource Name> <resource Id>
terraform import yandex_logging_group.group1 ...
```
