---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a single member for a single IAM binding for a Key Management Service.
---

# {{.Name}} ({{.Type}})

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex KMS Asymmetric Encryption Key.

~> Roles controlled by `yandex_kms_asymmetric_encryption_key_iam_binding` should not be assigned using `yandex_kms_asymmetric_encryption_key_iam_member`.

## Example usage

{{ tffile "examples/kms_asymmetric_encryption_key_iam_member/r_kms_asymmetric_encryption_key_iam_member_1.tf" }}

## Argument Reference

The following arguments are supported:

* `asymmetric_encryption_key_id` - (Required) The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Asymmetric Encryption Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/kms/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

KMS Asymmetric Encryption Key IAM member resource can be imported using the `asymmetric_encryption_key_id` resource role and member ID (account).

{{ codefile "shell" "examples/kms_asymmetric_encryption_key_iam_member/import.sh" }}
