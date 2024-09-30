---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud"
description: |-
  Retrieve Yandex RM Cloud details.
---


# yandex_resourcemanager_cloud




Use this data source to get cloud details. For more information, see [Cloud](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#cloud).

```terraform
# Get folder by ID
data "yandex_resourcemanager_folder" "my_folder_1" {
  folder_id = "folder_id_number_1"
}

# Get folder by name in specific cloud
data "yandex_resourcemanager_folder" "my_folder_2" {
  name     = "folder_name"
  cloud_id = "some_cloud_id"
}

output "my_folder_1_name" {
  value = data.yandex_resourcemanager_folder.my_folder_1.name
}

output "my_folder_2_cloud_id" {
  value = data.yandex_resourcemanager_folder.my_folder_2.cloud_id
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
