---
layout: "yandex"
page_title: "Yandex: yandex_compute_disk_iam_binding"
sidebar_current: "docs-yandex-compute-disk-iam-binding"
description: |-
  Allows management of a single IAM binding for a Disk.
---

# yandex\_compute\_disk\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Disk.

## Example Usage

```hcl
resource "yandex_compute_disk" "disk1" {
  name     = "disk-name"
  type     = "network-ssd"
  zone     = "ru-central1-a"
  image_id = "ubuntu-16.04-v20180727"

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_disk_iam_binding" "editor" {
  disk_id = "${data.yandex_compute_disk.disk1.id}"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Required) ID of the disk to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one
  `yandex_compute_disk_iam_binding` can be used per role.

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
These bindings can be imported using the `disk_id` and role, e.g.

```
$ terraform import yandex_compute_disk_iam_binding.editor "disk_id editor"
```
