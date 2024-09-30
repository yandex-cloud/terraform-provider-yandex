---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud"
description: |-
  Allows management of the Cloud resource.
---


# yandex_resourcemanager_cloud




Allows creation and management of Cloud resources for an existing Yandex.Cloud Organization. See [the official documentation](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of clouds may take up to 30 minutes as it requires a lot of communication between cloud services.

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

* `organization_id` - (Optional) Yandex.Cloud Organization that the cloud belongs to. If value is omitted, the default provider Organization ID is used.

* `name` - (Optional) The name of the Cloud.

* `description` - (Optional) A description of the Cloud.

* `labels` - (Optional) A set of key/value label pairs to assign to the Cloud.
