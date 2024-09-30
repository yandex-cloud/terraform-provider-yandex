---
subcategory: "Organization Manager"
page_title: "Yandex: yandex_organizationmanager_os_login_settings"
description: |-
  Allows management of OsLogin Settings within an existing Yandex.Cloud Organization.
---


# yandex_organizationmanager_os_login_settings




```terraform
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
```

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
