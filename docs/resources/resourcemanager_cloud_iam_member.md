---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud_iam_member"
description: |-
  Allows management of a single member for a single IAM binding on a Yandex Resource Manager cloud.
---


# yandex_resourcemanager_cloud_iam_member




Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Resource Manager cloud.

~> **Note:** Roles controlled by `yandex_resourcemanager_cloud_iam_binding` should not be assigned using `yandex_resourcemanager_cloud_iam_member`.

~> **Note:** When you delete `yandex_resourcemanager_cloud_iam_binding` resource, the roles can be deleted from other users within the cloud as well. Be careful!

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

* `cloud_id` - (Required) ID of the cloud to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account. This member resource can be imported using the `cloud id`, role, and account, e.g.

```
$ terraform import yandex_resourcemanager_cloud_iam_member.my_project "cloud_id viewer foo@example.com"
```
