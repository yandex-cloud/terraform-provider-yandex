---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about Yandex Cloud Lockbox secret version.
---

# {{.Name}} ({{.Type}})

Get information about Yandex Cloud Lockbox secret version. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).

## Example usage

{{ tffile "examples/lockbox_secret_version/d_lockbox_secret_version_1.tf" }}

If you're creating the secret in the same project, then you should indicate `version_id`, since otherwise you may refer to a wrong version of the secret (e.g. the first version, when it is still empty).

{{ tffile "examples/lockbox_secret_version/d_lockbox_secret_version_2.tf" }}

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID.
* `version_id` - (Optional) The Yandex Cloud Lockbox secret version ID.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `entries` - List of entries in the Yandex Cloud Lockbox secret version.

The `entries` block contains:

* `key` - The key of the entry.
* `text_value` - The text value of the entry.
