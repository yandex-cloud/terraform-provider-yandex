---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Yandex Cloud Lockbox secret version (with values hashed in state).
---

# {{.Name}} ({{.Type}})

Yandex Cloud Lockbox secret version resource (with values hashed in state). For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).

## Example usage

{{ tffile "examples/lockbox_secret_version_hashed/r_lockbox_secret_version_hashed_1.tf" }}

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID where to add the version.
* `description` - (Optional) The Yandex Cloud Lockbox secret version description.
* `key_<NUMBER>` - (Optional) Each of the entry keys in the Yandex Cloud Lockbox secret version.
* `text_value_<NUMBER>` - (Optional) Each of the entry values in the Yandex Cloud Lockbox secret version.

The `<NUMBER>` can range from `1` to `10`. If you only need one entry, use `key_1`/`text_value_1`. If you need a second entry, use `key_2`/`text_value_2`, and so on.


## Import

~> Import for this resource is not implemented yet.

