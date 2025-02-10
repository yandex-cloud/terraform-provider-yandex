---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret_version"
description: |-
  Get information about Yandex Cloud Lockbox secret version.
---

# yandex_lockbox_secret_version (Data Source)

Get information about Yandex Cloud Lockbox secret version. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).

## Example usage

```terraform
//
// Get information about existing Lockbox Secret Version.
//
data "yandex_lockbox_secret_version" "my_secret_version" {
  secret_id  = "some-secret-id"
  version_id = "some-version-id" # if you don't indicate it, by default refers to the latest version
}

output "my_secret_entries" {
  value = data.yandex_lockbox_secret_version.my_secret_version.entries
}
```

If you're creating the secret in the same project, then you should indicate `version_id`, since otherwise you may refer to a wrong version of the secret (e.g. the first version, when it is still empty).

```terraform
//
// Get information about existing Lockbox Secret Version.
//
resource "yandex_lockbox_secret" "my_secret" {
  # ...
}

resource "yandex_lockbox_secret_version" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
  # ...
}

data "yandex_lockbox_secret_version" "my_version" {
  secret_id  = yandex_lockbox_secret.my_secret.id
  version_id = yandex_lockbox_secret_version.my_version.id
}

output "my_secret_entries" {
  value = data.yandex_lockbox_secret_version.my_version.entries
}
```

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
