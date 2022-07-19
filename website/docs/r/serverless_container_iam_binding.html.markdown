---
layout: "yandex"
page_title: "Yandex: yandex_serverless_container_iam_binding"
sidebar_current: "docs-yandex-serverless-container-iam-binding"
description: |-
Allows management of a single IAM binding for a [Yandex Serverless Container](https://cloud.yandex.com/docs/serverless-containers/).
---

## yandex\_serverless\_container\_iam\_binding

```hcl
resource "yandex_serverless_container_iam_binding" "container-iam" {
  container_id = "your-container-id"
  role        = "serverless.containers.invoker"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `container_id` - (Required) The [Yandex Serverless Container](https://cloud.yandex.com/docs/serverless-containers/) ID to apply a binding to.

* `role` - (Required) The role that should be applied.

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)
