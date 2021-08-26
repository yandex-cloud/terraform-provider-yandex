---
layout: "yandex"
page_title: "Yandex: yandex_organizationmanager_organization_iam_binding"
sidebar_current: "docs-yandex-organizationmanager-organization-iam-binding"
description: |-
 Allows management of a single IAM binding for a Yandex Organization Manager organization.
---

# yandex\_organizationmanager\_organization\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex.Cloud Organization Manager organization.

## Example Usage

```hcl
resource "yandex_organizationmanager_organization_iam_binding" "editor" {
  organization_id = "some_organization_id"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) ID of the organization to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one
    `yandex_organizationmanager_organization_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}**: A unique federated user ID.

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `organization_id` and role, e.g.

```
$ terraform import yandex_organizationmanager_organization_iam_binding.viewer "organization_id viewer"
```
