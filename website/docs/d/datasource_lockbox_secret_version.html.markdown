---
layout: "yandex"
page_title: "Yandex: yandex_lockbox_secret_version"
sidebar_current: "docs-yandex-datasource-lockbox-secret-version"
description: |-
  Get information about Yandex Cloud Lockbox secret version.
---

# yandex\_lockbox\_secret\_version

Get information about Yandex Cloud Lockbox secret version. For more information,
see [the official documentation](https://cloud.yandex.com/en/docs/lockbox/).

## Example Usage

```hcl
data "yandex_lockbox_secret_version" "my_secret_version" {
  secret_id = "some ID"
}

output "my_secret_entries" {
  value = data.yandex_lockbox_secret.my_secret_version.entries
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID.
* `version_id` - (Optional) The Yandex Cloud Lockbox secret version ID.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `entries` - List of entries in the Yandex Cloud Lockbox secret version.

The `entries` block contains:

* `key` - The key of the entry.
* `text_value` - The text value of the entry.
