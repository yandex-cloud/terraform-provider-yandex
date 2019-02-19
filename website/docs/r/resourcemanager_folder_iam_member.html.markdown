---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_folder_iam_member"
sidebar_current: "docs-yandex-resourcemanager-folder-iam-member"
description: |-
 Allows management of a single member for a single IAM binding for a Yandex Resource Manager folder.
---

# yandex\_resourcemanager\_folder\_iam\_member

Allows creation and management of a single member for a single binding within
the IAM policy for an existing Yandex Resource Manager folder.

~> **Note:** This resource _must not_ be used in conjunction with
   `yandex_resourcemanager_folder_iam_policy` or they will conflict over what your policy should be. Similarly, roles controlled by `yandex_resourcemanager_folder_iam_binding`
   should not be assigned using `yandex_resourcemanager_folder_iam_member`.

## Example Usage

```hcl
data "yandex_resourcemanager_folder" "department1" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_member" "admin" {
  folder_id = "${data.yandex_resourcemanager.department1.name}"

  role   = "editor"
  member = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field.
  This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account.
This member resource can be imported using the `folder id`, role, and account, e.g.

```
$ terraform import yandex_resourcemanager_folder_iam_member.my_project "folder_id viewer foo@example.com"
```
