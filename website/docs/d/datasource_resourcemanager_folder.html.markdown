---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_folder"
sidebar_current: "docs-yandex-datasource-resourcemanager-folder"
description: |-
  Get information about a Yandex RM Folder.
---

# yandex\_resourcemanager\_folder

Use this data source to get information about a Yandex Resource Manager Folder. For more information, see
[the official documentation](https://cloud.yandex.ru/docs/resource-manager/concepts/resources-hierarchy#folder).

```hcl
# Get folder by ID
data "yandex_resourcemanager_folder" "my_folder_1" {
  folder_id = "folder_id_number_1"
}

# Search by fields
data "yandex_resourcemanager_folder" "my_folder_2" {
  folder_id = "folder_id_number_2"
}

output "my_folder_1_name" {
  value = "${data.yandex_resourcemanager_folder.my_folder_1.name}"
}

output "my_folder_2_cloud_id" {
  value = "${data.yandex_resourcemanager_folder.my_folder_2.cloud_id}"
}

```

## Argument Reference

The following arguments are supported:

* `folder_id` (Required) - ID of the folder.

## Attributes Reference

The following attributes are exported:

* `name` - Name of the Folder.
* `description` - Description of the folder.
* `cloud_id` - ID of the cloud that contains the folder.
* `status` - Current status of the folder.
* `labels` - A map of labels applied to this folder.
* `created_at` - Folder creation timestamp.

[Folder]: https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#folder
