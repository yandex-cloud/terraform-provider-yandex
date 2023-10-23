---
layout: "yandex"
page_title: "Yandex: yandex_datasphere_community_iam_binding"
sidebar_current: "docs-yandex-datasphere-community-iam-binding"
description: |-
Allows management of a single IAM binding for a Yandex Datasphere Community.
---

## yandex\_datasphere\_community\_iam\_binding

```hcl
resource "yandex_datasphere_community_iam_binding" "community-iam" {
  community_id = "your-datasphere-community-id"
  role        = "datasphere.communities.developer"
  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `community_id` - (Required) The Yandex Cloud Datasphere Community ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/en/docs/datasphere/security/)

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
    * **group:{group_id}**: A unique group ID.
    * **system:{allUsers|allAuthenticatedUsers}**: see [system groups](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group)
