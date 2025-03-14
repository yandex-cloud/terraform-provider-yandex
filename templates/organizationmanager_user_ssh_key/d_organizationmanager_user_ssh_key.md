---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud User SSH Key.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Cloud User SSH Key.

## Example usage

{{ tffile "examples/organizationmanager_user_ssh_key/d_organizationmanager_user_ssh_key_1.tf" }}

## Argument Reference

* `user_ssh_key_id` - (Required) ID of the user ssh key.

## Attributes Reference

The following attributes are exported:

* `organization_id` - Organization that the user ssh key belongs to.
* `subject_id` - Subject that the user ssh key belongs to.
* `name` - Name of the user ssh key.
* `data` - Data of the user ssh key.
* `fingerprint` - Auto generated fingerprint of the user ssh key.
* `created_at` - User ssh key creation timestamp.
* `expires_at` - User ssh key will be no longer valid after expiration timestamp.
