---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_cloud_iam_binding"
sidebar_current: "docs-yandex-resourcemanager-cloud-iam-binding"
description: |-
 Allows management of a single IAM  binding for a Yandex Resource Manager cloud.
---

# yandex\_resourcemanager\_cloud\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex Resource Manager cloud.

## Example Usage

```hcl
data "yandex_resourcemanager_cloud" "project1" {
  name = "Project 1"
}

resource "yandex_resourcemanager_cloud_iam_binding" "admin" {
  cloud_id = "${data.yandex_resourcemanager_cloud.project1.id}"

  role     = "editor"
  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Required) ID of the cloud to attach policy is to.

* `members` (Required) - An array of identities that will be granted the privilege in the `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: An unique user ID that represents a specific Yandex account.

* `role` - (Required) The role that should be applied. Only one
    `yandex_resourcemanager_cloud_iam_binding` can be used per role.

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `cloud_id` and role, e.g.

```
$ terraform import yandex_resourcemanager_cloud_iam_binding.viewer "cloud_id viewer"
```
