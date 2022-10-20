---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_group_iam_member"
sidebar_current: "docs-yandex-organizationmanager-group-iam-member"
description: |-
 Allows management of a single member for a single IAM binding on a Yandex.Cloud Organization Manager Group.
---

# yandex\_organizationmanager\_group\_iam\_member

Allows creation and management of a single member for a single binding within
the IAM policy for an existing Yandex Organization Manager Group.

## Example Usage

```hcl
resource "yandex_organizationmanager_group_iam_member" "editor" {
  group_id = "some_group_id"
  role     = "editor"
  member   = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required) ID of the organization to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field.
  This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}**: A unique federated user ID.

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account.
This member resource can be imported using the `group_id`, role, and account, e.g.

```
$ terraform import yandex_organizationmanager_group_iam_member.my_project "group_id viewer foo@example.com"
```
