---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_group_membership"
description: |-
  Allows management of members of Yandex Cloud Organization Manager Group.
---

# yandex_organizationmanager_group_membership (Resource)

Allows members management of a single Yandex Cloud Organization Manager Group. For more information, see [the official documentation](https://yandex.cloud/docs/organization/manage-groups#add-member).

~> Multiple `yandex_organizationmanager_group_iam_binding` resources with the same group id will produce inconsistent behavior!

## Example usage

```terraform
resource "yandex_organizationmanager_group_membership" "group" {
  group_id = "sdf4*********3fr"
  members = [
    "xdf********123"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required, Forces new resource) The Group to add/remove members to/from.
* `members` - A set of members of the Group. Each member is represented by an id.
