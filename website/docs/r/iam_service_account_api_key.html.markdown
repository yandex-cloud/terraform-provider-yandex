---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_account_api_key"
sidebar_current: "docs-yandex-iam-service-account-api-key"
description: |-
 Allows management of a Yandex.Cloud IAM service account API key.
---

# yandex\_iam\_service\_account\_api\_key

Allows management of a [Yandex.Cloud IAM service account API key](https://cloud.yandex.com/docs/iam/concepts/authorization/api-key).
The API key is a private key used for simplified authorization in the Yandex.Cloud API. API keys are only used for [service accounts](https://cloud.yandex.com/docs/iam/concepts/users/service-accounts).

API keys do not expire. This means that this authentication method is simpler, but less secure. Use it if you can't automatically request an [IAM token](https://cloud.yandex.com/docs/iam/concepts/authorization/iam-token).

## Example Usage

This snippet creates an API key.

```hcl
resource "yandex_iam_service_account_api_key" "sa-api-key" {
  service_account_id = "some_sa_id"
  description        = "api key for authorization"
  pgp_key            = "keybase:keybaseusername"
}
```

## Argument Reference

The following arguments are supported:

* `service_account_id` - (Required) ID of the service account to an API key for.

- - -

* `description` - (Optional) The description of the key.

* `pgp_key` - (Optional) An optional PGP key to encrypt the resulting secret key material. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `secret_key` - The secret key. This is only populated when no `pgp_key` is provided.

* `encrypted_secret_key` - The encrypted secret key, base64 encoded. This is only populated when `pgp_key` is supplied.

* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the secret key. This is only populated when `pgp_key` is supplied.

* `created_at` - Creation timestamp of the static access key.
