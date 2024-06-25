---
layout: "yandex"
page_title: "Yandex: yandex_compute_filesystem_iam_binding"
sidebar_current: "docs-yandex-compute-filesystem-iam-binding"
description: |-
Allows management of a single IAM binding for a Filesystem.
---

# yandex\_compute\_filesystem\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Filesystem.

## Example Usage

```hcl
resource "yandex_compute_filesystem" "fs1" {
  name  = "fs-name"
  type  = "network-ssd"
  zone  = "ru-central1-a"
  size  = 10

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_filesystem_iam_binding" "editor" {
  filesystem_id = "${data.yandex_compute_filesystem.fs1.id}"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `filesystem_id` - (Required) ID of the filesystem to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one
  `yandex_compute_filesystem_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **federatedUser:{federated_user_id}**: A unique federated user ID.
    * **federatedUser:{federated_user_id}:**: A unique SAML federation user account ID.
    * **group:{group_id}**: A unique group ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `filesystem_id` and role, e.g.

```
$ terraform import yandex_compute_filesystem_iam_binding.editor "filesystem_id editor"
```
