---
layout: "yandex"
page_title: "Yandex: yandex_lockbox_secret"
sidebar_current: "docs-yandex-lockbox-secret"
description: |-
  Manages Yandex Cloud Lockbox secret.
---

# yandex\_lockbox\_secret

Yandex Cloud Lockbox secret resource. For more information, see
[the official documentation](https://cloud.yandex.com/en/docs/lockbox/).

## Example Usage

Use `yandex_lockbox_secret_version` to add entries to the secret.

```hcl
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret"
}
```

## Argument Reference

The following arguments are supported:

* `deletion_protection` - (Optional) Whether the Yandex Cloud Lockbox secret is protected from deletion.
* `description` - (Optional) A description for the Yandex Cloud Lockbox secret.
* `folder_id` - (Optional) ID of the folder that the Yandex Cloud Lockbox secret belongs to.
  It will be deduced from provider configuration if not set explicitly.
* `kms_key_id` - (Optional) The KMS key used to encrypt the Yandex Cloud Lockbox secret.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Cloud Lockbox secret.
* `name` - (Optional) Name for the Yandex Cloud Lockbox secret.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `created_at` - The Yandex Cloud Lockbox secret creation timestamp.
* `status` - The Yandex Cloud Lockbox secret status.
