---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret"
description: |-
  Manages Yandex Cloud Lockbox secret.
---


# yandex_lockbox_secret




Yandex Cloud Lockbox secret resource. For more information, see [the official documentation](https://cloud.yandex.com/en/docs/lockbox/).

```terraform
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret"
}

resource "yandex_lockbox_secret_version_hashed" "my_version" {
  secret_id    = yandex_lockbox_secret.my_secret.id
  key_1        = "key1"
  text_value_1 = "sensitive value 1" // in Terraform state, these values will be stored in hash format
  key_2        = "k2"
  text_value_2 = "sensitive value 2"
  // etc. (up to 10 entries)
}
```

Use `yandex_lockbox_secret_version` to add entries to the secret.

```terraform
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret with passowrd"

  password_payload_specification {
    password_key = "some_password"
    length       = 12
  }
}

resource "yandex_lockbox_secret_version" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
}
```

The created secret will contain a version with the generated password. You can use `yandex_lockbox_secret_version` to create new versions.

## Argument Reference

The following arguments are supported:

* `deletion_protection` - (Optional) Whether the Yandex Cloud Lockbox secret is protected from deletion.
* `description` - (Optional) A description for the Yandex Cloud Lockbox secret.
* `folder_id` - (Optional) ID of the folder that the Yandex Cloud Lockbox secret belongs to. It will be deduced from provider configuration if not set explicitly.
* `kms_key_id` - (Optional) The KMS key used to encrypt the Yandex Cloud Lockbox secret.
* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Cloud Lockbox secret.
* `name` - (Optional) Name for the Yandex Cloud Lockbox secret.
* `password_payload_specification` - (Optional) Payload specification for password generation.

The `password_payload_specification` block contains:

* `password_key` - (Required) The key with which the generated password will be placed in the secret version.
* `length` - (Optional) Length of generated password. Default is 36.
* `include_uppercase` - (Optional) Use capital letters in the generated password. Default is true.
* `include_lowercase` - (Optional) Use lowercase letters in the generated password. Default is true.
* `include_digits` - (Optional) Use digits in the generated password. Default is true.
* `include_punctuation` - (Optional) Use punctuations (`!"#$%&'()*+,-./:;<=>?@[\]^_`{|}~`) in the generated password. Default is true.
* `included_punctuation` - (Optional) String of specific punctuation characters to use. Requires `include_punctuation = true`. Default is empty.
* `excluded_punctuation` - (Optional) String of punctuation characters to exclude from the default. Requires `include_punctuation = true`. Default is empty.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - The Yandex Cloud Lockbox secret creation timestamp.
* `status` - The Yandex Cloud Lockbox secret status.
