---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_asymmetric_signature_key_iam_binding"
description: |-
  Allows management of a single IAM binding for a Key Management Service.
---

# yandex_kms_asymmetric_signature_key_iam_binding (Resource)

Allows creation and management of a single binding within IAM policy for an existing Yandex KMS Asymmetric Signature Key.

~> Roles controlled by `yandex_kms_asymmetric_signature_key_iam_binding` should not be assigned using `yandex_kms_asymmetric_signature_key_iam_member`.

~> When you delete `yandex_kms_asymmetric_signature_key_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!

## Example usage

```terraform
//
// Create a new KMS Assymetric Signature Key and new IAM Binding for it.
//
resource "yandex_kms_asymmetric_signature_key" "your-key" {
  name = "asymmetric-signature-key-name"
}

resource "yandex_kms_asymmetric_signature_key_iam_binding" "viewer" {
  asymmetric_signaturen_key_id = yandex_kms_asymmetric_signature_key.your-key.id
  role                         = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `asymmetric_signature_key_id` - (Required) The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Asymmetric Signature Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/kms/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

KMS Asymmetric Signature Key IAM binding resource can be imported using the `asymmetric_signature_key_id` and resource role.

```shell
# terraform import yandex_kms_asymmetric_signature_key_iam_binding.<resource Name> "<resource Id> <resource Role>"
terraform import yandex_kms_asymmetric_signature_key_iam_binding.viewer "abjjf**********p3gp8 viewer"
```
