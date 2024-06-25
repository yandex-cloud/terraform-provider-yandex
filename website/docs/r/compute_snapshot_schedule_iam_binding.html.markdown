---
layout: "yandex"
page_title: "Yandex: yandex_compute_snapshot_schedule_iam_binding"
sidebar_current: "docs-yandex-compute-snapshot-schedule-iam-binding"
description: |-
Allows management of a single IAM binding for a Snapshot Schedule.
---

# yandex\_compute\_snapshot_schedule\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Snapshot Schedule.

## Example Usage

```hcl
resource "yandex_compute_snapshot_schedule" "schedule1" {
  schedule_policy {
	expression = "0 0 * * *"
  }

  retention_period = "12h"

  snapshot_spec {
	  description = "retention-snapshot"
  }

  disk_ids = ["test_disk_id", "another_test_disk_id"]
}

resource "yandex_compute_snapshot_schedule_iam_binding" "editor" {
  snapshot_schedule_id = "${data.yandex_compute_snapshot_schedule.schedule1.id}"

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_schedule_id` - (Required) ID of the snapshot schedule to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one
  `yandex_compute_snapshot_schedule_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **federatedUser:{federated_user_id}**: A unique federated user ID.
    * **federatedUser:{federated_user_id}:**: A unique SAML federation user account ID.
    * **group:{group_id}**: A unique group ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `snapshot_schedule_id` and role, e.g.

```
$ terraform import yandex_compute_snapshot_schedule_iam_binding.editor "snapshot_schedule_id editor"
```
