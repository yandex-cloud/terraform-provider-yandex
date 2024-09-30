---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret_version"
description: |-
  Manages Yandex Cloud Lockbox secret version.
---


# yandex_lockbox_secret_version




Yandex Cloud Lockbox secret version resource. For more information, see [the official documentation](https://cloud.yandex.com/en/docs/lockbox/).

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

## Argument Reference

The following arguments are supported:

* `entries` - (Optional) List of entries in the Yandex Cloud Lockbox secret version. Must be omitted for secrets with a payload specification.
* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID where to add the version.
* `description` - (Optional) The Yandex Cloud Lockbox secret version description.

The `entries` block contains:

* `key` - (Required) The key of the entry.
* `text_value` - (Optional) The text value of the entry.
* `command` - (Optional) The command that generates the text value of the entry.

Note that either `text_value` or `command` is required.

The `command` block contains:

* `path` - (Required) The path to the script or command to execute.
* `args` - (Optional) List of arguments to be passed to the script/command.
* `env` - (Optional) Map of environment variables to set before calling the script/command.
