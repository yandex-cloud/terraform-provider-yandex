---
layout: "yandex"
page_title: "Yandex: yandex_container_registry_iam_binding"
sidebar_current: "docs-yandex-container-registry-iam-binding"
description: |-
 Allows management of a single IAM binding for a [Yandex Container Registry](https://cloud.yandex.com/docs/container-registry/).
---

## yandex\_container\_registry\_iam\_binding

Allows creation and management of a single binding within IAM policy for
an existing Yandex Container Registry.

## Example Usage

```hcl
resource yandex_container_registry your-registry {
  folder_id = "your-folder-id"
  name      = "registry-name"
}

resource "yandex_container_registry_iam_binding" "puller" {
  registry_id = yandex_container_registry.your-registry.id
  role        = "container-registry.images.puller"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Required) The [Yandex Container Registry](https://cloud.yandex.com/docs/container-registry/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/container-registry/security/).

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)

## Import

IAM binding imports use space-delimited identifiers; first the resource in question and then the role.
These bindings can be imported using the `registry_id` and role, e.g.

```
$ terraform import yandex_container_registry_iam_binding.puller "registry_id container-registry.images.puller"
```
