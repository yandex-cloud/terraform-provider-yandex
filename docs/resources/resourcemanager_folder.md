---
subcategory: "Resource Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of the Cloud Folder.
---

# {{.Name}} ({{.Type}})

Allows creation and management of Cloud Folders for an existing Yandex Cloud. See [the official documentation](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of folders may take up to 30 minutes as it requires a lot of communication between cloud services.

## Example usage

{{ tffile "examples/resourcemanager_folder/r_resourcemanager_folder_1.tf" }}

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) Cloud that the resource belongs to. If value is omitted, the default provider Cloud ID is used.

* `name` - (Optional) The name of the Folder.

* `description` - (Optional) A description of the Folder.

* `labels` - (Optional) A set of key/value label pairs to assign to the Folder.
