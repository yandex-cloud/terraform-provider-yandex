---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_snapshot_iam_binding"
description: |-
  Allows management of a single IAM binding for a Snapshot.
---


# yandex_compute_snapshot_iam_binding




Allows creation and management of a single binding within IAM policy for an existing Snapshot.

## Example usage

```terraform
resource "yandex_compute_snapshot" "snapshot1" {
  name           = "test-snapshot"
  source_disk_id = "test_disk_id"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_snapshot_iam_binding" "editor" {
  snapshot_id = data.yandex_compute_snapshot.snapshot1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `snapshot_id` - (Required) ID of the snapshot to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one `yandex_compute_snapshot_iam_binding` can be used per role.

* `members` - (Required) An array of identities that will be granted the privilege in the `role`. Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **federatedUser:{federated_user_id}**: A unique federated user ID.
  * **federatedUser:{federated_user_id}:**: A unique SAML federation user account ID.
  * **group:{group_id}**: A unique group ID.
  * **system:group:federation:{federation_id}:users**: All users in federation.
  * **system:group:organization:{organization_id}:users**: All users in organization.
  * **system:allAuthenticatedUsers**: All authenticated users.
  * **system:allUsers**: All users, including unauthenticated ones.

  Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `snapshot_id` and role, e.g.

```
$ terraform import yandex_compute_snapshot_iam_binding.editor "snapshot_id editor"
```
