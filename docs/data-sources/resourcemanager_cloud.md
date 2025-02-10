---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud"
description: |-
  Retrieve Yandex RM Cloud details.
---

# yandex_resourcemanager_cloud (Data Source)

Use this data source to get cloud details. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#cloud).

## Example usage

```terraform
//
// Get information about existing Cloud.
//
data "yandex_resourcemanager_cloud" "my_cloud" {
  name = "foo-cloud"
}

output "cloud_create_timestamp" {
  value = data.yandex_resourcemanager_cloud.my_cloud.created_at
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) ID of the cloud.
* `name` - (Optional) Name of the cloud.

~> Either `cloud_id` or `name` must be specified.

## Attributes Reference

The following attributes are returned:

* `name` - Name of the cloud.
* `description` - Description of the cloud.
* `created_at` - Cloud creation timestamp.
