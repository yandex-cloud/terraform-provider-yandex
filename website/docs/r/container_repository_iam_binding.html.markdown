---
layout: "yandex"
page_title: "Yandex: yandex_container_repository_iam_binding"
sidebar_current: "docs-yandex-container-repository-iam-binding"
description: |-
 Allows management of a single IAM binding for a [Yandex Container Repository](https://cloud.yandex.com/docs/container-registry/concepts/repository).
---

## yandex\_container\_repository\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex Container Repository.

## Example Usage

```hcl
resource yandex_container_registry your-registry {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource yandex_container_repository repo-1 {
  name      = "${yandex_container_registry.your-registry.id}/repo-1"
}

resource "yandex_container_repository_iam_binding" "puller" {
  repository_id = yandex_container_repository.repo-1.id
  role        = "container-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}

data "yandex_container_repository" "repo-2" {
  name = "some_repository_name"
}

resource "yandex_container_repository_iam_binding" "pusher" {
  repository_id = yandex_container_repository.repo-2.id
  role        = "container-registry.images.pusher"

  members = [
    "serviceAccount:your-service-account-id",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `repository_id` - (Required) The [Yandex Container Repository](https://cloud.yandex.com/docs/container-registry/concepts/repository) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/container-registry/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `repository_id` and role, e.g.

```
$ terraform import yandex_container_repository_iam_binding.puller "repository_id container-registry.images.puller"
```
