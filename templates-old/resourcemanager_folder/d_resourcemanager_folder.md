---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex RM Folder.
---

# {{.Name}} ({{.Type}})

Use this data source to get information about a Yandex Resource Manager Folder. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#folder).

## Example usage

{{ tffile "examples/resourcemanager_folder/d_resourcemanager_folder_1.tf" }}

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
