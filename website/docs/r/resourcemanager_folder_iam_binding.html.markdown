---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_folder_iam_binding"
sidebar_current: "docs-yandex-resourcemanager-folder-iam-binding"
description: |-
 Allows management of a single IAM binding for a Yandex Resource Manager folder.
---

# yandex\_resourcemanager\_folder\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex Resource Manager folder.

~> **Note:** This resource _must not_ be used in conjunction with
   `yandex_resourcemanager_folder_iam_policy` or they will conflict over what your policy
   should be.

~> **Note:** When you delete `yandex_resourcemanager_folder_iam_binding` resource,
   the roles can be deleted from other users within the folder as well. Be careful!

## Example Usage

```hcl
data "yandex_resourcemanager_folder" "project1" {
  folder_id = "some_folder_id"
}

resource "yandex_resourcemanager_folder_iam_binding" "admin" {
  folder_id = "${data.yandex_resourcemanager_folder.project1.id}"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Required) ID of the folder to attach a policy to.

* `role` - (Required) The role that should be assigned. Only one
    `yandex_resourcemanager_folder_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege that is specified in the `role` field.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: An email address that represents a specific Yandex account. For example, ivan@yandex.ru or joe@example.com.
  * **serviceAccount:{service_account_id}**: A unique service account ID.

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `folder_id` and role, e.g.

```
$ terraform import yandex_resourcemanager_folder_iam_binding.viewer "folder_id viewer"
```
