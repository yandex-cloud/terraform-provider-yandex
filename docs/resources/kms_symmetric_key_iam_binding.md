---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_symmetric_key_iam_binding"
description: |-
  Allows management of a single IAM binding for a [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/).
---


# yandex_kms_symmetric_key_iam_binding




Allows creation and management of a single binding within IAM policy for an existing Yandex KMS Symmetric Key.

## Example usage

```terraform
resource "yandex_kms_symmetric_key" "your-key" {
  folder_id = "your-folder-id"
  name      = "symmetric-key-name"
}

resource "yandex_kms_symmetric_key_iam_binding" "viewer" {
  symmetric_key_id = yandex_kms_symmetric_key.your-key.id
  role             = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `symmetric_key_id` - (Required) The [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/) Symmetric Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/kms/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **federatedUser:{user_id}**: A unique user ID that represents a specific user account from an identity federation, like Active Directory.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `symmetric_key_id` and role, e.g.

```
$ terraform import yandex_kms_symmetric_key_iam_binding.viewer "symmetric_key_id viewer"
```
