---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_service_account_api_key"
description: |-
  Allows management of a Yandex Cloud IAM service account API key.
---

# yandex_iam_service_account_api_key (Resource)

Allows management of a [Yandex Cloud IAM service account API key](https://yandex.cloud/docs/iam/concepts/authorization/api-key). The API key is a private key used for simplified authorization in the Yandex Cloud API. API keys are only used for [service accounts](https://yandex.cloud/docs/iam/concepts/users/service-accounts).

API keys do not expire. This means that this authentication method is simpler, but less secure. Use it if you can't automatically request an [IAM token](https://yandex.cloud/docs/iam/concepts/authorization/iam-token).

## Example usage

```terraform
//
// Create a new IAM Service Account API Key.
//
resource "yandex_iam_service_account_api_key" "sa-api-key" {
  service_account_id = "aje5a**********qspd3"
  description        = "api key for authorization"
  scopes             = ["yc.ydb.topics.manage", "yc.ydb.tables.manage"]
  expires_at         = "2024-11-11T00:00:00Z"
  pgp_key            = "keybase:keybaseusername"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account to an API key for.

---

* `description` - (Optional) The description of the key.

* `scopes` - (Optional) The list of scopes of the key.

* `expires_at` - (Optional) The key will be no longer valid after expiration timestamp.

* `pgp_key` - (Optional) An optional PGP key to encrypt the resulting secret key material. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.

* `output_to_lockbox` - (Optional) Used to store the sensitive values into a Lockbox secret, to avoid leaking them to the Terraform state.

The `output_to_lockbox` block contains:

* `secret_id` - (Required) ID of the Lockbox secret where to store the sensible values.
* `entry_for_secret_key` - (Required) Entry where to store the value of `secret_key`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `secret_key` - The secret key. This is only populated when neither `pgp_key` nor `output_to_lockbox` are provided.

* `encrypted_secret_key` - The encrypted secret key, base64 encoded. This is only populated when `pgp_key` is supplied.

* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the secret key. This is only populated when `pgp_key` is supplied.

* `created_at` - Creation timestamp of the static access key.

* `output_to_lockbox_version_id` - ID of the Lockbox secret version that contains the value of `secret_key`. This is only populated when `output_to_lockbox` is supplied. This version will be destroyed when the IAM key is destroyed, or when `output_to_lockbox` is removed.

## Import

~> Import for this resource is not implemented yet.

