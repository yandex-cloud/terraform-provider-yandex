---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_folder"
description: |-
  Allows management of the Cloud Folder.
---

# yandex_resourcemanager_folder (Resource)

Allows creation and management of Cloud Folders for an existing Yandex Cloud. See [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of folders may take up to 30 minutes as it requires a lot of communication between cloud services.

## Example usage

```terraform
//
// Create a new Folder.
//
resource "yandex_resourcemanager_folder" "folder1" {
  cloud_id = "my_cloud_id"
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) Cloud that the resource belongs to. If value is omitted, the default provider Cloud ID is used.

* `name` - (Optional) The name of the Folder.

* `description` - (Optional) A description of the Folder.

* `labels` - (Optional) A set of key/value label pairs to assign to the Folder.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_resourcemanager_folder.<resource Name> <resource Id>
terraform import yandex_resourcemanager_folder.my_foldeer b1g5r**********dqmsp
```
