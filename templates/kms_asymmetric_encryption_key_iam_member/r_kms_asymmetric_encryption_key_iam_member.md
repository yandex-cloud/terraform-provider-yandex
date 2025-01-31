---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single member for a single IAM binding for a [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/).
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex KMS Asymmetric Encryption Key.

~> Roles controlled by `yandex_kms_asymmetric_encryption_key_iam_binding` should not be assigned using `yandex_kms_asymmetric_encryption_key_iam_member`.

## Example usage

{{ tffile "examples/kms_asymmetric_encryption_key_iam_member/r_kms_asymmetric_encryption_key_iam_member_1.tf" }}

## Argument Reference

The following arguments are supported:

* `asymmetric_encryption_key_id` - (Required) The [Yandex Key Management Service](https://cloud.yandex.com/docs/kms/) Asymmetric Encryption Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/kms/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account. This member resource can be imported using the `asymmetric_encryption_key_id`, role, and account, e.g.

```
$ terraform import {{.Name}}.viewer "asymmetric_encryption_key_id viewer foo@example.com"
```
