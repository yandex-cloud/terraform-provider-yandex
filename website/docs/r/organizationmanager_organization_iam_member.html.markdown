---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_organization_iam_member"
sidebar_current: "docs-yandex-organizationmanager-organization-iam-member"
description: |-
 Allows management of a single member for a single IAM binding on a Yandex.Cloud Organization Manager organization.
---

# yandex\_organizationmanager\_organization\_iam\_member

Allows creation and management of a single member for a single binding within
the IAM policy for an existing Yandex Organization Manager organization.

~> **Note:** Roles controlled by `yandex_organizationmanager_organization_iam_binding`
   should not be assigned using `yandex_organizationmanager_organization_iam_member`.

~> **Note:** When you delete `yandex_organizationmanager_organization_iam_binding` resource,
   the roles can be deleted from other users within the organization as well. Be careful!

## Example Usage

```hcl
resource "yandex_organizationmanager_organization_iam_member" "editor" {
  organization_id = "some_organization_id"
  role            = "editor"
  member          = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) ID of the organization to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field.
  This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}**: A unique federated user ID.

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account.
This member resource can be imported using the `organization id`, role, and account, e.g.

```
$ terraform import yandex_organizationmanager_organization_iam_member.my_project "organization_id viewer foo@example.com"
```
