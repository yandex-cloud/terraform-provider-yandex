---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_folder"
description: |-
  Get information about a Yandex RM Folder.
---

# yandex_resourcemanager_folder (Data Source)

Use this data source to get information about a Yandex Resource Manager Folder. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#folder).

## Example usage

```terraform
//
// Get information about existing Folder.
//
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

* `folder_id` (Optional) - ID of the folder.

* `name` (Optional) - Name of the folder.

~> Either `folder_id` or `name` must be specified.

* `cloud_id` - (Optional) Cloud that the resource belongs to. If value is omitted, the default provider cloud is used.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the folder.
* `cloud_id` - ID of the cloud that contains the folder.
* `status` - Current status of the folder.
* `labels` - A map of labels applied to this folder.
* `created_at` - Folder creation timestamp.
