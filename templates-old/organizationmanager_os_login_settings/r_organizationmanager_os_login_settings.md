---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of OsLogin Settings within an existing Yandex Cloud Organization.
---

# {{.Name}} ({{.Type}})

## Example usage

{{ tffile "examples/organizationmanager_os_login_settings/r_organizationmanager_os_login_settings_1.tf" }}

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) The organization to manage it's OsLogin Settings.
* `user_ssh_key_settings` - (Optional) The structure is documented below.
* `ssh_certificate_settings` - (Optional) The structure is documented below.

The `user_ssh_key_settings` block supports:
* `enabled` - Enables or disables usage of ssh keys assigned to a specific subject.
* `allow_manage_own_keys` - If set to true subject is allowed to manage own ssh keys without having to be assigned specific permissions.

The `ssh_certificate_settings` block supports:
* `enabled` - Enables or disables usage of ssh certificates signed by trusted CA.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/organizationmanager_os_login_settings/import.sh" }}
