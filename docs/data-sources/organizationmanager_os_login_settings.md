---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_os_login_settings"
description: |-
  Get information about a Yandex Cloud OsLogin Settings.
---

# yandex_organizationmanager_os_login_settings (Data Source)

## Example usage

```terraform
data "yandex_organizationmanager_os_login_settings" "my_os_login_settings_settings" {
  organization_id = "some_organization_id"
}

output "my_organization_ssh_certificates_enabled" {
  value = "data.yandex_organizationmanager_os_login_settings.my_os_login_settings.ssh_certificate_settings.0.enabled"
}
```

## Argument Reference

* `organization_id` - (Required) ID of the organization.

## Attributes Reference

The following attributes are exported:

* `user_ssh_key_settings` - The structure is documented below.
* `ssh_certificate_settings` - The structure is documented below.

The `user_ssh_key_settings` block supports:
* `enabled` - Enables or disables usage of ssh keys assigned to a specific subject.
* `allow_manage_own_keys` - If set to true subject is allowed to manage own ssh keys without having to be assigned specific permissions.

The `ssh_certificate_settings` block supports:
* `enabled` - Enables or disables usage of ssh certificates signed by trusted CA.
