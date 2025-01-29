---
subcategory: "Cloud Organization"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of User SSH Keys within an existing Yandex Cloud Organization and Subject.
---

# {{.Name}} ({{.Type}})

Allows management of User SSH Keys within an existing Yandex Cloud Organization and Subject.

## Example usage

{{ tffile "examples/organizationmanager_user_ssh_key/r_organizationmanager_user_ssh_key_1.tf" }}

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) Organization that the user ssh key belongs to.
* `subject_id` - (Required) Subject that the user ssh key belongs to.
* `data` - (Required) Data of the user ssh key.
* `name` - (Optional) Name of the user ssh key.
* `expires_at` - (Optional) User ssh key will be no longer valid after expiration timestamp.
