---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_user_ssh_key"
description: |-
  Allows management of User SSH Keys within an existing Yandex Cloud Organization and Subject.
---

# yandex_organizationmanager_user_ssh_key (Resource)

Allows management of User SSH Keys within an existing Yandex Cloud Organization and Subject.

## Example usage

```terraform
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) Organization that the user ssh key belongs to.
* `subject_id` - (Required) Subject that the user ssh key belongs to.
* `data` - (Required) Data of the user ssh key.
* `name` - (Optional) Name of the user ssh key.
* `expires_at` - (Optional) User ssh key will be no longer valid after expiration timestamp.
