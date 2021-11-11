---
layout: "yandex"
page_title: "Yandex: yandex_function_iam_binding"
sidebar_current: "docs-yandex-function-iam-binding"
description: |-
 Allows management of a single IAM binding for a [Yandex Cloud Function](https://cloud.yandex.com/docs/functions/).
---

## yandex\_function\_iam\_binding

```hcl
resource "yandex_function_iam_binding" "function-iam" {
  function_id = "your-function-id"
  role        = "serverless.functions.invoker"

  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `function_id` - (Required) The [Yandex Cloud Function](https://cloud.yandex.com/docs/functions/) ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/docs/functions/security/)

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.
  * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)
