---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_symmetric_key_iam_member"
description: |-
  Allows management of a single member for a single IAM binding for a Key Management Service.
---

# yandex_kms_symmetric_key_iam_member (Resource)

Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex KMS Symmetric Key.

~> Roles controlled by `yandex_kms_symmetric_key_iam_binding` should not be assigned using `yandex_kms_symmetric_key_iam_member`.

## Example usage

```terraform
//
// Create a new KMS Symmetric Key and new IAM Member for it.
//
resource "yandex_kms_symmetric_key" "your-key" {
  name = "symmetric-key-name"
}

resource "yandex_kms_symmetric_key_iam_member" "viewer" {
  symmetric_key_id = yandex_kms_symmetric_key.your-key.id
  role             = "viewer"

  member = "userAccount:foo_user_id"
}
```

## Argument Reference

The following arguments are supported:

* `symmetric_key_id` - (Required) The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Symmetric Key ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://yandex.cloud/docs/kms/security/).

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field. This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **federatedUser:{user_id}**: A unique user ID that represents a specific user account from an identity federation, like Active Directory.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_kms_symmetric_key_iam_member.<resource Name> "<resource Id> <resource Role> <subject Id>"
terraform import yandex_kms_symmetric_key_iam_member.viewer "abjjf**********p3gp8 viewer foo@example.com"
```
