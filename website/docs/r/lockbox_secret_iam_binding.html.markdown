---
layout: "yandex"
page_title: "Yandex: yandex_lockbox_secret_iam_binding"
sidebar_current: "docs-yandex-lockbox-secret-iam-binding"
description: |-
Allows management of a single IAM binding for a [Lockbox Secret](https://cloud.yandex.com/docs/lockbox/).
---

## yandex\_lockbox\_secret\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex Lockbox Secret.

## Example Usage

```hcl
resource "yandex_lockbox_secret" "your-secret" {
  name      = "secret-name"
}

resource "yandex_lockbox_secret_iam_binding" "viewer" {
  secret_id = yandex_lockbox_secret.your-secret.id
  role             = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The [Yandex Lockbox Secret](https://cloud.yandex.com/docs/lockbox/) Secret ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/lockbox/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `secret_id` and role, e.g.

```
$ terraform import yandex_lockbox_secret_iam_binding.viewer "secret_id viewer"
```
