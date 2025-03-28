---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages Yandex Cloud Lockbox secret version.
---

# {{.Name}} ({{.Type}})

Yandex Cloud Lockbox secret version resource. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).

## Example usage

{{ tffile "examples/lockbox_secret_version/r_lockbox_secret_version_1.tf" }}

{{ tffile "examples/lockbox_secret_version/r_lockbox_secret_version_1.tf" }}

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

## Import

~> Import for this resource is not implemented yet.

