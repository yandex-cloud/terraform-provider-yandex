---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_group_membership"
sidebar_current: "docs-yandex-organizationmanager-group-membership"
description: |- Allows management of members of Yandex.Cloud Organization Manager Group.
---

# yandex\_organizationmanager\_group\_membership

Allows members management of a single Yandex.Cloud Organization Manager Group. For more information, see [the official documentation](https://cloud.yandex.com/en-ru/docs/organization/manage-groups#add-member).

~> **Note:** Multiple `yandex_organizationmanager_group_iam_binding` resources with the same group id will produce inconsistent behavior!

## Example Usage

```hcl
resource "yandex_organizationmanager_group_membership" group {
  group_id = "sdf4*********3fr"
  members  = [
    "xdf********123"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required, Forces new resource) The Group to add/remove members to/from.
* `members` - A set of members of the Group. Each member is represented by an id.
