---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_gpu_cluster_iam_binding"
description: |-
  Allows management of a single IAM binding for a GPU Cluster.
---


# yandex_compute_gpu_cluster_iam_binding




Allows creation and management of a single binding within IAM policy for an existing GPU Cluster.

## Example usage

```terraform
resource "yandex_compute_gpu_cluster" "cluster1" {
  name              = "gpu-cluster-name"
  interconnect_type = "infiniband"
  zone              = "ru-central1-a"

  labels = {
    environment = "test"
  }
}

resource "yandex_compute_gpu_cluster_iam_binding" "editor" {
  gpu_cluster_id = data.yandex_compute_gpu_cluster.cluster1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `gpu_cluster_id` - (Required) ID of the gpu cluster to attach the policy to.

* `role` - (Required) The role that should be assigned. Only one `yandex_compute_gpu_cluster_iam_binding` can be used per role.

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

IAM binding imports use space-delimited identifiers; first the resource in question and then the role. These bindings can be imported using the `gpu_cluster_id` and role, e.g.

```
$ terraform import yandex_compute_gpu_cluster_iam_binding.editor "gpu_cluster_id editor"
```
