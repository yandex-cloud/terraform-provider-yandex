---
layout: "yandex"
page_title: "Yandex: yandex_iam_role"
sidebar_current: "docs-yandex-datasource-iam-role"
description: |-
  Generates an IAM role that can be referenced by other resources, applying
  the role to them.
---

# yandex\_iam\_role

Generates an [IAM] role document that may be referenced by and applied to
other Yandex.Cloud Platform resources, such as the `yandex_resourcemanager_folder` resource. For more information, see
[the official documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/roles).

```hcl
data "yandex_iam_role" "admin" {
  binding {
    role = "admin"

    members = [
      "userAccount:user_id_1"
    ]
  }
}
```

This data source is used to define [IAM] roles in order to apply them to other resources.
Currently, defining a role through a data source and referencing that role
from another resource is the only way to apply an IAM role to a resource.

## Argument Reference

The following arguments are supported:

* `binding` (Required) - A nested configuration block (described below)
  that defines a binding to be included in the policy document. Multiple
  `binding` arguments are supported.

Each role document configuration must have one or more `binding` blocks. Each block accepts the following arguments:

* `role` (Required) - The role/permission that will be granted to the members.
  See the [IAM Roles] documentation for a complete list of roles.

* `members` (Required) - An array of identities that will be granted the privilege in the `role`.
  Each entry can have one of the following values:
  * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
  * **serviceAccount:{service_account_id}**: A unique service account ID.

## Attributes Reference

The following attribute is exported:

* `role_data` - The above bindings serialized in a format suitable for
  referencing from a resource that supports IAM.

[IAM]: https://cloud.yandex.com/docs/iam/
[IAM Roles]: https://cloud.yandex.com/docs/iam/concepts/access-control/roles
