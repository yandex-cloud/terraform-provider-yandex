---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_cloud_iam_member"
sidebar_current: "docs-yandex-resourcemanager-cloud-iam-member"
description: |-
 Allows management of a single member for a single IAM binding on a Yandex Resource Manager cloud.
---

# yandex\_resourcemanager\_cloud\_iam\_member

Allows creation and management of a single member for a single binding within
the IAM policy for an existing Yandex Resource Manager cloud.

~> **Note:** Roles controlled by `yandex_resourcemanager_cloud_iam_binding`
   should not be assigned using `yandex_resourcemanager_cloud_iam_member`.

~> **Note:** When you delete `yandex_resourcemanager_cloud_iam_binding` resource,
   the roles can be deleted from other users within the cloud as well. Be careful!

## Example Usage

```hcl
data "yandex_resourcemanager_cloud" "department1" {
  name = "Department 1"
}

resource "yandex_resourcemanager_cloud_iam_member" "admin" {
  cloud_id = "${data.yandex_resourcemanager_cloud.department1.id}"
  role     = "editor"
  member   = "userAccount:user_id"
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Required) ID of the cloud to attach a policy to.

* `role` - (Required) The role that should be assigned.

* `member` - (Required) The identity that will be granted the privilege that is specified in the `role` field.
  This field can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.

## Import

IAM member imports use space-delimited identifiers; the resource in question, the role, and the account.
This member resource can be imported using the `cloud id`, role, and account, e.g.

```
$ terraform import yandex_resourcemanager_cloud_iam_member.my_project "cloud_id viewer foo@example.com"
```
