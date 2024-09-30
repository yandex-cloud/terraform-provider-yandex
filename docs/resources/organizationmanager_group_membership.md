---
subcategory: "Organization Manager"
page_title: "Yandex: yandex_organizationmanager_group_membership"
description: |-
  Allows management of members of Yandex.Cloud Organization Manager Group.
---


# yandex_organizationmanager_group_membership




Allows members management of a single Yandex.Cloud Organization Manager Group. For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/organization/manage-groups#add-member).

~> **Note:** Multiple `yandex_organizationmanager_group_iam_binding` resources with the same group id will produce inconsistent behavior!

```terraform
resource "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  organization_id = "some_organization_id"
  subject_id      = "some_subject_id"
  data            = "ssh_key_data"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required, Forces new resource) The Group to add/remove members to/from.
* `members` - A set of members of the Group. Each member is represented by an id.
