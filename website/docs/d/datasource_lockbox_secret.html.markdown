---
layout: "yandex"
page_title: "Yandex: yandex_lockbox_secret"
sidebar_current: "docs-yandex-datasource-lockbox-secret"
description: |-
  Get information about Yandex Cloud Lockbox secret.
---

# yandex\_lockbox\_secret

Get information about Yandex Cloud Lockbox secret. For more information,
see [the official documentation](https://cloud.yandex.com/en/docs/lockbox/).

## Example Usage

```hcl
data "yandex_lockbox_secret" "my_secret" {
  secret_id = "some ID"
}

output "my_secret_created_at" {
  value = data.yandex_lockbox_secret.my_secret.created_at
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `created_at` - The Yandex Cloud Lockbox secret creation timestamp.
* `current_version` - Information about the current version of the Yandex Cloud Lockbox secret.
* `deletion_protection` - Whether the Yandex Cloud Lockbox secret is protected from deletion.
* `description` - The Yandex Cloud Lockbox secret description.
* `folder_id` - ID of the folder that the Yandex Cloud Lockbox secret belongs to.
* `kms_key_id` - The KMS key used to encrypt the Yandex Cloud Lockbox secret (if an explicit key was used).
* `labels` - A set of key/value label pairs assigned to the Yandex Cloud Lockbox secret.
* `name` - The Yandex Cloud Lockbox secret name.
* `status` - The Yandex Cloud Lockbox secret status.

The `current_version` block contains:

* `created_at` - The version creation timestamp.
* `description` - The version description.
* `destroy_at` - The version destroy timestamp.
* `id` - The version ID.
* `payload_entry_keys` - List of keys that the version contains (doesn't include the values).
* `secret_id` - The secret ID the version belongs to (it's the same as the `secret_id` argument indicated above)
* `status` - The version status.
