---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud"
description: |-
  Retrieve Yandex RM Cloud details.
---


# yandex_resourcemanager_cloud




Use this data source to get cloud details. For more information, see [Cloud](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#cloud).

## Example usage

```terraform
data "yandex_resourcemanager_cloud" "foo" {
  name = "foo-cloud"
}

output "cloud_create_timestamp" {
  value = data.yandex_resourcemanager_cloud.foo.created_at
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) ID of the cloud.
* `name` - (Optional) Name of the cloud.

~> **NOTE:** Either `cloud_id` or `name` must be specified.

## Attributes Reference

The following attributes are returned:

* `name` - Name of the cloud.
* `description` - Description of the cloud.
* `created_at` - Cloud creation timestamp.