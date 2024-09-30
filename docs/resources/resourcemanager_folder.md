---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_folder"
description: |-
  Allows management of the Cloud Folder.
---


# yandex_resourcemanager_folder




Allows creation and management of Cloud Folders for an existing Yandex Cloud. See [the official documentation](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of folders may take up to 30 minutes as it requires a lot of communication between cloud services.

```terraform
data "yandex_resourcemanager_folder" "project1" {
  folder_id = "my_folder_id"
}

data "yandex_iam_policy" "admin" {
  binding {
    role = "editor"

    members = [
      "userAccount:some_user_id",
    ]
  }
}

resource "yandex_resourcemanager_folder_iam_policy" "folder_admin_policy" {
  folder_id   = data.yandex_folder.project1.id
  policy_data = data.yandex_iam_policy.admin.policy_data
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) Cloud that the resource belongs to. If value is omitted, the default provider Cloud ID is used.

* `name` - (Optional) The name of the Folder.

* `description` - (Optional) A description of the Folder.

* `labels` - (Optional) A set of key/value label pairs to assign to the Folder.
