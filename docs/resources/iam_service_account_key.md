---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_service_account_key"
description: |-
  Allows management of a Yandex Cloud IAM service account key.
---

# yandex_iam_service_account_key (Resource)

Allows management of [Yandex Cloud IAM service account authorized keys](https://cloud.yandex.com/docs/iam/concepts/authorization/key). Generated pair of keys is used to create a [JSON Web Token](https://tools.ietf.org/html/rfc7519) which is necessary for requesting an [IAM Token](https://cloud.yandex.com/docs/iam/concepts/authorization/iam-token) for a [service account](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts).

## Example usage

```terraform
resource "yandex_iam_service_account_key" "sa-auth-key" {
  service_account_id = "some_sa_id"
  description        = "key for service account"
  key_algorithm      = "RSA_4096"
  pgp_key            = "keybase:keybaseusername"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account to create a pair for.

---

* `description` - (Optional) The description of the key pair.

* `format` - (Optional) The output format of the keys. `PEM_FILE` is the default format.

* `key_algorithm` - (Optional) The algorithm used to generate the key. `RSA_2048` is the default algorithm. Valid values are listed in the [API reference](https://cloud.yandex.com/docs/iam/api-ref/Key).

* `pgp_key` - (Optional) An optional PGP key to encrypt the resulting private key material. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.

* `output_to_lockbox` - (Optional) Used to store the sensible values into a Lockbox secret, to avoid leaking them to the Terraform state.

The `output_to_lockbox` block contains:

* `secret_id` - (Required) ID of the Lockbox secret where to store the sensible values.
* `entry_for_private_key` - (Required) Entry where to store the value of `private_key`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `public_key` - The public key.

* `private_key` - The private key. This is only populated when neither `pgp_key` nor `output_to_lockbox` are provided.

* `encrypted_private_key` - The encrypted private key, base64 encoded. This is only populated when `pgp_key` is supplied.

* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the private key. This is only populated when `pgp_key` is supplied.

* `created_at` - Creation timestamp of the static access key.

* `output_to_lockbox_version_id` - ID of the Lockbox secret version that contains the value of `private_key`. This is only populated when `output_to_lockbox` is supplied. This version will be destroyed when the IAM key is destroyed, or when `output_to_lockbox` is removed.
