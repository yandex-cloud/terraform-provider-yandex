---
subcategory: "IAM (Identity and Access Management)"
page_title: "Yandex: yandex_iam_service_account_static_access_key"
description: |-
  Allows management of a Yandex.Cloud IAM service account static access key.
---


# yandex_iam_service_account_static_access_key




Allows management of [Yandex.Cloud IAM service account static access keys](https://cloud.yandex.com/docs/iam/operations/sa/create-access-key). Generated pair of keys is used to access [Yandex Object Storage](https://cloud.yandex.com/docs/storage) on behalf of service account.

Before using keys do not forget to [assign a proper role](https://cloud.yandex.com/docs/iam/operations/sa/assign-role-for-sa) to the service account.

## Example usage

```terraform
resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = "some_sa_id"
  description        = "static access key for object storage"
  pgp_key            = "keybase:keybaseusername"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account which is used to get a static key.

---

* `description` - (Optional) The description of the service account static key.

* `pgp_key` - (Optional) An optional PGP key to encrypt the resulting secret key material. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.

* `output_to_lockbox` - (Optional) Used to store the sensible values into a Lockbox secret, to avoid leaking them to the Terraform state.

The `output_to_lockbox` block contains:

* `secret_id` - (Required) ID of the Lockbox secret where to store the sensible values.
* `entry_for_access_key` - (Required) Entry where to store the value of `access_key`.
* `entry_for_secret_key` - (Required) Entry where to store the value of `secret_key`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `access_key` - ID of the static access key. This is only populated when `output_to_lockbox` is not provided.

* `secret_key` - Private part of generated static access key. This is only populated when neither `pgp_key` nor `output_to_lockbox` are provided.

* `encrypted_secret_key` - The encrypted secret, base64 encoded. This is only populated when `pgp_key` is supplied.

* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the secret key. This is only populated when `pgp_key` is supplied.

* `created_at` - Creation timestamp of the static access key.

* `output_to_lockbox_version_id` - ID of the Lockbox secret version that contains the values of `access_key` and `secret_key`. This is only populated when `output_to_lockbox` is supplied. This version will be destroyed when the IAM key is destroyed, or when `output_to_lockbox` is removed.
