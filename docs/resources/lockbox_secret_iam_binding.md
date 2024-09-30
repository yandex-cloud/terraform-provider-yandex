---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret_iam_binding"
description: |-
  Allows management of a single IAM binding for a [Lockbox Secret](https://cloud.yandex.com/docs/lockbox/).
---


# yandex_lockbox_secret_iam_binding




Allows creation and management of a single binding within IAM policy for an existing Yandex Lockbox Secret.

```terraform
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret"
}

resource "yandex_lockbox_secret_version_hashed" "my_version" {
  secret_id    = yandex_lockbox_secret.my_secret.id
  key_1        = "key1"
  text_value_1 = "sensitive value 1" // in Terraform state, these values will be stored in hash format
  key_2        = "k2"
  text_value_2 = "sensitive value 2"
  // etc. (up to 10 entries)
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The [Yandex Lockbox Secret](https://cloud.yandex.com/docs/lockbox/) Secret ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/lockbox/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `secret_id` and role, e.g.

```
$ terraform import yandex_lockbox_secret_iam_binding.viewer "secret_id viewer"
```
